package graphrag_test

import (
	"context"
	"errors"
	"testing"

	"github.com/LarsArtmann/CV/graphrag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVectorRoundTrip(t *testing.T) {
	t.Parallel()

	original := graphrag.Vector{0.5, -1.25, 3.75, 0}

	decoded, err := graphrag.DecodeVector(original.Encode())
	require.NoError(t, err)

	assert.InDeltaSlice(t, []float32(original), []float32(decoded), 1e-6, "encode/decode must round-trip")
}

func TestDecodeVectorRejectsMisalignedBlob(t *testing.T) {
	t.Parallel()

	_, err := graphrag.DecodeVector([]byte{1, 2, 3})
	require.ErrorIs(t, err, graphrag.ErrVectorDecode)
}

func TestCosineProperties(t *testing.T) {
	t.Parallel()

	identical := graphrag.Vector{1, 2, 3}
	orthogonal := graphrag.Vector{0, 1, 0}
	opposite := graphrag.Vector{-1, 0, 0}
	unit := graphrag.Vector{1, 0, 0}

	assert.InDelta(t, 1.0, graphrag.Cosine(identical, identical), 1e-9)
	assert.InDelta(t, 0.0, graphrag.Cosine(unit, orthogonal), 1e-9)
	assert.InDelta(t, -1.0, graphrag.Cosine(unit, opposite), 1e-9)
	assert.InDelta(t, 0.0, graphrag.Cosine(unit, graphrag.Vector{0, 0, 0}), 1e-9, "zero vector must yield 0, not NaN")
	assert.InDelta(t, 0.0, graphrag.Cosine(unit, graphrag.Vector{1}), 1e-9, "dimension mismatch must yield 0")
}

func TestHashTextIsStable(t *testing.T) {
	t.Parallel()

	assert.Len(t, graphrag.HashText("kubernetes"), 64, "SHA-256 hex digest")
	assert.NotEqual(t, graphrag.HashText("kubernetes"), graphrag.HashText("kubernetes!"))
}

func TestHashProviderEmbedsDeterministically(t *testing.T) {
	t.Parallel()

	provider := graphrag.NewHashProvider()
	require.Equal(t, graphrag.ProviderHash, provider.Name())

	first, err := provider.Embed(t.Context(), []string{"platform engineer go kubernetes", "react frontend css"})
	require.NoError(t, err)
	require.Len(t, first, 2)
	assert.Equal(t, graphrag.HashProviderDims, first[0].Dims())

	second, err := provider.Embed(t.Context(), []string{"platform engineer go kubernetes", "react frontend css"})
	require.NoError(t, err)

	for i := range first {
		assert.InDeltaSlice(t, []float32(first[i]), []float32(second[i]), 1e-9, "hash embeddings must be deterministic")
	}

	distant, err := provider.Embed(t.Context(), []string{"completely unrelated words about cooking pasta"})
	require.NoError(t, err)
	assert.Less(t, graphrag.Cosine(first[0], distant[0]), graphrag.Cosine(first[0], second[0]),
		"related text must score above unrelated text")
}

func TestNewProviderDefaultsToHash(t *testing.T) {
	t.Parallel()

	provider, err := graphrag.NewProvider(graphrag.EmbeddingConfig{})
	require.NoError(t, err)
	assert.Equal(t, graphrag.ProviderHash, provider.Name())

	_, err = graphrag.NewProvider(graphrag.EmbeddingConfig{Provider: "nope"})
	require.ErrorIs(t, err, graphrag.ErrUnknownProvider)
}

func TestOpenAICompatProviderValidation(t *testing.T) {
	t.Parallel()

	_, err := graphrag.NewOpenAICompatProvider(graphrag.EmbeddingConfig{Model: "m"})
	require.ErrorIs(t, err, graphrag.ErrEmbedBaseURLRequired)

	_, err = graphrag.NewOpenAICompatProvider(graphrag.EmbeddingConfig{BaseURL: "https://x.test/v1"})
	require.ErrorIs(t, err, graphrag.ErrEmbedModelRequired)

	_, err = graphrag.NewOpenAICompatProvider(graphrag.EmbeddingConfig{BaseURL: "://broken", Model: "m"})
	require.ErrorIs(t, err, graphrag.ErrEmbedBaseURLRequired)
}

// failingProvider isolates provider-failure propagation in Build.
type failingProvider struct{}

func (failingProvider) Name() string  { return "failing" }
func (failingProvider) Model() string { return "v1" }
func (failingProvider) Embed(_ context.Context, _ []string) ([]graphrag.Vector, error) {
	return nil, errors.New("boom")
}

func TestBuildPropagatesProviderErrors(t *testing.T) {
	t.Parallel()

	_, err := graphrag.Build(
		t.Context(),
		failingProvider{},
		nil,
		[]graphrag.Document{
			{ID: "job:1", Kind: graphrag.KindJob, Label: "J", Text: "some text", Attrs: graphrag.NodeAttrs{}},
		},
		nil,
		graphrag.BuildOptions{},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}
