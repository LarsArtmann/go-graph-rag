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
  `encoding/json` only, NOT json/v2).
- `graph.go` — `Node`/`Edge`/`Graph` + the `NodeKind`/`Relation` TYPES.
  Domain kind/relation vocabularies are caller-side (CV keeps its own in
  `internal/graphvocab`).
- `search.go` — `Searcher` hybrid retrieval + `SearcherOptions` policy
  (ReferenceKind / DocumentKinds / DetailKind) + `SimilarPairs` dedup scan.
- `store.go` — SQLite graph + embedding cache (namespaced
  content-hash x provider x model), snapshot load, health.
- `vector.go` — cosine math + Vector type.

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
```

CI (`.github/workflows/go-test.yml`) runs exactly these on every push.

## Gotchas

- The dev shell exports `GOEXPERIMENT=jsonv2` machine-wide; CI does not.
  Before pushing, pristine-check with
  `env -u GOEXPERIMENT GOTOOLCHAIN=go1.26.7 go build ./...` — it is the only
  guard against experiment-gated stdlib sneaking in.
- golangci-lint is pinned to v2.13.2 in CI (parity with CV's pin). Bump both
  repos together; the config requires the v2 binary.
- Markdown is dprint-formatted (`dprint.json`: table alignment, `_em_`
  style); Go is tab-indented gofmt (`.editorconfig`).

## Upstream sync

The CV repo consumes this module from the proxy (`v0.x`). The CV-side
mounting points: `internal/graphvocab` (kind/relation constants),
`internal/features/graphrag/service` (SearcherOptions wiring, tracing/counting
decoration), `internal/config` (`GraphRAGConfig` alias).
