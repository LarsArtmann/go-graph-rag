# Metaengine / System Adoption: PRO/CONTRA Deep Research

|          |                                                                                                                                                                                                                                                                                                                                                     |
| -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Date     | 2026-09-15                                                                                                                                                                                                                                                                                                                                          |
| Question | Should `go-graph-rag` adopt `go-cqrs-lite/metaengine` and/or `go-cqrs-lite/system` (as store, as query layer, or at all)?                                                                                                                                                                                                                           |
| Method   | Primary sources only: both module trees read file-by-file (`metaengine/`, `system/`, engine subpackages), `go.mod` dependency analysis, repo `FEATURES.md`/`ROADMAP.md`/ADR-0123/ADR-0135 status, git tags, plus CV-side SUPERB plan context. Every claim below carries a file:line or doc citation.                                                |
| Verdict  | **Adopt neither into the SDK. `system` is a category error for a library; `metaengine` is capable but wrong-shaped (event-fed read-model planner, brute-force vectors, label/weight-less graph edges, breaks the dependency-light non-negotiable). Adoption belongs at the CV application layer, where SUPERB already plans it with proper gates.** |

## 1. TL;DR

Both modules are the author's own, actively maintained, and impressive: `metaengine` (v4.13.0) is a
cost-based storage planner with 10 ADTs — including Vector and Graph — across 10 engine drivers,
backed by ≥355 test files (soak to 10M events, fuzz, race, restart safety). `system` (v4.7.0) is the
deployer-driven composition root that wires deciders, dispatchers, projections, and engines from a
declarative `DomainConfig` + `DeploymentConfig` (with `cqrs.yaml` support). ADR-0123 designates this
pair as the **only supported v5 path** — the old `stack/*` presets are deprecated and die at v5.

Neither solves a problem `go-graph-rag` has:

- The SDK's write path is `Build(documents, edges)` — bulk, caller-authored corpus. metaengine's
  write path is `Apply(eventType, payload)` dispatched through per-query **folds**. Using it means
  faking events to satisfy an event-sourcing abstraction the SDK does not have.
- metaengine's vector search is **brute-force O(N·D) in every shipped engine** (guidance: <10K
  vectors; `vector_search.go:152-156`). The SDK already does a linear cosine scan in ~120 lines
  (`vector.go`). The roadmap's actual scale need — ANN/HNSW (`sqlite-vec`, `hannoy`) — is explicitly
  NOT provided by metaengine; adopting it buys zero scale headroom.
- metaengine's graph edges are `Edge{From, To}` — **no relation labels, no weights**
  (`types.go:45-48`). The SDK's `Edge{Source, Target, Relation, Weight}` plus derived
  `RelationSimilar` edges are the core of hybrid ranking; they would have to be side-stored in Maps.
- `system` is an **application** composition root (owns lifecycle, buses, projectionhost, config
  loading). A library must not own its consumer's composition root — that inverts control and forces
  every consumer's infrastructure choices.
- Both are marked **🧪 Experimental** in the repo's `FEATURES.md` (`:1442-1464`), and v4→v5 churn is
  observable today (`On`/`OnTyped` removal-pending, `RecordAwareFold` compat shim after a data-race
  fix). A v0.x SDK must not pin an experimental planner.

The rational path: the SDK stays hand-rolled and dependency-light. When CV adopts
`system`+`metaengine` for its evented funnel core (SUPERB plan, T01/T29), `go-graph-rag` integrates
at the **application layer** — a query handler plus a rebuild command — with zero SDK dependency
changes. Revisit SDK-layer adoption only at the documented triggers (§9).

## 2. What each module actually is (verified snapshot)

### 2.1 `metaengine/v4` — cost-based storage planner for event-sourced data

