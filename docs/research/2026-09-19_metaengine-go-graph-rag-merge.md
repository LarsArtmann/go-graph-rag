# Smart Merge Analysis: metaengine ⇄ go-graph-rag

|          |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Date     | 2026-09-19                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| Question | Could `go-cqrs-lite/metaengine` and `go-graph-rag` be smartly merged — move logic from metaengine into go-graph-rag, and make parts of metaengine depend on go-graph-rag?                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| Method   | Both module trees read file-by-file on 2026-09-19. Vector math compared function-by-function (`graphrag/vector.go` vs `metaengine/vector_search.go:147-242`, `vector_binary.go`), licenses read (both `LICENSE` files), module graphs read (`metaengine/go.mod`, `graphadapter/go.mod`, `dgraphengine/go.mod`), json/v2 census run (`rg -l encoding/json/v2`, 66 non-test files). Every claim carries a file:line or measurement.                                                                                                                                                                                                                                             |
| Verdict  | **No merge — in either direction — and no shared leaf module today.** The apparent duplication is ~80 lines of math that is deliberately NOT the same math (float64 similarity vs float32 distance semantics, different dim-mismatch behavior, different on-disk formats), each pinned by its own test suite. A dependency from metaengine to go-graph-rag inverts the infrastructure/application layering and freezes a v0 SDK's API as a cross-repo contract mid-metaengine-v5-churn. Moving metaengine code into the SDK imports `encoding/json/v2` (66 files; CI-red) and the `any`-typed edge model the SDK rejected. Re-evaluate a leaf module only at the §9 triggers. |

## 1. TL;DR

"Smart merge" assumes the two repos share logic worth sharing. Measured, they don't:
the shared surface is one cosine/dot/euclidean block plus a float32 blob encoder —
and both differ semantically where it matters (§3). Everything else is
complementary, not duplicated (2026-09-19 comparison doc): metaengine owns storage
ADTs and engine abstraction, the SDK owns embeddings, retrieval policy, and context
rendering. The three concrete merge shapes all price out negative today:

- **P1 — metaengine depends on go-graph-rag:** layering inversion + API freeze +
  version treadmill. Rejected.
- **P2 — move metaengine logic into go-graph-rag:** json/v2 contamination (the
  recurring-regression class that already slipped three times) + semantics conflict
  pinned by tests on both sides. Rejected.
- **P3 — third leaf module (fleet-idiomatic):** correct shape, premature now; the
  net-shared surface is ~80 lines and parity is enforced per-repo by stronger
  instruments (adttest matrix, SDK tests) than shared code would provide under
  version skew. Deferred with explicit triggers.

## 2. Decomposing "smart merge" into testable proposals

| #  | Shape                                                                                 | Dependency direction    | Module blast radius                   |
| -- | ------------------------------------------------------------------------------------- | ----------------------- | ------------------------------------- |
| P1 | Parts of metaengine (engine modules?) import `go-graph-rag`                           | infra → application     | metaengine core and/or engine modules |
| P2 | Move metaengine's vector/graph primitives into the SDK; then P1 becomes possible      | code moves into the SDK | SDK (grows), metaengine (shrinks)     |
| P3 | Extract shared math/index into a leaf module both import                              | both → leaf             | new module, both repos bump           |
| P0 | Status quo: two implementations, per-repo tests, this document records the divergence | none                    | none                                  |

Module topology facts (verified): metaengine core is its own module
(`go-cqrs-lite/metaengine/v4`, go 1.27.1); engines and adapters are separate modules
(`graphadapter/go.mod` pulls `graph/v4 v4.3.0`, `metaengine/v4 v4.13.0`,
`record/v4 v4.5.0`); the fleet already runs a many-small-modules pattern
(`go-sse`, `go-codec`, `go-branded-id`, `graph/v4`). So P1/P3 are mechanically easy.
Ease is not the problem. §3-§7 are.

