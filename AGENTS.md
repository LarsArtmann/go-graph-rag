# AGENTS.md — go-graph-rag

Standalone GraphRAG SDK extracted from the CV repo's `graphrag/` module
(2026-09-15). Module `github.com/larsartmann/go-graph-rag`, Go floor 1.26.7.

## What lives here

- `build.go` — `Build`: documents + caller edges in, embedded graph out
  (nodes, edges, derived `RelationSimilar` edges, vectors).
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
   `GOEXPERIMENT=jsonv2`. The one `omitempty`-on-float (`NodeAttrs.Score`)
   is documented for the v2-engine no-op behavior.
3. **stdlib-only except `samber/lo` + `modernc.org/sqlite`**
   (`testify` in tests). Adding a dependency needs a hard reason.
4. **No `primitives` / CV-repo imports.** The extraction cut exactly that
   dependency (`sleepContext` is inlined in `embed_openai.go`).
5. **golangci config is CV's, minus experiments/goheader.** Keep the two in
   sync when tightening linters; the v2 config format is required.

## Commands

```bash
GOTOOLCHAIN=go1.26.7 go build ./... && go vet ./... && go test ./...
golangci-lint run ./...
go mod tidy && git diff --exit-code go.mod go.sum   # tidy drift gate
```

CI (`.github/workflows/go-test.yml`) runs exactly these on every push.

## Upstream sync

The CV repo consumes this module from the proxy (`v0.x`). The CV-side
mounting points: `internal/graphvocab` (kind/relation constants),
`internal/features/graphrag/service` (SearcherOptions wiring, tracing/counting
decoration), `internal/config` (`GraphRAGConfig` alias).
