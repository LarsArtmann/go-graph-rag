package graphrag_test

import (
	"context"
	"fmt"
	"os"

	graphrag "github.com/larsartmann/go-graph-rag"
)

// exampleCorpus is a tiny domain-neutral knowledge base: three articles
// (two about Go, one about Rust) plus a hub tag with no text of its own.
func exampleCorpus() ([]graphrag.Document, []graphrag.Edge) {
	docs := []graphrag.Document{
		{
			ID:    "article:go",
			Kind:  "article",
			Label: "Go concurrency",
			Text:  "go concurrency channels goroutines",
		},
		{
			ID:    "article:go-patterns",
			Kind:  "article",
			Label: "Go patterns",
			Text:  "go concurrency channels patterns",
		},
		{
			ID:    "article:rust",
			Kind:  "article",
			Label: "Rust ownership",
			Text:  "rust ownership borrowing lifetimes",
		},
		{ID: "tag:go", Kind: "tag", Label: "Go"},
	}

	edges := []graphrag.Edge{
		{Source: "article:go", Target: "tag:go", Relation: "tagged", Weight: 1},
		{Source: "article:go-patterns", Target: "tag:go", Relation: "tagged", Weight: 1},
	}

	return docs, edges
}

// ExampleBuild embeds a small corpus with the offline hash provider and
// prints the resulting graph: caller-supplied edges plus the derived
// similar_to edge between near-duplicate articles.
func ExampleBuild() {
	docs, edges := exampleCorpus()

	result, err := graphrag.Build(
		context.Background(),
		graphrag.NewHashProvider(),
		nil, // no cache: pure in-memory build
		docs,
		edges,
		graphrag.BuildOptions{SimilarThreshold: 0.5},
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%d nodes, %d embedded\n", len(result.Nodes), result.Embedded)

	for _, edge := range result.Edges {
		fmt.Printf("%s -%s(%.2f)-> %s\n", edge.Source, edge.Relation, edge.Weight, edge.Target)
	}
	// Output:
	// 4 nodes, 3 embedded
	// article:go -similar_to(0.73)-> article:go-patterns
	// article:go -tagged(1.00)-> tag:go
	// article:go-patterns -tagged(1.00)-> tag:go
}

// ExampleSearcher_Search ranks articles by vector similarity and expands
// every hit with its graph neighborhood into an LLM-ready context block.
func ExampleSearcher_Search() {
	docs, edges := exampleCorpus()
	provider := graphrag.NewHashProvider()

	result, err := graphrag.Build(
		context.Background(),
		provider,
		nil,
		docs,
		edges,
		graphrag.BuildOptions{SimilarThreshold: 0.5},
	)
	if err != nil {
		panic(err)
	}

	searcher := graphrag.NewSearcher(result.Nodes, result.Edges, result.Vectors)

	res, err := searcher.Search(context.Background(), provider, "go channels", graphrag.SearchOptions{TopK: 2})
	if err != nil {
		panic(err)
	}

	fmt.Print(res.ContextText)
	// Output:
	// Retrieval context for query: "go channels"
	//
	// 1. [article] Go concurrency (similarity 0.54)
	//    related: [tag] Go (tagged 1.00)
	//    related: [article] Go patterns (similar_to 0.73)
	//
	// 2. [article] Go patterns (similarity 0.54)
	//    related: [tag] Go (tagged 1.00)
	//    related: [article] Go concurrency (similar_to 0.73)
}

// ExampleSearcher_SimilarPairs reports every embedded node pair whose
// cosine similarity clears a floor: the generic near-duplicate detector.
func ExampleSearcher_SimilarPairs() {
	docs, edges := exampleCorpus()

	result, err := graphrag.Build(
		context.Background(),
		graphrag.NewHashProvider(),
		nil,
		docs,
		edges,
		graphrag.BuildOptions{SimilarThreshold: 0.5},
	)
	if err != nil {
		panic(err)
	}

	searcher := graphrag.NewSearcher(result.Nodes, result.Edges, result.Vectors)

	for _, pair := range searcher.SimilarPairs(0.5) {
		fmt.Printf("%s ~ %s: %.2f\n", pair.A.ID, pair.B.ID, pair.Score)
	}
	// Output:
	// article:go ~ article:go-patterns: 0.73
}

// ExampleOpenStore persists a built graph and reads it back: the store
// doubles as the embedding cache during Build, ReplaceGraph swaps the
// derived index in one transaction, and LoadGraph/Stats read the snapshot.
func ExampleOpenStore() {
	store, err := graphrag.OpenStore(":memory:")
	if err != nil {
		panic(err)
	}
	defer func() { _ = store.Close() }()

	docs, edges := exampleCorpus()

	result, err := graphrag.Build(
		context.Background(),
		graphrag.NewHashProvider(),
		store, // the store is also the embedding cache
		docs,
		edges,
		graphrag.BuildOptions{SimilarThreshold: 0.5},
	)
	if err != nil {
		panic(err)
	}

	if err := store.ReplaceGraph(result.Nodes, result.EmbeddingHash, result.Edges); err != nil {
		panic(err)
	}

	snapshot, err := store.LoadGraph()
	if err != nil {
		panic(err)
	}

	stats, err := store.Stats()
	if err != nil {
		panic(err)
	}

	fmt.Printf(
		"%d nodes, %d edges, %d cached embeddings\n",
		len(snapshot.Nodes), len(snapshot.Edges), stats.Embeddings,
	)
	fmt.Println(snapshot.Nodes[0].ID, "->", snapshot.Nodes[0].Label)
	// Output:
	// 4 nodes, 3 edges, 3 cached embeddings
	// article:go -> Go concurrency
}

// ExampleNewSearcherWithOptions assigns search-time roles: one kind is the
// reference every hit is measured against (it never appears in hits), one
// kind lists the ranked documents, and one kind renders as a compact label
// list.
func ExampleNewSearcherWithOptions() {
	docs, edges := exampleCorpus()
	docs = append(docs, graphrag.Document{
		ID:    "brief:go",
		Kind:  "brief",
		Label: "Go brief",
		Text:  "go concurrency essentials idiomatic goroutine patterns",
	})
	edges = append(edges, graphrag.Edge{
		Source: "brief:go", Target: "tag:go", Relation: "briefs", Weight: 1,
	})

	provider := graphrag.NewHashProvider()

	result, err := graphrag.Build(
		context.Background(),
		provider,
		nil,
		docs,
		edges,
		graphrag.BuildOptions{SimilarThreshold: 0.5},
	)
	if err != nil {
		panic(err)
	}

	searcher := graphrag.NewSearcherWithOptions(
		result.Nodes,
		result.Edges,
		result.Vectors,
		graphrag.SearcherOptions{
			ReferenceKind: "brief",
			DocumentKinds: []graphrag.NodeKind{"article"},
			DetailKind:    "tag",
		},
	)

	res, err := searcher.Search(context.Background(), provider, "go channels", graphrag.SearchOptions{TopK: 2})
	if err != nil {
		panic(err)
	}

	fmt.Print(res.ContextText)
	// Output:
	// Retrieval context for query: "go channels"
	//
	// 1. [article] Go concurrency (similarity 0.54, ref match 0.37)
	//    tag: Go
	//    related: [article] Go patterns (similar_to 0.73)
	//
	// 2. [article] Go patterns (similarity 0.54, ref match 0.51)
	//    tag: Go
	//    related: [article] Go concurrency (similar_to 0.73)
}

// ExampleNewProvider selects an embedding provider by configuration: the
// offline default, any OpenAI-compatible endpoint, or a typed error for
// unknown names.
func ExampleNewProvider() {
	offline, err := graphrag.NewProvider(graphrag.EmbeddingConfig{})
	if err != nil {
		panic(err)
	}

	fmt.Println(offline.Name(), offline.Model())

	network, err := graphrag.NewProvider(graphrag.EmbeddingConfig{
		Provider: graphrag.ProviderOpenAICompat,
		BaseURL:  "https://api.openai.com/v1",
		APIKey:   os.Getenv("OPENAI_API_KEY"),
		Model:    "text-embedding-3-small",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(network.Name(), network.Model())

	_, err = graphrag.NewProvider(graphrag.EmbeddingConfig{Provider: "nobody"})
	fmt.Println(err)
	// Output:
	// hash v1
	// openai-compat text-embedding-3-small
	// graphrag: unknown embedding provider: "nobody" (want "hash" or "openai-compat")
}