## 3. The duplication inventory — it is divergence, not duplication

The entire candidate-for-moving surface, compared behavior-by-behavior:

| Function pair                             | go-graph-rag (`vector.go`)                                                                             | metaengine (`vector_search.go` / `vector_binary.go`)                                                                          | Same function?              |
| ----------------------------------------- | ------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------- | --------------------------- |
| Cosine                                    | `Cosine(a, b Vector) float64` — **similarity**, float64 accumulation (`:97-115`)                       | `cosineDistance` — **float32 distance** (`1 - sim`), float32 accumulation (`:191`)                                            | ✗                           |
| Mismatched dimensions                     | Return 0 ("no signal", `:98`)                                                                          | Truncate to shorter vector and compute (`:199-202`)                                                                           | ✗ different results         |
| Zero-magnitude input                      | Similarity 0 (`:110-112`)                                                                              | Distance 1 (same similarity outcome, different type/scale)                                                                    | ~                           |
| Dot / Euclidean                           | Not implemented (SDK ranks cosine-only)                                                                | `dotProduct` (negated for ascending sort), `euclideanDistance`, string-metric dispatch (`:178-242`)                           | ✗ SDK lacks                 |
| Top-k                                     | Policy-level inside `Searcher` (TopK + MinScore + RefMatch + expansion)                                | `TopKNearest` — bare sort + truncate (`:166`)                                                                                 | ✗ different layers          |
| Normalization                             | `Normalize()` — L2, float64 accumulation (`:39-60`)                                                    | none (cosine normalizes implicitly)                                                                                           | ✗ SDK-only                  |
| f32 blob encoding                         | `Encode`/`DecodeVector` — raw LE f32, length = bytes/4, **the SQLite cache on-disk format** (`:62-92`) | `EncodeVectorBinary` with a **uint32 dim header** + `DecodeVectorAuto` (JSON/binary/F32 sniffing) (`vector_binary.go:37-107`) | ✗ different on-disk formats |
| JSON decode                               | none (SDK never stores JSON vectors)                                                                   | `DecodeVectorJSON` — **imports `encoding/json/v2`** (`:153-163`)                                                              | ✗                           |
| Metadata filters / counters / collections | none (policy filters by NodeKind instead)                                                              | `VectorFilterBackend`, `VectorCounter` (`vector_search.go:84-118`)                                                            | ✗ ME-only                   |
| HashText / cache keys                     | SHA-256 content hash (`:119-123`)                                                                      | none (caller-managed)                                                                                                         | ✗ SDK-only                  |

Net shared surface if a merge "unified" everything: **~80 lines of accumulation
loops**. Net semantic changes required: at least five (similarity-vs-distance,
dim-mismatch, metric set, blob framing, filter model), each pinned by tests — the
adttest matrix asserts cross-engine numerical identity inside metaengine, and the
SDK's `example_test.go` output blocks plus unit tests pin its degradation semantics.
Unification breaks one side's contract in every direction. The two implementations
diverge because their domains diverge (LLM-ranking vs storage-ADT parity), not
because anyone forgot to DRY them.

## 4. P1 — make parts of metaengine depend on go-graph-rag

### PRO

- Same owner, same license (both `PROPRIETARY LICENSE`, (c) 2026 Lars Artmann) — no
  governance or licensing friction; dependency direction is license-neutral.
- Module isolation makes it surgically containable: only an engine or adapter module
  would take the dep, never metaengine core's dependency graph.
- Fleet-idiomatic — `graphadapter` already demonstrates engine-side modules
  importing sibling leaf modules (`graph/v4`).
- metaengine's `dgraphengine/graphrag_test.go` hand-rolls a GraphRAG pipeline; an
  integration module could replace the sketch with the SDK's `Searcher` and get
  policy ranking + context rendering for free.

### CONTRA

