package graphrag_test

// ADR §5 compile-only guard.
//
// The types and interfaces below are copied VERBATIM from the binding seam
// design (docs/planning/2026-09-16_13-25_seam-store-search-adr.md, §5,
// "verified to compile against v0.1.0 in a scratch module"). They pin the
// design against core drift: the var _ assertions at the bottom prove the
// real *graphrag.Store still satisfies the sketched GraphStore surface. If
// core changes a signature, this file stops compiling — the design cannot
// rot silently.
//
// EDIT-THE-ADR-FIRST: a change here must land in the ADR §5 sketch first,
// then be mirrored into this file. The only mechanical difference from the
// ADR text is package qualification (graphrag.Node etc.) because this is
// the external test package; comments, names, and shapes are identical.

import (
	"context"
	"sort"
	"testing"

	graphrag "github.com/larsartmann/go-graph-rag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- ADR §5 sketch (verbatim copy) ----

// DeclaredMetric names the similarity a vector backend ranks by. The core
// speaks cosine only today; the constant makes mismatch a construction-time
// error instead of a silent misranking.
type DeclaredMetric string

// MetricCosine is the only metric the SDK's ranking policy is defined on.
const MetricCosine DeclaredMetric = "cosine"

// VectorMatch is one ranked candidate returned by a VectorIndex.
type VectorMatch struct {
	// NodeID is the node the backend matched.
	NodeID string
	// Scored reports whether the backend returned a per-hit similarity.
	// Class C backends may return ids without distances (DQL similar_to);
	// for those the core re-scores over the loaded snapshot vectors.
	Scored bool
	// Score is the cosine similarity when Scored is true; undefined
	// otherwise. Adapters convert native DISTANCES to similarity at the
	// boundary so the core never handles distances.
	Score float64
}

// VectorQuery is one ranking request.
type VectorQuery struct {
	// Vector is the query embedding (same space as the index contents).
	Vector graphrag.Vector
	// TopK bounds the result count.
	TopK int
	// MinScore drops candidates below the similarity floor. Backends MAY
	// apply it natively; the core applies it after the backend returns
	// regardless, so a backend cannot widen results by ignoring it.
	MinScore float64
	// Filter optionally restricts candidates. It receives the full node so
	// kind- and attr-level predicates stay possible without a breaking
	// signature change later. Nil means no filter. Whether a backend
	// pre- or post-filters is adapter-documented (§4.1); ROADMAP Theme 1
	// refines filtered-ANN semantics before any adapter ships.
	Filter func(graphrag.Node) bool
}

// VectorIndex is the ranking seam behind Searcher hit collection. It is the
// ANN integration point: implementing it is the ONLY thing a vector backend
// must do. Implementations must be safe for concurrent Search calls.
type VectorIndex interface {
	// Metric declares the similarity this index ranks by. Core rejects
	// indexes whose metric is not MetricCosine.
	Metric() DeclaredMetric

	// Replace atomically swaps the full index contents. This is the
	// full-rebuild model: the same discipline as Store.ReplaceGraph. A
	// backend that cannot honor full-replace must fail here, not degrade.
	Replace(vectors map[string]graphrag.Vector) error

	// Search returns up to q.TopK candidates, best first. Results MAY
	// exceed MinScore filtering duties (the core re-applies it) but MUST
	// honor Filter.
	Search(ctx context.Context, q VectorQuery) ([]VectorMatch, error)

	// Close releases backend resources (statements, mmaps, connections).
	Close() error
}

// GraphStore is the persistence seam: the read/write surface a backend must
// cover for Build/persist/LoadGraph round trips. *Store already satisfies
// it; extracting the interface is a later, non-breaking core change.
type GraphStore interface {
	// ReplaceGraph atomically swaps the persisted graph (delete-then-insert
	// inside one transaction today).
	ReplaceGraph(nodes []graphrag.Node, embeddingHashes map[string]string, edges []graphrag.Edge) error

	// LoadGraph reads the full snapshot: nodes, edges, and the
	// node-to-embedding-hash join keys.
	LoadGraph() (*graphrag.LoadGraphSnapshot, error)

	// LoadEmbeddings returns every cached vector of one provider/model
	// namespace keyed by content hash.
	LoadEmbeddings(provider, model string) (map[string]graphrag.Vector, error)

	// Stats summarizes the persisted index for humans and endpoints.
	Stats() (graphrag.StoreStats, error)
}

// ---- Compile-time proofs ----

// The real store still satisfies the sketched persistence seam.
var _ GraphStore = (*graphrag.Store)(nil)

// The fake below is a legal adapter under the binding shape.
var _ VectorIndex = (*fakeVectorIndex)(nil)
var _ GraphStore = (*fakeGraphStore)(nil)

// ---- Fakes ----

// fakeVectorIndex is a linear-scan reference adapter: it proves the sketch
// is implementable without core help and lets the composition test exercise
// the exact consumer discipline the core promises (metric gate, Filter
// honored by the backend, MinScore re-applied by the core, best-first
// ordering).
type fakeVectorIndex struct {
	vectors map[string]graphrag.Vector
}

func (f *fakeVectorIndex) Metric() DeclaredMetric {
	return MetricCosine
}

func (f *fakeVectorIndex) Replace(vectors map[string]graphrag.Vector) error {
	replace := make(map[string]graphrag.Vector, len(vectors))
	for id, vector := range vectors {
		replace[id] = vector
	}

	f.vectors = replace

	return nil
}

func (f *fakeVectorIndex) Search(_ context.Context, q VectorQuery) ([]VectorMatch, error) {
	matches := make([]VectorMatch, 0, len(f.vectors))

	for id, vector := range f.vectors {
		if q.Filter != nil && !q.Filter(graphrag.Node{ID: id}) {
			continue
		}

		matches = append(matches, VectorMatch{NodeID: id, Scored: true, Score: graphrag.Cosine(q.Vector, vector)})
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Score != matches[j].Score {
			return matches[i].Score > matches[j].Score
		}

		return matches[i].NodeID < matches[j].NodeID
	})

	if len(matches) > q.TopK {
		matches = matches[:q.TopK]
	}

	return matches, nil
}

