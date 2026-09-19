package graphrag_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	graphrag "github.com/larsartmann/go-graph-rag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// embedServer spins an /embeddings endpoint with per-request behaviour and
// counts requests. Behaviour receives the decoded request and returns the
// raw response body plus status.
type embedServer struct {
	server        *httptest.Server
	requests      atomic.Int64
	lastModel     atomic.Value // string
	lastAuth      atomic.Value // string
	lastUserAgent atomic.Value // string
	respond       func(req graphragEmbeddingsRequest, requestNumber int64) (int, string)
}

// graphragEmbeddingsRequest mirrors the provider's wire request for asserts.
type graphragEmbeddingsRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

func newEmbedServer(
	t *testing.T,
	respond func(req graphragEmbeddingsRequest, requestNumber int64) (int, string),
) *embedServer {
	t.Helper()

	embed := &embedServer{respond: respond}
	embed.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		number := embed.requests.Add(1)

		var req graphragEmbeddingsRequest

		if err := json.Unmarshal(readAll(t, r), &req); err != nil {
			w.WriteHeader(http.StatusBadRequest)

			return
		}

		embed.lastModel.Store(req.Model)

		auth := r.Header.Get("Authorization")
		if auth != "" {
			embed.lastAuth.Store(auth)
		}

		if ua := r.Header.Get("User-Agent"); ua != "" {
			embed.lastUserAgent.Store(ua)
		}

		status, body := respond(req, number)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(embed.server.Close)

	return embed
}

func readAll(t *testing.T, r *http.Request) []byte {
	t.Helper()

	body, err := io.ReadAll(r.Body)
	require.NoError(t, err)

	return body
}

// embedBody renders an OpenAI-style response; order controls the data-array
// order (index fields stay input-correct when shuffled=true to prove the
// provider reorders by index).
func embedBody(dims int, shuffled bool, vectors ...[]float64) string {
	type datum struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	}

	data := make([]datum, 0, len(vectors))
	for i, vec := range vectors {
		v := vec
		if v == nil {
			v = make([]float64, dims)
			for j := range v {
				v[j] = float64(i + j)
			}
		}

		data = append(data, datum{Embedding: v, Index: i})
	}

	if shuffled && len(data) == 2 {
		data[0], data[1] = data[1], data[0]
	}

	response := struct {
		Data  []datum        `json:"data"`
		Model string         `json:"model"`
		Usage map[string]int `json:"usage"`
	}{Data: data, Model: "test-model", Usage: map[string]int{}}

	body, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}

	return string(body)
}

func newTestProvider(t *testing.T, url string, retries int) *graphrag.OpenAICompatProvider {
	t.Helper()

	provider, err := graphrag.NewOpenAICompatProvider(graphrag.EmbeddingConfig{
		Provider:   graphrag.ProviderOpenAICompat,
		BaseURL:    url,
		APIKey:     "test-key",
		Model:      "test-model",
		Timeout:    5 * time.Second,
		MaxRetries: retries,
		MaxChars:   100,
	})
	require.NoError(t, err)

	return provider
}

func TestOpenAICompatProvider_EmbedsInInputOrder(t *testing.T) {
	t.Parallel()

	embed := newEmbedServer(t, func(_ graphragEmbeddingsRequest, _ int64) (int, string) {
		return http.StatusOK, embedBody(4, true, nil, nil)
	})

	provider := newTestProvider(t, embed.server.URL, 0)

	vectors, err := provider.Embed(t.Context(), []string{"alpha", "beta"})
	require.NoError(t, err)
	require.Len(t, vectors, 2, "one vector per input, reordered by index despite shuffled data array")

	assert.InDeltaSlice(t, []float64{0, 0.26726, 0.53452, 0.80178}, vectors[0], 1e-4,
		"first input's L2-normalized vector must come first")
	assert.InDeltaSlice(t, []float64{0.18257, 0.36515, 0.54772, 0.73030}, vectors[1], 1e-4)

	assert.Equal(t, "test-model", embed.lastModel.Load(), "model must be sent verbatim")
	assert.Equal(t, "Bearer test-key", embed.lastAuth.Load(), "API key rides the Authorization header")
	assert.Equal(t, "go-graph-rag/"+graphrag.UserAgentVersion, embed.lastUserAgent.Load(),
		"User-Agent carries the stamped SDK version")
}

