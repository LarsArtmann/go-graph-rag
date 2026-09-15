# Domain Language

> Ubiquitous language for the go-graph-rag SDK. Bounded context: this SDK is
> domain-neutral; wherever a term below has a consumer-side counterpart, the
> boundary is called out.

## Glossary

| Term                   | Definition                                                                                                                        | Where used                                    |
| ---------------------- | --------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------- |
| Document               | The indexing unit: one graph node plus the text to embed for it. Empty text = vector-less node.                                   | `build.go` (`Document`)                       |
| Node                   | One entity in the knowledge graph: ID, kind, label, attrs.                                                                        | `graph.go` (`Node`)                           |
| NodeKind               | The type of an entity (a string). The vocabulary is caller-defined constants.                                                     | `graph.go` (`NodeKind`)                       |
| Kind (vocabulary)      | The set of `NodeKind` constants a consumer defines for its domain. The SDK ships none.                                            | caller-side; CV keeps its own in `graphvocab` |
| Relation               | The type of an edge (a string). Caller-defined; all relations traverse both directions.                                           | `graph.go` (`Relation`)                       |
| RelationSimilar        | The one SDK-derived relation: links two embedded nodes of the same kind whose cosine clears the build threshold; weight = cosine. | `graph.go`, `build.go`                        |
| Edge                   | A typed, weighted, directed link between two nodes.                                                                               | `graph.go` (`Edge`)                           |
| Graph                  | Immutable in-memory view over nodes and edges with deterministic iteration.                                                       | `graph.go` (`Graph`)                          |
| Hub                    | An embedded node of a kind NOT in `DocumentKinds`; fills top-k slots after documents.                                             | `search.go` (`rankHits`)                      |
| Reference node         | The `ReferenceKind` node (e.g. the profile in a matching graph); never a hit; source of `RefMatch`.                               | `search.go` (`SearcherOptions.ReferenceKind`) |
| RefMatch               | A hit's cosine similarity to the reference node. Replaces the CV-era `CVMatch`.                                                   | `search.go` (`Hit.RefMatch`)                  |
| Document kinds (tier)  | The kinds ranked as primary documents in the two-tier ordering. Empty = every node is a document.                                 | `search.go` (`SearcherOptions.DocumentKinds`) |
| DetailKind             | The kind whose related nodes render as a compact label list in context output.                                                    | `search.go` (`SearcherOptions.DetailKind`)    |
| Two-tier ranking       | Documents first (score order), hubs fill remaining slots; prevents short labels crowding out documents.                           | `search.go` (`rankHits`)                      |
| Provider               | Turns texts into dense vectors; names the cache namespace with `Name()`/`Model()`.                                                | `embed.go` (`Provider`)                       |
| Embedding cache        | Persisted vectors keyed by content hash x provider x model so unchanged text never re-embeds.                                     | `store.go` (`Cache`, `graphrag_embeddings`)   |
| Content hash           | SHA-256 hex of an embedded text; the cache key.                                                                                   | `vector.go` (`HashText`)                      |
| SimilarPair            | Two nodes whose vectors clear a runtime threshold; the near-duplicate detector's result unit.                                     | `search.go` (`SimilarPair`)                   |
| Snapshot               | The full read model loaded from the store: nodes, edges, node-to-hash linkage.                                                    | `store.go` (`LoadGraphSnapshot`)              |
| Reindex (full rebuild) | Replacing the whole persisted graph in one transaction; the only write mode.                                                      | `store.go` (`ReplaceGraph`)                   |
| Context block          | The deterministic, LLM-ready text a `Search` renders from its hits.                                                               | `search.go` (`RenderContext`)                 |

## Bounded contexts

- **SDK vs consumer:** kind/relation vocabularies, feature gating (`Enabled`),
  and search policy VALUES live with the consumer; the SDK owns the TYPES and
  the mechanics. CV's `internal/graphvocab` is the reference consumer
  implementation.
- `RefMatch` is the SDK-neutral rename of the CV-era `CVMatch` (wire key
  `refMatch`); consumers migrating from the in-repo CV module should treat the
  two as the same concept.

## Deprecated terms

- `CVMatch`: pre-extraction name for `RefMatch`; exists only in CV migration
  notes, not in this codebase.
