package graphrag

import (
	"context"
	"fmt"
	"sort"

	"github.com/samber/lo"
)

// Index-building defaults.
const (
	// DefaultSimilarTopK is how many nearest neighbors each embedded node
	// links to via RelationSimilar edges.
	DefaultSimilarTopK = 3

	// DefaultSimilarThreshold is the minimum cosine similarity for a
	// RelationSimilar edge. The default suits the offline hash provider,
	// whose raw-count cosines run lower than neural embeddings; with a
	// real embedding model, calibrate upward (0.85+) and set
	// BuildOptions.SimilarThreshold explicitly.
	DefaultSimilarThreshold = 0.75
)

// ErrDimsInconsistent is returned when a provider mixes vector
// dimensionalities within one build.
type ErrDimsInconsistent struct { //nolint:errname // names the failure precisely
	First, Second int
}

func (e *ErrDimsInconsistent) Error() string {
	return fmt.Sprintf("graphrag: inconsistent vector dimensions: %d then %d", e.First, e.Second)
}

// Document is the indexing unit: one graph node plus the text to embed for
// it. Documents with an empty Text produce a node without a vector, so the
// graph can still traverse through them (vector-less hubs stay reachable
// via graph expansion only).
type Document struct {
	ID    string
	Kind  NodeKind
	Label string
	Text  string
	Attrs NodeAttrs
}

// BuildOptions tunes index construction.
type BuildOptions struct {
	// SimilarTopK is the per-node nearest-neighbor count for
	// RelationSimilar edges (0 selects DefaultSimilarTopK).
	SimilarTopK int

	// SimilarThreshold is the minimum cosine for a RelationSimilar edge
	// (0 selects DefaultSimilarThreshold).
	SimilarThreshold float64
}

// BuildResult reports what a Build produced.
type BuildResult struct {
	// Nodes and Edges form the persisted graph (caller edges plus the
	// computed RelationSimilar edges), deterministically ordered.
	Nodes []Node
	Edges []Edge

	// Vectors maps node ID to its embedding, for in-memory search warm-up.
	Vectors map[string]Vector

	// EmbeddingHash maps node ID to its content hash, for cache joins.
	EmbeddingHash map[string]string

	// Provider and Model identify the vector space of Vectors.
	Provider string
	Model    string

	// Embedded counts provider calls made; CacheHits counts reused vectors.
	Embedded  int
	CacheHits int
}

// Build embeds the embeddable documents (through the cache), assembles the
// graph, and derives RelationSimilar edges between all embedded nodes whose
// cosine similarity clears the threshold. It is a pure function of its
// inputs: same documents, same provider, same edges out.
func Build(
	ctx context.Context,
	provider Provider,
	cache Cache,
	docs []Document,
	edges []Edge,
	opts BuildOptions,
) (*BuildResult, error) {
	topK := max(opts.SimilarTopK, 0)
	if topK == 0 {
		topK = DefaultSimilarTopK
	}

	threshold := opts.SimilarThreshold
	if threshold <= 0 {
		threshold = DefaultSimilarThreshold
	}

	vectors, hashes, embedded, hits, err := resolveVectors(ctx, provider, cache, docs)
	if err != nil {
		return nil, err
	}

	// Duplicate document IDs (e.g. the same hub reached from two different
	// documents) collapse onto their FIRST occurrence so insertion order
	// stays deterministic and persistence never hits a unique constraint.
	seen := make(map[string]struct{}, len(docs))
	nodes := make([]Node, 0, len(docs))

	for _, doc := range docs {
		if _, dup := seen[doc.ID]; dup {
			continue
		}

		seen[doc.ID] = struct{}{}
		nodes = append(nodes, Node{ID: doc.ID, Kind: doc.Kind, Label: doc.Label, Attrs: doc.Attrs})
	}

	allEdges := append(sortEdgesClone(edges), similarEdges(nodes, vectors, topK, threshold)...)

	return &BuildResult{
		Nodes:         nodes,
		Edges:         mergeEdges(allEdges),
		Vectors:       vectors,
		EmbeddingHash: hashes,
		Provider:      provider.Name(),
		Model:         provider.Model(),
		Embedded:      embedded,
		CacheHits:     hits,
	}, nil
}

// resolveVectors fills the vector table from cache and provider, returning
// vectors by node ID and content hashes by node ID.
func resolveVectors(
	ctx context.Context,
	provider Provider,
	cache Cache,
	docs []Document,
) (map[string]Vector, map[string]string, int, int, error) {
	vectors := make(map[string]Vector, len(docs))
	hashes := make(map[string]string, len(docs))
	cacheHits := 0

	misses := make([]Document, 0)
	missHashes := make(map[string]string)

	for _, doc := range docs {
		if doc.Text == "" {
			continue
		}

		hash := HashText(doc.Text)
		hashes[doc.ID] = hash

		if cached, ok := lookupCache(cache, provider, hash); ok {
			vectors[doc.ID] = cached
			cacheHits++

			continue
		}

		if _, seen := missHashes[hash]; !seen {
			misses = append(misses, doc)
			missHashes[hash] = doc.ID
		}
	}

	fresh, err := embedAll(ctx, provider, misses)
	if err != nil {
		return nil, nil, 0, 0, err
	}

	entries := make([]EmbeddingEntry, 0, len(fresh))

	for hash, vector := range fresh {
		entries = append(entries, EmbeddingEntry{Hash: hash, Vector: vector})
	}

	if cache != nil && len(entries) > 0 {
		if putErr := cache.PutEmbeddings(provider.Name(), provider.Model(), entries); putErr != nil {
			return nil, nil, 0, 0, fmt.Errorf("graphrag: persist embeddings: %w", putErr)
		}
	}

	for _, doc := range misses {
		if vector, ok := fresh[HashText(doc.Text)]; ok {
			vectors[doc.ID] = vector
		}
	}

	return vectors, hashes, len(fresh), cacheHits, nil
}

