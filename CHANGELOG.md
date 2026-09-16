# Changelog

All notable changes to this project are documented in this file.

## [Unreleased]

### Fixed

- Restored the json-v1 build contract on `master`: a post-v0.1.0 push had
  switched `embed_openai.go` and `store.go` to `encoding/json/v2`, which only
  compiles with `GOEXPERIMENT=jsonv2` and broke CI on pristine toolchains.
  The tagged v0.1.0 tree was never affected.

### Added

- Repository tooling: grouped dependabot updates (Go modules + Actions),
  dprint formatting config, `.editorconfig`, `.gitattributes`, `.gitignore`,
  and CONTRIBUTING.md.
- Living project docs: FEATURES.md, TODO_LIST.md, ROADMAP.md, and
  docs/DOMAIN_LANGUAGE.md.
- Runnable godoc examples (`ExampleBuild`, `ExampleSearcher_Search`,
  `ExampleSearcher_SimilarPairs`, `ExampleNewProvider`) with
  output-verified comments; the README quick-start snippet is now
  compile-verified against the module.
- `UserAgentVersion` const: the OpenAI-compat User-Agent is version-stamped
  (was a hardcoded string); asserted by httptest.
- Benchmark suite (`bench_test.go`) for Build / Search / SimilarPairs over a
  synthetic corpus, with reference numbers recorded in-file; the quadratic
  pairwise cost (~373ms per 1000-doc build) now has measured backing.
- SECURITY.md with GitHub private vulnerability reporting enabled.
- Env-gated live-endpoint smoke test (`GRAPHRAG_LIVE_EMBED_URL` family);
  skips cleanly offline.
- `.githooks/pre-push` pristine-build guard (blocks pushes of trees that
  only build under `GOEXPERIMENT=jsonv2`); install via
  `git config core.hooksPath .githooks`.
- Adoption research decision records: `docs/research/2026-09-15_dgraph-adoption.md`
  (verdict: do not adopt into the SDK; revisit as a consumer-owned backend
  at the ~50k trigger) and `docs/research/2026-09-15_metaengine-system-adoption.md`
  (adopt neither into the SDK; CV app-layer adoption per SUPERB T01/T29),
  plus the owner-decision package (license posture, v0.2.0 cadence) under
  `docs/planning/` and a `docs/research/` index.
- Tag protection: a `protect-tags` ruleset (repo rulesets, enforcement
  active) covers `refs/tags/*` with deletion and non-fast-forward rules and
  no bypass actors, so tags can no longer be moved or deleted silently.
- Release automation (`.github/workflows/release.yml`): a future `v*` tag
  push publishes a GitHub Release with notes extracted from the matching
  `CHANGELOG.md` section; an empty extraction fails the run. Drafted only —
  no release executed; the v0.2.0 cut stays owner-gated.
- Seam ADR `docs/planning/2026-09-16_13-25_seam-store-search-adr.md`: the
  binding design for pluggable persistence (`GraphStore`) and vector ranking
  (`VectorIndex`) — three backend classes (embedded ANN libs / metaengine
  projection / consumer-owned server DBs), compiling Go interface sketch,
  CGO-in-adapters-only stance, measured dependency-isolation numbers
  (metaengine: +27 modules; +system: 6.6×), and trigger-gated adapter plan.

### Changed

- golangci config: dropped CV-specific depguard rules; experiment build-tags
  and goheader re-removed so the SDK lints exactly as it builds (without
  experiments).
- `KindUnknown` stays exported; documented as the zero-value placeholder
  contract (appears only on `SimilarPair` placeholders for vectors whose
  node vanished).
- CI: dprint format check step (SHA-pinned `dprint/check`); dependabot
  actions bump merged (checkout v7.0.1, setup-go v7.0.0).
- Repository: `master` branch protection requires the `build-test-lint`
  check; topics, pkg.go.dev homepage, and private vulnerability reporting
  configured.
- Full test suite verified clean under `-race`.

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
