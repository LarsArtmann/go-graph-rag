# Roadmap

> Long-term direction and raw ideas. Items here are NOT actionable tasks.
> When an idea is refined into bounded work, it moves to TODO_LIST.md.

## Themes

### 1. Retrieval at scale

The index is a full-rebuild, linear-scan design that is deliberately simple.
Scale changes the math, not the API. The integration point for everything in
this theme is decided: `docs/planning/2026-09-16_13-25_seam-store-search-adr.md`
(VectorIndex/GraphStore seams, three backend classes, adapter modules keep core
dependency-free).

Raw ideas:

- ANN/HNSW vector backend behind the seam ADR's `VectorIndex` interface
  (candidates: sqlite-vec, hannoy, in-process HNSW port)
- Incremental indexing: upsert/delete of single documents instead of the
  `ReplaceGraph` full swap
- Corpus-size triggers that flip the implementation automatically (~50k
  nodes is the measured revisit point: Build/SimilarPairs scale quadratically,
  ~370ms per 1000-doc build, ~15min extrapolated at 50k — `bench_test.go`)
- Optional pruning of stale `RelationSimilar` edges on rebuild
- Shared ANN benchmark harness (recall + p50/p95 latency vs the brute-force
  scan) so backend candidates are comparable
- Metadata-filtered vector-search semantics: kind filters + `MinScore`
  composed with ANN, preserving today's linear-scan behavior
- ADR for the ANN winner: choice, migration path, and `RelationSimilar`
  derivation at ANN scale
- Parallel pairwise scan with a deterministic merge (shard by row index;
  the existing comparators are total orders, so output stays
  byte-identical) — expected ~cores× on the measured quadratic paths
  (review: `docs/architecture-understanding/2026-09-19_13-30_architecture-review.html`)
- Dot-of-normalized internal fast path for pairwise cosine (every stored
  vector is normalized at ingest; a pure dot skips ~2/3 of the per-pair
  arithmetic) — owner-gated: shifts user-visible scores in the last
  decimals and requires re-pinning golden tests

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
  every persisted artifact stays rebuildable from source data. Evaluated
  2026-09-15 (`docs/research/2026-09-15_dgraph-adoption.md`): verdict — do
  not adopt into the SDK; revisit as a consumer-owned backend at the ~50k
  trigger. Any backend enters only as an isolated adapter module per the
  seam ADR (`docs/planning/2026-09-16_13-25_seam-store-search-adr.md`);
  core `go.mod` never grows.
