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
- `graph.go` — `Node`/`Edge`/`Graph` + the `NodeKind`/`Relation`/`NodeID` TYPES.
  Domain kind/relation vocabularies are caller-side (CV keeps its own in
  `internal/graphvocab`). Node identity is the strong `NodeID` (BD003
  hardening, 2026-10-15): `Node.ID`, `Edge.Source`/`Target`, and
  `Document.ID` share the one namespace (a document's ID becomes its node's
  ID); JSON wire shapes are unchanged and untyped string literals still
  assign directly — only `string` variables need `graphrag.NodeID(...)`.
- `search.go` — `Searcher` hybrid retrieval + `SearcherOptions` policy
  (ReferenceKind / DocumentKinds / DetailKind) + `SimilarPairs` dedup scan.
- `store.go` — SQLite graph + embedding cache (namespaced
  content-hash x provider x model), snapshot load, health.
- `vector.go` — cosine math + Vector type.
- `bdd/` — BDD test-kit SUB-MODULE (own `go.mod`, `github.com/larsartmann/go-graph-rag/bdd`):
  builders, demo corpus, real-machinery `Harness`, Gomega matchers, `FailingProvider`;
  ginkgo/gomega live here and ONLY here. Requires the SDK by published version
  with a `replace => ../` for local dev. Decisions + interface sketch:
  `docs/planning/2026-10-15_05-02_bdd-submodule-adr.md`. First `bdd/v*` tag is
  owner-gated; until then consumers must use a sibling checkout.
- Tests: `example_test.go` (runnable godoc examples — keep the `// Output:`
  blocks truthful, `go test` enforces them), `bench_test.go` (reference
  numbers recorded in the file doc comment),
  `embed_openai_live_test.go` (env-gated, skips offline).
- `.githooks/pre-push` — three-gate guard for BOTH modules (pristine build
  [bdd adds `go vet`: `go build` skips `_test.go` files], tidy drift, dprint
  check); install once per clone with
  `git config core.hooksPath .githooks`. All bdd gate arms proven by planted
  failures 2026-10-15.
- `.github/workflows/release.yml` — GitHub Release automation on `v*` tags
  (notes extracted from the matching CHANGELOG section; empty extraction fails
  the run). Proven end-to-end on the v0.3.0 tag (2026-09-19, green run,
  correct notes) after the v0.2.0 awk-extraction failure (fixed `350b815`).

## Non-negotiables

1. **No CV-domain vocabulary in prod code.** No `KindJob`, no job/company
   semantics, no CV-calibrated defaults in doc comments. Fixtures in tests
   may use any vocabulary.
2. **`encoding/json` (v1) only.** POLICY, no longer build-enforced: since the
   Go 1.27.x toolchains, `encoding/json/v2` graduated into stdlib and imports
   compile WITHOUT `GOEXPERIMENT` (verified 2026-10-15 on go1.27.1, core and
   bdd plants) — the pristine-build gate can no longer discriminate a json/v2
   import. Keep v1 for API stability/wire-compat reasons. The one
   `omitempty`-on-float (`NodeAttrs.Score`) is documented for the v2-engine
   no-op behavior.
3. **stdlib-only except `samber/lo` + `modernc.org/sqlite`**
   (`testify` in tests). Adding a dependency needs a hard reason. Test-only
   deps (ginkgo, gomega) live in the `bdd/` sub-module, never core. Backend
   work is governed by the seam ADR
   (`docs/planning/2026-09-16_13-25_seam-store-search-adr.md`): adapters own
   their dependencies in separate modules; core `go.mod` never grows; core
   stays CGO-free.
4. **No `primitives` / CV-repo imports.** The extraction cut exactly that
   dependency (`sleepContext` is inlined in `embed_openai.go`).
5. **golangci config mirrors CV's, minus CV-specific depguard rules.**
   Since the Go 1.27.1 alignment (af3be2c, 2026-09-19, owner-requested) the
   config ALSO carries CV's experiment build-tags (`goexperiment.*`) and
   goheader, so the linter type-checks the experiment-enabled surface. Keep
   the two configs in sync when tightening linters; the v2 config format is
   required. From `bdd/` the linter discovers this parent config
   automatically (verified 2026-10-15).

## Commands

```bash
GOTOOLCHAIN=go1.27.1 go build ./... && go vet ./... && go test ./...
GOTOOLCHAIN=go1.27.1 golangci-lint run ./...   # without the prefix the
                                                 # dev-shell GOTOOLCHAIN=local
                                                 # cannot load go >= 1.27.1
go mod tidy && git diff --exit-code go.mod go.sum   # tidy drift gate
nix run nixpkgs#dprint -- check                     # markdown/json/yaml fmt
brandid-lint -rules BD001,BD003 .                   # strong-ID gate (BD003
                                                    # report-only; -fix only
                                                    # patches BD001 stubs)
# bdd sub-module (same loop, from bdd/; golangci finds the parent config):
(cd bdd && env -u GOEXPERIMENT GOTOOLCHAIN=go1.27.1 go test ./...)
(cd bdd && GOTOOLCHAIN=go1.27.1 golangci-lint run ./...)
(cd bdd && go mod tidy && cd .. && git diff --exit-code bdd/go.mod bdd/go.sum)
```

