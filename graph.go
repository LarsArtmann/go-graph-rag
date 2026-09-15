package graphrag

import (
	"slices"
	"sort"
	"time"
)

// NodeKind enumerates the entity types in the knowledge graph.
type NodeKind string

// KindUnknown is NOT a graph kind: it is the zero value a placeholder node
// carries when a vector's node vanished (search.go nodeOrPlaceholder) —
// named here so the empty string stays grep-able. Domain kinds are the
// caller's vocabulary: define your own NodeKind constants per domain.
const (
	KindUnknown NodeKind = ""
)

// Relation enumerates the edge types in the knowledge graph.
type Relation string

// The rest of the edge-relation vocabulary is the caller's domain
// vocabulary: define your own Relation constants per domain. RelationSimilar
// is the one SDK-derived relation, computed at index time from vector
// cosine similarity. All relations are traversed in both directions by the
// Searcher.
const (
	// RelationSimilar links two embedded nodes whose vectors are close;
	// its weight is the cosine similarity.
	RelationSimilar Relation = "similar_to"
)

// NodeAttrs carries the per-kind display metadata. Fields left empty are
// simply not applicable to the node's kind; keeping one typed struct (rather
// than per-kind payloads) keeps the storage layer and wire contracts flat.
type NodeAttrs struct {
	URL          string     `json:"url,omitempty"`
	Source       string     `json:"source,omitempty"`
	Location     string     `json:"location,omitempty"`
	RemotePolicy string     `json:"remotePolicy,omitempty"`
	Status       string     `json:"status,omitempty"`
	// Score is a display hint (e.g. a match score); zero means absent.
	// Under GOEXPERIMENT=jsonv2 the omitempty option is a no-op on
	// numerics, so the key may be emitted as 0 there; both engines decode
	// it back to the same zero.
	Score        float64    `json:"score,omitempty"`
	PostedAt     *time.Time `json:"postedAt,omitempty"`
}

// Node is one entity in the knowledge graph.
type Node struct {
	ID    string    `json:"id"`
	Kind  NodeKind  `json:"kind"`
	Label string    `json:"label"`
	Attrs NodeAttrs `json:"attrs"`
}

// Edge is one typed, weighted relation between two nodes.
type Edge struct {
	Source   string   `json:"source"`
	Target   string   `json:"target"`
	Relation Relation `json:"relation"`
	Weight   float64  `json:"weight"`
}

// Graph is an immutable in-memory view over nodes and edges with
// deterministic iteration everywhere: accessors sort by ID (nodes) and by
// (relation, target) or (relation, source) (edges), so renders and exports
// are byte-stable across processes.
type Graph struct {
	nodes    map[string]Node
	outEdges map[string][]Edge
	inEdges  map[string][]Edge
}

// NewGraph indexes the given nodes and edges. Duplicate edges (same source,
// target, and relation) collapse onto the maximum weight so that repeated
// indexing stays idempotent.
func NewGraph(nodes []Node, edges []Edge) *Graph {
	graph := &Graph{
		nodes:    make(map[string]Node, len(nodes)),
		outEdges: make(map[string][]Edge),
		inEdges:  make(map[string][]Edge),
	}

	for _, node := range nodes {
		graph.nodes[node.ID] = node
	}

	merged := mergeEdges(edges)

	for _, edge := range merged {
		graph.outEdges[edge.Source] = append(graph.outEdges[edge.Source], edge)
		graph.inEdges[edge.Target] = append(graph.inEdges[edge.Target], edge)
	}

	for id := range graph.outEdges {
		sortEdges(graph.outEdges[id])
	}

	for id := range graph.inEdges {
		sortEdges(graph.inEdges[id])
	}

	return graph
}

// Node returns the node with the given ID.
func (g *Graph) Node(id string) (Node, bool) {
	node, ok := g.nodes[id]

	return node, ok
}

// Nodes returns all nodes sorted by ID.
func (g *Graph) Nodes() []Node {
	all := make([]Node, 0, len(g.nodes))

	for _, node := range g.nodes {
		all = append(all, node)
	}

	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })

	return all
}

// NodesByKind returns the nodes of one kind sorted by ID.
func (g *Graph) NodesByKind(kind NodeKind) []Node {
	matched := make([]Node, 0)

	for _, node := range g.nodes {
		if node.Kind == kind {
			matched = append(matched, node)
		}
	}

	sort.Slice(matched, func(i, j int) bool { return matched[i].ID < matched[j].ID })

	return matched
}

// OutEdges returns the edges leaving id, deterministically ordered.
func (g *Graph) OutEdges(id string) []Edge {
	return slices.Clone(g.outEdges[id])
}

// InEdges returns the edges entering id, deterministically ordered.
func (g *Graph) InEdges(id string) []Edge {
	return slices.Clone(g.inEdges[id])
}

// EdgeCount returns the number of distinct edges.
func (g *Graph) EdgeCount() int {
	count := 0

	for _, edges := range g.outEdges {
		count += len(edges)
	}

	return count
}

// NodeCount returns the number of nodes.
func (g *Graph) NodeCount() int {
	return len(g.nodes)
}

// mergeEdges collapses duplicate (source, target, relation) keys onto the
// maximum weight, returning edges sorted deterministically.
func mergeEdges(edges []Edge) []Edge {
	type edgeKey struct {
		source   string
		target   string
		relation Relation
	}

	maximum := make(map[edgeKey]Edge, len(edges))

	for _, edge := range edges {
		key := edgeKey{source: edge.Source, target: edge.Target, relation: edge.Relation}

		existing, ok := maximum[key]
		if !ok || edge.Weight > existing.Weight {
			maximum[key] = edge
		}
	}

	merged := make([]Edge, 0, len(maximum))

	for _, edge := range maximum {
		merged = append(merged, edge)
	}

	sortEdges(merged)

	return merged
}

// sortEdges orders edges by relation, then target, then source.
func sortEdges(edges []Edge) {
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Relation != edges[j].Relation {
			return edges[i].Relation < edges[j].Relation
		}

		if edges[i].Target != edges[j].Target {
			return edges[i].Target < edges[j].Target
		}

		return edges[i].Source < edges[j].Source
	})
}