| Fact            | Evidence                                                                                                                                                                                             |
| --------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Version         | v4.13.0 (14 releases v4.0.0→v4.13.0; git tags)                                                                                                                                                       |
| Status          | 🧪 Experimental — repo `FEATURES.md:1442-1463` (root + every engine submodule)                                                                                                                       |
| Core model      | Developer declares `Query[Q,R]` + fold functions; operator provides `Engine`s with cost profiles; `Plan()` assigns queries to engines (`planner.go:138-210`)                                         |
| Write path      | `Store.Apply(ctx, eventType, payload)` dispatches all matching folds (`store.go:416-418`); full context via `ApplyRecord` (`:444-459`)                                                               |
| ADTs            | Map / Set / Counter / Graph / Multimap / Log / Scan / Vector / Search / Spatial — the fold return type IS the ADT (`README.md:185-200`)                                                              |
| Engines (10)    | memory (built-in), sqlite, turso, postgres, mysql, pebble, bbolt, badger, duckdb, dgraph — database/sql-style blank-import registration (`register.go:11-31`, ADR-0123 §3)                           |
| Durability      | strict / normal / relaxed tiers; engines that cannot honor a tier must reject it (`durability.go:21-93`); surfaced in Doctor                                                                         |
| Observability   | Doctor, ExplainPlan, health quarantine/failover (ADR-0137), engine stats, `otelobserver` counters                                                                                                    |
| SSE             | `ServeSSE[V]` / `Watcher[V]` — materialized read-model values to browsers, in-memory ring replay (`sse.go`, skill ADR-0091 table)                                                                    |
| Atomicity       | Per-engine `Transactional.RunInTx`; **cross-engine 2PC NOT supported** (`store.go:461-468`)                                                                                                          |
| Test mass       | 155 test files in root package alone, ≥355 across subpackages; 10M-event soak with heap assertions, fold-classifier fuzz, catch-up stress, restart/concurrency idempotency                           |
| v5 churn        | `On`/`OnTyped` deprecated→removed v5 (`fold.go:293-309`); `NsPerRead` (`engine.go:34-37`); `RecordAwareFold` compat shim after Record-by-value race fix (`record_fold.go:14-22`)                     |
| Deps (consumer) | direct: `dedup`, `id`, `record`, `metaengine/sqliteengine`, `go-error-family`, `go-sse`; transitive: `branded-id`, `ulid/v2`, `modernc.org/sqlite` + libc chain — ~10 modules vs the SDK's current 3 |

### 2.2 Vector capability (the overlap that matters)

- **Brute force everywhere shipped.** Memory index: "computes distances on every search —
  O(N·D) per query… Suitable for small collections (<10K vectors)… for production scale, use an
  engine with ANN search (HNSW, PQ)" (`vector_search.go:152-156`). sqliteengine: pure-Go path
  scores in Go, libSQL path pushes k-NN into SQL via `vector_distance_*` — "same O(N) complexity"
  (`sqliteengine/vector.go:12-27`). The `VectorBackend` interface _permits_ HNSW/PQ implementations
  (`:59-61`); none exist.
- **Metrics:** cosine, dot, euclidean (`vector_search.go:283-294`); dot negated so ascending sort =
  nearest-first (`:248-254`). The SDK needs cosine only (`vector.go`).
- **Shape:** `Embedding{ID string; Values []float32; Metadata map[string]any}`
  (`vector_search.go:23-27`); collection is a per-call argument. Filtered search exists and
  pre-filters BEFORE ranking (`VectorFilterBackend`, `:84-98`).
- **Performance insight already absorbed:** the binary-payload codec exists because JSON decode was
  "~17us/vector on pebble vs the ~90ns in-RAM ceiling" (`vector_binary.go:10-24`). The SDK already
  stores vectors as BLOBs (`store.go:27`) — same lesson, no dependency needed.

### 2.3 Graph capability

- Operations: `GraphAddEdge` (idempotent), `GraphRemoveEdge` (tombstone, idempotent), directed
  `GraphNeighbors(node, depth)` (depth-limited BFS) and `GraphNeighborsUndirected`
  (`memory_graph.go:13-156`; `execute.go:218,414-425`).
- sqliteengine implements neighbors via **recursive CTEs** with a driver probe + iterative-BFS
  fallback (`sqliteengine/graph.go:11-59`); add is `INSERT OR IGNORE` (`engine.go:152`).
- **Edges carry no relation label and no weight** — `Edge{From, To any}` (`types.go:45-48`).
  Engines without native graph degrade to multimap BFS, O(N·degree^depth), with a planner warning
  (`graph_fallback.go:8-36`).

### 2.4 `system/v4` — deployer-driven composition root

