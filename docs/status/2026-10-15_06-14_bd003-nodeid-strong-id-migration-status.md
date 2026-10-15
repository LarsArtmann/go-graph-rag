# Status Report: BD003 Strong-ID Migration (`NodeID`)

- **Date/time:** 2026-10-15 06:14 CEST
- **Session scope:** fix all 18 `brandid-lint` BD003 (type-safety) findings across core + `bdd/` sub-module; verify every project gate; amend affected ADRs ADR-first; document in CHANGELOG + AGENTS.md.
- **Outcome:** `brandid-lint -rules BD001,BD003 .` exits 0 with zero findings. All gates green. No broken state.
- **Format note:** single Markdown file at the user-demanded path per the standing `.md` dispatch exception; the brutal-self-review answers are folded in here instead of a separate HTML report at `docs/reviews/`.

## Direct answers first

### What did I forget?

1. **No benchmark sanity run.** The change swaps map key types (`string` → `NodeID`) on hot paths (`BuildResult.Vectors`, `Searcher.vectors`). Near-zero risk (same underlying kind), but I never ran even a 100ms `BenchmarkBuild`/`BenchmarkSimilarPairs` sanity pass, and per the AGENTS.md benchstat protocol the `bench_test.go` reference table was not re-measured. Unverified-by-measurement claim of "no perf impact".
2. **CV-side compile impact is unknown.** CV consumes this module from the proxy; callers passing `string` variables (not literals) into `Document.ID`/`Edge` endpoints will fail to compile after this change. I did not (and per instructions should not) check the CV repo. The CHANGELOG entry documents the migration, but the actual consumer work is unquantified.
3. **Dangling uncommitted docs.** `AGENTS.md` and `CHANGELOG.md` edits sit uncommitted in the working tree (harness forbids me committing; the daemon had not picked them up at report time). The daemon committed all code changes in 4 batches (`5914fb1`, `b7daf2c`, `88fc373`, `196b88d`) — verified, but the docs were still dangling when I finished.
4. **`bdd/go.mod` require-sync reminder not recorded anywhere actionable.** The next `bdd/v*` tag must bump its core require to the release containing this change. The convention lives in a go.mod comment only; no TODO row exists.
5. **`docs/DOMAIN_LANGUAGE.md` not extended.** `NodeID` is a new ubiquitous-language term and belongs in the glossary next to `NodeKind`.
6. **TODO_LIST.md not harvested** from the next-steps brainstorm (section f below). Status-report skill says section f is primary HARVEST input; user said wait, so it waits.
7. **BD002 never investigated.** The linter auto-emits "BD002 directive findings"; I never looked at what directives exist or whether any fired.
8. **Pre-push hook installation not verified** for this clone (`git config core.hooksPath .githooks` is per-clone; AGENTS.md warns every push runs the three gates).
9. **Ginkgo `DescribeTable` reflection gotcha not documented for kit consumers.** Entry values must be explicitly typed (`graphrag.NodeID("x")`) once params are named types — string literals panic at runtime via Ginkgo's reflect-based invocation. I hit this, fixed it, and documented it nowhere.
10. **Final bdd test pass was cache-served.** The last full-module verification printed `ok ... (cached)`. Content was unchanged since the fresh green run, but "verified" is technically one run stale.

### What could I have done better?

