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
go get github.com/larsartmann/go-graph-rag@v0.3.2
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
same SQLite file as the graph, so full rebuilds are cheap. Cold builds over a
network provider embed concurrently — set `EmbeddingConfig.EmbedConcurrency`
(default 1) to raise throughput when the endpoint allows it.

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

## BDD test kit

Consumers who test in Ginkgo/Gomega style can use the optional, composable test
kit — its own module, so the SDK's dependency budget never grows:

```bash
go get github.com/larsartmann/go-graph-rag/bdd   # first bdd/v* tag pending (owner-gated)
```

```go
import graphragbdd "github.com/larsartmann/go-graph-rag/bdd"
```

The kit provides exactly what consumer test suites hand-roll today: `Doc`/`Edge`
builders, a deterministic demo corpus with a matching `SearcherOptions` policy,
a `Harness` over the real machinery (provider, optional SQLite store, `Build`,
`Search`, `Reload` from disk), Gomega matchers (`HaveTopHit`, `ContainHit`,
`HaveNoHitsOfKind`, ...), and a `FailingProvider` for error paths:

```go
docs, edges := graphragbdd.DemoCorpus()
harness := graphragbdd.NewHarness(GinkgoT(), docs, edges,
    graphragbdd.WithSearcherOptions(graphragbdd.DemoSearcherOptions()))

_, err := harness.Build(ctx)
Expect(err).NotTo(HaveOccurred())

result, err := harness.Search(ctx, "hybrid search combines vector similarity")
Expect(err).NotTo(HaveOccurred())
Expect(result).To(graphragbdd.HaveTopHit("article:retrieval"))
```

Design and scope: [`docs/planning/2026-10-15_05-02_bdd-submodule-adr.md`](docs/planning/2026-10-15_05-02_bdd-submodule-adr.md).
The kit's own 47-spec suite doubles as a black-box acceptance suite over the SDK.

## Design constraints (deliberate)

- **Dependency-light**: stdlib-only except `samber/lo` and `modernc.org/sqlite`.
  `encoding/json` (v1) only, as a stability policy — since Go 1.27.x a v2
  import compiles with or without `GOEXPERIMENT`, so this is a rule the team
  keeps, not a build the compiler enforces.
- **Strong IDs**: node identity is the defined type `NodeID`; untyped string
  literals still assign directly, and the JSON wire shapes are unchanged.
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
