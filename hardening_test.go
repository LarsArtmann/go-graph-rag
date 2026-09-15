package graphrag_test

import (
	"math"
	"strings"
	"sync"
	"testing"

	"github.com/LarsArtmann/CV/graphrag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHashProviderProperty pins the hash embedder's contract: fixed
// dimensionality, normalized output, determinism, and near-dup cosine
// bounds (the corpus-observed ~0.56 for near-duplicates justifies the
// default 0.75 similar threshold).
func TestHashProviderProperty(t *testing.T) {
	t.Parallel()

	provider := graphrag.NewHashProvider()

	t.Run("fixed dimensionality and unit norm", func(t *testing.T) {
		t.Parallel()

		vectors, err := provider.Embed(t.Context(), []string{"go kubernetes platform", "react css frontend"})
		require.NoError(t, err)
		require.Len(t, vectors, 2)

		for _, v := range vectors {
			assert.Equal(t, graphrag.HashProviderDims, v.Dims())

			norm := 0.0

			for _, c := range v {
				norm += float64(c) * float64(c)
			}

			assert.InDelta(t, 1.0, math.Sqrt(norm), 1e-6, "embeddings must be L2-normalized")
		}
	})

	t.Run("near-duplicate cosine bounds", func(t *testing.T) {
		t.Parallel()

		// The hash embedder is token-overlap based: near-identical postings
		// score high (0.85+), unrelated text scores near zero. The contract
		// is a clear, monotone gap between related and unrelated text — the
		// corpus-observed ~0.56 near-dup band is for real-world postings
		// with different wording, not synthetic near-copies.
		base := "kubernetes go platform infrastructure observability remote linux containers networking service mesh"
		variants := []string{
			base,
			base + " devops ci cd terraform",
			strings.ToUpper(base),
			"kubernetes go platform infrastructure observability remote linux containers",
		}

		vectors, err := provider.Embed(t.Context(), variants)
		require.NoError(t, err)

		var relatedScores []float64

		for i := 1; i < len(vectors); i++ {
			cos := graphrag.Cosine(vectors[0], vectors[i])
			relatedScores = append(relatedScores, cos)
			assert.Greater(t, cos, 0.5, "variant %d must be recognizably similar to the base", i)
		}

		unrelated, err := provider.Embed(t.Context(), []string{"baking croissants pastry kitchen butter flour sugar"})
		require.NoError(t, err)

		unrelatedScore := graphrag.Cosine(vectors[0], unrelated[0])

		for _, cos := range relatedScores {
			assert.Greater(
				t,
				cos,
				unrelatedScore+0.4,
				"every related variant (%.3f) must outscore unrelated text (%.3f) by a clear margin",
				cos,
				unrelatedScore,
			)
		}
	})
}

// TestBuildNodeIDCollisionPolicy pins the duplicate-ID collapse: the same
// node reached from a job and a CV project lands ONCE, first occurrence
// wins, and persistence never hits a unique constraint.
func TestBuildNodeIDCollisionPolicy(t *testing.T) {
	t.Parallel()

	docs := []graphrag.Document{
		{
			ID:    "skill:go",
			Kind:  graphrag.KindSkill,
			Label: "Go",
			Text:  "go programming language",
			Attrs: graphrag.NodeAttrs{},
		},
		{ID: "job:1", Kind: graphrag.KindJob, Label: "Go Engineer", Text: "go kubernetes", Attrs: graphrag.NodeAttrs{}},
		{
			ID:    "skill:go",
			Kind:  graphrag.KindSkill,
			Label: "Go (duplicate)",
			Text:  "go programming language",
			Attrs: graphrag.NodeAttrs{},
		},
	}

	result, err := graphrag.Build(
		t.Context(),
		graphrag.NewHashProvider(),
		nil,
		docs,
		nil,
		graphrag.BuildOptions{},
	)
	require.NoError(t, err)

	require.Len(t, result.Nodes, 2, "duplicate ID must collapse onto its first occurrence")
	assert.Equal(t, "Go", result.Nodes[0].Label, "first occurrence's label wins")

	// Persisting the collapsed result must not hit a unique constraint.
	store, err := graphrag.OpenStore(":memory:")
	require.NoError(t, err)

	defer func() { _ = store.Close() }()

	require.NoError(t, store.ReplaceGraph(result.Nodes, result.EmbeddingHash, result.Edges))

	snapshot, err := store.LoadGraph()
	require.NoError(t, err)
	require.Len(t, snapshot.Nodes, 2)
}

// TestPutEmbeddingsConcurrentRace hammers the embedding cache from parallel
// writers; the store serializes via its single connection, so every write
// must land and reads must never race.
func TestPutEmbeddingsConcurrentRace(t *testing.T) {
	t.Parallel()

	store, err := graphrag.OpenStore(":memory:")
	require.NoError(t, err)

	defer func() { _ = store.Close() }()

	const writers = 8

	const perWriter = 20

	var wg sync.WaitGroup
	wg.Add(writers)

	for w := range writers {
		go func(w int) {
			defer wg.Done()

			for i := range perWriter {
				text := strings.Repeat("text", w+i+1)
				entry := graphrag.EmbeddingEntry{
					Hash:   graphrag.HashText(text),
					Vector: graphrag.Vector{float32(w), float32(i), 1}.Normalize(),
				}

				if err := store.PutEmbeddings("hash", "v1", []graphrag.EmbeddingEntry{entry}); err != nil {
					t.Errorf("writer %d: %v", w, err)

					return
				}
			}
		}(w)
	}

	wg.Wait()

	// Every distinct hash must be readable back.
	for w := range writers {
		for i := range perWriter {
			text := strings.Repeat("text", w+i+1)
			hash := graphrag.HashText(text)

			_, found, err := store.CachedEmbedding("hash", "v1", hash)
			require.NoError(t, err)
			assert.True(t, found, "hash %s must survive concurrent writes", hash)
		}
	}
}

// TestSearcherEdgeCases pins the retrieval edges: empty corpus, single hit,
// and threshold boundaries.
func TestSearcherEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("empty corpus returns no hits", func(t *testing.T) {
		t.Parallel()

		searcher := graphrag.NewSearcher(nil, nil, map[string]graphrag.Vector{})
		require.Equal(t, 0, searcher.NodeCount())
		require.Equal(t, 0, searcher.VectorCount())

		result, err := searcher.Search(
			t.Context(),
			graphrag.NewHashProvider(),
			"anything",
			graphrag.SearchOptions{},
		)
		require.NoError(t, err)
		assert.Empty(t, result.Hits)
		assert.Contains(t, result.ContextText, "anything")
	})

	t.Run("single hit returns exactly one", func(t *testing.T) {
		t.Parallel()

		docs := []graphrag.Document{
			{
				ID:    "job:only",
				Kind:  graphrag.KindJob,
				Label: "Only Role",
				Text:  "go kubernetes platform",
				Attrs: graphrag.NodeAttrs{},
			},
		}

		result, err := graphrag.Build(t.Context(), graphrag.NewHashProvider(), nil, docs, nil, graphrag.BuildOptions{})
		require.NoError(t, err)

		searcher := graphrag.NewSearcher(result.Nodes, result.Edges, result.Vectors)

		got, err := searcher.Search(t.Context(), graphrag.NewHashProvider(), "go kubernetes", graphrag.SearchOptions{})
		require.NoError(t, err)
		require.Len(t, got.Hits, 1)
		assert.Equal(t, "job:only", got.Hits[0].Node.ID)
	})

	t.Run("threshold boundary is inclusive", func(t *testing.T) {
		t.Parallel()

		docs := []graphrag.Document{
			{ID: "job:a", Kind: graphrag.KindJob, Label: "A", Text: "alpha beta gamma", Attrs: graphrag.NodeAttrs{}},
			{
				ID:    "job:b",
				Kind:  graphrag.KindJob,
				Label: "B",
				Text:  "alpha beta gamma delta",
				Attrs: graphrag.NodeAttrs{},
			},
		}

		result, err := graphrag.Build(t.Context(), graphrag.NewHashProvider(), nil, docs, nil, graphrag.BuildOptions{})
		require.NoError(t, err)

		searcher := graphrag.NewSearcher(result.Nodes, result.Edges, result.Vectors)

		// At threshold 0 both pairs qualify; at 0.99 none do.
		pairs := searcher.SimilarPairs(0)
		assert.NotEmpty(t, pairs, "threshold 0 admits every pair")

		pairs = searcher.SimilarPairs(0.99)
		assert.Empty(t, pairs, "threshold 0.99 excludes every pair")
	})
}

// TestRenderContextGolden pins the exact LLM-ready context block for a
// known retrieval, so a formatting change fails loudly instead of silently
// changing what the model sees.
func TestRenderContextGolden(t *testing.T) {
	t.Parallel()

	hits := []graphrag.Hit{
		{
			Node:  graphrag.Node{ID: "job:k8s", Kind: graphrag.KindJob, Label: "Platform Engineer (K8s)"},
			Score: 0.82,
			Related: []graphrag.RelatedNode{
				{
					Node:     graphrag.Node{ID: "skill:go", Kind: graphrag.KindSkill, Label: "Go"},
					Relation: graphrag.RelationRequires,
					Weight:   1,
				},
				{
					Node:     graphrag.Node{ID: "company:acme", Kind: graphrag.KindCompany, Label: "ACME"},
					Relation: graphrag.RelationAt,
					Weight:   1,
				},
			},
		},
	}

	got := graphrag.RenderContext("kubernetes go", hits)

	const want = `Retrieval context for query: "kubernetes go"

1. [job] Platform Engineer (K8s) (similarity 0.82)
   skills: Go
   related: [company] ACME (at 1.00)
`

	assert.Equal(t, want, got, "context block must be byte-deterministic")
}
