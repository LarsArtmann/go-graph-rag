package graphrag_test

import (
	"testing"

	"github.com/larsartmann/go-graph-rag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreEmbeddingCacheRoundTrip(t *testing.T) {
	t.Parallel()

	store, err := graphrag.OpenStore(":memory:")
	require.NoError(t, err)

	defer func() { _ = store.Close() }()

	vector := graphrag.Vector{0.1, 0.2, 0.3}
	hash := graphrag.HashText("cached text")

	_, found, err := store.CachedEmbedding("hash", "v1", hash)
	require.NoError(t, err)
	require.False(t, found)

	require.NoError(
		t,
		store.PutEmbeddings("hash", "v1", []graphrag.EmbeddingEntry{{Hash: hash, Vector: vector.Normalize()}}),
	)

	cached, found, err := store.CachedEmbedding("hash", "v1", hash)
	require.NoError(t, err)
	require.True(t, found)
	assert.InDeltaSlice(t, []float32(vector.Normalize()), []float32(cached), 1e-6)

	_, found, err = store.CachedEmbedding("hash", "v2", hash)
	require.NoError(t, err)
	require.False(t, found, "model namespaces must not bleed into each other")
}

func TestStoreGraphPersistenceRoundTrip(t *testing.T) {
	t.Parallel()

	store, err := graphrag.OpenStore(":memory:")
	require.NoError(t, err)

	defer func() { _ = store.Close() }()

	posted := []graphrag.Node{
		{ID: "job:1", Kind: kindJob, Label: "Platform Engineer", Attrs: graphrag.NodeAttrs{
			URL:      "https://jobs.test/1",
			Source:   "greenhouse",
			Location: "Remote",
			Status:   "evaluated",
			Score:    4.2,
		}},
		{ID: "skill:go", Kind: kindSkill, Label: "Go", Attrs: graphrag.NodeAttrs{}},
	}

	edges := []graphrag.Edge{
		{Source: "job:1", Target: "skill:go", Relation: relRequires, Weight: 1},
	}

	hashes := map[string]string{"job:1": graphrag.HashText("job one text")}

	require.NoError(t, store.ReplaceGraph(posted, hashes, edges))

	snapshot, err := store.LoadGraph()
	require.NoError(t, err)
	require.Len(t, snapshot.Nodes, 2)
	require.Len(t, snapshot.Edges, 1)
	assert.Equal(t, posted[0].ID, snapshot.Nodes[0].ID)
	assert.Equal(t, posted[0].Attrs.URL, snapshot.Nodes[0].Attrs.URL)
	assert.InDelta(t, 4.2, snapshot.Nodes[0].Attrs.Score, 1e-9)
	assert.Equal(t, hashes["job:1"], snapshot.EmbeddingHash["job:1"])

	stats, err := store.Stats()
	require.NoError(t, err)
	assert.Equal(t, 2, stats.Nodes)
	assert.Equal(t, 1, stats.Edges)
	assert.Equal(t, 1, stats.ByKind["job"])

	require.NoError(t, store.ReplaceGraph(nil, nil, nil))

	stats, err = store.Stats()
	require.NoError(t, err)
	assert.Equal(t, 0, stats.Nodes, "ReplaceGraph is a full swap")
}

