# AGENTS.md — go-graph-rag

Standalone GraphRAG SDK extracted from the CV repo's `graphrag/` module
(2026-09-15). Module `github.com/larsartmann/go-graph-rag`, Go floor 1.27.1 (raised from 1.26.7 on 2026-09-19, owner-requested commit af3be2c).

## What lives here

- `build.go` — `Build`: documents + caller edges in, embedded graph out
  (nodes, edges, derived `RelationSimilar` edges, vectors).
- `config.go` — koanf-tagged `Config`/`EmbeddingConfig` so a consumer's root
  config can alias them; `NewProvider` consumes `EmbeddingConfig`.
- `embed.go` / `embed_hash.go` / `embed_openai.go` — provider seam +
  deterministic offline hasher + OpenAI-compatible client (stdlib
  `encoding/json` only, NOT json/v2). The UA header is stamped from
  `UserAgentVersion` — bump it when cutting a release.
- `graph.go` — `Node`/`Edge`/`Graph` + the `NodeKind`/`Relation` TYPES.
  Domain kind/relation vocabularies are caller-side (CV keeps its own in
  `internal/graphvocab`).
- `search.go` — `Searcher` hybrid retrieval + `SearcherOptions` policy
  (ReferenceKind / DocumentKinds / DetailKind) + `SimilarPairs` dedup scan.
- `store.go` — SQLite graph + embedding cache (namespaced
  content-hash x provider x model), snapshot load, health.
- `vector.go` — cosine math + Vector type.
- Tests: `example_test.go` (runnable godoc examples — keep the `// Output:`
  blocks truthful, `go test` enforces them), `bench_test.go` (reference
  numbers recorded in the file doc comment),
  `embed_openai_live_test.go` (env-gated, skips offline).
- `.githooks/pre-push` — three-gate guard (pristine build, tidy drift,
  dprint check); install once per clone with
  `git config core.hooksPath .githooks`.
- `.github/workflows/release.yml` — GitHub Release automation on `v*` tags
  (notes extracted from the matching CHANGELOG section; empty extraction fails
  the run). Proven end-to-end on the v0.3.0 tag (2026-09-19, green run,
  correct notes) after the v0.2.0 awk-extraction failure (fixed `350b815`).

## Non-negotiables

1. **No CV-domain vocabulary in prod code.** No `KindJob`, no job/company
   semantics, no CV-calibrated defaults in doc comments. Fixtures in tests
   may use any vocabulary.
2. **`encoding/json` (v1) only.** The SDK must build without
   `GOEXPERIMENT=jsonv2` (json/v2 is experiment-gated; importing it turns
   CI red). The one `omitempty`-on-float (`NodeAttrs.Score`) is
   documented for the v2-engine no-op behavior.
3. **stdlib-only except `samber/lo` + `modernc.org/sqlite`**
   (`testify` in tests). Adding a dependency needs a hard reason. Backend
   work is governed by the seam ADR
   (`docs/planning/2026-09-16_13-25_seam-store-search-adr.md`): adapters own
   their dependencies in separate modules; core `go.mod` never grows; core
   stays CGO-free.
4. **No `primitives` / CV-repo imports.** The extraction cut exactly that
   dependency (`sleepContext` is inlined in `embed_openai.go`).
5. **golangci config mirrors CV's, minus CV-specific depguard rules.**
   Since the Go 1.27.1 alignment (af3be2c, 2026-09-19, owner-requested) the
   config ALSO carries CV's experiment build-tags (`goexperiment.*`) and
   goheader, so the linter type-checks the experiment-enabled surface;
   the `env -u GOEXPERIMENT` pristine BUILD gate below remains the actual
   json-v2 regression guard. Keep the two configs in sync when tightening
   linters; the v2 config format is required.

## Commands

```bash
GOTOOLCHAIN=go1.27.1 go build ./... && go vet ./... && go test ./...
GOTOOLCHAIN=go1.27.1 golangci-lint run ./...   # without the prefix the
                                                 # dev-shell GOTOOLCHAIN=local
                                                 # cannot load go >= 1.27.1
go mod tidy && git diff --exit-code go.mod go.sum   # tidy drift gate
nix run nixpkgs#dprint -- check                     # markdown/json/yaml fmt
```

CI (`.github/workflows/go-test.yml`) runs exactly these on every push
(including the dprint check) and master is branch-protected on the
`build-test-lint` check. Occasional local extras: `go test ./... -race`
(clean as of 2026-09-15), `go test -bench . -benchtime 100ms ./...`, and the
opt-in live smoke test (`GRAPHRAG_LIVE_EMBED_URL`/`_KEY`/`_MODEL` env vars,
see `embed_openai_live_test.go`).

Benchmark reference numbers follow a benchstat protocol:
`go test -bench . -count 10` summarized with benchstat (pinned
`golang.org/x/perf v0.0.0-20260908200009`); the current table, machine, and
date live in the `bench_test.go` doc comment — re-record there whenever
numbers are re-measured.