| Fact            | Evidence                                                                                                                                                                                                                                                  |
| --------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Version         | v4.7.0 (git tags)                                                                                                                                                                                                                                         |
| Status          | 🧪 Experimental — `FEATURES.md:1464`; notably NOT marked experimental inside the module itself (grep: zero hits in `system/*.go` + README) — doc inconsistency worth knowing                                                                              |
| Core model      | Consumer declares `DomainConfig` (Commands/Queries/Projections/Evolutions/coeffect gate, `config_types.go:19-114`); operator declares `DeploymentConfig` (Engines/Instances/Buses/durability/priority, `:127-159`); `system.New` wires everything         |
| Config loading  | `cqrs.yaml` via koanf + `CQRS_`-prefixed env overrides (`config_loader.go:14-107`); operators swap engines/DSNs/pragmas/durability without recompiling                                                                                                    |
| Driver registry | Lives in metaengine; `system` bridges via `metaengine.LookupDriver` (`driver_registry.go:10-41`); unknown driver names fail at construction                                                                                                               |
| Roles           | `RoleSourceOfTruth` / `RoleProjections` / commands/queries/snapshots per-engine binding; duplicate roles error (`roles.go`, `config_types.go:325-391`)                                                                                                    |
| Known caveats   | Bus drivers: only `gochannel` in-process supported (README `:243-245`); durability-conflict rule (`:422-425`); EventAdapter.Save atomicity ladder — AtomicAppender→Transactional→**racy** fallback ("do NOT rely on it under concurrency", `doc.go:1-31`) |
| Test mass       | 3 core tests + integration suites (sqlite lifecycle, badger, postgres env-gated, shutdown ordering) — far thinner than metaengine's                                                                                                                       |
| Deps            | ~21 direct requires (koanf, go-codec, watermill, pebble, badger, pgx, otter, 12 go-cqrs-lite modules…); indirect pulls otel, prometheus client, sentry, cbor — the largest dep tree in the ecosystem                                                      |

### 2.5 Ecosystem trajectory (ADR-0123, ADR-0135)

- **ADR-0123 (Proposed, 2026-08-09):** v5 unification — `system.System` becomes the single
  composition root; `stack.Bundle` and all `stack/*` presets are deleted; the driver registry moves
  fully into metaengine with blank-import self-registration; `OnRecord`/`OnRecordTyped` become the
  only fold constructors. Implication: anyone building on the old tiers must migrate to
  system+metaengine by v5 — this is why CV's SUPERB plan targets them.
- **ADR-0135 (Accepted, 2026-09-07):** Turso materialized views — scalar aggregates exact; grouped
  views carry an upstream correctness caveat (silently wrong SUMs after a second transaction).

## 3. Fit against `go-graph-rag`'s actual shape

| Dimension         | `go-graph-rag` today                                                                                        | metaengine offers                                                                                                          | Fit                                                                                 |
| ----------------- | ----------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| Write model       | `Build(ctx, provider, cache, docs, edges)` — bulk corpus, full rebuild in one tx (`build.go`, `store.go`)   | `Apply(eventType, payload)` through per-query folds; events are the source of truth                                        | ✗ conceptual inversion — would require fake events (`DocumentIndexed`, `EdgeAdded`) |
| Vector search     | In-memory linear cosine scan over embedded nodes (~120 LOC, `vector.go` + `search.go`)                      | Brute-force O(N·D) in all shipped engines; <10K guidance; cosine/dot/euclidean; pre-filtering                              | ≈ parity only — no gain, more deps                                                  |
| Scale path        | Roadmap: ANN/HNSW (`sqlite-vec`, `hannoy`) behind the `Searcher` seam; ~50k-node trigger (`ROADMAP.md`)     | No ANN engine shipped; `VectorBackend` permits one in future                                                               | ✗ does not unlock the roadmap                                                       |
| Graph edges       | `Edge{Source, Target, Relation, Weight}` + derived `RelationSimilar`; relation vocabularies are caller-side | `Edge{From, To}` — no label, no weight                                                                                     | ✗ loses typed/weighted semantics; metadata would flee to side Maps                  |
| Graph reads       | Neighborhood expansion fused with two-tier ranking + context rendering (`search.go`)                        | `GraphNeighbors(node, depth)` BFS / recursive CTEs                                                                         | ≈ primitive exists, composition missing — planner has no expand-and-rank notion     |
| Persistence       | One SQLite file, rebuildable derived index, single writer, `synchronous=OFF` (`store.go:19-47`)             | 10 engines, durability tiers, health/failover, quarantine, Doctor                                                          | ✓ genuinely better ops story — for problems the SDK doesn't have yet                |
| Embedding cache   | Namespaced `content_hash × provider × model` BLOB table (`store.go:22-30`)                                  | Map ADT could hold it; binary codec lesson already absorbed                                                                | ✗ trivial gain                                                                      |
| Dependency policy | stdlib + `samber/lo` + `modernc.org/sqlite` — README-marketed non-negotiable                                | ~10 consumer-visible modules (cqrs-lite core, go-sse, go-error-family, …); system: 30+ with otel/prometheus/sentry         | ✗ breaks non-negotiable #3                                                          |
| Consumers         | Arbitrary public-SDK consumers (CV today, more later); CV consumes via module proxy                         | Designed for event-sourced applications; PROPRIETARY license (same author — fine for Lars, a constraint for third parties) | ≈                                                                                   |

