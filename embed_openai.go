//nolint:tagliatelle // the /embeddings wire contract is OpenAI's snake_case API; these structs pin it verbatim
package graphrag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
)

// OpenAI-compatible provider defaults and limits.
const (
	// defaultEmbedTimeout bounds one embeddings request when the
	// configuration leaves Timeout at zero.
	defaultEmbedTimeout = 30 * time.Second

	// defaultEmbedRetryBackoff is the linear backoff base between retries.
	defaultEmbedRetryBackoff = time.Second

	// embedBatchTexts caps how many texts ride in one /embeddings request.
	// OpenAI documents a 2048-input ceiling; batches stay far below it so
	// per-request latency and retry cost stay bounded.
	embedBatchTexts = 64

	// embedErrorBodyLimit bounds how much of an error response is embedded
	// into returned errors (diagnostics without unbounded memory).
	embedErrorBodyLimit = 2048

	// embedResponseBodyLimit caps the accepted response size; embedding
	// payloads for one batch stay in the low megabytes, so 64 MiB leaves
	// generous headroom while still bounding memory.
	embedResponseBodyLimit = 64 << 20
)

// Sentinel errors for the OpenAI-compatible provider.
var (
	// ErrEmbedBaseURLRequired is returned when no base URL is configured.
	ErrEmbedBaseURLRequired = errors.New("graphrag: openai-compat embedding provider requires base_url")

	// ErrEmbedModelRequired is returned when no model is configured.
	ErrEmbedModelRequired = errors.New("graphrag: openai-compat embedding provider requires model")

	// ErrEmbedEmptyResponse is returned when the endpoint answers without data.
	ErrEmbedEmptyResponse = errors.New("graphrag: embeddings endpoint returned no data")

	// ErrEmbedCountMismatch is returned when the endpoint answers with a
	// different number of vectors than inputs were sent.
	ErrEmbedCountMismatch = errors.New("graphrag: embeddings endpoint returned mismatched vector count")

	// ErrEmbedDimsMismatch is returned when vectors in one response
	// disagree on dimensionality.
	ErrEmbedDimsMismatch = errors.New("graphrag: embeddings endpoint returned inconsistent dimensions")
)

// OpenAICompatProvider calls any OpenAI-compatible POST /embeddings endpoint
// (OpenAI, Mistral, Ollama, vLLM, LM Studio). It is safe for concurrent use.
type OpenAICompatProvider struct {
	baseURL      string
	apiKey       string
	model        string
	maxChars     int
	maxRetries   int
	retryBackoff time.Duration
	httpClient   *http.Client
}

// NewOpenAICompatProvider validates the configuration and builds the client.
func NewOpenAICompatProvider(cfg EmbeddingConfig) (*OpenAICompatProvider, error) {
	if cfg.BaseURL == "" {
		return nil, ErrEmbedBaseURLRequired
	}

	if cfg.Model == "" {
		return nil, ErrEmbedModelRequired
	}

	parsed, err := url.Parse(cfg.BaseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("%w: %q is not a valid base URL", ErrEmbedBaseURLRequired, cfg.BaseURL)
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultEmbedTimeout
	}

	maxRetries := max(cfg.MaxRetries, 0)

	return &OpenAICompatProvider{
		baseURL:      strings.TrimSuffix(cfg.BaseURL, "/"),
		apiKey:       cfg.APIKey,
		model:        cfg.Model,
		maxChars:     cfg.MaxChars,
		maxRetries:   maxRetries,
		retryBackoff: defaultEmbedRetryBackoff,
		httpClient:   &http.Client{Timeout: timeout},
	}, nil
}

// Name identifies the provider in the embedding cache namespace.
func (p *OpenAICompatProvider) Name() string {
	return ProviderOpenAICompat
}

// Model identifies the model whose vectors are cached.
func (p *OpenAICompatProvider) Model() string {
	return p.model
}

// embeddingsRequest is the /embeddings wire request.
type embeddingsRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// embeddingsResponse is the /embeddings wire response.
type embeddingsResponse struct {
	Data  []embeddingDatum `json:"data"`
	Model string           `json:"model"`
	Usage embeddingsUsage  `json:"usage"`
}

