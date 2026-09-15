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
