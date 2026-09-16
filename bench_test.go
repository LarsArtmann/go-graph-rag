package graphrag_test

import (
	"fmt"
	"path/filepath"
	"testing"

	graphrag "github.com/larsartmann/go-graph-rag"
)

// Benchmarks measure the offline end-to-end paths with the hash provider
// (1024-dim vectors) over synthetic corpora of unique texts with overlapping
// vocabularies. Reference numbers, 2026-09-16, AMD Ryzen AI MAX+ 395
// (x86_64, single goroutine does the work), Go 1.26.7,
// `go test -bench . -count 10` (StoreRoundTrip: -count 3), summarized with
// benchstat (golang.org/x/perf v0.0.0-20260908200009) — treat as orders of
// magnitude, not absolutes; ± figures are the benchstat spread:
//
//	Build/docs=100             3.961 ms/op  ± 2%  (4,950 pairwise cosines)
//	Build/docs=1000            369.5 ms/op  ± 2%  (499,500 pairwise cosines)
//	Search/docs=100            101.1 µs/op  ± 1%
//	Search/docs=1000           948.6 µs/op  ± 1%
//	SimilarPairs/docs=1000     391.1 ms/op  ± 1%
//	StoreRoundTrip/docs=1000   21.6 ms/op         (ReplaceGraph + LoadGraph
//	                               + LoadEmbeddings on a real SQLite file)
//	StoreRoundTrip/docs=10000  197.4 ms/op
//
// The Build/SimilarPairs scaling is quadratic in embedded-node count (10x
// docs -> ~90x time): at ~50k nodes a full rebuild extrapolates to ~15
// minutes. StoreRoundTrip shows persistence is NOT the bottleneck (10k
// nodes round-trip in ~0.2s against a ~37s rebuild) — the rebuild cost is
// pairwise cosine compute. That is the measured trigger behind ROADMAP's
// "revisit around ~50k nodes" item (ANN/HNSW backend, incremental
// indexing).

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

// BenchmarkStoreRoundTrip measures the persistence loop on a real SQLite
// file: ReplaceGraph (full derived-index swap) plus LoadGraph and
// LoadEmbeddings (the snapshot a Searcher warm-starts from). The build
// happens once outside the loop; the loop is pure store I/O over BLOB
// vectors. Reference numbers live in the file-top doc comment.
func BenchmarkStoreRoundTrip(b *testing.B) {
	for _, size := range []int{1000, 10000} {
		b.Run(fmt.Sprintf("docs=%d", size), func(b *testing.B) {
			docs, edges := benchCorpus(size)
			provider := graphrag.NewHashProvider()

			store, err := graphrag.OpenStore(filepath.Join(b.TempDir(), "bench.db"))
			if err != nil {
				b.Fatal(err)
			}
			defer func() { _ = store.Close() }()

			// The store is the cache, as in the real pipeline: the setup
			// build populates the embedding table that LoadEmbeddings reads.
			result, err := graphrag.Build(b.Context(), provider, store, docs, edges, graphrag.BuildOptions{})
			if err != nil {
				b.Fatal(err)
			}

			for b.Loop() {
				if err := store.ReplaceGraph(result.Nodes, result.EmbeddingHash, result.Edges); err != nil {
					b.Fatal(err)
				}

				snapshot, err := store.LoadGraph()
				if err != nil {
					b.Fatal(err)
				}

				vectors, err := store.LoadEmbeddings(provider.Name(), provider.Model())
				if err != nil {
					b.Fatal(err)
				}

				if len(snapshot.Nodes) != size || len(vectors) != size {
					b.Fatalf("expected %d nodes and vectors, got %d/%d", size, len(snapshot.Nodes), len(vectors))
				}
			}
		})
	}
}
