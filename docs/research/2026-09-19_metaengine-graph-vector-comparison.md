# Metaengine Graph/Vector Capabilities vs go-graph-rag: Comparison

|          |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Date     | 2026-09-19                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| Question | How do `go-graph-rag`'s capabilities relate to the graph and vector capabilities of `go-cqrs-lite/metaengine` — overlap, gap, or duplication?                                                                                                                                                                                                                                                                                                                                                                                                             |
| Method   | Both module trees read file-by-file on 2026-09-19: `metaengine/` core (`vector_search.go`, `dispatch.go`, `graph_fallback.go`, `graphadapter/`, per-engine vector/graph files, `dgraphengine/`) and this SDK (`graph.go`, `search.go`, `store.go`, `build.go`, `embed*.go`). Every claim carries a file:line cite verified that day. metaengine pinned at `v4` (`metaengine/go.mod`, go 1.27.1); this SDK at v0.2 (`embed_openai.go:47`).                                                                                                                 |
| Verdict  | **Complementary layers, not competitors.** metaengine ships graph + vector as storage ADTs (k-NN and traversal primitives across 10 engine drivers, event-fed index building, no embeddings, no retrieval policy). go-graph-rag productizes the GraphRAG vertical on top of that layer's job: real embedding providers + cache, typed graph semantics, hybrid ranking policy, LLM-ready context rendering. Exactly one capability overlaps (the hand-rolled pipeline in `dgraphengine/graphrag_test.go`), and this SDK supersedes it. No adoption change. |

## 1. TL;DR

The two share a topic, not a job. metaengine answers "where do graph edges and vectors
_live_" — as foldable ADTs inside a cost-planned, engine-swappable CQRS read-model
store. This SDK answers "how do I get from documents to an answerable GraphRAG index"
— embedding, building, persisting, searching, and rendering context. Reading them as
alternatives is a category error; reading them as layers is accurate.

The single true overlap: `metaengine/dgraphengine/graphrag_test.go` demonstrates a
GraphRAG pipeline by hand (`SearchInsert` → `GraphAddEdge` → `SearchQuery` →
`GraphNeighbors` → assemble) — with full-text search standing in for embeddings. This
SDK is that sketch industrialized: real embeddings instead of text scoring, typed
graph semantics instead of bare `{From, To}`, policy-driven hybrid ranking instead of
glue code.

Consistent with the 2026-09-15 adoption research (adopt neither metaengine nor system
into the SDK) and the seam ADR (adapters in separate modules, core CGO-free).

## 2. Capability matrix

| Dimension          | go-graph-rag (v0.2)                                                                                                                           | metaengine graph                                                               | metaengine vector                                                                                      |
| ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------ |
| Core abstraction   | GraphRAG vertical: `Build` → `Store` → `Searcher` → `ContextText`                                                                             | `graphBackend` ADT + multimap fallback                                         | `VectorBackend` / `VectorFilterBackend` / `VectorCounter` ADTs                                         |
| Graph model        | `Node{Kind,Title,Attrs}`, `Edge{Relation,Weight,Direction}`, in/out indexes (`graph.go`)                                                      | `Edge{From any, To any}` — no relations, weights, node records (`types.go:60`) | —                                                                                                      |
| Traversal          | Expansion only: neighbors of vector hits, capped per hit                                                                                      | Depth-N BFS, undirected variant, edge tombstones (ADR-0114)                    | —                                                                                                      |
| Vector index       | Brute-force in-memory cosine over a loaded snapshot (`vector.go`)                                                                             | —                                                                              | k-NN on every engine, metric param (cosine/euclidean/dot), pre-ranked metadata-filtered k-NN, counters |
| ANN                | None (roadmap: sqlite-vec/hannoy behind the seam)                                                                                             | dgraph native `@reverse` for graph                                             | **None shipped** — brute force everywhere; dgraph native ANN and pgvector are documented future paths  |
| Embeddings         | Yes: provider seam, offline hash fallback, OpenAI client, content-hash × provider × model SQLite cache                                        | No — caller supplies `float32`                                                 | No — caller supplies `float32` (`Embedding{ID, Values, Metadata}`)                                     |
| Retrieval policy   | `SearcherOptions` (ReferenceKind/DocumentKinds/DetailKind), RefMatch, TopK, SimilarPairs dedup, deterministic context rendering (`search.go`) | none                                                                           | none — raw top-k IDs + distances                                                                       |
| Index construction | Caller corpus: `Build(documents, edges)` derives `RelationSimilar` edges                                                                      | Fold targets: graph edges written by event folds                               | Fold inputs: `Embedding`/`IndexedText` events build indexes (event-sourced)                            |
| Persistence        | Own SQLite (CGO-free, `modernc.org/sqlite`), replace/load snapshot + embedding cache                                                          | Per-engine: memory, sqlite, pg, pebble, badger, duckdb, iroh, dgraph           | Same engine fleet (`VectorBackend` is an engine capability)                                            |
| Scale guidance     | Well under ~50k nodes (owner decision 2026-09-15)                                                                                             | O(degree^depth) native on dgraph; degraded multimap fallback elsewhere         | Brute-force O(N·D); guidance <10k vectors (2026-09-15 research, `vector_search.go:152`)                |

