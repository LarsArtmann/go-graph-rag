# Atlas — ecosystem Graph RAG with a git time dimension

**Date:** 2026-11-13
**Status:** PROPOSED (design exploration, not started)
**Home:** future consumer repo (working name `atlas`); recorded here because the design is expressed against this SDK's current API. When the consumer repo starts, move/rewrite as its ADR and keep this as the originating record.

## Premise

The goal: a Graph RAG presentation of all LarsArtmann programming projects with a git-based time dimension — one knowledge graph spanning every repo (files, docs, commits, dependencies), searchable semantically, and aware of _when_ things happened, changed, or co-evolved.

The SDK already carries the hook this needs: `NodeAttrs.PostedAt *time.Time` (`graph.go`). The entire design therefore builds as a **consumer repo on top of `go-graph-rag`, with zero core changes**.

## Architecture: one live graph + rebuildable as-of snapshots

### Tier 1 — Ecosystem graph (event time lives in the graph)

A harvester CLI walks all repos and produces `Document`s plus caller edges:

| Node kind | Embedded text              | Time carrier                                      |
| --------- | -------------------------- | ------------------------------------------------- |
| `project` | README summary             | repo first-commit date                            |
| `file`    | source/head (filtered)     | last-commit date via `git log -1 -- path`         |
| `doc`     | ADRs, CHANGELOG, TODO_LIST | doc mtime from git                                |
| `commit`  | subject + body             | `PostedAt` = author date, `Location` = short hash |
| `person`  | _(hub, no vector)_         | —                                                 |

Caller edges:

- `project —contains→ file/doc`
- `commit —touches→ file` (from `git log --name-status`; no vectors, so cheap)
- `commit —authored_by→ person`
- `project —depends_on→ project` from `go.mod` requires — the ecosystem's real dependency web (go-aichat ← InboxClean, go-graph-rag ← CV, ...). This cross-repo edge set is the killer feature: the real dependency graph Lars's projects already form.

The SDK auto-derives `similar_to` edges across everything, giving a cross-repo semantic map; `SimilarPairs` doubles as a cross-repo duplication detector.

### Tier 2 — Time queries without touching the SDK

`SearcherOptions` has kind filters (`ReferenceKind`/`DocumentKinds`/`DetailKind`) but no time filter. Do not add one to core: wrap the `Searcher` (the CV decoration pattern — tracing/counting decorators) and post-filter `Hits` on `Node.Attrs.PostedAt ∈ [since, until]`, keeping `Related` expansion intact. Surface:

```bash
atlas search "event store error handling" --since 2026-09
```

### Tier 3 — Time travel via as-of rebuilds

Full bitemporal attrs (valid-from/valid-to on every node) would be the "pure" model but is invasive and explodes node count with per-blob file versions. Cheaper and just as useful:

```bash
atlas build --as-of 2026-09-01
```

checks out that date into a `git worktree`, builds a **second graph into its own SQLite file** (`Store` is one graph per DSN — separate files just work), and points `Build`'s `Cache` seam at the shared long-lived cache DB so unchanged texts re-embed nothing (cache is namespaced by content-hash × provider × model).

"What changed" = node-set diff between two snapshot files; cross-snapshot `SimilarPairs` = semantic drift detection.

## Presentation surfaces (pick 1–2)

1. **CLI search** — cheapest, ships the value first.
2. **Web UI** — force-directed graph colored by project + a timeline scrubber that filters on `PostedAt`; the "watch your ecosystem grow" view. templ + HTMX fits the house stack.
3. **Chat** — `SearchResult.ContextText` is already LLM-ready; go-aichat + this graph = "ask your codebase archive".

## Build order and scale guards

1. Harvester + live graph + CLI search (offline hash provider first, real OpenAI-compatible endpoint after — `embed_openai.go` takes any base URL, local llama.cpp included).
2. Time decorator, `--since`/`--until`.
3. As-of snapshots + diff.
4. Web timeline presentation.

**Scale guard:** similar-edge derivation is pairwise over embedded nodes, so do not embed every commit. Roughly 30 repos × files + selective commits (doc/config-touching, merges, releases, or a 12-month window) keeps the corpus in the ~5–10k embedded-node sweet spot. `commit —touches→ file` edges are cheap because they carry no vectors. Commit selection policy should be config, decided by measured numbers.

## SDK gaps (informational — none blocking, all consumer-solvable)

- No time-range filter in `SearchOptions` → solved by the Tier-2 decorator; a core `Since/Until` filter could be a later, separately-justified addition.
- No graph naming in `Store` (one graph per SQLite file) → as-of snapshots use separate DSNs; a graph-name column would be convenience, not necessity.
- No incremental indexing (full rebuild each run) → fine at this corpus size; already ROADMAP fuel for this SDK.

## Execution notes (when the owner says go)

- Load the `stack-selection` skill first (new-project foundation).
- Scaffold the consumer repo; write this design there as its ADR per the planning convention.
- Everything runs locally: SQLite + optional local OpenAI-compatible embedding endpoint; no cloud dependency required.
