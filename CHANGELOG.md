# Changelog

All notable changes to this project are documented in this file.

## [Unreleased]

### Added

- **BDD test-kit sub-module `bdd/`** (`github.com/larsartmann/go-graph-rag/bdd`, own go.mod): Ginkgo/Gomega test kit with `Doc`/`Edge` builders, a deterministic demo corpus + kind policy, a `Harness` over the real SDK machinery (hash provider by default, optional real SQLite store with `Reload` from disk), Gomega matchers (`HaveTopHit`, `ContainHit`, `HaveHitIDs`, `HaveNoHitsOfKind`, `HaveSimilarEdge`), and a `FailingProvider` for error paths. The kit keeps ginkgo/gomega out of the core dependency budget (AGENTS non-negotiable #3) and dogfoods itself via its own 43-spec black-box suite. Scope and decisions: `docs/planning/2026-10-15_05-02_bdd-submodule-adr.md`. Consumers cannot fetch it from the proxy until the first `bdd/v*` tag lands (owner-gated).

### Changed

- **Node identity is now the strong type `NodeID`** (brandid-lint BD003 type-safety sweep):
  `Node.ID`, `Edge.Source`/`Edge.Target`, and `Document.ID` carry `graphrag.NodeID` (documents and
  nodes share one ID namespace: a document's ID becomes its node's ID), and every ID-taking surface
  follows — `Graph.Node`/`OutEdges`/`InEdges`, `BuildResult.Vectors` (`map[NodeID]Vector`) and
  `.EmbeddingHash`, `NewSearcher`/`NewSearcherWithOptions`, `Store.ReplaceGraph`,
  `LoadGraphSnapshot.EmbeddingHash`, plus the BDD kit's `Doc`/`Edge`/`WeightedEdge` builders and
  `HaveTopHit`/`ContainHit`/`HaveHitIDs`/`HaveSimilarEdge`/`SimilarEdgeScore`. JSON wire shapes are
  byte-identical (wire goldens unchanged), untyped string literals still assign directly
  (`Node{ID: "job:1"}` keeps compiling); only callers passing `string` variables add an explicit
  `graphrag.NodeID(...)` conversion. The seam ADR §5 sketch was amended in step (ADR-first), so
  `VectorIndex.Replace` and `VectorMatch.NodeID` are `NodeID`-typed before any adapter exists.

### Fixed

- **Dead allocation in the concurrent-embed collector** (`embed_openai.go`): a 2026-10-11 refactor switched slot-flattening to `lo.Flatten` but left the preceding `make([]Vector, 0, len(texts))` allocation in place, immediately overwritten — flagged by `ineffassign` and caught while re-baselining CI.
- **CI Lint step green again**: `setup-go`'s `go-version: "1.27"` resolved to Go 1.27.2 on the runner, whose export-data format the pinned golangci-lint v2.13.2 cannot decode (every typecheck import failed; master red 2026-10-11..15). The workflow now pins `go-version: "1.27.1"` to match the `GOTOOLCHAIN` pin; the two pins must move together with the golangci-lint pin (CV parity rule).
- **Stale public-face docs re-synced** (2026-10-15 docs-health pass): README's install command now pins `v0.3.2`, its dependency-light constraint states the json/v1 rule as policy (json/v2 graduated — the compiler no longer discriminates), and the `NodeID` strong-ID constraint is surfaced; `SECURITY.md`'s supported-versions table now says `v0.3.x` (was still `v0.1.x`, contradicting the file's own latest-line policy); `docs/DOMAIN_LANGUAGE.md` gains the `NodeID` term.

## [0.3.2] - 2026-10-15

### Fixed

- **Duplicate-ID documents with distinct texts no longer lose their cached vectors.** The fresh-vector map was keyed through the ID-to-hash map, so same-ID documents with different texts (e.g. a skill hub reached from a JD extraction and from the CV with different label casing) overwrote each other's vectors: the provider embedded both, the cache persisted only one, `BuildResult.Embedded` undercounted, and every subsequent rebuild re-embedded the overwritten texts forever. The per-miss hashes now travel positionally beside the missed documents, so every distinct text is embedded once, cached once, and served from the cache on every later build. Regression pinned by `TestBuildDuplicateIDDistinctTextCachesAll`.

## [0.3.1] - 2026-09-30

### Fixed

- **Wire golden fixtures re-pinned for the Go 1.27 toolchain's `encoding/json/v2`**: numeric `omitempty` fields (e.g. `NodeAttrs.Score`) now serialize their zero value (`"score": 0`), so the three `testdata/wire/*.golden` fixtures gain the explicit zero. Wire output is textually different but semantically identical — both engines decode back to the same zero (see the `NodeAttrs.Score` doc comment). No code change.

## [0.3.0] - 2026-09-19

The hardening + throughput release: opt-in parallel embedding for
network-bound cold builds, a Go 1.27.1 language floor, fail-fast cache
reads, and structural immutability for `Searcher` inputs. Consumers need a
Go 1.27.1+ toolchain to build against this version; no exported signature
changed.

### Added

- `EmbeddingConfig.EmbedConcurrency` (koanf `embed_concurrency`, default 1):
  opt-in bounded worker pool that runs /embeddings batch requests in
  parallel for network-bound cold builds. Values at or below 1 take the
  identical serial path; results keep input order regardless of completion
  order; mid-flight cancellation aborts the whole call. Overlap, order
  preservation, cancellation, and race-freedom are pinned by tests, and the
  pool overhead is benched against an in-process endpoint.
- ADR seam guard: the §5 interface sketch (`VectorIndex`, `GraphStore`,
  `DeclaredMetric`, `VectorMatch`, `VectorQuery`) is copied verbatim into a
  compile-only test file with `var _` assertions proving the real `*Store`
  still satisfies the persistence seam, plus fake implementations and a
  composition smoke test. Core drift against the binding design now breaks
  the build instead of rotting silently.
- `.githooks/pre-push` now runs three gates (pristine build without
  `GOEXPERIMENT=jsonv2`, `go mod tidy` drift, dprint format check); every
  blocking path proven by planted-failure dry-run pushes.

### Changed

- Go language floor raised from 1.26.7 to 1.27.1 (owner-requested,
  `af3be2c`); CI toolchain pins, the pre-push hook, and local commands
  aligned. golangci-lint config now type-checks the experiment-enabled
  surface (`goexperiment.*` build tags) and enables goheader, matching CV.

### Fixed

- Release automation matched no CHANGELOG section on any tag: `awk -v` eats
  the backslash escapes in the header pattern before the regex engine sees
  them, so the v0.2.0 tag-push run failed at extraction and the release was
  created manually. Extraction now uses a plain-substring match that needs no
  escaping (`350b815`); this is the first tag cut with the fixed workflow.
- Cache READ errors are no longer swallowed: `Build` now fails fast with a
  wrapped error, symmetric with the fatal write path. A silently degraded
  cache caused paid re-embeds nobody could see; the alternative (a
  `CacheErrors` counter with degraded-mode continuation) was considered and
  rejected until real telemetry demand exists.
- `Searcher` now deep-clones caller inputs: the vectors map (including each
  vector's contents) and `SearcherOptions.DocumentKinds`. The immutability
  promise is structural — mutating caller copies after construction can no
  longer change search behavior.
- `resolveVectors` hashed every missed text twice (classification + result
  assignment); the hash recorded during cache lookup is now reused.

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