## 3. Graph capabilities, precisely

| Aspect                   | go-graph-rag                                                                                                                                | metaengine                                                                                                                                                                     |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Edge payload             | `Edge{Source, Target, Relation, Weight}`; relations/weights feed hybrid ranking                                                             | `Edge{From any, To any}` (`types.go:60`) — deliberately label-free at the storage boundary; typed folding lives one layer up                                                   |
| Dispatch contract        | none needed (in-memory `Graph`)                                                                                                             | Unexported `graphBackend` per ADR-0113; consumers use `graphadapter.Adapter` (wraps a `graph.MemoryDriver`) — `dispatch.go:9-13`                                               |
| Optional contracts       | —                                                                                                                                           | `graphEdgeRemover` (tombstone deletion), `undirectedGraphBackend` (bidirectional traversal) — `dispatch.go`                                                                    |
| Degradation              | n/a                                                                                                                                         | Engines without native graph degrade to MultimapBackend: edges as multimap entries, BFS via iterative `MultiGet`, O(N·degree^depth), `DiagLevelDegraded` — `graph_fallback.go` |
| Native traversal         | `Graph.OutEdges`/`InEdges` over a merged in-memory index; expansion in `Searcher` is 1-hop per hit, capped by `DefaultMaxRelatedPerHit = 6` | Depth-N neighbors; dgraph native via `@reverse` at O(degree^depth), no degradation (`dgraphengine/README.md`)                                                                  |
| Semantics this SDK needs | Relation-typed, weighted, directed edges + derived `RelationSimilar` edges from vector similarity (`build.go`)                              | Would have to be side-stored (Maps) — the 2026-09-15 adoption doc's central objection, unchanged                                                                               |

## 4. Vector capabilities, precisely

