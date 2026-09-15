# go-graph-rag

Semantic retrieval primitives in Go: embedding providers, a typed knowledge
graph, a SQLite-backed vector + graph store, and hybrid retrieval that blends
vector similarity with graph expansion (GraphRAG).

```go
import graphrag "github.com/larsartmann/go-graph-rag"
```

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

## Design constraints (deliberate)

- **Dependency-light**: stdlib-only except `samber/lo` and `modernc.org/sqlite`.
  No `encoding/json/v2` (builds without `GOEXPERIMENT=jsonv2`).
- **Full rebuild, not incremental**: `Reindex` replaces the whole graph in one
  transaction; the embedding cache makes that cheap. Revisit around ~50k nodes.
- **One writer per store file**: take an ownership lease in the caller if more
  than one process might open the same DSN.
- **Two-tier ranking is by design**: assert hub _reachability_, never
  `hits[0]` primacy, for hub labels.

## Status

v0.x: the API is stable enough to consume but may still change before v1
(ANN/HNSW backend and incremental indexing are open design decisions).

## License

PROPRIETARY. Copyright (c) 2026 Lars Artmann. All rights reserved.
