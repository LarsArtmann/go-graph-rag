package graphrag

import (
	"context"
	"fmt"
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
	Node    Node          `json:"node"`
	Score   float64       `json:"score"`
	CVMatch float64       `json:"cvMatch"`
	Related []RelatedNode `json:"related"`
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

// Searcher answers queries against an in-memory snapshot of the index. It
// is immutable and safe for concurrent use.
type Searcher struct {
	graph    *Graph
	vectors  map[string]Vector
	cvVector Vector
	hasCV    bool
}

// NewSearcher warm-starts a Searcher from a loaded index snapshot. Nodes
// without vectors (skills, companies) still participate through graph
// expansion. If exactly one CV-kind node carries a vector, it becomes the
// CVMatch reference.
func NewSearcher(nodes []Node, edges []Edge, vectors map[string]Vector) *Searcher {
	searcher := &Searcher{
		graph:    NewGraph(nodes, edges),
		vectors:  vectors,
		cvVector: nil,
		hasCV:    false,
	}

	for _, node := range nodes {
		if node.Kind != KindCV {
			continue
		}

		if vector, ok := vectors[node.ID]; ok {
			searcher.cvVector = vector
			searcher.hasCV = true
		}
	}

	return searcher
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
		ContextText: RenderContext(query, hits),
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

// rankHits scores every embedded node (except the CV reference itself) and
// keeps the top-k above the floor. Ranking is two-tier: document nodes
// (jobs, projects) come first in score order, hub nodes (skills, companies)
// fill the remaining slots. Hubs embed short labels whose cosine to a
// multi-term query is structurally higher than a full document's, so without
// the tiering a bare skill name would crowd out the very jobs it is a label
// for; with it, hub-only queries ("terraform") still surface the hub once no
// document matches.
func (s *Searcher) rankHits(query Vector, topK int, minScore float64) []Hit {
	documents := make([]Hit, 0, len(s.vectors))
	hubs := make([]Hit, 0, len(s.vectors))

	for id, vector := range s.vectors {
		node, ok := s.graph.Node(id)
		if !ok || node.Kind == KindCV {
			continue
		}

		score := Cosine(query, vector)
		if score < minScore {
			continue
		}

		hit := Hit{Node: node, Score: score, CVMatch: s.cvScore(vector), Related: nil}

		if node.Kind == KindJob || node.Kind == KindProject {
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

// cvScore measures a node's similarity to the indexed CV.
func (s *Searcher) cvScore(vector Vector) float64 {
	if !s.hasCV {
		return 0
	}

	return Cosine(s.cvVector, vector)
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

// RenderContext produces the deterministic LLM-ready text block for a query
// and its hits. Ordering is fully determined (hit rank, then relation, then
// label) so identical retrieval yields byte-identical context.
func RenderContext(query string, hits []Hit) string {
	var builder strings.Builder

	builder.WriteString("Retrieval context for query: \"")
	builder.WriteString(query)
	builder.WriteString("\"\n")

	for i, hit := range hits {
		fmt.Fprintf(&builder, "\n%d. [%s] %s (similarity %.2f", i+1, hit.Node.Kind, hit.Node.Label, hit.Score)

		if hit.CVMatch > 0 {
			fmt.Fprintf(&builder, ", CV match %.2f", hit.CVMatch)
		}

		builder.WriteString(")\n")

		writeRelated(&builder, hit.Related)
	}

	return builder.String()
}

// writeRelated renders one hit's neighborhood compactly.
func writeRelated(builder *strings.Builder, related []RelatedNode) {
	skills := make([]string, 0)
	others := make([]string, 0)

	for _, entry := range related {
		if entry.Node.Kind == KindSkill {
			skills = append(skills, entry.Node.Label)

			continue
		}

		detail := fmt.Sprintf("[%s] %s (%s %.2f)", entry.Node.Kind, entry.Node.Label, entry.Relation, entry.Weight)
		others = append(others, detail)
	}

	if len(skills) > 0 {
		fmt.Fprintf(builder, "   skills: %s\n", strings.Join(skills, ", "))
	}

	for _, other := range others {
		fmt.Fprintf(builder, "   related: %s\n", other)
	}
}