| Aspect                | go-graph-rag                                                                                                                | metaengine                                                                                                                                                                                                                                                                                                                               |
| --------------------- | --------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Input type            | SDK embeds node content itself (`EmbeddingConfig` → `NewProvider`); vectors are `Vector` keyed by node ID                   | `Embedding{ID, Values []float32, Metadata map[string]any}` (`vector_search.go:23`) — caller-embedded, optional metadata                                                                                                                                                                                                                  |
| Query surface         | `Searcher`: hybrid — vector top-k (`DefaultTopK = 5`), RefMatch similarity, graph expansion, dedup, `ContextText` rendering | `VectorSearch(ctx, col, query, k, metric)` → `[]VectorResult{ID, Distance}` (`vector_search.go:62`); metrics cosine/euclidean/dot via shared `VectorDistance` (`vector_search.go:147`)                                                                                                                                                   |
| Filtered search       | Policy-level (kind roles via `SearcherOptions`), not metadata filters                                                       | `VectorFilterBackend.VectorSearchFiltered`: filters applied BEFORE ranking so k results are the k nearest matching (`vector_search.go:84`); AND semantics shared via `VectorMatchesFilters`                                                                                                                                              |
| Introspection         | `Store.Stats`                                                                                                               | `VectorCounter`: per-collection counts + collection enumeration without payload transfer (`vector_search.go:105`)                                                                                                                                                                                                                        |
| Implementation status | One brute-force cosine scan, in-memory, ~120 lines (`vector.go`)                                                            | Every shipped engine brute-force today: memory (`memory_engine.go:304`), sqlite (`sqliteengine/vector.go:78`), pg (`pgengine/vector.go:27`, comment points to pgvector for production scale), badger (`badgerengine/vector.go:27`), dgraph streams nodes over gRPC + Go-side scoring, declared degraded (`dgraphengine/vector.go:14-20`) |
| Encoding helpers      | SQLite BLOBs via store schema                                                                                               | `EncodeVectorBinary`/`DecodeVectorAuto` F32 framing (`vector_binary.go:37`)                                                                                                                                                                                                                                                              |
| What it does NOT have | Filtered k-NN, cross-engine backends, event-fed index building                                                              | Hybrid ranking, reference-node matching, graph expansion, dedup scan, context assembly, embedding cache — the entire retrieval half                                                                                                                                                                                                      |

## 5. The overlap point

`metaengine/dgraphengine/graphrag_test.go` (`TestGraphRAG_SearchThenGraphTraverse`)
builds the canonical pipeline on the only engine with native graph + search parity:

1. INDEX: `SearchInsert` entity descriptions + `GraphAddEdge` relationships
2. RETRIEVE: `SearchQuery` by text
3. EXPAND: `GraphNeighbors` around each hit
4. ASSEMBLE: dedupe into a context window

go-graph-rag supersedes this sketch at every step: embeddings replace text scoring,
`Relation`/`Weight`-typed edges replace bare pairs, `SearcherOptions` policy replaces
ad-hoc glue, and the deterministic `ContextText` renderer replaces manual assembly.
The sketch is evidence the layer split works, not competition.

## 6. Relationship to existing decisions

This document introduces no new adoption decision; it is the capability snapshot the
other three documents implicitly assume:

- `2026-09-15_metaengine-system-adoption.md` — verdict "adopt neither into the SDK".
  Every argument there (event-fed write path mismatch, brute-force vectors, label-free
  edges, experimental v4→v5 churn) is re-confirmed by the file:line evidence in §3-§4.
- `2026-09-15_dgraph-adoption.md` — scenario B (consumer-owned dgraph backend at the
  ~50k-node trigger) is exactly where metaengine's dgraphengine becomes the natural
  adapter target rather than a direct `dgo` dependency.
- `docs/planning/2026-09-16_13-25_seam-store-search-adr.md` — the `VectorIndex`
  (`:157`) and `GraphStore` (`:179`) seams are the designed integration point. Note
  the impedance mismatch to solve if a metaengine adapter is ever attempted: the seam
  is full-rebuild (`Replace(map[string]Vector)`), while metaengine `VectorBackend` is
  incremental (`VectorInsert`); a wrapper must buffer or diff.

## 7. Revisit triggers

| Trigger                                                                             | Action                                                                                                    |
| ----------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| metaengine ships an ANN-capable engine (HNSW/PQ) behind `VectorBackend`             | Re-evaluate engine delegation behind the `VectorIndex` seam (shared with the 2026-09-15 adoption doc §9)  |
| Seam ADR implementation starts (trigger-gated per ADR §9)                           | Re-read §4 as the adapter-feasibility baseline; resolve the Replace-vs-Insert mismatch explicitly         |
| Dgraph native ANN uncouples metric-at-schema coupling or starts returning distances | The dgraphengine vector path stops being degraded; scenario B gets cheaper                                |
| CV adopts metaengine/system at the application layer (SUPERB T01/T29)               | go-graph-rag integrates as a query handler + rebuild command above metaengine, per the 2026-09-15 verdict |

---

_Point-in-time comparison snapshot (2026-09-19). Never rewrite; annotate inline or
archive. Re-verify file:line cites against both module trees before acting on any
claim here — line numbers drift._