## 4. PRO adopting `metaengine` (steelmanned)

1. **Vector + Graph + Search ADTs exist and are tested.** Filtered k-NN with pre-ranking filters,
   per-collection namespacing, depth-limited neighbor queries with SQLite recursive-CTE
   implementations — real, working overlap with the SDK's two cores.
2. **Ten engines behind one interface.** memory for tests, sqlite (already a dep), pebble (~7x
   faster MapGet per module docs), postgres/mysql/turso (encrypted, materialized views), duckdb,
   dgraph, badger, bbolt. Engine swaps without API change — a stronger version of the SDK's
   "pluggable backend someday" idea.
3. **Operational maturity the SDK will not build for years.** Doctor, ExplainPlan, health-driven
   quarantine and failover (ADR-0137), durability tiers that refuse to lie, SSE `Watcher` streaming,
   restart-idempotency and 10M-event soak tests.
4. **Same author, same toolchain (go 1.26.7), same house style.** No supply-chain trust gap, no
   version-skew drama, mutual fix turnaround is one Slack message to yourself.
5. **v5 alignment.** ADR-0123 makes metaengine the center of the v5 read-model world; adopting early
   would ride the gradient instead of fighting it.
6. **CV synergy.** If CV's SUPERB core lands on system+metaengine, an SDK that speaks metaengine
   natively could project its graph into CV's read-model world (SSE-updating search results,
   watchers) with no app-side glue.

## 5. CONTRA adopting `metaengine`

