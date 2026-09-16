# AGENTS.md — go-graph-rag

Standalone GraphRAG SDK extracted from the CV repo's `graphrag/` module
(2026-09-15). Module `github.com/larsartmann/go-graph-rag`, Go floor 1.26.7.

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
- `.githooks/pre-push` — pristine-build guard; install once per clone with
  `git config core.hooksPath .githooks`.

## Non-negotiables

1. **No CV-domain vocabulary in prod code.** No `KindJob`, no job/company
   semantics, no CV-calibrated defaults in doc comments. Fixtures in tests
   may use any vocabulary.
2. **`encoding/json` (v1) only.** The SDK must build without
   `GOEXPERIMENT=jsonv2` (json/v2 is experiment-gated in Go 1.26; importing
   it turns CI red). The one `omitempty`-on-float (`NodeAttrs.Score`) is
   documented for the v2-engine no-op behavior.
3. **stdlib-only except `samber/lo` + `modernc.org/sqlite`**
   (`testify` in tests). Adding a dependency needs a hard reason.
4. **No `primitives` / CV-repo imports.** The extraction cut exactly that
   dependency (`sleepContext` is inlined in `embed_openai.go`).
5. **golangci config mirrors CV's, minus CV-specific depguard rules,
   experiment build-tags, and goheader** (the SDK lints exactly as it
   builds: without experiments). Keep the two in sync when tightening
   linters; the v2 config format is required.

## Commands

```bash
GOTOOLCHAIN=go1.26.7 go build ./... && go vet ./... && go test ./...
golangci-lint run ./...
go mod tidy && git diff --exit-code go.mod go.sum   # tidy drift gate
nix run nixpkgs#dprint -- check                     # markdown/json/yaml fmt
```

CI (`.github/workflows/go-test.yml`) runs exactly these on every push
(including the dprint check) and master is branch-protected on the
`build-test-lint` check. Occasional local extras: `go test ./... -race`
(clean as of 2026-09-15), `go test -bench . -benchtime 100ms ./...`, and the
opt-in live smoke test (`GRAPHRAG_LIVE_EMBED_URL`/`_KEY`/`_MODEL` env vars,
see `embed_openai_live_test.go`).

## Gotchas

- The dev shell exports `GOEXPERIMENT=jsonv2` machine-wide; CI does not.
  Before pushing, pristine-check with
  `env -u GOEXPERIMENT GOTOOLCHAIN=go1.26.7 go build ./...` — it is the only
  guard against experiment-gated stdlib sneaking in.
- golangci-lint is pinned to v2.13.2 in CI (parity with CV's pin). Bump both
  repos together; the config requires the v2 binary.
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
- Next work session: the seam design itself (ADR + interface sketch, f.6).
- metaengine dep-tree quantification: run soon (owner 2026-09-16), not
  trigger-gated.
- CGO stance for the seam ADR: core stays CGO-free; CGO is acceptable inside
  optional adapter modules only.

## Upstream sync

The CV repo consumes this module from the proxy (`v0.x`). The CV-side
mounting points: `internal/graphvocab` (kind/relation constants),
`internal/features/graphrag/service` (SearcherOptions wiring, tracing/counting
decoration), `internal/config` (`GraphRAGConfig` alias).
