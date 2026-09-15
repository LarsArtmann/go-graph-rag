package graphrag_test

import (
	"fmt"
	"testing"

	graphrag "github.com/larsartmann/go-graph-rag"
)

// Benchmarks measure the offline end-to-end paths with the hash provider
// (1024-dim vectors) over synthetic corpora of unique texts with overlapping
// vocabularies. Reference numbers, 2026-09-15, x86_64 (32 threads, single
// goroutine does the work), Go 1.26.7, -benchtime=100ms — treat as orders of
// magnitude, not absolutes:
//
//	Build/docs=100       ~4.2 ms/op    (4,950 pairwise cosines)
//	Build/docs=1000      ~373 ms/op    (499,500 pairwise cosines)
//	Search/docs=100      ~111 µs/op
//	Search/docs=1000     ~927 µs/op
//	SimilarPairs/docs=1000 ~408 ms/op
//
// The Build/SimilarPairs scaling is quadratic in embedded-node count (10x
// docs -> ~90x time): at ~50k nodes a full rebuild extrapolates to ~15
// minutes. That is the measured trigger behind ROADMAP's "revisit around
// ~50k nodes" item (ANN/HNSW backend, incremental indexing).

// benchCorpus generates n deterministic documents over a fixed vocabulary:
// overlapping token sets exercise similar-edge derivation without any
// randomness, so runs are comparable across machines and time.
func benchCorpus(n int) ([]graphrag.Document, []graphrag.Edge) {
	vocab := []string{
		"go", "concurrency", "channels", "patterns", "ownership",
		"search", "vector", "graph", "sqlite", "embedding",
		"retrieval", "index", "hybrid", "rank", "context",
	}

	docs := make([]graphrag.Document, 0, n)
	for i := range n {
		// The leading note-number keeps every text distinct (content-hash
		// dedup would otherwise collapse the corpus to ~15 vectors) while
		// the vocabulary strides keep overlapping token sets.
		text := fmt.Sprintf(
			"note%d %s %s %s %s",
			i,
			vocab[i%len(vocab)],
			vocab[(i*3+1)%len(vocab)],
			vocab[(i*7+2)%len(vocab)],
			vocab[(i*11+5)%len(vocab)],
		)
		docs = append(docs, graphrag.Document{
			ID:    fmt.Sprintf("doc:%06d", i),
			Kind:  "article",
			Label: fmt.Sprintf("Doc %d", i),
			Text:  text,
		})
	}

	edges := make([]graphrag.Edge, 0, n/10)
	for i := range n {
		if i%10 != 0 || i+1 >= n {
			continue
		}

		edges = append(edges, graphrag.Edge{
			Source:   fmt.Sprintf("doc:%06d", i),
			Target:   fmt.Sprintf("doc:%06d", i+1),
			Relation: "related",
			Weight:   1,
		})
	}

	return docs, edges
}

func BenchmarkBuild(b *testing.B) {
	for _, size := range []int{100, 1000} {
		b.Run(fmt.Sprintf("docs=%d", size), func(b *testing.B) {
			docs, edges := benchCorpus(size)
			provider := graphrag.NewHashProvider()

			for b.Loop() {
				result, err := graphrag.Build(b.Context(), provider, nil, docs, edges, graphrag.BuildOptions{})
				if err != nil {
					b.Fatal(err)
				}

				if len(result.Nodes) != size {
					b.Fatalf("expected %d nodes, got %d", size, len(result.Nodes))
				}
			}
		})
	}
}

func BenchmarkSearch(b *testing.B) {
	for _, size := range []int{100, 1000} {
		b.Run(fmt.Sprintf("docs=%d", size), func(b *testing.B) {
			docs, edges := benchCorpus(size)
			provider := graphrag.NewHashProvider()

			result, err := graphrag.Build(b.Context(), provider, nil, docs, edges, graphrag.BuildOptions{})
			if err != nil {
				b.Fatal(err)
			}

			searcher := graphrag.NewSearcher(result.Nodes, result.Edges, result.Vectors)

			for b.Loop() {
				res, err := searcher.Search(b.Context(), provider, "go concurrency channels", graphrag.SearchOptions{})
				if err != nil {
					b.Fatal(err)
				}

				if len(res.Hits) == 0 {
					b.Fatal("expected hits")
				}
			}
		})
	}
}

func BenchmarkSimilarPairs(b *testing.B) {
	docs, edges := benchCorpus(1000)

	result, err := graphrag.Build(b.Context(), graphrag.NewHashProvider(), nil, docs, edges, graphrag.BuildOptions{})
	if err != nil {
		b.Fatal(err)
	}

	searcher := graphrag.NewSearcher(result.Nodes, result.Edges, result.Vectors)

	for b.Loop() {
		if pairs := searcher.SimilarPairs(0.5); len(pairs) == 0 {
			b.Fatal("expected similar pairs")
		}
	}
}
