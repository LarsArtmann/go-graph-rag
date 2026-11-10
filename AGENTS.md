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
  `encoding/json/v2`). The UA header is stamped from
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
- CI (`.github/workflows/go-test.yml`) — authoritative gate on every push/PR for BOTH
  modules: json/v2 policy check (bare `encoding/json` v1-shim import anywhere fails;
  polarity flipped 2026-11-10 with the "go v2" ruling, planted-failure proven both ways),
  build+vet, `go test -race`, golangci-lint, tidy drift, dprint. The former
  `.githooks/pre-push` local guard was deliberately dropped by the owner
  (`2002882`, 2026-11-02): opt-in local gates are false comfort; CI is
  unbypassable.
- `.github/workflows/release.yml` — GitHub Release automation on `v*` tags
  (notes extracted from the matching CHANGELOG section; empty extraction fails
  the run). Proven end-to-end on the v0.3.0 tag (2026-09-19, green run,
  correct notes) after the v0.2.0 awk-extraction failure (fixed `350b815`).

## Non-negotiables

1. **No CV-domain vocabulary in prod code.** No `KindJob`, no job/company
   semantics, no CV-calibrated defaults in doc comments. Fixtures in tests
   may use any vocabulary.
2. **`encoding/json/v2` is the canonical engine** (owner ruling "go v2",
   2026-11-10). On Go 1.27 the v1 package is a compatibility shim implemented
   on top of v2 (`v2_encode.go`/`v2_decode.go` inside `encoding/json`), so the
   SDK imports v2 directly and the bare v1 import is FORBIDDEN everywhere —
   CI's `json/v2 policy check` enforces it. The engines emit different bytes
   (v1 omits zero numerics under `omitempty` where v2 writes them —
   `NodeAttrs.Score`; v1 sorts map keys where v2 uses iteration order — no Go
   map crosses a prod wire path). History: the inverse v1-only policy ran
   2026-09-15..2026-11-10.
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
  The former pre-push hook was dropped by the owner (`2002882`, 2026-11-02:
  CI enforces the same checks and cannot be forgotten or bypassed) — the
  pristine `env -u GOEXPERIMENT` build and `go mod tidy` drift checks are now
  HABITS documented above, and CI's json/v1 policy check mechanizes the one
  gate no build tool provides.
- golangci-lint pinned v2.14.0 in CI = local dev-shell binary (2026-10-15;
  v2.13.2 could not decode Go 1.27.2 export data). CV still pins v2.13.2 —
  bump it there in its own session (bump-together rule).