1. **Wasted two tool calls on `multiedit`.** It failed twice with an internal validation error before I switched to single `edit` calls. Should have adapted after the first failure, not the second.
2. **Should have run the cheap bench sanity pass** (see forget #1) instead of arguing from theory in the final summary.
3. **Deviated from the linter's literal suggestions without flagging it up front.** Using one `NodeID` instead of the suggested `DocumentID`/`OutEdgesID`/`expandID`/`AID`/`BID`/`DocID` was a domain-first judgment call, validated empirically (exit 0) — but the deviation-and-rationale only appeared in the final summary, not before the work.
4. **Lint tooling tension left implicit.** I ran `golangci-lint` directly per the project AGENTS.md canonical commands (CI parity), while the buildflow skill says "run buildflow, not the underlying tool". Project AGENTS.md wins for this repo, but the choice deserves one explicit line in AGENTS.md rather than staying tribal knowledge.
5. **ADR amendment is a judgment call worth airing.** Planning docs are "point-in-time, never rewrite" — I edited the seam ADR §5 sketch text (its own EDIT-THE-ADR-FIRST governance demands exactly that, and the amendment is dated/annotated), but a purist reading of the policy would have annotated-only. The BDD ADR got the same treatment. Defensible; not silent.
6. **Could have written the failing-state tests first** (e.g., a NodeID JSON round-trip) before the sweep. The wire goldens + godoc examples already pin behavior, so this is a nicety, not a gap.

### What could I still improve?

- Type-model next step: `ContentHash` for cache keys (BD003's rule only names `id`/`ID` identifiers, so content hashes slipped through — see section f).
- Glossary, README selling point, migration note for consumers (section f).
- Process: bench sanity as a standing verification step for perf-adjacent changes; faster tool-failure adaptation; harvest discipline.

## a) FULLY DONE

1. All 18 BD003 findings fixed; `brandid-lint -rules BD001,BD003 .` exit 0, zero findings (BD001 also clean).
2. `NodeID` defined type added in `graph.go` with doc comment (namespace rule: a document's ID becomes its node's ID; wire shape unchanged; untyped literals still assign).
3. Core prod migration: `Node.ID`, `Edge.Source`/`Edge.Target`, `Document.ID`, `Graph.Node`/`OutEdges`/`InEdges`, `Graph` internal maps, `mergeEdges` key, `BuildResult.Vectors` (`map[NodeID]Vector`), `BuildResult.EmbeddingHash`, `resolveVectors`, `similarEdges`, `pairKey`, `Searcher.vectors`, `NewSearcher`/`NewSearcherWithOptions`, `cloneVectors`, `SimilarPairs`, `expand`, `nodeOrPlaceholder`, `Store.ReplaceGraph`, `LoadGraphSnapshot.EmbeddingHash`, `loadNodes`.
4. Root test migration: `graph_test.go`, `search_test.go`, `store_test.go`, `hardening_test.go`, `bench_test.go`, `adrsketch_test.go` (compile-pin mirror of the amended ADR sketch).
5. `bdd/` sub-module migration: `builders.go`, `matchers.go`, `harness.go` (`joinVectors`), `build_spec_test.go` (including typed `DescribeTable` entries after the reflection panic), `matchers_spec_test.go`, `search_spec_test.go`. All 43 specs pass.
6. Seam ADR (`docs/planning/2026-09-16_13-25_seam-store-search-adr.md`) §5 sketch amended ADR-first with a dated inline annotation; `VectorIndex.Replace` and `VectorMatch.NodeID` are now `NodeID`-typed in the binding design before any adapter exists.
7. BDD kit ADR (`docs/planning/2026-10-15_05-02_bdd-submodule-adr.md`) §3 sketch updated + dated annotation.
8. `CHANGELOG.md` [Unreleased] "Changed" entry: full surface list, wire-compat guarantee, migration note.
9. `AGENTS.md`: `graph.go` bullet documents the `NodeID` policy; `brandid-lint` added to the Commands block.
10. Verification battery, all green: pristine build (`env -u GOEXPERIMENT GOTOOLCHAIN=go1.27.1 go build ./...`) both modules, `go vet ./...`, `go test ./...` (core), `go test ./... -race` (core), `golangci-lint run ./...` 0 issues (both modules), `go mod tidy` + `git diff --exit-code go.mod go.sum` (both modules), `dprint check`, `brandid-lint`.
11. Wire-contract proof: `testdata/wire/*.golden` untouched and passing; godoc example `// Output:` blocks unchanged and passing — JSON shapes and rendered context are byte-identical.
12. Daemon-commit verification: all 4 auto-commit batches inspected; every gate re-run on top of the final committed state.

## b) PARTIALLY DONE

1. **Documentation of the change:** CHANGELOG + AGENTS.md done; DOMAIN_LANGUAGE.md glossary, README, consumer migration note not touched.
2. **Release readiness:** code is release-quality (all gates green) but unversioned — no v0.4.0 cut, `bdd/go.mod` require not synced to a new core tag, no `bdd/v*` tag, CV-side bump untouched (pre-existing owner package Phase 8).
3. **Self-review:** folded into this report per the explicit one-file instruction; the canonical HTML self-review at `docs/reviews/` was intentionally not produced this session.
4. **Bench story:** benchmarks compile and run in the suite; reference numbers neither re-measured nor re-verified against the new types.
5. **Daemon state:** code fully committed by the daemon (verified); two doc files dangling uncommitted at report time.

## c) NOT STARTED (identified this session, zero work done)

1. `ContentHash` strong type for embedding-cache keys (BD003 blind spot: the rule only flags `id`/`ID`-named strings).
2. `brandid-lint` as a pre-push or CI gate (owner decision; currently a local-only command).
3. TODO_LIST.md harvest from section f.
4. Benchmark re-measure + `bench_test.go` table re-record (benchstat protocol) — only if the owner wants post-change numbers.
5. v0.4.0 release cut with this change; `bdd` require sync + first `bdd/v*` tag (owner-gated).
6. CV-side adoption (owner package Phase 8): `internal/graphvocab`, `internal/features/graphrag/service`, `internal/config` touchpoints.
7. `NodeID` glossary entry in `docs/DOMAIN_LANGUAGE.md`.
8. Kit-consumer gotcha note (typed `DescribeTable` entries) in `bdd/doc.go`.
9. BD002 directive-rule investigation.
10. CI-green confirmation for the daemon's 4 commits on master (branch protection runs `build-test-lint`; I did not check CI state per scope limits).

## d) TOTALLY FUCKED UP

**Nothing.** No broken state, no data loss, no reverted work, no planted-failure cleanup needed. The two closest calls, both caught immediately and resolved:

1. Two aborted `multiedit` calls (tool-internal validation error; atomic no-op, zero damage) — cost: two wasted tool calls.
2. Ginkgo `DescribeTable` panic in 3 bdd specs (reflection requires typed entry values) — caught by the test run on first execution, fixed in one edit, all 43 specs green.

## e) WHAT WE SHOULD IMPROVE

1. **Verification completeness over argument-from-theory:** run the cheap measurement (bench sanity) when a change touches hot-path data structures, even when the risk "should" be zero.
2. **Tool-failure adaptation speed:** after one `multiedit` validation failure, switch strategy immediately.
3. **State deviations from tool suggestions explicitly and early** — the `NodeID`-vs-`DocumentID` design call deserved a stated rationale before execution, not only after.
4. **One canon for lint runs:** project AGENTS.md commands vs buildflow delegation — record the chosen rule in AGENTS.md (one line) so no future session re-litigates it.
5. **Split-brain check on ADR sketches:** the seam sketch is compile-pinned by `adrsketch_test.go`; the BDD kit ADR sketch has no pin and drifted once already (today). Consider a small pin test or accept documented drift risk.
6. **Kit ergonomics:** document the typed-entries requirement for Ginkgo tables in the bdd kit docs before a consumer hits the panic.
7. **Daemon discipline:** keep re-verifying daemon commits that land mid-session (worked this time; AGENTS.md gotcha exists because it sometimes does not).
8. **Type-model hardening roadmap:** `ContentHash` next, then decide policy for `NodeAttrs.Source` (display metadata) vs `Edge.Source` (node ID) name collision — a naming-review candidate, not a rename to do casually.
9. **Consumer empathy:** a short "upgrading to NodeID" note (README or docs/) beats a CHANGELOG paragraph for the one affected caller pattern (`string` variables).

## f) Up to 50 things we should get done next

Brainstorm, not a commitment list — most items below the first ~10 are ROADMAP fuel for `docs-health` HARVEST routing. Ordered roughly by impact.

1. Commit (or daemon-collect) the dangling `AGENTS.md` + `CHANGELOG.md` doc edits.
2. Confirm CI green on the daemon's 4 master commits (`build-test-lint` check).
3. Decide + cut v0.4.0 (breaking-for-`string`-variables change in v0.x) with release notes from the CHANGELOG section; bump `UserAgentVersion` per AGENTS.md.
4. Sync `bdd/go.mod` require to the new core tag; cut the first `bdd/v*` tag (owner-gated) so consumers can fetch the kit.
5. CV-side bump: fold `NodeID` into owner package Phase 8; convert `internal/graphvocab`/service/config touchpoints where `string` variables cross the boundary.
6. Bench sanity run (`-benchtime 100ms`) on `BenchmarkBuild`/`BenchmarkSimilarPairs`/`BenchmarkStoreRoundTrip` to confirm map-key typing is perf-neutral; schedule a full benchstat re-measure only if numbers move.
7. Introduce `ContentHash` strong type for `EmbeddingEntry.Hash`, `Cache` seam, `HashText` outputs, and `Store` cache columns-boundary.
8. Add `NodeID` (and later `ContentHash`) to `docs/DOMAIN_LANGUAGE.md` glossary.
9. Add a short "upgrading to NodeID" consumer note (README section or `docs/`).
10. README: add strong-ID typing as a selling point; re-verify the quickstart snippets copy-paste-run.
11. Document the Ginkgo typed-entries gotcha in `bdd/doc.go` (kit consumers will hit it).
12. Owner decision: promote `brandid-lint -rules BD001,BD003` into `.githooks/pre-push` and/or CI.
13. Investigate BD002 (directive findings): what directives exist, whether any fire here.
14. Harvest this report's section f into `TODO_LIST.md` (docs-health HARVEST, routing rigor applied).
15. Consider a compile-pin test for the BDD kit ADR §3 sketch (mirrors the seam ADR pattern) — or document why not.
16. Add a `NodeID` godoc example (repo-side godoc matters; pkg.go.dev is license-blocked until the Q1 license decision).
17. Q1 license decision (pre-existing, AGENTS.md): unblocks pkg.go.dev godoc rendering for everything, including the new type docs.
18. Metaengine dep-tree quantification (pre-existing owner decision 2026-09-16, still pending).
19. Naming review candidate: `NodeAttrs.Source` (display metadata) vs `Edge.Source` (node reference) — same word, two meanings.
20. Decide policy for remaining stringly-typed spaces: `NodeAttrs.URL`/`Location`/`RemotePolicy` stay display strings (likely fine) — record the deliberate non-typing.
21. Property test: Build determinism under `map[NodeID]...` iteration (byte-stable nodes/edges across runs) — current tests cover determinism at fixed inputs; a property test would broaden it.
22. Explicit JSON round-trip test for `Node`/`Edge` with `NodeID` under both engines (goldens pin the v2 engine's output; a v1-vs-v2 decode-equals test would complete the story).
23. Consider `slices.Sort` migration for the remaining `sort.Slice` call sites (readability; graph.go, search.go, adrsketch fake) — mechanical, low priority.
24. `bdd` kit: matcher for "hit list contains IDs in any order" (complement to rank-pinned `HaveHitIDs`) if consumers ask.
25. `bdd` kit: `Doc` overload accepting `WithText`-style options for symmetry (only if a real consumer asks — YAGNI today).
26. Pre-push hook: verify `git config core.hooksPath .githooks` is installed in this clone before the next push.
27. AGENTS.md: one-line rule for golangci-lint direct-vs-buildflow (see e.4).
28. Re-run the planted-failure drills if the pre-push guard gains a fourth gate (brandid-lint) — prove the blocking path.
29. Check whether `brandid-lint` supports SARIF output wired into CI annotations (`-format sarif` exists per `-h`).
30. Consider `go-branded-id`-style validation (non-empty IDs) at `Build` entry: error on empty `Document.ID` — currently silently becomes a node with empty ID. Owner call: invariant enforcement vs permissiveness.
31. `SimilarPairs` placeholder nodes: reconsider rendering (`Label: string(id)`, `KindUnknown`) — is a vector-without-node a data hole worth surfacing louder? (AGENTS.md already calls KindUnknown a "data hole".)
32. `docs/research/README.md`: no row needed this session (no research artifact) — confirmed correct, keep the discipline.
33. FEATURES.md: consider a "strong ID types" row under the type-system inventory.
34. ROADMAP: record the "distinct `DocumentID` if documents ever split from nodes (chunking)" trigger so the one-namespace decision is revisited deliberately, not by accident.
35. bdd kit: example in `doc.go` still shows literal-style calls — still valid; optionally annotate with the typed-variable pattern for completeness.
36. Sweep `string` concatenation helpers (`pairKey`) for allocation-friendly variants (`slices.Concat`, `strings.Join`) — micro; only with a bench reason.
37. Error messages: `store.go` interpolates `NodeID` with `%s` — verify all user-facing messages render identically post-change (tests cover the paths; a grep audit is cheap).
38. CI: consider adding the bdd module's race test to the workflow (currently core parity only — verify; if already covered, close as no-op).
39. `hardening_test.go` fixture table: other `id string`-style tables elsewhere? (grep audit for future sweeps).
40. Toolchain: keep `GOTOOLCHAIN=go1.27.1` + golangci v2.13.2 pins moving together (AGENTS.md rule) — nothing to do now, just the standing reminder.
41. Upstream: held patches (tobi/qmd#959, charmbracelet/crush#3846) — no reminder pings per owner decision; listed for completeness only.
42. `bdd` Harness: `Search` variadic options — consider `WithSearchOptions` for symmetry with `WithBuildOptions` (only with consumer demand).
43. Consider `Store.Stats` `ByKind` keys staying `string` (kind as map key in a stats DTO) — deliberate; record next to 20.
44. Docs: annotate the seam ADR §1 "current state" table row (`map[string]Vector` reference) if it reads misleading after this change — the amendment note already covers it; check reader experience.
45. Evaluate whether `VectorQuery.Filter`/`VectorMatch` (ADR sketch) should move into core as compiled types now that signatures are NodeID-stable — trigger-gated, note only.
46. Housekeeping: after v0.4.0, annotate this status report (docs-health ANNOTATE) once its items resolve; archive per policy.
47. bdd suite runtime is 8ms — plenty of headroom; if specs grow, keep sub-second (kit philosophy).
48. Consider `go vet` copylocks/print checks on the new type (already covered by vet run — close as done).
49. Keep `bench_test.go` doc-comment protocol intact: if item 6 triggers re-measure, re-record machine/date/table together.
50. Celebrate the wire goldens: they turned a type-sweep across 21 files into a zero-drama change — preserve the golden discipline for any future wire event.

## g) Questions I cannot figure out myself

1. **Release intent:** should this land as **v0.4.0 now**, or batch with the CV-side Phase 8 bump? And do you want the bench table **re-measured (benchstat, -count 10) and re-recorded** as part of that release, or is the existing table (pre-NodeID) fine to keep?
2. **CV-side caller shape:** do CV callers pass `string` _variables_ (not literals) into `Document.ID` / `Edge` endpoints — and if so, should CV adopt `graphrag.NodeID` inside `internal/graphvocab`, or convert at the boundary? I cannot see the CV repo from this session (and was told not to research it).
3. **Gate + scope policy:** should `brandid-lint -rules BD001,BD003` become a **pre-push/CI gate**, and is the ID sweep **done by policy** (content hashes, `NodeAttrs.*` metadata, and `ByKind` stat keys stay plain strings), or do you want a follow-up sweep (e.g. a `ContentHash` type)?

---

_Point-in-time snapshot, 2026-10-15 06:14 CEST. Section f is HARVEST fuel for TODO_LIST.md on your instruction; awaiting further instructions._
