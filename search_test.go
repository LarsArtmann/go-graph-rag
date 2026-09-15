package graphrag_test

import (
	"testing"

	graphrag "github.com/larsartmann/go-graph-rag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildFixture assembles a small realistic index: two platform jobs, one
// unrelated job, skills, and a CV node with its own vector.
func buildFixture(t *testing.T) *graphrag.Searcher {
	t.Helper()

	docs := []graphrag.Document{
		{
			ID:    "job:k8s",
			Kind:  kindJob,
			Label: "Platform Engineer (K8s)",
			Text:  "kubernetes go platform infrastructure observability remote",
			Attrs: graphrag.NodeAttrs{},
		},
		{
			ID:    "job:sre",
			Kind:  kindJob,
			Label: "Site Reliability Engineer",
			Text:  "kubernetes go platform infrastructure incident response reliability",
			Attrs: graphrag.NodeAttrs{},
		},
		{
			ID:    "job:pastry",
			Kind:  kindJob,
			Label: "Pastry Chef",
			Text:  "croissant baking pastry kitchen",
			Attrs: graphrag.NodeAttrs{},
		},
		{
			ID:    "cv:en",
			Kind:  kindCV,
			Label: "CV",
			Text:  "go kubernetes platform infrastructure cloud",
			Attrs: graphrag.NodeAttrs{},
		},
		{ID: "skill:go", Kind: kindSkill, Label: "Go", Text: "", Attrs: graphrag.NodeAttrs{}},
		{ID: "skill:kubernetes", Kind: kindSkill, Label: "Kubernetes", Text: "", Attrs: graphrag.NodeAttrs{}},
		{ID: "company:acme", Kind: kindCompany, Label: "ACME", Text: "", Attrs: graphrag.NodeAttrs{}},
		{
			ID:    "project:p1",
			Kind:  kindProject,
			Label: "Infra Automation",
			Text:  "kubernetes go terraform automation",
			Attrs: graphrag.NodeAttrs{},
		},
	}

	edges := []graphrag.Edge{
		{Source: "job:k8s", Target: "skill:go", Relation: relRequires, Weight: 1},
		{Source: "job:k8s", Target: "skill:kubernetes", Relation: relRequires, Weight: 1},
		{Source: "job:k8s", Target: "company:acme", Relation: relAt, Weight: 1},
		{Source: "job:sre", Target: "skill:go", Relation: relPrefers, Weight: 1},
		{Source: "project:p1", Target: "skill:kubernetes", Relation: relUses, Weight: 1},
	}

	result, err := graphrag.Build(t.Context(), graphrag.NewHashProvider(), nil, docs, edges, graphrag.BuildOptions{})
	require.NoError(t, err)

	return graphrag.NewSearcherWithOptions(result.Nodes, result.Edges, result.Vectors, graphrag.SearcherOptions{
		ReferenceKind: kindCV,
		DocumentKinds: []graphrag.NodeKind{kindJob, kindProject},
		DetailKind:    kindSkill,
	})
}

func TestSearchRanksRelevantJobsFirst(t *testing.T) {
	t.Parallel()

	searcher := buildFixture(t)
	require.Equal(t, 8, searcher.NodeCount())

	result, err := searcher.Search(
		t.Context(),
		graphrag.NewHashProvider(),
		"kubernetes go infrastructure",
		graphrag.SearchOptions{MinScore: 0.1},
	)
	require.NoError(t, err)
	require.NotEmpty(t, result.Hits)

	assert.Equal(t, "job:k8s", result.Hits[0].Node.ID, "the k8s job must outrank the pastry chef")
	assert.NotContains(t, topHitIDs(result), "job:pastry", "zero-overlap jobs must not appear above the floor")

	assert.Greater(
		t,
		result.Hits[0].RefMatch,
		0.0,
		"reference similarity must be populated when a reference node exists",
	)

	skillSeen, companySeen := false, false

	for _, related := range result.Hits[0].Related {
		if related.Node.Kind == kindSkill {
			skillSeen = true
		}

		if related.Node.Kind == kindCompany {
			companySeen = true
		}
	}

	assert.True(t, skillSeen, "graph expansion must surface required skills")
	assert.True(t, companySeen, "graph expansion must surface the company")

	assert.Contains(t, result.ContextText, "Platform Engineer (K8s)")
	assert.Contains(t, result.ContextText, "skill:")
}

func topHitIDs(result *graphrag.SearchResult) []string {
	ids := make([]string, 0, len(result.Hits))

	for _, hit := range result.Hits {
		ids = append(ids, hit.Node.ID)
	}

	return ids
}

func TestSearchDeterministicContext(t *testing.T) {
	t.Parallel()

	searcher := buildFixture(t)
	provider := graphrag.NewHashProvider()

	first, err := searcher.Search(t.Context(), provider, "kubernetes go", graphrag.SearchOptions{})
	require.NoError(t, err)

	second, err := searcher.Search(t.Context(), provider, "kubernetes go", graphrag.SearchOptions{})
	require.NoError(t, err)

	assert.Equal(t, first.ContextText, second.ContextText, "identical retrieval must render byte-identical context")
	assert.Equal(t, first.Hits[0].Node.ID, second.Hits[0].Node.ID)
}

func TestSimilarPairsFindsNearDuplicates(t *testing.T) {
	t.Parallel()

	searcher := buildFixture(t)

	pairs := searcher.SimilarPairs(0.5)
	require.NotEmpty(t, pairs)

	var platformPairSeen bool

	for i, pair := range pairs {
		assert.NotEqual(t, pair.A.ID, pair.B.ID)
		assert.GreaterOrEqual(t, pair.Score, 0.5)

		if i > 0 {
			assert.GreaterOrEqual(t, pairs[i-1].Score, pair.Score, "pairs must be sorted by descending score")
		}

		assert.NotContains(t, []string{pair.A.ID, pair.B.ID}, "job:pastry", "unrelated jobs never reach the threshold")

		pairIDs := pair.A.ID + "|" + pair.B.ID
		if pairIDs == "job:k8s|job:sre" {
			platformPairSeen = true
		}
	}

	assert.True(t, platformPairSeen, "the two platform roles must be detected as near-duplicates")
}

func TestSearchRespectsMinScoreAndTopK(t *testing.T) {
	t.Parallel()

	searcher := buildFixture(t)

	result, err := searcher.Search(
		t.Context(),
		graphrag.NewHashProvider(),
		"kubernetes go",
		graphrag.SearchOptions{TopK: 1, MinScore: 0.99, MaxRelatedPerHit: 2},
	)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(result.Hits), 1)

	if len(result.Hits) == 1 {
		assert.LessOrEqual(t, len(result.Hits[0].Related), 2)
	}
}
