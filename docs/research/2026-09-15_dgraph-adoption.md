# Dgraph Adoption: PRO/CONTRA Deep Research

|          |                                                                                                                                                                                            |
| -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Date     | 2026-09-15                                                                                                                                                                                 |
| Question | Should `go-graph-rag` adopt Dgraph (as store, as backend, or at all)?                                                                                                                      |
| Method   | Primary sources only: GitHub API, official docs (docs.dgraph.io), pkg.go.dev, official forum (discuss.dgraph.io) posts. Every external claim below is cited; unverified items are labeled. |
| Verdict  | **Do not adopt into the SDK. Revisit only as a consumer-owned backend at the ~50k-node scale trigger.**                                                                                    |

## 1. TL;DR

Dgraph is alive again under new ownership (Istari Digital, since 2025-10) and is technically
capable: it bundles exactly what a GraphRAG stack needs (property graph + HNSW vector search +
fulltext) in one Go-native, Apache-2.0 database with an official Go client (`dgo/v250`).

It is still the wrong default for this SDK. `go-graph-rag` is a dependency-light, embedded,
rebuildable-index library (`store.go`: single SQLite file, one writer, stdlib-only except
`samber/lo` + `modernc.org/sqlite`). Dgraph is a distributed client/server database: adopting it
inverts the deployment model (every SDK consumer would have to operate a Zero + Alpha cluster),
pulls gRPC + protobuf into `go.mod` (against the non-negotiable dependency rule), and breaks the
"every persisted artifact is a rebuildable single file" promise. Its stewardship history
(three owners in three years, Hypermode's cloud shut down, community trust scars) is a real risk
signal, not FUD.

The rational path: keep SQLite (and the roadmap's `sqlite-vec` / `hannoy` / in-process HNSW
candidates) inside the SDK. If the CV-side consumer ever demonstrably outgrows ~50k nodes or
needs server-side multi-hop queries, expose a pluggable store seam and let _the consumer_ run
Dgraph. That decision trigger is already documented in `ROADMAP.md`; nothing in this research
moves it earlier.

## 2. What Dgraph is today (verified snapshot)

### 2.1 Stewardship timeline

| Date       | Event                                                                                                                                                     | Source                                              |
| ---------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------- |
| 2023-11-06 | "Dgraph Labs is becoming part of Hypermode" (acquisition)                                                                                                 | discuss.dgraph.io topic 19040                       |
| 2024-06-06 | v24.0.0 released: vector support lands (`similar_to`, `float32vector`, GraphQL vectors)                                                                   | GitHub release v24.0.0                              |
| 2024-11-21 | Hypermode: "The future of Dgraph is open, serverless, and AI-ready"                                                                                       | discuss.dgraph.io topic 19611                       |
| 2025-01-09 | Repos migrate `dgraph-io` to `hypermodeinc`                                                                                                               | discuss.dgraph.io topic 19690                       |
| 2025-09    | Hypermode winds down: `modus`, `modusGraph`, most org repos archived; Dgraph Cloud shut down (user reports Dec 2025)                                      | hypermodeinc org repos; discuss topic 20021         |
| 2025-10-07 | v25.0.0 released: "remove enterprise license completely from dgraph", "remove EE license and oss build" (one OSS build, Apache-2.0)                       | GitHub release v25.0.0                              |
| 2025-10-24 | Istari Digital acquires Dgraph ("the new stewards of Dgraph"); engineering continuity promised                                                            | discuss.dgraph.io topic 20021 + PR Newswire release |
| 2025-10-27 | Repos migrated back to `dgraph-io` "as we want Dgraph to stay an open-source project with its own place"                                                  | Raphael Derbier, topic 20021                        |
| 2025-11-05 | 2026 roadmap: maintained OSS project; focus on "security, scalability and distribution"; GraphQL layer becomes community effort; no cloud hosting planned | Raphael Derbier, topic 20021                        |
| 2025-12-02 | dgraph.io domain + docs transfer from Hypermode to Istari                                                                                                 | discuss.dgraph.io topic 20034                       |
| 2026-07/08 | v25.4.0 (2026-07-30) and v25.4.1 (2026-08-24) released; Go toolchain upgraded to 1.27.0 (2026-08-24)                                                      | GitHub releases; commit log                         |
| 2026-09-10 | Latest commit on `main` (query fix + dependency bumps)                                                                                                    | GitHub commits API                                  |

### 2.2 Project vitals (measured 2026-09-15)

| Metric            | Value                                                                                 | Source              |
| ----------------- | ------------------------------------------------------------------------------------- | ------------------- |
| Repo              | `dgraph-io/dgraph` (not archived, active)                                             | GitHub API          |
| Stars / forks     | 21,798 / 1,605                                                                        | GitHub API          |
| Open issues+PRs   | 99                                                                                    | GitHub API          |
| License (main)    | Apache-2.0 (`LICENSE.txt`)                                                            | GitHub contents API |
| Latest release    | v25.4.1 (2026-08-24); v24.1.9 maintenance line still patched                          | GitHub releases     |
| Go client         | `github.com/dgraph-io/dgo/v250` v250.0.0 (pkg.go.dev, Apache-2.0, 23 known importers) | pkg.go.dev          |
| Client deps       | `google.golang.org/grpc` + `google.golang.org/protobuf` (test-only testify)           | dgo go.mod          |
| Platforms         | Linux/amd64 + Linux/arm64 only (Mac/Windows support dropped 2021)                     | README              |
| Read-only mirror? | No: `hypermodeinc/dgraph` now redirects to `dgraph-io/dgraph`                         | GitHub API          |

## 3. Capability assessment vs `go-graph-rag` needs

What the SDK does today (see `build.go`, `search.go`, `store.go`): embed documents, rank by
cosine (linear scan in memory), expand hits via caller edges, render a deterministic LLM-ready
context block, persist everything as one rebuildable SQLite file.

### 3.1 Vector search

| Need (SDK)                | Dgraph answer (verified from docs.dgraph.io v25.4)                                                                                                          |
| ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Store embeddings          | `float32vector` predicate type (since v24.0.0, 2024-06)                                                                                                     |
| ANN index                 | HNSW: `@index(hnsw(metric:"cosine" \| "euclidean" \| "dotproduct", exponent, maxLevels, efConstruction, efSearch))`; default euclidean                      |
| kNN query                 | DQL `similar_to(predicate, topK, "[...]" \| $var)`; GraphQL auto-generated `querySimilar<Type>ByEmbedding(by:, topK:, vector:)` returning `vector_distance` |
| Cosine scoring            | Supported as an index metric                                                                                                                                |
| Hybrid graph expansion    | Native: kNN hits are nodes; edges/predicates traversable in the same DQL query (N-hop = N network hops per docs)                                            |
| Metadata filtering on ANN | Not verified either way in this research; treat as an open spike question                                                                                   |

So the core GraphRAG loop (vector hit + neighborhood) is genuinely expressible server-side.

### 3.2 Retrieval-adjacent features we would otherwise hand-roll

- `recurse` traversal and k-shortest-path queries (DQL docs)
- term/fulltext/ngram/trigram string indexes, facets on edges
- upsert blocks with conditional mutation (`@if`) and client-side retry semantics (`ErrAborted`)
- snapshot isolation + Raft replication; docs state Jepsen-tested ("gold standard" is Dgraph's
  wording, a vendor claim)
- multi-tenancy via namespaces (v25)

### 3.3 Operational footprint (the cost side)

| Aspect         | Fact (docs.dgraph.io architecture page)                                                                                                                                |
| -------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Minimum deploy | 2 processes: 1 Zero (control plane) + 1 Alpha (data plane); dev quick start via `dgraph/standalone` Docker image                                                       |
| Prod HA rec.   | 3 Zero + 3 Alpha (replicas=3); sharded prod: 9+ Alphas                                                                                                                 |
| Prod sizing    | Alpha: 8+ cores, 16GB+ RAM, 3000+ IOPS; Zero: 2-4 cores, 4GB                                                                                                           |
| Consistency    | Linearizable reads/writes, snapshot isolation transactions                                                                                                             |
| Backups        | Docs (v25.1/v25.4) still list binary backups under "enterprise features"; exports (RDF/JSON) are OSS. Note: docs lag the v25 EE unification; an observed inconsistency |
| Upgrades       | Major-version client coupling: `dgo/v230` (v23), `dgo/v240` (v24), `dgo/v250` (v25)                                                                                    |

## 4. PRO (arguments for adopting)

1. **Graph + vector + text in one engine.** Dgraph is one of the few Go-native databases where
   `similar_to` kNN, edge traversal, and term/fulltext filtering compose in a single query. Our
   `Searcher` two-phase design (rank, then expand) maps directly onto it, and at scale the
   expansion stops being an in-memory join.
2. **Headroom for the ~50k-node wall.** The README names ~50k nodes as the full-rebuild,
   linear-scan revisit point. Dgraph's predicate sharding + Raft replication is designed for
   terabyte-scale (docs tagline) and removes the full-rebuild model via incremental mutations.
3. **Maintenance risk has receded, materially.** After the Istari acquisition (2025-10) the
   project shipped v25.3.x and v25.4.x through 2026-08, unified all enterprise features into the
   Apache-2.0 OSS build, migrated repos back to the neutral `dgraph-io` org, and committed to
   "a maintained open source project by its own" (maintainer quote, topic 20021). It is not a
   zombie project: last commit 2026-09-10.
4. **Go-native end to end.** Server in Go, official client `dgo/v250` on pkg.go.dev (Apache-2.0,
   version-matched), single-image dev quick start (`dgraph/standalone`), Helm chart and k8s
   manifests for prod. No JVM, no polyglot ops burden.
5. **Vectors are first-class, not bolted on.** Since v24 (2024-06), with tunable HNSW
   (`exponent`, `efConstruction`, `efSearch`) and three metrics; v25 shipped vector performance
   work. Embeddings can also be declared directly in the GraphQL schema (`@embedding`).
6. **Enterprise features became free.** v25 removed the enterprise/oss binary split ("remove
   enterprise license completely from dgraph"): encryption-at-rest, ACL, backups, CDC are part of
   the one Apache-2.0 build now (docs sections still label some as enterprise; docs lag).

## 5. CONTRA (arguments against adopting)

1. **It violates the SDK's founding constraint.** `README.md` / `AGENTS.md`: stdlib-only except
   `samber/lo` + `modernc.org/sqlite`; adding a dependency needs a hard reason. `dgo/v250` drags
   gRPC + protobuf into `go.mod` of a v0 library, and the real cost is not the import: it is the
   server. An SDK cannot require consumers to operate a two-process distributed database as its
   default storage.
2. **It breaks the rebuildable-index promise.** `Store` is explicitly "a derived, rebuildable
   index" in a single SQLite file with one writer. Dgraph introduces cluster state, snapshot
   cadence, rebalancing windows (predicate moves go read-only), and a backup story the SDK cannot
   abstract away. Losing the file today costs a rebuild; losing a cluster is an incident.
3. **Governance volatility is proven, not hypothetical.** Three owners in three years (Dgraph
   Labs, Hypermode, Istari Digital). Hypermode archived nearly all its repos and shut down its
   Dgraph Cloud mid-2025; users had to migrate off a hosted service with days of notice (user
   accounts in topic 20021). Community scar tissue is on the record: "constantly changing CEOs,
   then the acquisition with Hypermode are all signs screaming stay away" (2023, topic 19111).
4. **The new owner's incentives are narrow.** Istari Digital is an aerospace/DoD engineering-data
   platform ("GitHub for Planes"). Its stated roadmap is customer-driven (security, scalability,
   distribution; possible future GQL/openCypher), it explicitly plans no cloud offering, and the
   GraphQL layer is now community-maintained. Continuity is real but concentrated: one company's
   roadmap, not a foundation.
5. **Operational weight is categorically larger than SQLite.** Minimum 2 processes, prod
   recommendation 6+, Alpha sized at 8+ cores / 16GB+ / 3000+ IOPS, Linux-only. Every consumer of
   this SDK (CLI tools included) would inherit that footprint or split the architecture.
6. **Vector search is young.** Shipped in v24.0.0 (2024-06): roughly two years old. Whether ANN
   queries compose cleanly with rich metadata filters was not verified in this research, and the
   SDK would be the layer to absorb any rough edges.
7. **Client version churn.** `dgo` import paths are hard-coupled to server majors (`/v230`,
   `/v240`, `/v250`). A consumer upgrading Dgraph majors must also change Go imports; a v0 SDK
   pinning these transfers churn to every downstream.
8. **What we would actually gain is smaller than it looks.** The SDK's differentiators
   (two-tier ranking, `SearcherOptions` role policy, deterministic LLM-ready context rendering)
   live in `search.go` and would remain ours regardless. Dgraph would only replace the cosine
   scan and the neighbor lookup: the two cheapest parts of the design, at a scale we do not have.

## 6. Adoption scenarios

| Scenario                                                   | Verdict   | Why                                                                                                                                                                                                                              |
| ---------------------------------------------------------- | --------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| A. Replace the default `Store`/`Cache` inside the SDK      | **NO**    | Contradicts dependency rule, embedded model, and rebuildability; forces a DB server on every consumer of a retrieval library                                                                                                     |
| B. Optional pluggable backend behind a `Store`/search seam | **LATER** | Viable only after the seam exists (it does not today; `Cache` is the only seam), and only if a consumer demands it; cost: adapter + CI matrix + dgo dep isolation                                                                |
| C. Consumer-side (CV repo) deployment at real scale        | **MAYBE** | The only defensible adoption point: CV owns its infra and can run Zero+Alpha; but the roadmap's lighter candidates (sqlite-vec, hannoy, in-process HNSW) preserve single-file simplicity and likely clear the 50k-node bar first |

## 7. Recommendation

1. Keep the non-goal as written (`ROADMAP.md`: "External vector databases: SQLite-only until
   scale demonstrably hurts").
2. Next storage investments, in roadmap order: `sqlite-vec` or `hannoy` ANN behind the existing
   `Searcher` seam, then incremental indexing. Both keep the single-file, rebuildable model.
3. If a consumer ever needs server-side multi-hop graph queries (true graph analytics, not
   1-hop expansion) or sustained >50k-node corpora, reopen Dgraph as scenario B/C and run a
   timeboxed spike first (Section 8).
4. Track two Dgraph-side signals as tripwires for the "alive" judgment: release cadence after
   v25.4.x, and whether a second meaningful steward (or foundation) ever materializes.

### Decision triggers to reopen this report

- Sustained corpus above ~50k nodes where full rebuild hurts in practice
- A product need for multi-hop graph queries or k-shortest paths at query time
- ANN in SQLite (sqlite-vec/hannoy) measured insufficient (recall or latency) at real corpus size
- A consumer explicitly requires Dgraph compatibility

## 8. What an adoption would look like (spike sketch, for scenario B)

- Server: `docker run -p 8080:8080 -p 9080:9080 dgraph/standalone:latest` (dev), 3+3 cluster (prod).
- Client: `go get github.com/dgraph-io/dgo/v250`; `dgo.Open("dgraph://localhost:9080")`.
- Schema (per node kind): `label: string @index(term) .`, `embedding: float32vector @index(hnsw(metric:"cosine")) .`, `related: [uid] @reverse .`
- Build path: `SetSchema` + batched `Txn.Mutate` (JSON), embeddings still produced by the SDK's
  `Provider` seam; Dgraph stores vectors, it does not compute them.
- Search path: DQL `similar_to(embedding, $k, $queryVec)` + 1-hop `related` expansion in the same
  query; SDK keeps scoring policy (`SearcherOptions`), reranking, and context rendering.
- Failure handling: `ErrAborted` retry loop; write conflicts expected under concurrent builders,
  which the SDK's one-writer model does not currently surface.

## 9. Verification status

| Claim                                                                       | Status                                   | Source (checked 2026-09-15)                                                                            |
| --------------------------------------------------------------------------- | ---------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| Repo active, Apache-2.0, 21,798 stars, latest v25.4.1                       | Verified                                 | GitHub API (`repos/dgraph-io/dgraph`), `LICENSE.txt` contents                                          |
| Vector support shipped in v24.0.0 (2024-06-06)                              | Verified                                 | v24.0.0 release body (`feat(vector)` PRs)                                                              |
| HNSW metrics cosine/euclidean/dotproduct, `similar_to` DQL, GraphQL kNN     | Verified                                 | docs.dgraph.io: dql/predicate-indexing, dql/query/functions, graphql/queries/vector-similarity         |
| Istari Digital acquisition 2025-10, repos back to dgraph-io, no cloud plans | Verified                                 | discuss.dgraph.io topics 20021, 20034 (maintainer posts, quoted)                                       |
| v25 removed enterprise license / EE+OSS split                               | Verified                                 | v25.0.0 release body (PR titles quoted)                                                                |
| dgo/v250 module, gRPC+protobuf deps, version-matched import paths           | Verified                                 | dgo go.mod + README (raw), pkg.go.dev module page                                                      |
| Min cluster 1 Zero + 1 Alpha; prod 3+3; Alpha 8+ cores/16GB/3000+ IOPS      | Verified                                 | docs.dgraph.io v25.1 installation/dgraph-architecture                                                  |
| Linearizable reads, snapshot isolation, "tested via Jepsen"                 | Verified as vendor claim                 | docs architecture page + design-concepts/transactions-concept                                          |
| Fortune 500 / terabyte-scale production use                                 | Vendor claim, not independently verified | README, docs homepage tagline                                                                          |
| Binary backups listed as enterprise feature while v25 unified EE into OSS   | Verified inconsistency (docs lag)        | docs v25.1 admin/enterprise-features vs v25.0.0 release notes                                          |
| Hypermode cloud shutdown mid-2025                                           | Verified via secondary accounts          | archived hypermodeinc repos (GitHub) + user posts in topic 20021 (not an official Hypermode statement) |
| ANN-with-metadata-filter support in Dgraph                                  | Unverified                               | Not tested in this research; flagged as open spike question                                            |

## 10. Sources

- GitHub API: `repos/dgraph-io/dgraph` (metadata, commits, releases v24.0.0/v24.0.2/v25.0.0/v25.3.x/v25.4.x, contents of `LICENSE.txt`, `README.md`); `repos/dgraph-io/dgo` (go.mod, README); `orgs/hypermodeinc/repos`
- pkg.go.dev: `github.com/dgraph-io/dgo/v250`
- docs.dgraph.io (v25.4/v25.1): `installation/dgraph-architecture`, `graphql/schema/directives/search`, `graphql/schema/directives/embedding`, `graphql/queries/vector-similarity`; DQL `predicate-indexing` + `query/functions` + `learn/howto/similarity-search` via `dgraph-io/dgraph-docs` repo
- discuss.dgraph.io: topics 19040 (Hypermode acquisition), 19611 (future of Dgraph), 19690 (repo migration), 19837 (unification), 20021 (Istari acquisition + roadmap, quoted), 20034 (domain transfer), 19111 (community history thread, quoted), 17697 ("The Good, The Bad, The Ugly", GraphQL-era gap catalog)
- Local: `README.md`, `ROADMAP.md`, `AGENTS.md`, `store.go`, `search.go` (this repository)