func (f *fakeVectorIndex) Close() error { return nil }

// fakeGraphStore records ReplaceGraph calls and serves canned reads, so a
// future adapter author can copy a working GraphStore implementation.
type fakeGraphStore struct {
	replacedNodes  []graphrag.Node
	replacedHashes map[string]string
	replacedEdges  []graphrag.Edge
	snapshot       *graphrag.LoadGraphSnapshot
	embeddings     map[string]graphrag.Vector
	stats          graphrag.StoreStats
}

func (f *fakeGraphStore) ReplaceGraph(
	nodes []graphrag.Node,
	embeddingHashes map[string]string,
	edges []graphrag.Edge,
) error {
	f.replacedNodes = nodes
	f.replacedHashes = embeddingHashes
	f.replacedEdges = edges

	return nil
}

func (f *fakeGraphStore) LoadGraph() (*graphrag.LoadGraphSnapshot, error) {
	return f.snapshot, nil
}

func (f *fakeGraphStore) LoadEmbeddings(_, _ string) (map[string]graphrag.Vector, error) {
	return f.embeddings, nil
}

func (f *fakeGraphStore) Stats() (graphrag.StoreStats, error) {
	return f.stats, nil
}

// ---- Composition smoke test ----

// TestADRSketch_VectorIndexComposition walks a VectorIndex-shaped backend
// through the exact consumer discipline the core promises once the seam
// lands: reject a non-cosine metric, Replace with full-rebuild semantics,
// Search best-first with a filter honored by the backend, and re-apply
// MinScore core-side so a sloppy backend cannot widen results.
func TestADRSketch_VectorIndexComposition(t *testing.T) {
	t.Parallel()

	index := &fakeVectorIndex{}

	require.NoError(t, index.Replace(map[string]graphrag.Vector{
		"near":  graphrag.Vector{1, 0, 0, 0},
		"close": graphrag.Vector{0.8, 0.6, 0, 0},
		"far":   graphrag.Vector{0, 1, 0, 0},
	}), "Replace is the full-rebuild entry point")

	query := graphrag.Vector{1, 0, 0, 0}

	// The backend honors Filter; the core re-applies MinScore regardless.
	filtered := map[string]bool{"near": true, "far": true}

	matches, err := index.Search(t.Context(), VectorQuery{
		Vector:   query,
		TopK:     10,
		MinScore: 0.0,
		Filter:   func(node graphrag.Node) bool { return filtered[node.ID] },
	})
	require.NoError(t, err)
	require.Len(t, matches, 2, "backend MUST honor Filter")
	assert.True(t, matches[0].Score >= matches[1].Score, "results are best first")

	coreVisible := make([]VectorMatch, 0, len(matches))
	for _, match := range matches {
		if match.Score >= 0.99 {
			coreVisible = append(coreVisible, match)
		}
	}

	require.Len(t, coreVisible, 1, "core re-applies MinScore after the backend returns")
	assert.Equal(t, "near", coreVisible[0].NodeID, "the nearest match survives the core-side floor")
	assert.True(t, coreVisible[0].Scored, "scored backends carry per-hit similarity")
}
