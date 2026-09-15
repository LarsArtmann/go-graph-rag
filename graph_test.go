package graphrag_test

import (
	"testing"

	"github.com/LarsArtmann/CV/graphrag"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphDeterministicOrdering(t *testing.T) {
	t.Parallel()

	nodes := []graphrag.Node{
		{ID: "job:b", Kind: graphrag.KindJob, Label: "Second", Attrs: graphrag.NodeAttrs{}},
		{ID: "job:a", Kind: graphrag.KindJob, Label: "First", Attrs: graphrag.NodeAttrs{}},
		{ID: "skill:go", Kind: graphrag.KindSkill, Label: "Go", Attrs: graphrag.NodeAttrs{}},
	}

	edges := []graphrag.Edge{
		{Source: "job:a", Target: "skill:go", Relation: graphrag.RelationRequires, Weight: 1},
		{Source: "job:b", Target: "skill:go", Relation: graphrag.RelationPrefers, Weight: 1},
		{Source: "job:a", Target: "skill:go", Relation: graphrag.RelationRequires, Weight: 0.5},
	}

	graph := graphrag.NewGraph(nodes, edges)

	assert.Equal(t, []string{"job:a", "job:b", "skill:go"}, nodeIDs(graph.Nodes()))
	assert.Equal(t, 2, graph.EdgeCount(), "duplicate edge keys collapse onto max weight")

	jobARequires, ok := graph.Node("job:a")
	require.True(t, ok)
	assert.Equal(t, "First", jobARequires.Label)

	_, ok = graph.Node("missing")
	require.False(t, ok)

	out := graph.OutEdges("job:a")
	require.Len(t, out, 1)
	assert.InDelta(t, 1.0, out[0].Weight, 1e-9, "max weight must survive the merge")

	in := graph.InEdges("skill:go")
	require.Len(t, in, 2)
	assert.Equal(t, graphrag.RelationPrefers, in[0].Relation, "edges sort by relation first")
}

func nodeIDs(nodes []graphrag.Node) []string {
	return lo.Map(nodes, func(node graphrag.Node, _ int) string { return node.ID })
}

func TestGraphEdgeCloneSafety(t *testing.T) {
	t.Parallel()

	graph := graphrag.NewGraph(
		[]graphrag.Node{{ID: "a", Kind: graphrag.KindSkill, Label: "A", Attrs: graphrag.NodeAttrs{}}},
		[]graphrag.Edge{{Source: "x", Target: "a", Relation: graphrag.RelationUses, Weight: 1}},
	)

	out := graph.OutEdges("x")
	require.Len(t, out, 1)

	out[0].Weight = 99

	assert.InDelta(t, 1.0, graph.OutEdges("x")[0].Weight, 1e-9, "callers must not mutate graph internals")
}

func TestGraphNodesByKind(t *testing.T) {
	t.Parallel()

	graph := graphrag.NewGraph([]graphrag.Node{
		{ID: "job:a", Kind: graphrag.KindJob, Label: "A", Attrs: graphrag.NodeAttrs{}},
		{ID: "skill:b", Kind: graphrag.KindSkill, Label: "B", Attrs: graphrag.NodeAttrs{}},
		{ID: "job:c", Kind: graphrag.KindJob, Label: "C", Attrs: graphrag.NodeAttrs{}},
	}, nil)

	assert.Equal(t, []string{"job:a", "job:c"}, nodeIDs(graph.NodesByKind(graphrag.KindJob)))
	assert.Empty(t, graph.NodesByKind(graphrag.KindCV))
}