func TestBuildEndToEndWithCache(t *testing.T) {
	t.Parallel()

	store, err := graphrag.OpenStore(":memory:")
	require.NoError(t, err)

	defer func() { _ = store.Close() }()

	provider := graphrag.NewHashProvider()

	docs := []graphrag.Document{
		{
			ID:    "job:k8s",
			Kind:  kindJob,
			Label: "K8s Platform Engineer",
			Text:  "kubernetes go platform infrastructure remote",
			Attrs: graphrag.NodeAttrs{},
		},
		{
			ID:    "job:chef",
			Kind:  kindJob,
			Label: "K8s Infrastructure Role",
			Text:  "kubernetes go platform infrastructure devops",
			Attrs: graphrag.NodeAttrs{},
		},
		{
			ID:    "job:pastry",
			Kind:  kindJob,
			Label: "Pastry Chef",
			Text:  "baking croissants pastry kitchen butter",
			Attrs: graphrag.NodeAttrs{},
		},
		{ID: "skill:go", Kind: kindSkill, Label: "Go", Text: "", Attrs: graphrag.NodeAttrs{}},
	}

	edges := []graphrag.Edge{
		{Source: "job:k8s", Target: "skill:go", Relation: relRequires, Weight: 1},
	}

	first, err := graphrag.Build(
		t.Context(),
		provider,
		store,
		docs,
		edges,
		graphrag.BuildOptions{SimilarThreshold: 0.5},
	)
	require.NoError(t, err)

	assert.Equal(t, 3, first.Embedded, "three embeddable documents")
	assert.Equal(t, 0, first.CacheHits)
	assert.Len(t, first.Nodes, 4)

	similarCount := countRelation(first.Edges, graphrag.RelationSimilar)
	assert.Equal(t, 1, similarCount, "only the two k8s jobs clear the explicit threshold")

	require.NoError(t, store.ReplaceGraph(first.Nodes, first.EmbeddingHash, first.Edges))

	second, err := graphrag.Build(
		t.Context(),
		provider,
		store,
		docs,
		edges,
		graphrag.BuildOptions{SimilarThreshold: 0.5},
	)
	require.NoError(t, err)
	assert.Equal(t, 3, second.CacheHits, "second build must be fully served from cache")
	assert.Equal(t, 0, second.Embedded)
}

func TestBuildSimilarEdgesLinkSameKindOnly(t *testing.T) {
	t.Parallel()

	docs := []graphrag.Document{
		{ID: "job:a", Kind: kindJob, Label: "A", Text: "alpha beta gamma", Attrs: graphrag.NodeAttrs{}},
		{ID: "job:b", Kind: kindJob, Label: "B", Text: "alpha beta gamma delta", Attrs: graphrag.NodeAttrs{}},
		{ID: "cv:x", Kind: kindCV, Label: "CV", Text: "alpha beta gamma", Attrs: graphrag.NodeAttrs{}},
		{
			ID:    "skill:one",
			Kind:  kindSkill,
			Label: "One",
			Text:  "beta gamma epsilon",
			Attrs: graphrag.NodeAttrs{},
		},
		{
			ID:    "skill:two",
			Kind:  kindSkill,
			Label: "Two",
			Text:  "beta gamma epsilon zeta",
			Attrs: graphrag.NodeAttrs{},
		},
	}

	result, err := graphrag.Build(
		t.Context(),
		graphrag.NewHashProvider(),
		nil,
		docs,
		nil,
		graphrag.BuildOptions{SimilarThreshold: 0.4},
	)
	require.NoError(t, err)

	var similar []graphrag.Edge

	for _, edge := range result.Edges {
		if edge.Relation == graphrag.RelationSimilar {
			similar = append(similar, edge)
		}
	}

	require.Len(t, similar, 2, "job:a↔job:b and skill:one↔skill:two clear the threshold; "+
		"cv:x (cosine 1.0 to job:a) and the job↔skill pairs (0.6) must stay unlinked: similar is same-kind only")

	for _, edge := range similar {
		require.Equal(t, kindOf(result, edge.Source), kindOf(result, edge.Target),
			"similar edges must never cross node kinds")
	}
}

// kindOf resolves a node's kind from a build result.
func kindOf(result *graphrag.BuildResult, id string) graphrag.NodeKind {
	for _, node := range result.Nodes {
		if node.ID == id {
			return node.Kind
		}
	}

	return ""
}

func countRelation(edges []graphrag.Edge, relation graphrag.Relation) int {
	count := 0

	for _, edge := range edges {
		if edge.Relation == relation {
			count++
		}
	}

	return count
}
