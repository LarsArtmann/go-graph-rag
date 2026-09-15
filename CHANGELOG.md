# Changelog

All notable changes to this project are documented in this file.

## [0.1.0] - 2026-09-15

Initial extraction from the CV repository's in-repo `graphrag/` module.

### Added

- `Build` index construction over caller-supplied documents and edges with
  derived `RelationSimilar` edges.
- Embedding providers: deterministic offline hasher and OpenAI-compatible
  HTTP client (timeout, retry with backoff, batched inputs).
- SQLite-backed graph store with a content-hash namespaced embedding cache.
- Hybrid `Searcher`: vector ranking with two-tier document/hub ordering,
  graph expansion per hit, deterministic LLM-ready context rendering, and a
  `SimilarPairs` near-duplicate scan.
- `SearcherOptions` policy: reference node (`RefMatch`), document kinds, and
  compact detail rendering — kind vocabularies are fully caller-defined.

### Changed (relative to the CV in-repo module)

- Domain-neutral API: CV node kinds (`job`/`company`/`skill`/`project`/`cv`)
  and CV relations moved to the consumer; `Hit.CVMatch` renamed
  `Hit.RefMatch` (wire key `refMatch`); `RenderContext` is a `Searcher`
  method.
- Builds without `GOEXPERIMENT=jsonv2` (stdlib `encoding/json`).
- Cut the `CV/primitives` dependency (context-aware sleep inlined).
- Node kinds / relations ship as types only; `SimilarThreshold` default is
  documented as hash-provider-calibrated, callers with neural embeddings
  should raise it via `BuildOptions`.