1. **Breaks the dependency non-negotiable (AGENTS.md #3, README "Design constraints").**
   "Dependency-light" is a marketed position of a freshly published public SDK, not an accident.
   metaengine drags cqrs-lite core modules (`record`, `id`, `dedup`), `go-error-family`, `go-sse`,
   and their transitive tree into every consumer's `go.sum` — including consumers that will never
   run an event-sourced system.
2. **Conceptual inversion.** metaengine's write path IS its design: folds over events, cost-planned
   per query. `Build(docs, edges)` has no event stream. You would either fake events (paying the
   abstraction and getting replay/catch-up machinery you cannot use — the store is deliberately a
   rebuildable derived index) or bypass folds and use engines directly (paying the module graph and
   getting ~an alternative SQLite schema).
3. **No ANN — the one thing the roadmap actually needs.** Every shipped vector path is O(N·D); the
   module's own guidance caps comfortable use at <10K vectors. The SDK already has the same
   complexity in less code. At the ~50k-node revisit point, metaengine is not the answer;
   `sqlite-vec`/`hannoy` behind the existing `Searcher` seam is.
4. **Graph ADT is semantically poorer than the SDK's needs.** No relation labels, no weights; the
   typed-edge + `RelationSimilar` + two-tier ranking pipeline would straddle two storage models
   (meta_graph_edges + side Maps), doubling write paths and losing single-edge atomicity.
5. **Experimental status + observable churn.** FEATURES.md marks it 🧪; v5 will remove `On`/`OnTyped`
   and reshape the driver registry (ADR-0123 Proposed); `RecordAwareFold` already needed a compat
   shim after a concurrency fix. A v0.x SDK stacking on an experimental v4 planner compounds two
   instability axes for its consumers.
6. **Correlated failure domain.** Same-author mono-ecosystem means one metaengine regression
   simultaneously hits CV's read models AND the SDK's search path. Today a go-graph-rag bug cannot
   come from go-cqrs-lite; adopting couples their release trains (the ecosystem-upgrade skill exists
   precisely because these sweeps are work).
7. **License perimeter.** Both PROPRIETARY. Harmless for Lars-owned consumption, but every future
   third-party `go get github.com/larsartmann/go-graph-rag` would inherit a proprietary transitive
   dependency they must evaluate — shrinking the public SDK's addressable audience.

## 6. PRO/CONTRA adopting `system`

**CON (for the SDK — near-categorical):**

1. **A library must not own the composition root.** `system.New` constructs dispatchers, buses,
   projectionhost, lifecycle, and config loading. If the SDK imported it, every consumer's
   infrastructure choices (engines, buses, durability, cqrs.yaml) would be shaped by a retrieval
   library. Control flow inverts: the app composes libraries, never the reverse.
2. **Largest dependency tree in the ecosystem** (watermill, pebble, badger, pgx, koanf, otel,
   prometheus, sentry indirect) — an order-of-magnitude violation of the dependency rule.
3. **Thinnest test mass of the candidate modules** (3 core tests + integrations) and the
   EventAdapter.Save racy-fallback caveat (`doc.go`) — the module itself says do not rely on Save
   under concurrency on third-party engines.
4. **No retrieval-relevant capability.** system wires CQRS topology; it contains nothing for
   embeddings, similarity, or graph ranking. There is no feature to want.

**PRO (for the CV application — where it actually belongs):**

1. Declared the v5-only composition root (ADR-0123); building new app wiring on `stack.Bundle` today
   is building on the deprecated path.
2. Domain/deployment separation is genuinely good design for an app: CV's operators could swap
   memory→sqlite→postgres engines via `cqrs.yaml` without touching domain code.
3. CV's SUPERB plan already scoped the honest adoption path: T01 (`system.New` spike over a store
   copy, Go/No-Go ADR) and T29 (read-model tier decision spike: metaengine vs `SQLViewStore` vs
   hand-rolled, WITH the v5-removal check) — with risk R3 explicitly naming "read-model tier chosen
   wrong" as a rewrite-of-the-rewrite hazard.

## 7. Decision matrix (the "and/or")

| Option                      | Verdict                 | Rationale                                                                                                                                                                      |
| --------------------------- | ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Neither (status quo)**    | ✅ KEEP                 | SDK stays dependency-light, stable-seamed, roadmap-aligned (ANN behind `Searcher`). CV adopts at app layer per SUPERB.                                                         |
| **metaengine only**         | ✗ REJECT                | Wrong write model, no ANN, poorer edges, dep-policy break, experimental churn. Buys ops maturity for problems not yet had.                                                     |
| **system only**             | ✗ REJECT                | Category error for a library; nothing retrieval-relevant inside.                                                                                                               |
| **Both**                    | ✗ REJECT                | Worst of both: system requires the metaengine world anyway; maximal dep tree; maximal coupling.                                                                                |
| **App-layer adoption (CV)** | ✅ PROCEED (as planned) | SUPERB T01/T29 gates, store-copy spike, tier decision matrix on measured read patterns — exactly how adoption should be evaluated. `go-graph-rag` needs zero changes for this. |

## 8. The middle paths (most of the value, none of the dependency)

1. **App-layer integration when CV adopts system+metaengine.** Register a CV query handler that
   calls `Searcher.Search`; run `Build` from a command/cron against the CV store copy; optionally
   project search results into a metaengine collection for `ServeSSE` live dashboards. The SDK's
   `Cache`/`Store` seams already support this (CV's `internal/features/graphrag/service` is the
   existing mounting point).
2. **ANN behind the existing seam.** When the ~50k-node trigger fires, implement `sqlite-vec` or
   `hannoy` behind `Searcher` — independent of metaengine, matching ROADMAP Theme 1.
3. **Isolated adapter module, if ever demanded.** If a second consumer explicitly wants metaengine
   projection of the graph, publish `go-graph-rag/metaengine` as a SEPARATE module (the
   ADR-0086-family dep-isolation pattern: optional module, optional dependency, core untouched).
   This is the ecosystem's own answer to heavy deps (pebbleengine, duckdbengine are separate for
   exactly this reason).
