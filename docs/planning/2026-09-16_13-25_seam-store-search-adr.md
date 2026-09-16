# ADR — The store/search seam: pluggable persistence and vector ranking

|          |                                                                                                                                                          |
| -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Date     | 2026-09-16                                                                                                                                               |
| Status   | Proposed → binding design for all future storage work (owner decision 2026-09-16: ADR + Go interface sketch BEFORE any code)                             |
| Question | How do embedded ANN libraries, metaengine engines, and raw server databases plug into `go-graph-rag` without ever forcing a breaking SDK change?         |
| Decision | Two small core interfaces — `VectorIndex` (ranking) and `GraphStore` (persistence) — introduced as ADDITIONS in v0.x, with the current linear scan and   |
|          | SQLite store as the reference implementations. Every backend lands in a separate adapter module. Core `go.mod` never grows. Core stays CGO-free.         |
| Method   | Seam surfaces re-read file-by-file (store.go, search.go, embed.go, build.go, graph.go); dependency trees measured in a scratch consumer module;          |
|          | semantics taken from the two research reports (cited inline). No seam code enters core in this ADR.                                                      |

## 1. Context

The SDK today is deliberately simple: `Build` embeds a caller-authored corpus, derives
`RelationSimilar` edges, and persists nodes/edges/vectors into one rebuildable SQLite file
(`store.go`). Search warm-starts an immutable in-memory `Searcher` from a snapshot and answers
queries with a linear cosine scan plus 1-hop graph expansion (`search.go`).

Two research evaluations (2026-09-15) converged on the same structural conclusion:

