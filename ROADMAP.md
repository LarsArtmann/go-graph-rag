# Roadmap

> Long-term direction and raw ideas. Items here are NOT actionable tasks.
> When an idea is refined into bounded work, it moves to TODO_LIST.md.

## Themes

### 1. Retrieval at scale

The index is a full-rebuild, linear-scan design that is deliberately simple.
Scale changes the math, not the API.

Raw ideas:

- ANN/HNSW vector backend behind the existing `Searcher` seam (candidates:
  sqlite-vec, hannoy, in-process HNSW port)
- Incremental indexing: upsert/delete of single documents instead of the
  `ReplaceGraph` full swap
- Corpus-size triggers that flip the implementation automatically (README
  names ~50k nodes as the revisit point)
- Optional pruning of stale `RelationSimilar` edges on rebuild

### 2. Provider ecosystem

The `Provider` seam is one interface; more backends are cheap.

Raw ideas:

- Additional first-class providers (local inference servers, other hosted
  embedding APIs)
- Provider health/capability reporting for multi-provider setups
- Batch-size and rate-limit hints surfaced per provider

### 3. Operational fitness

The store is a rebuildable derived index; operational hardening keeps that
promise cheap to honor.

Raw ideas:

- Metrics/tracing seams so consumers can decorate Build/Search without
  wrapping every call
- Store compaction and cache-eviction policy for long-lived files
- Backup/export story beyond "delete and rebuild"

### 4. API stability to v1

v0.x allows breaking changes; v1 freezes the surface.

Raw ideas:

- Settle `KindUnknown`'s public fate before v1 (see TODO_LIST)
- Codify the wire contract (`Hit`, `SearchResult` JSON keys) as frozen
- Document the migration path when ANN/incremental land

## Non-goals

Things we are deliberately NOT pursuing and why:

- **LLM/chat generation:** this is a retrieval SDK; prompt assembly and
  generation stay with consumers.
- **Domain vocabularies:** node kinds and relations are caller-defined; the
  SDK never ships semantic defaults.
- **Multi-writer stores:** one writer per store file by design; coordination
  leases belong to the caller.
- **External vector databases:** SQLite-only until scale demonstrably hurts;
  every persisted artifact stays rebuildable from source data.