// embeddingDatum is one vector plus its input position.
type embeddingDatum struct {
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// embeddingsUsage mirrors the token accounting block.
type embeddingsUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// Embed truncates each text to the configured character budget and embeds
// them in bounded batches, retrying transient failures with linear backoff.
func (p *OpenAICompatProvider) Embed(ctx context.Context, texts []string) ([]Vector, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	truncated := lo.Map(texts, func(text string, _ int) string { return truncateRunes(text, p.maxChars) })

	vectors := make([]Vector, 0, len(truncated))

	for start := 0; start < len(truncated); start += embedBatchTexts {
		end := min(start+embedBatchTexts, len(truncated))

		batch, err := p.embedBatch(ctx, truncated[start:end])
		if err != nil {
			return nil, err
		}

		vectors = append(vectors, batch...)
	}

	return vectors, nil
}

// sleepContext sleeps for d unless ctx is done first. It reports whether
// the full duration elapsed.
func sleepContext(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// embedBatch performs the retry loop around one HTTP round trip.
func (p *OpenAICompatProvider) embedBatch(ctx context.Context, batch []string) ([]Vector, error) {
	payload, err := json.Marshal(embeddingsRequest{Model: p.model, Input: batch})
	if err != nil {
		return nil, fmt.Errorf("graphrag: marshal embeddings request: %w", err)
	}

	var lastErr error

	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 && !sleepContext(ctx, p.retryBackoff*time.Duration(attempt)) {
			return nil, fmt.Errorf("graphrag: embeddings request cancelled: %w", ctx.Err())
		}

		vectors, batchErr := p.embedRequest(ctx, payload)
		if batchErr == nil {
			if len(vectors) != len(batch) {
				return nil, fmt.Errorf("%w: got %d vectors for %d inputs",
					ErrEmbedCountMismatch, len(vectors), len(batch))
			}

			return vectors, nil
		}

		lastErr = batchErr

		if !isRetryableEmbeddingErr(batchErr) {
			return nil, batchErr
		}
	}

	return nil, fmt.Errorf("graphrag: embeddings failed after %d attempts: %w", p.maxRetries+1, lastErr)
}

// retryableEmbeddingError marks transport-level failures worth retrying.
type retryableEmbeddingError struct{ cause error }

func (e retryableEmbeddingError) Error() string { return e.cause.Error() }
func (e retryableEmbeddingError) Unwrap() error { return e.cause }

// isRetryableEmbeddingErr reports whether err is worth another attempt.
func isRetryableEmbeddingErr(err error) bool {
	var retryable retryableEmbeddingError

	return errors.As(err, &retryable)
}

// embedRequest performs a single POST /embeddings and validates the shape of
// the answer before returning vectors in input order.
func (p *OpenAICompatProvider) embedRequest(ctx context.Context, payload []byte) ([]Vector, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/embeddings", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("graphrag: build embeddings request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "go-graph-rag/0.1")
	req.Header.Set("Accept", "application/json")

	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, retryableEmbeddingError{cause: fmt.Errorf("graphrag: embeddings request: %w", err)}
	}

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, embedResponseBodyLimit))
	closeErr := resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, newEmbedStatusError(resp.StatusCode, body)
	}

	if readErr != nil {
		return nil, fmt.Errorf("graphrag: read embeddings response: %w", readErr)
	}

	if closeErr != nil {
		return nil, fmt.Errorf("graphrag: close embeddings response: %w", closeErr)
	}

	return parseEmbeddings(body, resp.StatusCode)
}

// embedStatusError is a static error describing a non-2xx /embeddings
// response. 429 and 5xx are wrapped retryable so the caller can back off.
type embedStatusError struct {
	status int
	body   string
}

func (e *embedStatusError) Error() string {
	return fmt.Sprintf("graphrag: embeddings endpoint returned %s: %s", strconv.Itoa(e.status), e.body)
}

func newEmbedStatusError(status int, body []byte) error {
	snippet := string(body)
	if len(snippet) > embedErrorBodyLimit {
		snippet = snippet[:embedErrorBodyLimit]
	}

	wrapped := &embedStatusError{status: status, body: snippet}

	if status == http.StatusTooManyRequests || status >= http.StatusInternalServerError {
		return retryableEmbeddingError{cause: wrapped}
	}

	return wrapped
}

// parseEmbeddings decodes and validates one response payload.
func parseEmbeddings(body []byte, status int) ([]Vector, error) {
	var decoded embeddingsResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, fmt.Errorf("graphrag: decode embeddings response (status %s): %w", strconv.Itoa(status), err)
	}

	if len(decoded.Data) == 0 {
		return nil, ErrEmbedEmptyResponse
	}

	vectors := make([]Vector, len(decoded.Data)) //nolint:makezero // positional, index-assigned per provider datum
	dims := 0

	for _, datum := range decoded.Data {
		if datum.Index < 0 || datum.Index >= len(vectors) {
			return nil, fmt.Errorf("%w: index %d out of range", ErrEmbedCountMismatch, datum.Index)
		}

		if dims == 0 {
			dims = len(datum.Embedding)
		} else if dims != len(datum.Embedding) {
			return nil, ErrEmbedDimsMismatch
		}

		vector := make(Vector, len(datum.Embedding)) //nolint:makezero // length mirrors the provider answer
		for i, component := range datum.Embedding {
			vector[i] = float32(component)
		}

		vectors[datum.Index] = vector.Normalize()
	}

	for i, vector := range vectors {
		if vector == nil {
			return nil, fmt.Errorf("%w: missing index %d", ErrEmbedCountMismatch, i)
		}
	}

	return vectors, nil
}