- **Layering inversion.** metaengine is infrastructure (storage ADTs, engines,
  planner); go-graph-rag is a domain application (retrieval). Infrastructure
  depending on an application SDK bakes the inversion into the fleet permanently.
  Every future "which depends on which" question answers itself wrongly.
- **API freeze at the worst time.** The SDK is v0.x with CV pinned at v0.2.0 and the
  Q1 license decision pending; metaengine is mid v4→v5 churn (ADR-0123 deprecations,
  compat shims). Making the SDK the dependency target converts its internal choices
  (`Vector` type shape, cosine semantics) into a cross-repo public contract while it
  most needs freedom to change.
- **Version treadmill.** Independent release cadences mean skew reintroduces exactly
  the drift sharing aims to kill: metaengine pins SDK v0.2, SDK ships v0.3 math or
  API changes, metaengine's consumers run stale semantics until a lockstep bump —
  the full go-ecosystem-upgrade dance (bump, pristine builds, module-proxy
  verification) forever, for ~80 lines.
- **Nothing gained.** metaengine's vector stack is a strict superset for its use
  case (metrics, filters, counters, encodings, event-fed index building). It would
  import a dependency to receive a subset of what it already has and tests-pin.
- The 2026-09-15 metaengine-adoption verdict ("adopt neither; CV app-layer
  integration per SUPERB T01/T29") already chose the correct coupling point: the CV
  application layer, where both libraries meet without either depending on the other.

**Verdict: NO.**

## 5. P2 — move metaengine logic into go-graph-rag

### PRO

- One implementation of k-NN math in the fleet.
- The SDK would gain metaengine's richer primitives: euclidean/dot metrics,
  metadata-filtered k-NN, depth-N traversal, undirected variant, edge tombstones.
- Direction respects layering (application absorbs infra-grade helpers).

### CONTRA

- **json/v2 contamination — the recurring regression class.** 66 non-test files in
  metaengine import `encoding/json/v2`, including the exact candidate files
  (`vector_search.go`, `sqliteengine/vector.go`, `pebbleengine/vector.go`,
  `bboltengine/vector.go`). SDK non-negotiable #2 requires json/v1-only builds; the
  experiment-gated regression has already slipped into this repo three times
  (pre-v0.1.0, `e67bd9b`, `e4a9145`), each caught only by the pristine build. Every
  moved file is a scrub-and-retest, and one miss is a red CI master.
- **Semantics conflict (§3):** the moved math is not the SDK's math. Absorbing it
  either changes SDK ranking behavior (float32 accumulation, truncation on
  mismatched dims) or forks it again — recreating the duplication one directory over.
- **On-disk format collision:** metaengine's dim-headered blob vs the SDK's raw LE
  f32, which is a persisted SQLite cache format (`store.go` schema). Unifying
  encoders invalidates existing consumer caches.
- **The `any`-typed model rides along.** Depth-N traversal and tombstones are built
  on `Edge{From any, To any}` (`types.go:60`) — precisely the label-free model the
  2026-09-15 research rejected; the SDK's `Edge{Relation, Weight, Direction}` cannot
  flow through it without side-stores.
- **Solves a non-problem.** Multi-hop expansion is trigger-gated (~50k nodes / 1-hop
  insufficiency, dgraph research §7); metadata filtering has no requirement behind
  it; euclidean/dot have no consumer in the SDK's cosine-ranked pipeline.
- Breaks the SDK's dependency non-negotiables in spirit: the SDK stays
  dependency-light precisely so consumers (CV) get a stable, auditable core —
  absorbing an experimental module's primitives is the opposite motion.

**Verdict: NO.**

## 6. P3 — extract a shared leaf module (e.g. `go-vector`)

The fleet-correct shape: a tiny module holding pure math + blob codec, both repos
import it. PRO: single source, fleet-idiomatic, license-neutral, isolates the
skew. CONTRA:

- §3 stands: the leaf must expose BOTH semantics (two cosine functions, two blob
  formats) or break a consumer — at which point "shared" is a namespace, not a
  source of truth.
- Parity is already enforced by stronger instruments than shared code: inside
  metaengine, the adttest matrix asserts identical numerics across engines via the
  shared in-repo helpers; inside the SDK, example-output tests pin behavior.
  Cross-repo parity is asserted nowhere and required by nothing.
- Under version skew, shared code drifts anyway (pin v0.1 vs v0.2), so the module
  buys less than it appears to.
- Coordination cost per semantic change: bump leaf → tag → proxy verify → bump two
  consumers → pristine-check both (one with the machine-wide `GOEXPERIMENT=jsonv2`
  trap) → re-verify daemon commits. Recurring, for ~80 lines.
- The Go proverb applies verbatim: a little copying is better than a little
  dependency.

**Verdict: NOT YET.** Correct shape, wrong economics at the current surface size.
Trigger conditions in §9 make the flip objective.

## 7. Cost ledger (measured, not vibes)

| Item                                     | Measurement (2026-09-19)                                                                                              |
| ---------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| Net sharable code                        | ~80 lines (accumulation loops + truncation)                                                                           |
| Semantic changes a merge forces          | ≥5 (§3), each test-pinned on one side                                                                                 |
| json/v2 files needing scrub if P2        | 66 non-test files; 4 of them are the vector files themselves                                                          |
| Regression class activated by P2         | The experiment-gated build break that has recurred 3× and is caught only by pristine builds                           |
| Toolchain skew                           | SDK pins go1.26.7 (CI + pristine checks); metaengine modules are go 1.27.1 — leaf modules must target the lower floor |
| On-disk formats in conflict              | 2 (raw LE f32 cache blobs vs dim-headered binary)                                                                     |
| Consumers bearing the treadmill under P1 | metaengine engine modules + every future metaengine v5 consumer                                                       |
| Benefit either direction                 | None not already delivered by per-repo tests                                                                          |

## 8. Verdict

Keep P0. The merge question dissolves under measurement: the overlap is small,
deliberately divergent, and test-enforced on both sides; the coupling directions
either invert layering (P1), import contamination and rejected models (P2), or pay
fleet-coordination rent on 80 lines (P3). The smart move already happened — it is
the seam ADR: the SDK defined `VectorIndex`/`GraphStore` as the point where backends
(never libraries) plug in. When a backend integration becomes real, it is a new
adapter module, not a merge. Both 2026-09-15 verdicts remain intact and this
document adds the quantitative reason why: **there is no shared logic to merge,
only shared vocabulary.**

## 9. Revisit triggers

| Trigger                                                                                                                         | Re-evaluate                                                                                                                  |
| ------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| Seam ADR implementation starts AND an ANN backend lands (shared surface grows past ~300 lines: math + index plumbing + filters) | P3 leaf module becomes NPV-positive — re-run §3 inventory first                                                              |
| A third consumer needs both stacks' math with identical semantics                                                               | P3                                                                                                                           |
| CV app-layer adopts metaengine/system (SUPERB T01/T29)                                                                          | Not a merge trigger: integrate per 2026-09-15 (query handler + rebuild command above metaengine), dependency graph unchanged |
| SDK adds multi-hop expansion (dgraph research §7 trigger)                                                                       | Implement SDK-native typed traversal; do NOT import metaengine's `any`-edge BFS                                              |
| Q1 license decision lands (one repo goes OSS)                                                                                   | Re-check dependency-direction economics; today neutral (both PROPRIETARY)                                                    |
| metaengine v5 stabilizes and its vector semantics converge with the SDK's (similarity API, dim handling)                        | Re-run §3; a leaf module only makes sense if the functions become actually identical                                         |

---

_Point-in-time decision record (2026-09-19). Never rewrite; annotate inline or
archive. Line numbers and the json/v2 census were measured against both module
trees on this date — re-verify before acting._
