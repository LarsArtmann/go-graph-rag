# Changelog

All notable changes to this project are documented in this file.

## [Unreleased]

### Added

- Nothing yet.

### Changed

- Nothing yet.

### Fixed

- Release automation matched no CHANGELOG section on any tag: `awk -v` eats
  the backslash escapes in the header pattern before the regex engine sees
  them, so the v0.2.0 tag-push run failed at extraction and the release was
  created manually. Extraction now uses a plain-substring match that needs no
  escaping (`350b815`); the fixed workflow has not yet proven itself on a
  real tag.

## [0.2.0] - 2026-09-16

The public-SDK-face release: runnable godoc examples, the version-stamped
`UserAgentVersion` const, benchmark reference numbers, and a full living-doc
set are now what pkg.go.dev renders. No breaking changes; v0.1.0 consumers
can bump freely.

### Fixed

- `SECURITY.md` pointed at a 404 GitHub docs URL (the privately-report-a-
  vulnerability page moved); replaced with the verified live URL.
- README named the full-rebuild operation as a backticked `Reindex` — no
  such identifier exists; now names the real `ReplaceGraph`.
- Restored the json-v1 build contract: a post-v0.1.0 push had switched
  `embed_openai.go` and `store.go` to `encoding/json/v2`, which only
  compiles with `GOEXPERIMENT=jsonv2` and broke CI on pristine toolchains.
  The tagged v0.1.0 tree was never affected.
- The json-v1 contract regressed again (2026-09-16): `store.go`,
  `embed_openai.go`, and `embed_openai_test.go` came back as
  `encoding/json/v2` via auto-committed session work run under the
  machine-wide `GOEXPERIMENT=jsonv2` dev shell; the pre-push pristine-build
  gate blocked the push. Imports restored to `encoding/json`.

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
- Two more output-verified examples: `ExampleOpenStore` (persist +
  round-trip: store-as-cache, `ReplaceGraph`, `LoadGraph`/`Stats`) and
  `ExampleNewSearcherWithOptions` (search-time roles: `ReferenceKind`
  ref-match line, `DocumentKinds` two-tier ranking, `DetailKind` compact
  rendering); README's Search-time-roles section points at the latter.
- `UserAgentVersion` const: the OpenAI-compat User-Agent is version-stamped
  (was a hardcoded string); asserted by httptest.
- `Build` doc comment now states the identical-text rule: shared content
  hash = one provider call, first document in slice order gets the vector,
  siblings stay expansion-only until a warm-cache build fills them in.
- Benchmarks upgraded to a benchstat protocol (`-count 10`, spread ±1–2%)
  with the reference table rewritten in `bench_test.go`; new
  `BenchmarkStoreRoundTrip` (persist + snapshot load on a real SQLite file,
  1k/10k) shows persistence is not the rebuild bottleneck (10k nodes
  round-trip in ~0.2s against a ~37s rebuild).
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