func TestOpenAICompatProvider_RetriesTransientFailures(t *testing.T) {
	t.Parallel()

	embed := newEmbedServer(t, func(_ graphragEmbeddingsRequest, number int64) (int, string) {
		switch number {
		case 1:
			return http.StatusInternalServerError, `{"error":"boom"}`
		case 2:
			return http.StatusTooManyRequests, `{"error":"slow down"}`
		default:
			return http.StatusOK, embedBody(2, false, nil)
		}
	})

	provider := newTestProvider(t, embed.server.URL, 2)

	vectors, err := provider.Embed(t.Context(), []string{"alpha"})
	require.NoError(t, err)
	require.Len(t, vectors, 1)
	assert.Equal(t, int64(3), embed.requests.Load(), "500 and 429 are retryable: third attempt wins")
}

func TestOpenAICompatProvider_NonRetryableFailsFast(t *testing.T) {
	t.Parallel()

	embed := newEmbedServer(t, func(_ graphragEmbeddingsRequest, _ int64) (int, string) {
		return http.StatusBadRequest, `{"error":"invalid model"}`
	})

	provider := newTestProvider(t, embed.server.URL, 3)

	_, err := provider.Embed(t.Context(), []string{"alpha"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid model", "non-2xx body snippet is embedded for diagnostics")
	assert.Equal(t, int64(1), embed.requests.Load(), "4xx must NOT be retried")
}

func TestOpenAICompatProvider_CountMismatch(t *testing.T) {
	t.Parallel()

	embed := newEmbedServer(t, func(_ graphragEmbeddingsRequest, _ int64) (int, string) {
		return http.StatusOK, embedBody(2, false, nil)
	})

	provider := newTestProvider(t, embed.server.URL, 0)

	_, err := provider.Embed(t.Context(), []string{"alpha", "beta"})
	require.Error(t, err)
	assert.ErrorIs(t, err, graphrag.ErrEmbedCountMismatch)
}

func TestOpenAICompatProvider_DimsMismatch(t *testing.T) {
	t.Parallel()

	embed := newEmbedServer(t, func(_ graphragEmbeddingsRequest, _ int64) (int, string) {
		return http.StatusOK, embedBody(0, false, make([]float64, 3), make([]float64, 4))
	})

	provider := newTestProvider(t, embed.server.URL, 0)

	_, err := provider.Embed(t.Context(), []string{"alpha", "beta"})
	require.Error(t, err)
	assert.ErrorIs(t, err, graphrag.ErrEmbedDimsMismatch)
}

func TestOpenAICompatProvider_CancelDuringBackoff(t *testing.T) {
	t.Parallel()

	embed := newEmbedServer(t, func(_ graphragEmbeddingsRequest, _ int64) (int, string) {
		return http.StatusInternalServerError, `{}`
	})

	provider := newTestProvider(t, embed.server.URL, 5)

	ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
	defer cancel()

	_, err := provider.Embed(ctx, []string{"alpha"})
	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded) || errors.Is(err, ctx.Err()),
		"cancellation during the retry backoff must abort, not sleep out every retry: %v", err)
}

func TestOpenAICompatProvider_BatchesOver64Texts(t *testing.T) {
	t.Parallel()

	var sizes atomic.Value // []int

	embed := newEmbedServer(t, func(req graphragEmbeddingsRequest, _ int64) (int, string) {
		snapshot, _ := sizes.Load().([]int)
		sizes.Store(append(snapshot, len(req.Input)))

		return http.StatusOK, embedBody(0, false, make([][]float64, len(req.Input))...)
	})

	provider := newTestProvider(t, embed.server.URL, 0)

	texts := make([]string, 65)
	for i := range texts {
		texts[i] = "text"
	}

	vectors, err := provider.Embed(t.Context(), texts)
	require.NoError(t, err)
	require.Len(t, vectors, 65, "every text keeps its vector across batch boundaries")

	assert.Equal(t, []int{64, 1}, sizes.Load().([]int), "batches are capped at 64 texts per request")
}

// overlapTracker records the maximum number of in-flight requests the
// endpoint observed, proving (or refuting) actual batch overlap.
type overlapTracker struct {
	inflight atomic.Int64
	maxSeen  atomic.Int64
}

func (o *overlapTracker) enter() {
	current := o.inflight.Add(1)

	for {
		seen := o.maxSeen.Load()
		if current <= seen || o.maxSeen.CompareAndSwap(seen, current) {
			return
		}
	}
}

func (o *overlapTracker) exit() { o.inflight.Add(-1) }

