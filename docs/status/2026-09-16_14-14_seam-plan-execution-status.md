# Status Report — SUPERB seam-ADR plan execution (M1–M9 done, M10 paused at the push gate)

|         |                                                                                                                                             |
| ------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| Date    | 2026-09-16 14:14 CEST                                                                                                                       |
| Session | Executed `docs/planning/2026-09-16_12-15_SUPERB-seam-adr-execution-plan.md` (10 medium / 46 fine tasks) after the owner's "GET SHIT DONE"   |
| Verdict | M1–M9 executed with every fine-task verify gate green. M10 stopped at its designed stopping point: F10.2 done, F10.1/F10.3 await the owner  |
| Git     | Working tree CLEAN; master = 10 heuristic auto-daemon commits AHEAD of origin/master (nothing pushed); CI green on origin HEAD `bd14092`    |
| Honesty | 3 self-inflicted bugs found and fixed by the session's own gates (§d); 2 scope expansions beyond the plan (§b); 1 claim corrected (§g/note) |

## 1. What this session delivered in one paragraph

The seam ADR exists and is binding: `docs/planning/2026-09-16_13-25_seam-store-search-adr.md`
(300 lines) with the three backend classes (embedded ANN libs / metaengine projection /
consumer-owned server DBs), per-class vector semantics, a Go interface sketch (`VectorIndex` +
`GraphStore`) that COMPILES against the real SDK — and the real `*Store` provably satisfies the
sketched `GraphStore`. The metaengine rejection now has hard numbers: 31 → 58 → 205 modules (6.6×)
measured in a scratch consumer module. Tags are protected by an active ruleset (id 23541172).
`Build`'s identical-text rule is documented, two new output-verified godoc examples exist
(`ExampleOpenStore`, `ExampleNewSearcherWithOptions`), and benchmarks went from single-run
numbers to benchstat-grade (count=10, ±1–2%) plus a new store round-trip benchmark proving
persistence is NOT the bottleneck (10k nodes round-trip in 197ms vs ~37s rebuild). Docs-health
tooling got a code-span-aware annotate guard, a NEW completeness checker, permanent dotted-ID
support, and a regression suite. SECURITY.md had a real 404 link (fixed + verified), README had
a phantom `Reindex` identifier (fixed). Release automation is drafted and actionlint-clean. Both
external watch targets re-checked (still silent); `qmd embed` cleared the 173-doc backlog (1677
chunks, verified via MCP status: needs-embedding 0).

## a) FULLY DONE (every fine task's verify gate passed)

