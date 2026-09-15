package graphrag_test

import (
	"math"
	"os"
	"testing"
	"time"

	graphrag "github.com/larsartmann/go-graph-rag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenAICompatProvider_LiveSmoke exercises the real wire contract
// (auth, request shape, response parsing) against a live OpenAI-compatible
// endpoint — beyond the httptest doubles. It is strictly opt-in and skips
// cleanly when the environment is absent, so offline runs and CI never
// touch the network:
//
//	GRAPHRAG_LIVE_EMBED_URL=https://api.openai.com/v1 \
//	GRAPHRAG_LIVE_EMBED_KEY=sk-... \
//	GRAPHRAG_LIVE_EMBED_MODEL=text-embedding-3-small \
//	go test -run TestOpenAICompatProvider_LiveSmoke -v ./
//
// GRAPHRAG_LIVE_EMBED_MODEL defaults to text-embedding-3-small; the key may
// be empty for local endpoints (Ollama, LM Studio).
func TestOpenAICompatProvider_LiveSmoke(t *testing.T) {
	baseURL := os.Getenv("GRAPHRAG_LIVE_EMBED_URL")
	if baseURL == "" {
		t.Skip("GRAPHRAG_LIVE_EMBED_URL not set: live-endpoint smoke test is opt-in")
	}

	model := os.Getenv("GRAPHRAG_LIVE_EMBED_MODEL")
	if model == "" {
		model = "text-embedding-3-small"
	}

	provider, err := graphrag.NewProvider(graphrag.EmbeddingConfig{
		Provider: graphrag.ProviderOpenAICompat,
		BaseURL:  baseURL,
		APIKey:   os.Getenv("GRAPHRAG_LIVE_EMBED_KEY"),
		Model:    model,
		Timeout:  30 * time.Second,
	})
	require.NoError(t, err, "live config must construct a provider")

	vectors, err := provider.Embed(t.Context(), []string{"graph retrieval smoke test"})
	require.NoError(t, err, "live endpoint must answer a one-text embed")
	require.Len(t, vectors, 1, "one vector per input")

	dims := vectors[0].Dims()
	require.Positive(t, dims, "live embedding must have real dimensions")

	var normSquared float64
	for _, component := range vectors[0] {
		normSquared += float64(component) * float64(component)
	}

	require.Positive(t, normSquared, "embedding must be a non-zero vector")

	// Unit norm is asserted only loosely: OpenAI normalizes, but legitimately
	// unnormalized models exist (some bge/nomic deployments), so a hard
	// |v|==1 assert would false-fail valid endpoints.
	assert.InDelta(t, 1.0, math.Sqrt(normSquared), 0.1,
		"most embedding models return L2-normalized vectors")

	t.Logf("live smoke: model=%s dims=%d |v|=%.4f", model, dims, math.Sqrt(normSquared))
}