- The json-v2 regression is a STANDING daemon risk, not an event: after
  the three documented slips (pre-v0.1.0, `e67bd9b`, `e4a9145`), a FOURTH
  (`bfaa047`, 2026-11-10) was caught by release diff review during v0.4.0
  prep — and the same sweep then found `store.go` had carried json/v2
  CONTINUOUSLY since v0.3.1, shipping in every released 0.3.x and in
  v0.4.0 (reverted on master 2026-11-10, `[Unreleased]`; v0.4.1 decision
  owner-gated). Diff review only covers the delta; the ONLY complete
  catcher is a full-tree `grep -rn 'encoding/json/v2' --include='*.go' .`
  (`wire_golden_test.go` is the sanctioned exemption — it pins v1 bytes
  and documents v2's omitempty difference). CI enforces exactly this grep
  on every push/PR since 2026-11-10; local habit optional.
- pkg.go.dev RESOLVED (2026-11-10): godoc was hidden while the repo carried
  the PROPRIETARY LICENSE (all versions through v0.3.3 — their proxy zips
  are immutable, so their restriction is PERMANENT). v0.4.0 is the first
  MIT-carrying tag and its page was verified rendering the same day (full
  godoc, examples, License: MIT). The tagged-tree README still shows the
  pre-verification caveat (renders on pkg.go.dev); master's is updated.
- Markdown is dprint-formatted (`dprint.json`); Go is gofmt (`.editorconfig`).
- New research artifacts get a row in `docs/research/README.md` (artifact,
  date, question, verdict, reopen-triggers). The index stays research-only:
  planning docs (ADRs, SUPERB plans, owner packages) never get rows.
- `_test.go` FILENAME SUFFIX is what makes a file a test file — not the
  `package foo_test` clause. A spec file named `build_spec.go` declaring
  `package bdd_test` is a REGULAR source file and breaks the package with
  "found packages bdd (…) and bdd_test (…)". This cost a full debugging
  session 2026-10-15; spec files MUST end in `_spec_test.go` (or `_test.go`).
- Ginkgo forbids `go test -count=N` (N>1); re-seed via repeated `-count=1`
  runs or `ginkgo -repeat=N`. The bdd suite is order-independent (verified
  across seeds 2026-10-15).
- `buildflow`'s findings gate exits non-zero on accepted noise; do NOT treat
  the gate exit as a session regression. Mix (re-baselined 2026-10-15,
  binary 518739c, behind BuildFlow HEAD): go-structure-linter `agent-config`
  error while AGENTS.md exceeds 220 lines; `securitymd` 2 owner-gated
  errors (TODO_LIST); erraudit held at 0 by one `//nolint:legacyerrors`
  sentinel suppression — golangci warns "unknown linter legacyerrors" about
  it (permanent two-tool token disagreement; keep the directive). Flat
  single-package layout is deliberate (no `pkg/`/`internal/`); no
  `.buildflow.yml` skip entry on purpose (it would mute other rules too).
- CI is proven on ubuntu-26.04 (2026-10-15 green dispatch run — all steps
  pass); `go-test.yml` pins the runner ahead of the 2026-10-19 migration.
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

## Owner decisions (2026-11-10)

- **"go v2"** — `encoding/json/v2` is the canonical engine everywhere (see
  non-negotiable #2). Verified first: on go1.27.x the v1 package is a shim
  implemented over v2; the engines differ on omitempty-numerics and map-key
  order; no prod map crosses a wire path. Executed as **v0.5.0** (tag
  `9af2b9d`, CI green on the exact SHA; release.yml green + notes verified;
  proxy indexed; clean-dir `go get` prints `0.5`; CI gate polarity flipped
  with planted proof). Supersedes the v1-only policy AND the v0.4.1
  question. pkg.go.dev rendering pending at cut time (v0.4.0 rendered after
  ~40 min — same expectation).
- License posture (owner package Q1): DECIDED — **MIT**. PROPRIETARY LICENSE
  replaced with the MIT text (Copyright (c) 2026 Lars Artmann); README,
  CONTRIBUTING (inbound-grant line applied), TODO_LIST, CHANGELOG, and the
  owner-decision package synced in the same pass. pkg.go.dev godoc
  re-verification rides the next tagged release.
- Release: **v0.4.0 cut 2026-11-10** (tag at `656e42e`, CI green on the
  exact commit; release.yml green, notes correctly extracted from the
  `[0.4.0]` section; proxy indexed; clean-dir `go get` + compile + run
  verified; `UserAgentVersion` bumped to "0.4" in lockstep — test-enforced
  pairing). Discovery during prep: **v0.3.3 (2026-11-03) had silently
  shipped the NodeID/ContentHash breaking changes as an unlisted patch**
  (no CHANGELOG section, no UA bump; GH Releases for v0.3.1 + v0.3.3 need
  manual backfill — TODO_LIST row). json/v2 regression #4 (daemon commit
  `bfaa047`) caught by release diff review and reverted pre-tag. pkg.go.dev
  rendering of v0.4.0 was still pending at cut time.

## Upstream sync

The CV repo consumes this module from the proxy (`v0.x`; CV was pinned to
v0.3.2 as of 2026-10-15, and v0.3.3 + v0.4.0 shipped since — so the next
CV-facing event is the **CV-side bump to v0.4.0**, which also picks up the
NodeID/ContentHash breaking types that v0.3.3 carried unannounced;
migration guide `docs/upgrading.md`). The CV-side
mounting points: `internal/graphvocab` (kind/relation constants),
`internal/features/graphrag/service` (SearcherOptions wiring, tracing/counting
decoration), `internal/config` (`GraphRAGConfig` alias).