| Block                 | Fine tasks | Evidence                                                                                                                                                                                                                                                                                                                                                          |
| --------------------- | ---------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| M1 Seam ADR           | F1.1–F1.9  | ADR at `docs/planning/2026-09-16_13-25_seam-store-search-adr.md`; sketch compiled verbatim in /tmp (`env -u GOEXPERIMENT go build` OK) + `var _ GraphStoreQ = (*gr.Store)(nil)` proves the real store satisfies it; ROADMAP Theme 1 cross-links it; research README left research-only; dprint clean; all referenced paths exist                                  |
| M2 Dep quantification | F2.1–F2.5  | Measured: 31 mod/77 edges/60 go.sum → +metaengine+sqliteengine 58/168/94 → +system 205/1020/395; table in ADR §6; ONE inline annotation (quantified blockquote) on metaengine report CONTRA #1; /tmp module trashed; root `go.mod`/`go.sum` diff-clean                                                                                                            |
| M5 Tag protection     | F5.1–F5.4  | Ruleset `protect-tags` id 23541172: target=tag, enforcement=active, `refs/tags/*`, deletion + non_fast_forward, NO bypass actors; API echo verified; AGENTS research-index rule added; CHANGELOG entry                                                                                                                                                            |
| M3 Doc truth          | F3.1–F3.5  | `Build` doc comment documents identical-text → first-doc-wins rule incl. warm-cache nuance; `ExampleOpenStore` + `ExampleNewSearcherWithOptions` output-verified under `go test`; build/vet/test/lint green; README points at the roles example                                                                                                                   |
| M4 Benchmarks         | F4.1–F4.6  | benchstat (golang.org/x/perf v0.0.0-20260908200009): Build/100 3.961ms ±2%, Build/1000 369.5ms ±2%, Search 101.1µs/948.6µs ±1%, SimilarPairs 391.1ms ±1%; NEW `BenchmarkStoreRoundTrip` (1k/10k on real SQLite file): 21.6ms / 197.4ms; file doc comment rewritten with protocol + numbers; ROADMAP citation synced 373→370ms; gates green                        |
| M6 Tooling (core)     | F6.1–F6.3  | `outside_code_spans`/`already_annotated` guard fix; NEW `check-rows.py` per-row completeness checker; regression suite ALL GREEN: self-tests pass, archived plan = COMPLETE exit 0 (0 false positives), planted miss = 51/52 with row named (0 false negatives), F12.1 dry-run proceeds (the historical false trip is dead)                                       |
| M7 Docs close-out     | F7.1–F7.4  | dprint fmt+check clean; lychee 6/6 OK 0 errors after fix; F12.1 render check VERIFIED via GitHub markdown API (strikethrough renders, literal tildes protected inside code span); DOMAIN_LANGUAGE 20-row audit vs code — all accurate (SHA-256 confirmed in vector.go); SECURITY.md 404 fixed (new URL fetched and verified live); README `Reindex` phantom fixed |
| M9 Release automation | F9.1–F9.3  | `.github/workflows/release.yml`: on `v*` tag, extracts the tag's CHANGELOG section, FAILS on empty extraction, least-privilege `contents: write`, pinned checkout SHA; actionlint 1.7.12 + YAML parse clean; CHANGELOG entry states "drafted only, no release executed"                                                                                           |
| M8 Externals          | F8.1–F8.2  | qmd#959 OPEN + silent; crush#3846 OPEN + silent (it is an ISSUE, not a PR — see §g note); window ends 2026-09-22; `qmd embed` ran: 1677 chunks / 173 docs / 11m48s; MCP status verified "Needs embedding: 0"                                                                                                                                                      |
| M10 sync (part)       | F10.2      | TODO_LIST rewritten: 10 done rows deleted, watch row evidence updated with 2026-09-16 states, BLOCKED rows untouched; FEATURES PLANNED rows re-cited (line-number rot fixed → identifier refs) + seam row now says "design DONE, implementation trigger-gated"; AGENTS non-negotiable #3 now points at the ADR's adapter/dep rules                                |

## b) PARTIALLY DONE

1. **F6.4 (go-cqrs-lite skill wording) — fixed upstream, NOT propagated.** The installed skill
   resolves into the NIX STORE (`/nix/store/…-source/.agents/skills/go-cqrs-lite`) — read-only by
   design, correctly refused my first sed. The fix (`system` = Experimental per repo FEATURES.md
   ONLY, not module-level) is applied in the true source repo
   `/home/lars/projects/go-cqrs-lite/.agents/skills/go-cqrs-lite/references/modules.md` (verified:
   1 replacement). NOT committed there (no authorization; that repo also carries pre-existing
   uncommitted workflow changes that are NOT mine — left untouched per safety rules). The
   installed copy stays stale until the flake input is bumped.
2. **F10.2 (CHANGELOG) — missing entries.** Tag protection, seam ADR, and release automation ARE
   in `[Unreleased]`/Added. NOT yet added: the two examples, the `Build` doc rule, the
   benchstat/round-trip benchmark upgrade, the README/SECURITY fixes, the M6 tooling work
   (skill-repo side), the dep-quant annotation.