// newNumberedEmbedServer answers every input "tN" with the 2-dim embedding
// [N, 1]: after the provider's L2 normalization the first component is
// N/sqrt(N^2+1), so the assembled vector order is provable end to end.
// Requests sleep for delay and their overlap is tracked.
func newNumberedEmbedServer(t *testing.T, delay time.Duration, tracker *overlapTracker) *embedServer {
	t.Helper()

	return newEmbedServer(t, func(req graphragEmbeddingsRequest, _ int64) (int, string) {
		data := make([]map[string]any, 0, len(req.Input))

		for i, text := range req.Input {
			number, convErr := strconv.Atoi(strings.TrimPrefix(text, "t"))
			if convErr != nil {
				number = 1
			}

			data = append(data, map[string]any{
				"embedding": []float64{float64(number), 1},
				"index":     i,
			})
		}

		tracker.enter()
		defer tracker.exit()
		time.Sleep(delay)

		body, _ := json.Marshal(map[string]any{"data": data, "model": "test-model", "usage": map[string]int{}})

		return http.StatusOK, string(body)
	})
}

func newTestProviderWithConcurrency(t *testing.T, url string, concurrency int) *graphrag.OpenAICompatProvider {
	t.Helper()

	provider, err := graphrag.NewOpenAICompatProvider(graphrag.EmbeddingConfig{
		Provider:         graphrag.ProviderOpenAICompat,
		BaseURL:          url,
		APIKey:           "test-key",
		Model:            "test-model",
		Timeout:          5 * time.Second,
		MaxRetries:       0,
		MaxChars:         100,
		EmbedConcurrency: concurrency,
	})
	require.NoError(t, err)

	return provider
}

// numberedTexts builds "t1".."tn" — inputs whose encoded first components
// are all distinct.
func numberedTexts(n int) []string {
	texts := make([]string, 0, n)
	for i := range n {
		texts = append(texts, fmt.Sprintf("t%d", i+1))
	}

	return texts
}

func TestOpenAICompatProvider_ConcurrentBatchesOverlapAndKeepOrder(t *testing.T) {
	t.Parallel()

	tracker := &overlapTracker{}
	embed := newNumberedEmbedServer(t, 15*time.Millisecond, tracker)

	provider := newTestProviderWithConcurrency(t, embed.server.URL, 4)

	texts := numberedTexts(200) // 4 batches of 64

	vectors, err := provider.Embed(t.Context(), texts)
	require.NoError(t, err)
	require.Len(t, vectors, 200)

	for i, vector := range vectors {
		number := float64(i + 1)
		expected := number / math.Sqrt(number*number+1)

		assert.InDelta(t, expected, vector[0], 1e-4,
			"vector %d must be input %d's answer: completion order must never reorder results", i, i)
	}

	assert.GreaterOrEqual(t, tracker.maxSeen.Load(), int64(2), "batches must actually overlap at concurrency 4")
	assert.Equal(t, int64(4), embed.requests.Load(), "one request per batch")
}

func TestOpenAICompatProvider_DefaultConcurrencyIsSerial(t *testing.T) {
	t.Parallel()

	tracker := &overlapTracker{}
	embed := newNumberedEmbedServer(t, 5*time.Millisecond, tracker)

	provider := newTestProvider(t, embed.server.URL, 0) // EmbedConcurrency unset -> serial

	vectors, err := provider.Embed(t.Context(), numberedTexts(130)) // 3 batches
	require.NoError(t, err)
	require.Len(t, vectors, 130)

	for i, vector := range vectors {
		number := float64(i + 1)
		expected := number / math.Sqrt(number*number+1)

		assert.InDelta(t, expected, vector[0], 1e-4, "serial path must also preserve input order")
	}

	assert.Equal(t, int64(1), tracker.maxSeen.Load(), "default (unset) concurrency must take the serial path")
}

func TestOpenAICompatProvider_CancelDuringConcurrentEmbed(t *testing.T) {
	t.Parallel()

	tracker := &overlapTracker{}
	embed := newNumberedEmbedServer(t, 100*time.Millisecond, tracker)

	provider := newTestProviderWithConcurrency(t, embed.server.URL, 4)

	ctx, cancel := context.WithCancel(t.Context())

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	_, err := provider.Embed(ctx, numberedTexts(200))
	cancel()

	require.Error(t, err, "mid-flight cancellation must abort the concurrent embed")
}
