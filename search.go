package graphrag

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
)

// Search defaults.
const (
	// DefaultTopK is the number of vector hits returned when SearchOptions
	// leaves TopK at zero.
	DefaultTopK = 5

	// DefaultMaxRelatedPerHit caps the graph expansion attached to each
	// hit so context assembly stays bounded.
	DefaultMaxRelatedPerHit = 6
)

// RelatedNode is one graph neighbor reached during expansion.
type RelatedNode struct {
	Node     Node     `json:"node"`
	Relation Relation `json:"relation"`
	Weight   float64  `json:"weight"`
}

// Hit is one vector-similar node with its graph neighborhood.
type Hit struct {
	Node     Node          `json:"node"`
	Score    float64       `json:"score"`
	RefMatch float64       `json:"refMatch"`
	Related  []RelatedNode `json:"related"`
}

// SearchOptions tunes one Search call.
type SearchOptions struct {
	TopK             int
	MinScore         float64
	MaxRelatedPerHit int
}

// SearchResult is the hybrid retrieval outcome: vector hits, graph
// expansion, and a deterministic, LLM-ready context rendering.
type SearchResult struct {
	Query       string `json:"query"`
	Provider    string `json:"provider"`
	Model       string `json:"model"`
	Hits        []Hit  `json:"hits"`
	ContextText string `json:"contextText"`
}

// SimilarPair names two nodes whose vectors are at least a threshold apart.
type SimilarPair struct {
	A     Node    `json:"a"`
	B     Node    `json:"b"`
	Score float64 `json:"score"`
}

// SearcherOptions tunes which node kinds play which role at search time.
// The zero value is neutral behavior: no reference node, every embedded
// node ranked as a document, and full detail rendering for every related
// node. Domain kind vocabularies stay caller-side constants.
type SearcherOptions struct {
	// ReferenceKind names the kind of the reference node (e.g. the profile
	// document in a matching graph). Reference nodes never appear in hits;
	// every hit carries RefMatch instead: its cosine similarity to the
	// reference vector. If several embedded nodes carry the reference
	// kind, the last one wins. Empty disables the reference concept.
	ReferenceKind NodeKind

	// DocumentKinds lists the kinds ranked as primary documents. Embedded
	// nodes of every other kind become expansion hubs that fill the
	// remaining top-k slots after the document slots. Empty treats every
	// embedded node as a document.
	DocumentKinds []NodeKind

	// DetailKind names the kind whose related nodes render as a compact
	// label list in RenderContext output. Empty renders every related
	// node in full detail form.
	DetailKind NodeKind
}

// Searcher answers queries against an in-memory snapshot of the index. It
// is immutable and safe for concurrent use.
type Searcher struct {
	graph     *Graph
	vectors   map[string]Vector
	policy    SearcherOptions
	refVector Vector
	hasRef    bool
}

// NewSearcher warm-starts a neutral Searcher from a loaded index snapshot:
// no reference node, every embedded node ranked as a document. Nodes
// without vectors still participate through graph expansion. Use
// NewSearcherWithOptions to assign kind-specific search roles.
func NewSearcher(nodes []Node, edges []Edge, vectors map[string]Vector) *Searcher {
	//nolint:exhaustruct_v5 // the neutral policy IS the zero value
	return NewSearcherWithOptions(nodes, edges, vectors, SearcherOptions{})
}

// NewSearcherWithOptions warm-starts a Searcher whose kind roles come from
// opts. Nodes without vectors (e.g. hub labels) still participate through
// graph expansion. If opts.ReferenceKind is set, the last embedded node of
// that kind becomes the RefMatch reference. The vectors map (including each
// vector's contents) and opts.DocumentKinds are cloned: the Searcher is
// immutable, so mutating the caller's copies afterwards cannot change its
// behavior.
func NewSearcherWithOptions(nodes []Node, edges []Edge, vectors map[string]Vector, opts SearcherOptions) *Searcher {
	searcher := &Searcher{
		graph:     NewGraph(nodes, edges),
		vectors:   cloneVectors(vectors),
		policy:    opts,
		refVector: nil,
		hasRef:    false,
	}

	searcher.policy.DocumentKinds = slices.Clone(opts.DocumentKinds)

	if opts.ReferenceKind == "" {
		return searcher
	}

	for _, node := range nodes {
		if node.Kind != opts.ReferenceKind {
			continue
		}

		if vector, ok := searcher.vectors[node.ID]; ok {
			searcher.refVector = vector
			searcher.hasRef = true
		}
	}

	return searcher
}