4. **Cross-pollination, already banked.** Binary vector payloads (BLOB not JSON), namespaced cache
   keys, single-writer SQLite discipline — the SDK already embodies metaengine's hard-won lessons
   without importing the machinery.

## 9. Revisit triggers (what would flip the verdict)

| Trigger                                                                                                     | Flip                                                |
| ----------------------------------------------------------------------------------------------------------- | --------------------------------------------------- |
| metaengine ships an ANN-capable engine (HNSW/PQ) behind `VectorBackend`                                     | Reconsider engine delegation at the scale trigger   |
| go-graph-rag needs multi-engine persistence (pebble/pg) with one API and cannot afford its own backend seam | Reconsider metaengine-as-store                      |
| A consumer demands live SSE search results / watchable graph collections from the SDK itself                | Consider the isolated adapter module (§8.3)         |
| metaengine + system graduate from Experimental and v5 lands (ADR-0123 executed, fold churn settled)         | Re-run this evaluation on stable ground             |
| The SDK becomes event-fed (documents arrive as a domain event stream rather than bulk `Build`)              | The fold model starts to fit — re-evaluate honestly |

## 10. Evidence appendix

- Versions/status: git tags `metaengine/v4.13.0`, `system/v4.7.0`; `FEATURES.md:1442-1464` (🧪 both);
  `ROADMAP.md:396-415,494` (system shipped-list, v5 plan), `:590-598` (FoundationDB/Scylla ideas).
- Vector: `metaengine/vector_search.go:23-27,59-61,84-98,152-156,248-254,283-294`;
  `metaengine/sqliteengine/vector.go:12-27,61-70,157-169`; `metaengine/vector_binary.go:10-24`.
- Graph: `metaengine/types.go:39-48,58-61`; `metaengine/memory_graph.go:13-156`;
  `metaengine/sqliteengine/graph.go:11-59`, `graph_undirected.go:19-37`;
  `metaengine/sqliteengine/engine.go:152`; `metaengine/graph_fallback.go:8-36`;
  `metaengine/graph_cte_e2e_test.go:13-63`; `metaengine/store.go:789-815` (removal, no fallback).
- Store/planner: `metaengine/planner.go:10-16,62-68,101-120,138-210,292-303`;
  `metaengine/store.go:416-468`; `metaengine/durability.go:21-93`; `metaengine/execute.go:55-98`.
- Churn: `metaengine/fold.go:293-309`; `metaengine/engine.go:34-37`;
  `metaengine/record_fold.go:14-22`.
- Maturity: `soak_10m_test.go:15-35`; `fuzz_test.go:7-35`; `catchup_stress_test.go:39-50`;
  `restart_test.go:15-60`; 155 root test files (glob count), ≥355 across subpackages;
  `COOKBOOK.md`; `MIGRATION.md:6-90`.
- system: `system/README.md:5,243-245,422-435`; `system/doc.go:1-31`;
  `system/config_types.go:19-114,127-159,325-391`; `system/config_loader.go:14-107`;
  `system/driver_registry.go:10-41`; `system/system_test.go:74,165,220`.
- ADRs: `docs/adr/0123-v5-unification-single-composition-root.md` (Proposed);
  `docs/adr/0135-materialized-views-operator-option.md` (Accepted).
- go-graph-rag: `store.go:19-47` (schema, synchronous=OFF, single conn); `vector.go` (cosine);
  `search.go` (hybrid + two-tier); `ROADMAP.md` Theme 1 (ANN candidates, ~50k trigger);
  `README.md:90-99` (dependency-light, full-rebuild, one-writer constraints); `go.mod` (3 deps).
- CV context: `CV/go.mod` (direct: event/v4, id/v4, graph-rag v0.1.0);
  `CV/docs/planning/2026-09-14_19-13_SUPERB-evented-funnel-core-pareto-plan.md`
  (modules 2→≥8 row `:290`, T01 `:89`, T29 `:88`, R3 `:336`, v5-path row `:385`);
  `CV/docs/status/2026-09-15_17-09_graphrag-sdk-extraction-status.md` (SDK extraction context).

---

_Annotated with dependency graphs? No — this is a decision record. Re-run on the §9 triggers._