// lookupCache guards the nil-cache case (pure in-memory builds).
func lookupCache(cache Cache, provider Provider, hash string) (Vector, bool) {
	if cache == nil {
		return nil, false
	}

	cached, ok, err := cache.CachedEmbedding(provider.Name(), provider.Model(), hash)
	if err != nil || !ok {
		return nil, false
	}

	return cached, true
}

// embedAll embeds the missed documents in one provider pass, keyed by
// content hash (deduplicated) and dimension-checked.
func embedAll(ctx context.Context, provider Provider, misses []Document) (map[string]Vector, error) {
	if len(misses) == 0 {
		return map[string]Vector{}, nil
	}

	texts := lo.Map(misses, func(doc Document, _ int) string { return doc.Text })

	embedded, err := provider.Embed(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("graphrag: embed %d documents: %w", len(texts), err)
	}

	if len(embedded) != len(texts) {
		return nil, fmt.Errorf("%w: sent %d texts, got %d vectors", ErrEmbedCountMismatch, len(texts), len(embedded))
	}

	dims := 0
	fresh := make(map[string]Vector, len(embedded))

	for i, doc := range misses {
		vector := embedded[i]

		if dims == 0 {
			dims = vector.Dims()
		} else if dims != vector.Dims() {
			return nil, &ErrDimsInconsistent{First: dims, Second: vector.Dims()}
		}

		fresh[HashText(doc.Text)] = vector
	}

	return fresh, nil
}

// similarEdges computes RelationSimilar edges between embedded nodes of the
// SAME kind. For each node the topK nearest neighbors above the threshold
// get an edge; the edge is stored once per unordered pair with the pair's
// cosine as weight. The same-kind rule keeps the semantic honest: hub nodes
// are embedded too, but "similar" means near-duplicate documents or sibling
// hubs of one kind — never a document that merely mentions a hub (that is
// what caller-supplied edges express).
func similarEdges(nodes []Node, vectors map[string]Vector, topK int, threshold float64) []Edge {
	embedded := make([]Node, 0, len(vectors))

	for _, node := range nodes {
		if _, ok := vectors[node.ID]; ok {
			embedded = append(embedded, node)
		}
	}

	neighbors := make(map[string][]Edge)

	for i := range embedded {
		for j := i + 1; j < len(embedded); j++ {
			left, right := embedded[i], embedded[j]
			if left.Kind != right.Kind {
				continue
			}

			score := Cosine(vectors[left.ID], vectors[right.ID])

			if score < threshold {
				continue
			}

			neighbors[left.ID] = appendRanked(
				neighbors[left.ID],
				Edge{Source: left.ID, Target: right.ID, Relation: RelationSimilar, Weight: score},
				topK,
			)
			neighbors[right.ID] = appendRanked(
				neighbors[right.ID],
				Edge{Source: right.ID, Target: left.ID, Relation: RelationSimilar, Weight: score},
				topK,
			)
		}
	}

	edges := make([]Edge, 0)
	seen := make(map[string]struct{})

	for _, node := range embedded {
		for _, edge := range neighbors[node.ID] {
			pair := pairKey(edge.Source, edge.Target)
			if _, dup := seen[pair]; dup {
				continue
			}

			seen[pair] = struct{}{}

			edges = append(edges, edge)
		}
	}

	sortEdges(edges)

	return edges
}

// appendRanked keeps the topK heaviest similar-edges per node.
func appendRanked(current []Edge, candidate Edge, topK int) []Edge {
	current = append(current, candidate)

	if len(current) <= topK {
		return current
	}

	sort.Slice(current, func(i, j int) bool {
		if current[i].Weight != current[j].Weight {
			return current[i].Weight > current[j].Weight
		}

		return current[i].Target < current[j].Target
	})

	return current[:topK]
}

// pairKey orders an unordered pair so each similar-edge is stored once.
func pairKey(a, b string) string {
	if a < b {
		return a + "\x00" + b
	}

	return b + "\x00" + a
}

// sortEdgesClone returns a deterministically sorted copy of edges.
func sortEdgesClone(edges []Edge) []Edge {
	cloned := make([]Edge, 0, len(edges))
	cloned = append(cloned, edges...)
	sortEdges(cloned)

	return cloned
}