// cloneVectors deep-copies a vectors map so the Searcher owns its memory:
// the immutability promise covers both the map structure and vector
// contents, not just the map handle.
func cloneVectors(vectors map[string]Vector) map[string]Vector {
	cloned := make(map[string]Vector, len(vectors))

	for id, vector := range vectors {
		cloned[id] = slices.Clone(vector)
	}

	return cloned
}

// Search embeds the query, ranks embedded nodes by cosine similarity, and
// expands each hit with its graph neighborhood.
func (s *Searcher) Search(
	ctx context.Context,
	provider Provider,
	query string,
	opts SearchOptions,
) (*SearchResult, error) {
	topK := max(opts.TopK, 0)
	if topK == 0 {
		topK = DefaultTopK
	}

	maxRelated := max(opts.MaxRelatedPerHit, 0)
	if maxRelated == 0 {
		maxRelated = DefaultMaxRelatedPerHit
	}

	embedded, err := provider.Embed(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("graphrag: embed query: %w", err)
	}

	if len(embedded) != 1 {
		return nil, fmt.Errorf("%w: query embed returned %d vectors", ErrEmbedCountMismatch, len(embedded))
	}

	queryVector := embedded[0]
	hits := s.rankHits(queryVector, topK, opts.MinScore)

	for i := range hits {
		hits[i].Related = s.expand(hits[i].Node.ID, maxRelated)
	}

	return &SearchResult{
		Query:       query,
		Provider:    provider.Name(),
		Model:       provider.Model(),
		Hits:        hits,
		ContextText: s.RenderContext(query, hits),
	}, nil
}

// SimilarPairs returns every embedded node pair whose cosine similarity is
// at least minScore, sorted by descending score. This is the generic
// near-duplicate detector over the corpus — every embedded node participates
// (documents and hub nodes alike); callers wanting one kind filter the
// result.
func (s *Searcher) SimilarPairs(minScore float64) []SimilarPair {
	ids := make([]string, 0, len(s.vectors))

	for id := range s.vectors {
		ids = append(ids, id)
	}

	sort.Strings(ids)

	pairs := make([]SimilarPair, 0)

	for i := range ids {
		for j := i + 1; j < len(ids); j++ {
			score := Cosine(s.vectors[ids[i]], s.vectors[ids[j]])

			if score >= minScore {
				pairs = append(
					pairs,
					SimilarPair{A: s.nodeOrPlaceholder(ids[i]), B: s.nodeOrPlaceholder(ids[j]), Score: score},
				)
			}
		}
	}

	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Score != pairs[j].Score {
			return pairs[i].Score > pairs[j].Score
		}

		if pairs[i].A.ID != pairs[j].A.ID {
			return pairs[i].A.ID < pairs[j].A.ID
		}

		return pairs[i].B.ID < pairs[j].B.ID
	})

	return pairs
}

// NodeCount reports how many nodes the snapshot covers.
func (s *Searcher) NodeCount() int {
	return s.graph.NodeCount()
}

// VectorCount reports how many nodes carry vectors.
func (s *Searcher) VectorCount() int {
	return len(s.vectors)
}

// rankHits scores every embedded node (except the reference node itself)
// and keeps the top-k above the floor. Ranking is two-tier: document nodes
// come first in score order, hub nodes fill the remaining slots. Hubs embed
// short labels whose cosine to a multi-term query is structurally higher
// than a full document's, so without the tiering a bare label would crowd
// out the very documents it is a label for; with it, hub-only queries
// ("terraform") still surface the hub once no document matches.
func (s *Searcher) rankHits(query Vector, topK int, minScore float64) []Hit {
	documents := make([]Hit, 0, len(s.vectors))
	hubs := make([]Hit, 0, len(s.vectors))

	for id, vector := range s.vectors {
		node, ok := s.graph.Node(id)
		if !ok || (s.policy.ReferenceKind != "" && node.Kind == s.policy.ReferenceKind) {
			continue
		}

		score := Cosine(query, vector)
		if score < minScore {
			continue
		}

		hit := Hit{Node: node, Score: score, RefMatch: s.refScore(vector), Related: nil}

		if s.isDocument(node.Kind) {
			documents = append(documents, hit)
		} else {
			hubs = append(hubs, hit)
		}
	}

	sortHits(documents)
	sortHits(hubs)

	return append(documents, hubs...)[:min(topK, len(documents)+len(hubs))]
}

