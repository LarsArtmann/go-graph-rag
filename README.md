# go-graph-rag

[![go-test](https://github.com/LarsArtmann/go-graph-rag/actions/workflows/go-test.yml/badge.svg)](https://github.com/LarsArtmann/go-graph-rag/actions/workflows/go-test.yml)

Semantic retrieval primitives in Go: embedding providers, a typed knowledge
graph, a SQLite-backed vector + graph store, and hybrid retrieval that blends
vector similarity with graph expansion (GraphRAG).

API reference: [pkg.go.dev/github.com/larsartmann/go-graph-rag](https://pkg.go.dev/github.com/larsartmann/go-graph-rag)

> pkg.go.dev currently hides the rendered godoc ("Documentation not displayed
> due to license restrictions" — it does not recognize the PROPRIETARY
> LICENSE file). The documentation of record lives in this repo; runnable,
> output-verified examples are in [`example_test.go`](example_test.go).

## Installation

```bash
go get github.com/larsartmann/go-graph-rag@v0.2.0
```

```go
import graphrag "github.com/larsartmann/go-graph-rag"
```

## Quick start

```go
provider := graphrag.NewHashProvider() // offline default; see providers below

docs := []graphrag.Document{
	{ID: "article:go", Kind: "article", Label: "Go concurrency", Text: "go concurrency channels goroutines"},
	{ID: "article:rust", Kind: "article", Label: "Rust ownership", Text: "rust ownership borrowing lifetimes"},
	{ID: "tag:go", Kind: "tag", Label: "Go"}, // vector-less hub: reachable via graph expansion
}
edges := []graphrag.Edge{
	{Source: "article:go", Target: "tag:go", Relation: "tagged", Weight: 1},
}

result, err := graphrag.Build(ctx, provider, nil, docs, edges, graphrag.BuildOptions{})
if err != nil {
	return err
}

searcher := graphrag.NewSearcher(result.Nodes, result.Edges, result.Vectors)

res, err := searcher.Search(ctx, provider, "go channels",
	graphrag.SearchOptions{MinScore: 0.25}) // drop weak hits; 0 keeps everything
if err != nil {
	return err
}

fmt.Print(res.ContextText) // ranked hits with graph neighborhoods, LLM-ready
```

Runnable, output-verified versions of this flow live in
[`example_test.go`](example_test.go) (also rendered on pkg.go.dev).

## What it does

1. **Index** (`Build`): take `Document{ID, Kind, Label, Text, Attrs}` values
   plus caller-supplied edges, embed every document, derive `RelationSimilar`
   edges between near-duplicate vectors, and persist nodes, edges, and
   embedding cache in one SQLite database.
2. **Search** (`Searcher.Search`): embed the query, rank documents by cosine
   similarity, expand every hit with its graph neighborhood, and render a
   deterministic, LLM-ready context block.
3. **Compare** (`SimilarPairs`): generic near-duplicate detector over every
   embedded node pair.

## Embedding providers

| Provider                  | Type                   | Notes                                                                                                                                                                   |
| ------------------------- | ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `NewHashProvider`         | offline, deterministic | 1024-dim hashing trick; no network, byte-stable across runs; cosine scores run lower than neural embeddings, so calibrate `SimilarThreshold` accordingly (default 0.75) |
| `NewOpenAICompatProvider` | network                | any OpenAI-style `/embeddings` endpoint (base URL, model, API key, timeout, retries)                                                                                    |

Embeddings are cached namespaced by `content hash x provider x model` in the
same SQLite file as the graph, so full rebuilds are cheap.

## Search-time roles

The SDK is domain-neutral: node kinds and edge relations are caller-defined
constants over the `NodeKind` / `Relation` types. The `SearcherOptions`
policy assigns roles at search time:

- `ReferenceKind`: a reference node (e.g. the profile document in a matching
  graph) never appears in hits; every hit carries `RefMatch`, its cosine
  similarity to the reference.
- `DocumentKinds`: kinds ranked as primary documents; all other embedded
  nodes act as expansion hubs that fill the remaining top-k slots (two-tier
  ranking keeps short hub labels from crowding out real documents).
- `DetailKind`: related nodes of this kind render as a compact label list in
  `RenderContext` output.

See `ExampleNewSearcherWithOptions` in the godoc examples for a runnable
walkthrough of all three roles.

## Design constraints (deliberate)

- **Dependency-light**: stdlib-only except `samber/lo` and `modernc.org/sqlite`.
  No `encoding/json/v2` (builds without `GOEXPERIMENT=jsonv2`).
- **Full rebuild, not incremental**: `ReplaceGraph` swaps the whole persisted
  graph in one transaction; the embedding cache makes that cheap. Revisit
  around ~50k nodes.
- **One writer per store file**: take an ownership lease in the caller if more
  than one process might open the same DSN.
- **Two-tier ranking is by design**: assert hub _reachability_, never
  `hits[0]` primacy, for hub labels.

## Status

v0.x: the API is stable enough to consume but may still change before v1.
The scale path is decided, not open:
[`docs/planning/2026-09-16_13-25_seam-store-search-adr.md`](docs/planning/2026-09-16_13-25_seam-store-search-adr.md)
fixes the `VectorIndex`/`GraphStore` seams and the adapter-module rules;
ANN/HNSW and incremental indexing are implementation work, trigger-gated
(ADR §9), not design questions.

## Security

Found a vulnerability? Please report it privately — see
[SECURITY.md](SECURITY.md). Never open a public issue for security problems.

## License

PROPRIETARY. Copyright (c) 2026 Lars Artmann. All rights reserved.