CI (`.github/workflows/go-test.yml`) runs exactly these on every push for BOTH
modules (bdd steps: build+vet, `go test -race`, golangci-lint, tidy drift) and
master is branch-protected on the `build-test-lint` check. CI tests run with
`-race` (added 2026-10-15). Occasional local extras: `go test -bench .
-benchtime 100ms ./...`, and the opt-in live smoke test
(`GRAPHRAG_LIVE_EMBED_URL`/`_KEY`/`_MODEL` env vars, see
`embed_openai_live_test.go`).

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
- The json-v2 history: it slipped in three times (pre-v0.1.0, `e67bd9b`,
  `e4a9145`), always via daemon or parallel-session commits made under the
  machine-wide `GOEXPERIMENT=jsonv2` shell, and back then only the pristine
  build caught it. CORRECTION (2026-10-15): on go1.27.1 json/v2 imports
  compile WITHOUT the flag (graduated), so the pristine gate now guards
  general drift, not json/v2; the v1-only rule is policy (see
  non-negotiable #2). Still run the pristine check before EVERY push and
  re-verify any daemon commits that land on top of yours.
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
- `_test.go` FILENAME SUFFIX is what makes a file a test file — not the
  `package foo_test` clause. A spec file named `build_spec.go` declaring
  `package bdd_test` is a REGULAR source file and breaks the package with
  "found packages bdd (…) and bdd_test (…)". This cost a full debugging
  session 2026-10-15; spec files MUST end in `_spec_test.go` (or `_test.go`).
- Ginkgo forbids `go test -count=N` for N>1 ("Only -count=1 is allowed"); to
  re-run with fresh random seeds use repeated `go test -count=1` invocations
  or `ginkgo -repeat=N`. The bdd suite is order-independent (verified across
  seeds 2026-10-15).
- `buildflow`'s findings gate exits non-zero on accepted noise; do NOT treat
  the gate exit as a session regression. Current mix (buildflow binary newer
  than the 2026-09-15 era, re-baselined 2026-10-15): go-structure-linter
  errors `agent-config` (AGENTS.md older than 14d — self-resolves on AGENTS
  updates) and `race-condition-test` (FIXED 2026-10-15: CI now runs
  `-race`); `securitymd` 2 errors (SECURITY.md contact/practices — owner
  decision, TODO_LIST); `erraudit` root error `embed_openai.go:334`
  (`errors.As` → `AsType` modernization, TODO_LIST; core freeze kept it out
  of the kit delivery). The flat single-package layout stays deliberate —
  do NOT restructure into `pkg/` or `internal/`. No `.buildflow.yml` skip
  entry exists on purpose: skipping the whole step would mute the linter's
  other rules too.
- Git Town runs on this machine and can snapshot dirty working-tree files as
  dangling "Git Town WIP" commits, then leave the files reverted (happened
  2026-10-15: TODO_LIST/ROADMAP/status-report writes vanished mid-session and
  were recovered via `git restore --source=<wip-sha> -- <files>`). After any
  batch of file writes, re-check `git status` still shows them modified; if
  content vanishes, look in `git fsck --lost-found` before redoing the work.

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

## Owner decisions (2026-10-15)

- BDD kit (`bdd/` sub-module) DELIVERED in-tree: 43-spec suite green,
  `-race` clean, lint 0 issues, CI + pre-push gates wired and proven by
  planted failures. Tagging `bdd/v0.1.0` is owner-gated (TODO_LIST row);
  `release.yml` will not fire on `bdd/v*` tags without a trigger change.
- Core frozen during kit delivery (byte-identical `go.mod`/`go.sum`/files):
  the erraudit `errors.As`→`AsType` modernization and SECURITY.md
  contact/practices sections are queued in TODO_LIST, not applied.

## Upstream sync

The CV repo consumes this module from the proxy (`v0.x`; latest `v0.3.2` —
the 2026-10-15 consumer research observed CV already pinned to v0.3.2, so the
next CV-facing event is the unreleased-`NodeID` release decision tracked in
`TODO_LIST.md`, not a lagging bump). The CV-side
mounting points: `internal/graphvocab` (kind/relation constants),
`internal/features/graphrag/service` (SearcherOptions wiring, tracing/counting
decoration), `internal/config` (`GraphRAGConfig` alias).