3. **F8.3 (owner queue) — delivered as this report's §g.** The three queued questions are below.
4. **M6 scope expansion (undebated until now):** dotted-ID annotator support (the F1.1-style
   `/tmp` variant is now permanent in `annotate-rows.py`) and a rewritten self-test loader
   (importlib; the old documented one-liner could never import a hyphenated module). Both
   regression-tested. This exceeded F6.1–F6.3's letter; I judge it in M6's spirit ("every future
   annotate pass is trustworthy") — flagging it as an explicit scope call.

## c) NOT STARTED

1. **F10.1** — final pristine gate suite (`env -u GOEXPERIMENT` build/vet/test, golangci,
   tidy-drift, dprint) over the finished tree. All individual gates ran green at their task
   boundaries; the FINAL consolidated pass did not run (the last doc edits postdate them).
2. **F10.3** — detailed commits + push + CI watch. Superseded in part by reality: the
   auto-commit daemon already swept everything into **10 heuristic commits** (578 insertions, 13
   files — verified complete vs origin). Remaining decision: push the 10 as-is vs squash/reword
   into per-task detailed commits (local-only rewrite, needs owner approval), then push + watch
   `go-test`.

## d) TOTALLY FUCKED UP (self-inflicted, all caught by this session's own gates — which is the point of gates, but each was avoidable)

1. **I shipped a benchmark that could not pass.** `BenchmarkStoreRoundTrip` v1 built with
   `nil` cache — the embedding table stayed empty, `LoadEmbeddings` returned 0, the bench FAILED
   43s into the background run. I wrote the assertion without tracing where the rows come from.
   Fixed by passing the store as cache (the real pipeline shape). Cost: one wasted 43s run.
2. **My first completeness checker was wrong twice.** (a) The doubled-backtick regex
   (`[^`]+`) can't match spans CONTAINING backticks; (b) I required every struck cell to
   END with`~~`, which misreads the annotator's own output format (marker appended INSIDE the
   first cell after the closing tildes) — it flagged the fully-annotated archived plan as 0/15
   and 0/52 struck. I wrote a validator for a format I hadn't re-read. F6.3's regression-first
   discipline caught both before anything trusted the tool.
3. **Two blind fetches of rendered HTML.** My first two regex checks searched for `<code>` while
   GitHub emits `<code class="notranslate">` — both reported "no code spans" and I nearly
   concluded wrong; raw-HTML inspection settled it. The lesson I already banked
   (verify tool output against a fresher source) had to be re-learned at the tooling level.
4. **`FEATURES.md` edit-failure loop.** Hit the stale-read gotcha THREE times (two rejects + a
   wasted multiedit) even though the resume context names it explicitly. The file hadn't even
   changed — the tracker was stale from the prior session. Should have gone straight to
   view-then-edit in one breath.
5. **Two 422s on the ruleset API** (kv-array encoding mangled `rules` into an object;
   `target=push` rejects `ref_name` conditions). Read-the-enum-first would have saved both.
6. **The lychee "0 links" false green.** The offline pass reported "0 Total" and I initially read
   it as clean — the docs reference paths as CODE SPANS, not hyperlinks, so lychee scanned
   nothing. I replaced it with an explicit path-existence gate (and the later online pass then
   caught the real SECURITY.md 404). A gate that scans nothing must never be reported as a pass.
7. **Nothing here corrupted the repo** — every failure was caught by a verify gate this session
   built or ran, and the working tree is clean. But the pattern is consistent: I wrote/ran first
   and let the gate catch what a 30-second re-read would have prevented.

## e) WHAT WE SHOULD IMPROVE

1. **Regression-first for tooling:** run the new checker/annotator against a known artifact
   BEFORE declaring it done (F6.3 did this — make it the default order, not a numbered step).
2. **Read the producer's output format before writing its validator** (the annotator's
   marker-in-first-cell format was one `grep` away).
3. **One breath: view-then-edit** after ANY external write (daemon, dprint, script) — no
   interleaved bash between the view and the edit.
4. **APIs: check the enum/error docs before the first call** (ruleset targets, markdown endpoint).
5. **Gates must assert they measured something** (lychee 0-links ≡ scanner Files: 0 — same class
   as the gosec lesson already in AGENTS.md-global).
6. **Record CHANGELOG entries as work lands**, not batched at F10.2 (the batch is why entries
   are missing now).
7. **Check where skill symlinks resolve BEFORE editing** (`readlink -f` first would have routed
   F6.4 straight to the upstream repo and saved a read-only-fs error).
8. **Example-driven API naming:** the plan itself said `ExampleNewStore`, which `go vet` would
   reject (no `NewStore` exists) — plans should be vet-checked for identifier references at
   writing time, not execution time.
9. **The bench fixture cost (10k corpus ≈ 37s build per count) should be documented in-file**
   so future runs don't look hung.
10. **Plan-vs-reality drift log:** 2 of 46 fine tasks deviated (F3.2 renamed to
    `ExampleOpenStore`; F6.4 rerouted upstream). Deviations were right, but they should be
    written back into the plan file or its archived copy at execution time, not only reported.

## f) UP TO 50 THINGS TO GET DONE NEXT (ordered: this-session completion → owner gates → trigger-gated backlog)

1. Run F10.1: full pristine suite over the final tree (`env -u GOEXPERIMENT GOTOOLCHAIN=go1.26.7 go build/vet/test`, golangci, `go mod tidy && git diff --exit-code go.mod go.sum`, dprint check).
2. Decide push shape for the 10 daemon commits: push as-is vs squash into per-task detailed commits (local-only rewrite, owner call).
3. Push + watch `go-test` to green on the new HEAD (F10.3).
4. Complete CHANGELOG `[Unreleased]`: examples, `Build` doc rule, benchstat/round-trip upgrade, README/SECURITY fixes, tooling work (needs a commit first → daemon race).
5. Add the M6 tooling assets note to docs-health `SKILL.md` (check-rows.py exists; the skill body doesn't mention it yet).
6. Commit the docs-health tooling changes in `~/projects/SKILLS` (annotate-rows.py guard + dotted IDs, check-rows.py, test rewrite) — currently uncommitted there.
7. Commit the go-cqrs-lite `modules.md` wording fix in `/home/lars/projects/go-cqrs-lite` (owner call; unrelated uncommitted workflow changes live in that tree).
8. Bump the flake input that pins the go-cqrs-lite skill so the installed (store) copy receives the F6.4 fix.
9. Send the offered PR to tobi/qmd#959 before the 2026-09-22 window closes (still silent as of today).
10. Send the offered patch to charmbracelet/crush#3846 (issue, not PR) before 2026-09-22.
11. Owner gate Q1: license posture → then apply the CONTRIBUTING inbound-grant line.
12. Owner gate Q2: cut v0.2.0 → tag → `release.yml` fires → verify the release notes extraction end-to-end.
13. After v0.2.0: verify the new examples render on pkg.go.dev (the whole point of Q2).
14. Live smoke test once `GRAPHRAG_LIVE_EMBED_*` creds exist.
15. CV repo push (carries the `16fc16c7` extraction annotations).
16. CV-side cross-link decision for the metaengine report (owner taste).
17. Resolve the one lychee-reported redirecting URL to its final target.
18. benchstat-grade the round-trip bench (currently count=3, ±∞ spread): `-count 10` when a slow run is acceptable.
19. Document the 10k fixture-build cost (~37s per count) in the bench file so runs don't look hung.
20. Add `BenchmarkSimilarPairs/docs=100` for size symmetry with Build/Search.
21. Add tidy-drift + dprint check to `.githooks/pre-push` (it currently guards only the pristine build).
22. Record the benchstat invocation + pinned `golang.org/x/perf` version in AGENTS.md Commands.
23. Confirm dependabot's grouped Actions updates cover `release.yml`'s pinned checkout.
24. After push: `gh workflow view release` to confirm the workflow registered (F9.2's deferred half).
25. Verify CI's dprint step is happy with `release.yml` formatting (does dprint.json cover yaml?).
26. Compile the ADR sketch as a permanent (skipped or compile-only) test so it cannot rot against core types — ADR follow-up.
27. When any adapter work starts: define the `VectorIndex` filter-semantics doc contract first (ADR §4.1 promises adapter-documented pre/post-filter).
28. Cross-link the seam ADR from ROADMAP's non-goals (external-DB row) for symmetry.
29. FEATURES.md: add DONE rows for the round-trip benchmark + release automation if they count as features (owner taste).
30. Owner-decision package: add a pointer to the ADR under Q2 (release notes automation now exists).
31. gitleaks/codespell on-demand pass over the new docs (buildflow `-s`, this repo lints ad hoc).
32. ANN spike (`hannoy` / `sqlite-vec` / in-process HNSW) — trigger: ~50k corpus; the ADR is its umbrella.
33. Shared ANN benchmark harness (recall + p50/p95 vs the scan) — ROADMAP Theme 1.
34. Metadata-filtered ANN semantics (pre/post-filter contract refinement) — ROADMAP Theme 1.
35. ANN-winner ADR (choice, migration, `RelationSimilar` derivation at ANN scale) — under the seam ADR.
36. Incremental indexing design (upsert/delete vs `ReplaceGraph`) — ROADMAP Theme 1.
37. Index-backed Searcher constructor shape spike (`NewSearcherWithIndex`?) — first seam implementation step.
38. Promote the linear scan to the reference `VectorIndex` implementation (ADR migration step 1) when implementation is approved.
39. Wire-contract golden tests (`Hit`/`SearchResult` JSON frozen) — v1 approach work.
40. v1 freeze of the §5 interface signatures (ADR migration step 4).
41. Metrics/tracing seams (decorate Build/Search without wrapping) — ROADMAP Theme 3.
42. Store compaction + cache-eviction policy — ROADMAP Theme 3.
43. Backup/export story beyond "delete and rebuild" — ROADMAP Theme 3.
44. Dgraph trigger-gated sweep (if the ~50k/multi-hop trigger fires): §8 spike, dgo v240/v250, license history, second community source, release tripwires, CVE review, importer-count recheck, raw-notes appendix, lychee JSON-link tolerance.
45. Second SDK consumer → publish `go-graph-rag/metaengine` as an isolated adapter module (ADR §4.2/§5.2).
46. Corpus re-ask at the ~2027-09 horizon (owner decision recorded).
47. go-cqrs-lite doc-consistency sweep: after the modules.md fix lands, re-verify `system`/metaengine markers across FEATURES.md vs modules.md vs module code.
48. Searcher/GraphStore composition test fixture (fake GraphStore in testdata) to keep the seam honest before adapters exist — optional, cheap.
49. Plan-file hygiene: write the two deviations (F3.2 rename, F6.4 reroute) into the plan's archived copy when archiving.
50. Close the qmd/crush watch row at window end (2026-09-22) with a send-or-drop decision either way.

## g) QUESTIONS ONLY THE OWNER CAN ANSWER

_ANSWERED (inline, strikethrough per policy):_

1. ~~**Push shape for the 10 daemon commits?** The daemon committed everything as heuristic
   `chore: auto-commit` messages (578 insertions, 13 files, all verified present). Push them
   as-is, or squash/reword into per-task detailed commits first (local-only rewrite) before
   pushing?~~ → **ANSWERED: leave history intact, push as-is.** Rules: never rewrite what the
   daemon committed; the heuristic trail is expected behavior.
   _Executed: pushed `bd14092..f79c58a` (13 commits incl. the daemon's sweep of the CHANGELOG
   and this report; pre-push pristine guard passed); `go-test` run `35099876210` GREEN on the
   new HEAD._
2. ~~**Commit the two external skill-repo changes?** (a) `~/projects/SKILLS` docs-health tooling
   (guard fix, dotted IDs, check-rows.py, test rewrite); (b) go-cqrs-lite `modules.md` wording
   fix — and bump the flake input so the installed skill actually receives it. Yes to both,
   which, or leave for their own sessions?~~ → **ANSWERED: commit both in their own repos now**
   (go-cqrs-lite gets ONLY the modules.md change — its pre-existing workflow changes stay
   untouched), plus the flake-input bump so the installed skill receives the fix.
   _Executed: SKILLS repo — the daemon had already committed the bulk (`ff8cda2`, `db47b0e`);
   my follow-up fix round landed as `0d1aca6` with a detailed message. go-cqrs-lite — my
   explicit commit was blocked by its pre-commit `nix fmt` hook (which also flagged 17
   unrelated in-progress vector-file changes from a parallel work stream, left untouched);
   the daemon committed modules.md in `b13e17bba` instead. No lock-file bump exists: the
   installed skill is linked by the go-cqrs-lite devShell hook to the flake's own source, so
   it was re-linked to the fresh source (`3bdhizr9…`) and the installed modules.md now reads
   true (verified)._
3. ~~**The offered upstream patches (qmd#959, crush#3846): send them now, or hold past the
   2026-09-22 window?** Both upstreams are still silent after ~24h.~~ → **ANSWERED: hold past
   the window (no grace-period reminder pings); if asked "did you send it", the patch exists in
   the issue text — that's enough.** Row stays OPEN past 09-22 as a deliberate owner choice.

---

_F10.1 ran ALL GREEN (pristine build/vet/test, golangci 0 issues, tidy drift clean, dprint
check clean) before the push; F10.3 is COMPLETE: master = origin/master = `f79c58a`, CI green
(run `35099876210`). M1–M10 all executed._