## Gotchas

- The dev shell exports `GOEXPERIMENT=jsonv2` machine-wide; CI does not.
  The `.githooks/pre-push` guard now runs ALL THREE gates on every push
  (pristine build, tidy drift, dprint check — install with
  `git config core.hooksPath .githooks`), and the dprint/tidy gates ALSO
  catch non-experiment drift. All three blocking paths were proven by
  planted-failure dry-run pushes (2026-09-19).
- golangci-lint is pinned to v2.13.2 in CI (parity with CV's pin). Bump both
  repos together; the config requires the v2 binary.
- The json-v2 regression is RECURRING, not historical: it has slipped in
  three times (pre-v0.1.0, `e67bd9b`, `e4a9145`), always via daemon or
  parallel-session commits made under the machine-wide `GOEXPERIMENT=jsonv2`
  shell, and each time only the pristine build caught it. Run the pristine
  check before EVERY push and re-verify any daemon commits that land on top
  of yours.
- pkg.go.dev renders NO godoc for any version of this module
  ("Documentation not displayed due to license restrictions" — it does not
  recognize the PROPRIETARY LICENSE file), so the README's pkg.go.dev link
  lands on the restriction banner until the Q1 license decision lands.
  Verified live 2026-09-19 against v0.2.0.
- Markdown is dprint-formatted (`dprint.json`: table alignment, `_em_`
  style); Go is tab-indented gofmt (`.editorconfig`).
- New research artifacts get a row in `docs/research/README.md` (artifact,
  date, question, verdict, reopen-triggers). The index stays research-only:
  planning docs (ADRs, SUPERB plans, owner packages) never get rows.
- `buildflow`'s go-structure-linter reports 8 `root-package-files` findings
  (one per root `.go` file) on every run, and the buildflow findings gate
  exits non-zero because of them. Accepted policy (owner decision,
  2026-09-15): the flat single-package layout is deliberate for a v0 SDK.
  Do NOT restructure into `pkg/` or `internal/`; do NOT treat the gate exit
  as a session regression. No `.buildflow.yml` skip entry exists on
  purpose: skipping the whole step would mute the linter's other rules too.

## Owner decisions (2026-09-15)

- Corpus stays well under ~50k nodes for the next 12 months: storage work
  (ANN, incremental indexing) stays ROADMAP fuel, not near-term TODO work.
- Pluggable store/search seam: DESIGN approved (scenario B in
  `docs/research/2026-09-15_dgraph-adoption.md`); chosen, not started —
  live tracker is the High row in `TODO_LIST.md`.
- `docs/status/`, `docs/planning/`, `docs/research/` are point-in-time
  records: annotate inline (docs-health ANNOTATE) or `git mv` fully-done
  files to `<dir>/archived/`; never rewrite them.

## Owner decisions (2026-09-16)

- STATUS reports archive too: once every item in a `docs/status/` report is
  resolved (annotated), `git mv` it to `docs/status/archived/`.
- Seam-design deliverable (when it starts): ADR under `docs/planning/` WITH a
  Go interface sketch (seam signatures) — binding before any code lands.
- ROADMAP removes settled ideas entirely (decision trail lives in CHANGELOG +
  code docs); no struck-through zombies.
- Seam design DELIVERED 2026-09-16: ADR + compiling Go interface sketch at
  `docs/planning/2026-09-16_13-25_seam-store-search-adr.md`; implementation
  is trigger-gated (ADR §9), not scheduled.
- Daemon commits are pushed as-is, never rewritten. Upstream patches
  (tobi/qmd#959, charmbracelet/crush#3846) stay HELD past the 2026-09-22
  window — the patches live in the issue texts; no reminder pings.
- metaengine dep-tree quantification: run soon (owner 2026-09-16), not
  trigger-gated.
- CGO stance for the seam ADR: core stays CGO-free; CGO is acceptable inside
  optional adapter modules only.

## Owner decisions (2026-09-19)

- Release: v0.3.0 cut 2026-09-19 (hardening + throughput: EmbedConcurrency,
  Go 1.27.1 floor, fail-fast cache reads, Searcher defensive copies;
  release.yml green on the tag, notes extracted correctly, proxy indexed,
  clean-dir `go get` verified). v0.2.0 (tag `805aeba`) needed a manual
  release after the awk bug. pkg.go.dev still renders NO godoc (license
  restriction) — public face blocked on the Q1 license decision. CV-side
  bump to v0.3.0 pending (owner package Phase 8; can skip v0.2.0).

## Upstream sync

The CV repo consumes this module from the proxy (`v0.x`; latest `v0.3.0`,
CV-side bump pending — owner package Phase 8). The CV-side
mounting points: `internal/graphvocab` (kind/relation constants),
`internal/features/graphrag/service` (SearcherOptions wiring, tracing/counting
decoration), `internal/config` (`GraphRAGConfig` alias).