// sortHits orders hits by descending score, then by node ID for determinism.
func sortHits(hits []Hit) {
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}

		return hits[i].Node.ID < hits[j].Node.ID
	})
}

// isDocument reports whether kind belongs to the configured document
// kinds. With no DocumentKinds configured, every node is a document.
func (s *Searcher) isDocument(kind NodeKind) bool {
	if len(s.policy.DocumentKinds) == 0 {
		return true
	}

	return slices.Contains(s.policy.DocumentKinds, kind)
}

// refScore measures a node's similarity to the reference node.
func (s *Searcher) refScore(vector Vector) float64 {
	if !s.hasRef {
		return 0
	}

	return Cosine(s.refVector, vector)
}

// expand collects a node's outgoing and incoming edges as related nodes,
// ranked by weight then node ID, capped at max.
func (s *Searcher) expand(id string, limit int) []RelatedNode {
	edges := append(s.graph.OutEdges(id), s.graph.InEdges(id)...)

	related := make([]RelatedNode, 0, len(edges))

	for _, edge := range edges {
		neighborID := edge.Target
		if edge.Target == id {
			neighborID = edge.Source
		}

		node, ok := s.graph.Node(neighborID)
		if !ok {
			continue
		}

		related = append(related, RelatedNode{Node: node, Relation: edge.Relation, Weight: edge.Weight})
	}

	sort.Slice(related, func(i, j int) bool {
		if related[i].Weight != related[j].Weight {
			return related[i].Weight > related[j].Weight
		}

		return related[i].Node.ID < related[j].Node.ID
	})

	if len(related) > limit {
		related = related[:limit]
	}

	return related
}

// nodeOrPlaceholder keeps SimilarPair total even if a vector lost its node.
func (s *Searcher) nodeOrPlaceholder(id string) Node {
	if node, ok := s.graph.Node(id); ok {
		return node
	}

	//nolint:exhaustruct_v5 // placeholder for a vector whose node vanished; no attrs exist
	return Node{ID: id, Kind: KindUnknown, Label: id, Attrs: NodeAttrs{}}
}

// RenderContext produces the deterministic LLM-ready text block for a
// query and its hits. Ordering is fully determined (hit rank, then
// relation, then label) so identical retrieval yields byte-identical
// context. Rendering follows the searcher's kind policy (RefMatch line,
// compact detail-kind labels).
func (s *Searcher) RenderContext(query string, hits []Hit) string {
	var builder strings.Builder

	builder.WriteString("Retrieval context for query: \"")
	builder.WriteString(query)
	builder.WriteString("\"\n")

	for i, hit := range hits {
		fmt.Fprintf(&builder, "\n%d. [%s] %s (similarity %.2f", i+1, hit.Node.Kind, hit.Node.Label, hit.Score)

		if hit.RefMatch > 0 {
			fmt.Fprintf(&builder, ", ref match %.2f", hit.RefMatch)
		}

		builder.WriteString(")\n")

		s.writeRelated(&builder, hit.Related)
	}

	return builder.String()
}

// writeRelated renders one hit's neighborhood compactly. Related nodes of
// the configured DetailKind collapse into a comma-separated label list;
// everything else renders with kind, label, relation, and weight.
func (s *Searcher) writeRelated(builder *strings.Builder, related []RelatedNode) {
	details := make([]string, 0)
	others := make([]string, 0)

	for _, entry := range related {
		if s.policy.DetailKind != "" && entry.Node.Kind == s.policy.DetailKind {
			details = append(details, entry.Node.Label)

			continue
		}

		detail := fmt.Sprintf("[%s] %s (%s %.2f)", entry.Node.Kind, entry.Node.Label, entry.Relation, entry.Weight)
		others = append(others, detail)
	}

	if len(details) > 0 {
		fmt.Fprintf(builder, "   %s: %s\n", s.policy.DetailKind, strings.Join(details, ", "))
	}

	for _, other := range others {
		fmt.Fprintf(builder, "   related: %s\n", other)
	}
}