- `docs/research/2026-09-15_dgraph-adoption.md` — raw server databases are rejected for the SDK
  but viable as a CONSUMER-owned backend; its scenario B ("optional pluggable backend behind a
  Store/search seam — LATER") is the insurance this ADR buys. Today `Cache` is the only seam; the
  graph persistence and the ranking scan are hard-wired.
- `docs/research/2026-09-15_metaengine-system-adoption.md` — metaengine's uniform engine API is
  real but brute-force on vectors and label/weight-less on edges; adoption belongs at an
  application layer, or in an ISOLATED adapter module if a second consumer ever demands it
  (§8.3 of that report).

The measured scale horizon: Build/SimilarPairs scale quadratically, ~373ms per 1000-doc build,
~15min extrapolated at 50k nodes (`bench_test.go`). The ~50k trigger is far away (owner: corpus
"NO" until ~2027-09), which is exactly why the seam must be designed NOW and implemented later:
designing it at leisure is cheap; retrofitting it under scale pressure is how breaking changes
happen.

## 2. Decision drivers

1. **Dependency non-negotiable (AGENTS.md #3).** stdlib + `samber/lo` + `modernc.org/sqlite`, and
   that is the permanent core budget. Measured cost of breaking it: §6.
2. **CGO-free core.** `modernc.org/sqlite` keeps the build pure-Go. ANN libraries are the one
   place CGO tempts (sqlite-vec), so CGO is acceptable in OPTIONAL ADAPTER MODULES ONLY, never in
   core and never in a module a consumer pulls transitively by default.
3. **Rebuildable derived index.** Whatever persists, the SDK's promise stays: losing the store
   costs a full rebuild, and embeddings cached by content hash make that cheap. Backends that
   cannot honor full-replace semantics must say so at construction.
4. **Deterministic rendering and ranking policy are core IP.** Two-tier document/hub ranking,
   `SearcherOptions` roles, `RelationSimilar` semantics, and byte-stable context rendering
   (`RenderContext`) remain SDK-owned for every backend. Backends replace the SCAN, not the
   policy.
5. **Three backend classes must fit the SAME seam** (§4): they disagree on scores, filters, and
   metric coupling, so the seam must make those disagreements explicit instead of hiding them.

## 3. Extension points (re-read 2026-09-16)

| Surface                | Today                                                                                       | Seam                                       |
| ---------------------- | ------------------------------------------------------------------------------------------- | ------------------------------------------ |
| Embedding cache        | `Cache` interface (`store.go:60-66`), `Store` implements                                     | Unchanged — already pluggable              |
| Embedding provider     | `Provider` interface (`embed.go:32-43`), `NewProvider` factory                               | Unchanged — already pluggable              |
| Graph persistence      | Concrete `*Store` (`store.go:73`): `ReplaceGraph`/`LoadGraph`/`LoadEmbeddings`/`Stats`       | `GraphStore` interface (sketch §5) — `*Store` already satisfies it |
| Vector ranking         | `rankHits` linear scan over `map[string]Vector` (`search.go:239`)                            | `VectorIndex` interface (sketch §5)        |
| Searcher warm start    | `NewSearcher`/`NewSearcherWithOptions` over nodes+edges+vectors (`search.go:99-133`)         | Index-backed constructor added later, non-breaking |
| Vocabulary and wire    | `Node`/`Edge`/`Graph`/`Hit`/`SearchResult` (`graph.go`, `search.go`)                         | Never delegated — frozen at v1             |

## 4. The three backend classes

### 4.1 Class A — embedded SQLite + in-process ANN libraries

Candidates from ROADMAP Theme 1: `sqlite-vec`, `hannoy` (pure-Go HNSW), an in-process HNSW port.

| Property            | Semantics                                                                                                                                                            |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ANN                 | Native. This is the only class that actually unlocks the ~50k trigger.                                                                                                |
| Scores              | Returned per hit (cosine distance by the index; converted to cosine SIMILARITY at the adapter boundary so core never sees a distance). `VectorMatch.Scored = true`.  |
| Metric coupling     | Chosen at index creation, frozen for the index's lifetime. Adapter must declare it (`Metric()`) and refuse `Replace`/`Search` on mismatch with the core's expectation. |
| Filters             | Split by sub-candidate: `sqlite-vec` pre-filters in SQL (`WHERE` before the scan — exact candidate sets); HNSW-family libs post-filter after the graph walk — selective filters under post-filtering cost recall and MUST be mitigated by overfetch (fetch k', filter, trim to k). The adapter documents which semantics it implements; the core applies `MinScore` after the backend returns regardless. |
| Ops footprint       | Same process, same file model as today (vectors as BLOBs, index rebuilt on load or extension-managed). No new deployment story.                                        |
| CGO                 | `sqlite-vec` typically builds through CGO or a precompiled extension → adapter module carries it; `hannoy`/HNSW-port stay pure Go. Core never does.                    |

### 4.2 Class B — metaengine engines (adapter = projection, not replacement)

| Property        | Semantics                                                                                                                                                              |
| --------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ANN             | None shipped. Every engine's vector path is brute-force O(N·D) with a documented <10K-vector comfort cap (`vector_search.go:152-156`); `VectorBackend` PERMITS HNSW/PQ but no engine implements one. Adopting metaengine buys zero scale headroom over the core's own scan — same complexity, more code. |
| Scores          | Computed by the engine (cosine/dot/euclidean per query); `Scored = true`. Core accepts cosine only and rejects other metrics at the adapter boundary.                   |
| Metric coupling | Per query — the loosest of the three classes.                                                                                                                           |
| Filters         | `VectorFilterBackend` pre-filters BEFORE ranking — good semantics, exists today.                                                                                        |
| Graph model     | `Edge{From, To}` — no relation labels, no weights (`types.go:45-48`). The SDK's `Edge{Source, Target, Relation, Weight}` + derived `RelationSimilar` are the core of hybrid ranking and CANNOT be expressed; they would flee to side Maps, doubling write paths and losing single-edge atomicity. |
| Consequence     | An adapter is a PROJECTION: the SDK graph/types stay the source of truth; metaengine mirrors vector + neighbor lookups for consumers already living in that world (CV app-layer per SUPERB T01/T29). It never becomes the SDK's store. |
| Deps            | Measured in §6: +27 modules into every consumer graph for metaengine + sqliteengine alone.                                                                              |

### 4.3 Class C — raw server databases (Dgraph exemplar)

| Property        | Semantics                                                                                                                                                              |
| --------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ANN             | Native HNSW server-side (`@index(hnsw(metric:"cosine", ...))` since v24). Tunables (`exponent`, `efConstruction`, `efSearch`) are adapter/consumer domain.             |
| Metric coupling | Coupled in the SCHEMA at index time — changing metric means a schema migration. Declared via `Metric()`, enforced at construction.                                       |
| Scores          | Path-dependent: the GraphQL kNN surface returns `vector_distance`, but DQL `similar_to` returns matching uids ordered by distance without per-hit distances — uidS-WITHOUT-DISTANCES. The seam therefore allows `Scored = false`; the core then re-scores candidates by cosine over the loaded snapshot vectors before ranking. |
| Filters         | Server-side composition of metadata predicates with kNN was NOT verified in the 2026-09-15 research — an explicit spike question before any adapter is written.         |
| Concurrency     | Concurrent writers meet `ErrAborted` transaction retries; the SDK's one-writer model does not surface this today, so the adapter owns a retry policy.                   |
| Ops footprint   | Categorically different: 2 processes minimum (Zero + Alpha), prod 3+3, Linux-only, Alpha sized 8+ cores/16GB+/3000+ IOPS, client majors hard-coupled to server majors (`dgo/v240`, `dgo/v250`). CONSUMER-owned deployment (dgraph report scenario C); the SDK ships at most a pattern, not a default. |
| Non-negotiable  | `dgo` (gRPC + protobuf) NEVER enters core `go.mod` — see §6.                                                                                                            |

## 5. Interface sketch (binding shape, v0.x additions)

This sketch compiles against `go-graph-rag v0.1.0` in a scratch module (verified 2026-09-16,
`env -u GOEXPERIMENT go build`); it is a DESIGN ARTIFACT — none of it is merged into core yet.

```go
package graphrag

import "context"

// DeclaredMetric names the similarity a vector backend ranks by. The core
// speaks cosine only today; the constant makes mismatch a construction-time
// error instead of a silent misranking.
type DeclaredMetric string

// MetricCosine is the only metric the SDK's ranking policy is defined on.
const MetricCosine DeclaredMetric = "cosine"

// VectorMatch is one ranked candidate returned by a VectorIndex.
type VectorMatch struct {
	// NodeID is the node the backend matched.
	NodeID string
	// Scored reports whether the backend returned a per-hit similarity.
	// Class C backends may return ids without distances (DQL similar_to);
	// for those the core re-scores over the loaded snapshot vectors.
	Scored bool
	// Score is the cosine similarity when Scored is true; undefined
	// otherwise. Adapters convert native DISTANCES to similarity at the
	// boundary so the core never handles distances.
	Score float64
}

// VectorQuery is one ranking request.
type VectorQuery struct {
	// Vector is the query embedding (same space as the index contents).
	Vector Vector
	// TopK bounds the result count.
	TopK int
	// MinScore drops candidates below the similarity floor. Backends MAY
	// apply it natively; the core applies it after the backend returns
	// regardless, so a backend cannot widen results by ignoring it.
	MinScore float64
	// Filter optionally restricts candidates. It receives the full node so
	// kind- and attr-level predicates stay possible without a breaking
	// signature change later. Nil means no filter. Whether a backend
	// pre- or post-filters is adapter-documented (§4.1); ROADMAP Theme 1
	// refines filtered-ANN semantics before any adapter ships.
	Filter func(Node) bool
}

// VectorIndex is the ranking seam behind Searcher hit collection. It is the
// ANN integration point: implementing it is the ONLY thing a vector backend
// must do. Implementations must be safe for concurrent Search calls.
type VectorIndex interface {
	// Metric declares the similarity this index ranks by. Core rejects
	// indexes whose metric is not MetricCosine.
	Metric() DeclaredMetric

	// Replace atomically swaps the full index contents. This is the
	// full-rebuild model: the same discipline as Store.ReplaceGraph. A
	// backend that cannot honor full-replace must fail here, not degrade.
	Replace(vectors map[string]Vector) error

	// Search returns up to q.TopK candidates, best first. Results MAY
	// exceed MinScore filtering duties (the core re-applies it) but MUST
	// honor Filter.
	Search(ctx context.Context, q VectorQuery) ([]VectorMatch, error)

	// Close releases backend resources (statements, mmaps, connections).
	Close() error
}

// GraphStore is the persistence seam: the read/write surface a backend must
// cover for Build/persist/LoadGraph round trips. *Store already satisfies
// it; extracting the interface is a later, non-breaking core change.
type GraphStore interface {
	// ReplaceGraph atomically swaps the persisted graph (delete-then-insert
	// inside one transaction today).
	ReplaceGraph(nodes []Node, embeddingHashes map[string]string, edges []Edge) error

	// LoadGraph reads the full snapshot: nodes, edges, and the
	// node-to-embedding-hash join keys.
	LoadGraph() (*LoadGraphSnapshot, error)

	// LoadEmbeddings returns every cached vector of one provider/model
	// namespace keyed by content hash.
	LoadEmbeddings(provider, model string) (map[string]Vector, error)

	// Stats summarizes the persisted index for humans and endpoints.
	Stats() (StoreStats, error)
}
```

### 5.1 What core exports vs what adapters own

| Core (this module) owns                                            | Adapter modules own                                                        |
| ------------------------------------------------------------------ | -------------------------------------------------------------------------- |
| `VectorIndex`/`GraphStore`/`VectorQuery`/`VectorMatch`/`DeclaredMetric` types | Construction: `Open*` functions with backend-specific knobs (ef, metric params, DSNs) |
| The reference `VectorIndex`: today's linear scan, promoted          | Index-time metric and parameter choices, schema management (class C)        |
| Two-tier ranking, `SearcherOptions` roles, `MinScore` re-application | Pre- vs post-filter strategy + its documentation                            |
| Graph expansion, `RenderContext`, determinism guarantees            | Distance→similarity conversion; `Scored=false` handling inputs              |
| `Cache`/`Provider` seams (unchanged), SQLite `*Store` as default     | Retry policies (`ErrAborted`), health checks, CGO build tags               |
| `go.mod`: the permanent 3-dependency budget                         | ALL third-party dependencies (ANN libs, `dgo`, metaengine), in THEIR go.mod |

### 5.2 Planned module layout

```
github.com/larsartmann/go-graph-rag              core: types, seams, scan index, SQLite store
github.com/larsartmann/go-graph-rag/sqlitevec    class A (optional; CGO possible)
github.com/larsartmann/go-graph-rag/hannoy       class A (optional; pure Go)  [if chosen]
github.com/larsartmann/go-graph-rag/metaengine   class B projection (optional; on demand)
class C: pattern documented, adapter consumer-side (dgo stays with the consumer)
```

Adapter modules are published only when a trigger fires (§9). A missing adapter costs nothing;
a present one never touches core go.mod (ADR-0086-family dep isolation, the same pattern the
go-cqrs-lite ecosystem uses for pebbleengine/duckdbengine).

## 6. Dependency isolation — measured

Scratch consumer module in /tmp requiring `go-graph-rag v0.1.0` from the proxy, measured with
`go list -m all` + `go mod graph` + `go.sum` line count (Go 1.26.7, 2026-09-16, module trashed
afterwards; root repo verified diff-clean):

| Consumer scenario                              | modules (`go list -m all`) | `go mod graph` edges | `go.sum` lines |
| ---------------------------------------------- | ------------------------- | -------------------- | -------------- |
| `go-graph-rag` only (status quo)               | 31                        | 77                   | 60             |
| + `metaengine/v4` + `sqliteengine/v4`          | 58 (+27)                  | 168 (+91)            | 94 (+34)       |
| + `system/v4` (brings metaengine world)        | 205 (+174)                | 1020 (+943)          | 395 (+335)     |

What the +system column pulls into every consumer graph: `cockroachdb/pebble`,
`dgraph-io/badger`, `jackc/pgx`, `knadh/koanf`, `ThreeDotsLabs/watermill` + redisstream,
OpenTelemetry (sdk/metric/trace/otelhttp), Prometheus client, `getsentry/sentry-go`,
`fxamacker/cbor`, redis client, and — via test dependencies leaking into the module graph —
testcontainers-go with the Docker client, ginkgo/gomega, go-snaps, and templ. 6.6× the module
count of the status quo, for a component whose vector path is no faster than the SDK's own scan.

This is the hard number behind metaengine report CONTRA #1 and the reason the seam puts ALL
adapter dependencies in adapter modules: a consumer who never imports an adapter never pays for
one.

## 7. CGO stance

- Core: pure Go, forever. `GOEXPERIMENT`-free, CGO-free, buildable with
  `env -u GOEXPERIMENT go build ./...`.
- Class A adapters: CGO permitted where the candidate demands it (`sqlite-vec`); pure-Go
  candidates (`hannoy`, in-process HNSW) are preferred when recall/latency parity holds.
- Class B/C adapters: no CGO involved; their cost is dependency weight and ops, not toolchain.
- A CGO adapter must never become a transitive dependency of the core module or of another
  adapter.

## 8. Consequences

**Migration path.**

1. v0.x minor release: land the §5 interfaces + promote the linear scan to the reference
   `VectorIndex`; `NewSearcher` behavior unchanged (the in-memory path stays the default).
2. Add an index-backed Searcher constructor alongside the existing ones (additive, non-breaking).
   Exact constructor shape is decided at implementation time; the §5 types are the binding part.
3. Adapters land per class, trigger-gated (§9), each in its own module with its own go.mod.
4. v1 freezes the §5 interface signatures and the wire contract (`Hit`/`SearchResult` JSON).

**Non-breaking (any time, any version):** adding the interfaces; adding adapter modules; adding
constructors; documenting filter semantics; re-scoring behavior for `Scored = false`.

**Breaking (requires the v1 gate / major bump):** making `VectorIndex` required for `Searcher`;
removing the snapshot-load path (`LoadGraph` warm start); changing §5 type field semantics after
v1; making `Filter` mandatory; changing `MinScore` re-application guarantees.

**What we consciously give up:** the illusion that "pluggable" is free. Every backend class
documented here degrades SOME property the default path has (A: CGO or recall under filters;
B: scale and edge semantics; C: ops simplicity and score availability). The seam makes the
degradation explicit at the type level instead of discovering it in production.

## 9. Decision triggers (when adapters get built)

| Trigger                                                                                    | Unlocks                                    |
| ------------------------------------------------------------------------------------------ | ------------------------------------------ |
| Sustained corpus approaching ~50k nodes, or measured build/scan pain at real size           | Class A spike (`hannoy`/`sqlite-vec`/HNSW port) |
| Shared ANN benchmark harness shows a candidate beating the scan on recall AND p50/p95       | Class A adapter module                     |
| A second SDK consumer demands metaengine projection of the graph                            | Class B adapter module (§4.2 projection)   |
| A consumer explicitly requires server-side multi-hop queries or Dgraph compatibility        | Class C pattern doc; consumer-owned adapter |
| metaengine ships an ANN engine behind `VectorBackend`                                       | Re-run the metaengine evaluation           |
| The SDK becomes event-fed (documents arrive as a domain event stream)                       | Re-evaluate the write model honestly       |

Until one fires: no adapter code, no new dependencies, core unchanged. The ROADMAP Theme 1
ideas (ANN benchmark harness, metadata-filtered ANN semantics, ANN-winner ADR) are refinements
UNDER this ADR, not alternatives to it.

## 10. References

- `docs/research/2026-09-15_dgraph-adoption.md` — class C evidence, scenario B/C, spike sketch
- `docs/research/2026-09-15_metaengine-system-adoption.md` — class B evidence, CONTRA #1
  (quantified in §6), isolated-adapter middle path (§8.3)
- `ROADMAP.md` Theme 1 — ANN candidates, ~50k trigger, filtered-ANN semantics
- `bench_test.go` — the measured numbers behind the trigger
- `store.go`, `search.go`, `embed.go`, `build.go` — the re-read seam surfaces (§3)
