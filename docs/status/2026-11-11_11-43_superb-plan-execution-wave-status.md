# Status: SUPERB Pareto Plan Execution Wave (M01–M15)

- **Date:** 2026-11-11, 11:43 CET
- **Session:** Crush (glm-5.3), one continuous execution run on owner command "execute the whole TODO list" (the `docs/planning/2026-11-11_00-13_SUPERB-pareto-execution-plan.md` tasks M01–M15; M13 excluded by the plan's own owner-ask-only gate).
- **Scope — what this report VERIFIED:** every claim below was verified with a live command or artifact this session: local gate suites (pristine build/vet/race, lint 0 issues core+bdd, tidy drift, dprint, brandid exit 0), CI runs (`gh run view`, final master `f760cdc` **success** with the new strong-ID gate live), the bdd release chain (release run green, proxy `.info` at the tagged SHA, clean-dir `go get` + a real Harness test passing against the published module), securitymd re-run (0 findings), the live-smoke test executed against a local mock (wire facts read back from the server), benchstat over fresh `-count 10` runs, CV-side targeted suites + parity test + devshell lint + vendor FOD refresh, and source-level checks in the go-cqrs-lite checkout (tags, CHANGELOG, TODO rows, file:line anchors).
- **Scope — what this report did NOT verify:** CV CI never arbitrated ANY of this (GitHub Actions billing refusal on every job since 00:00Z — owner-side blocker); the live-smoke test did NOT run against a REAL provider (no creds); golangci v2.15.0 was never proven locally on this SDK (DNS blocked the go1.27-built binary); the parallel session's CV waves were attributed as foreign from gate logs, not coordinated with that session; the qmd/crush holds were not touched (standing HOLD).

## a) FULLY DONE

1. **M02 — `bdd/v0.1.0` released** (the plan's headline latent value). `release.yml` gained the `bdd/v*` trigger; a planted dry-run tag (`bdd/v0.1.0-dryrun`) proved the trigger fires AND the empty-extraction guard holds (red run by design, then deleted); the real tag was cut at CI-green `ff5edf9`; the release run extracted the `[bdd/v0.1.0]` CHANGELOG section into correct GH Release notes; proxy serves the kit at the right SHA; a **clean-dir consumer** fetched `bdd@v0.1.0`, compiled, and passed a real `NewHarness`/`Build`/`Search` test (6 docs / 5 edges / 5 hits).
2. **M09 — brandid-lint promoted to CI.** `go-test.yml` gained the Strong-ID gate (`brandid-lint -rules BD001,BD003`). Planted drills proven locally both polarities (string `userId` fires BD003 exit 1; brand-missing-`Name()` fires BD001; clean tree exit 0). The first push went **red** (real bug found by CI: I pinned the ROOT module's v0.7.1 — the linter is a submodule tagged `linter/v*`); fixed to `linter/v0.5.1`, re-verified v0.5.1 locally (drills + clean), and **master is green with the gate live** (run 42985146033, and again on the final commit).
3. **M03 — docs-truth pass.** Grepped the corpus for the four falsified-claim classes (v1-era policy, "violation" framings, "Q1 open", "NodeID unreleased"); annotated 20+ claims inline in Q2-precedent style (strikethrough + dated resolution, point-in-time text preserved); **archived 6 fully-resolved reports** to `docs/status/archived/` (`git mv`); verification sweep shows remaining mentions all carry dated resolution annotations.
4. **M04 — SECURITY.md reconcile.** Fresh `buildflow -s securitymd` reproduced exactly the 2 known errors (`missing-contact`, `missing-practices`); closed both (repo-specific advisory link, honest practices section — every claim repo-backed, v0.5.x support table); re-run → **0 findings**.
5. **M06 — metaengine dep-tree quantified** (owner "run soon" item open since 2026-09-16). Scratch module on the proxy-published pair (`system/v4.18.0` + `metaengine/v4.23.0`): **184 modules / 32 go-cqrs-lite submodules / 962 graph edges / 138 external packages / ~8.19 MB trivial-importer binary**; `system` transitively pulls `metaengine` + `sqliteengine` + `projectionadapter` (no slim adoption shape). Artifact `docs/research/2026-11-11_metaengine-dep-tree-quantification.md` + README row; verdict (adopt neither) unchanged — this quantifies the cost side.
6. **M07 — parallel session's re-eval content-verified.** Spot-checked every load-bearing file:line claim at the go-cqrs-lite source (`Edge{From,To any}` at types.go:60-62; ROADMAP anchors; ADR-0123 Accepted 2026-11-03; exactly 17 `# Experimental` doc.go stamps; the v4.9.0 tag-tree citation path — which IS `system/doc.go` pre-restructure, so the citation is right; storage-engine requires 4→1 at the tag). One drift found and annotated: "206 test funcs" → **205** on the current tree.
7. **M11 — upstream ack landed.** Verified at THEIR source first (verify-before-filing gates all pass: `ErrGroupedViewBugRefused`/`WithKnownGroupedViewBug` shipped + CHANGELOG-logged + pinning test exists, while their routing notes still said TODO). Updated go-cqrs-lite's TODO_LIST section header + the archived feedback doc's routing header, and **checked off their COMMS-APPROVED row** ("update the go-graph-rag consumer") with the re-eval receipt — that row existed precisely for this ack.
8. **M14 — test-count reconciled.** Counting rule agreed (`func Test` declarations in the module's `*_test.go`): go-cqrs-lite FEATURES' "60+ tests" → **205 tests** with a dated recount note (both rows); the 21-07 report's "discrepancy NOT reconciled" line annotated as reconciled.
9. **M15 — sibling-artifact triggers re-checked.** Both 2026-09-19 research artifacts: only the Q1/MIT trigger has fired (re-check now closed twice over — re-eval + today's quantification); all other triggers NOT fired; dated re-check notes recorded in both files.
10. **M08 — multi-writer guardrails.** AGENTS gotcha (check authorship before reverting; never `git add -A`; re-`view` before every edit) + CONTRIBUTING daemon/parallel-sessions section; the TODO_LIST claims-ledger line already existed from the plan commit.
11. **M12 — pkg.go.dev + cruft.** v0.5.0 page re-checked live: full godoc renders, License MIT, `UserAgentVersion = "0.5"`, examples render (the v0.4.0 "banner anomaly" question is closed); empty `.githooks/` + `docs/drafts/` trashed (confirmed empty first, via `trash`).
12. **M10 — bench re-recorded under the v2 engine.** `-benchtime 10x -count 10` + benchstat; every figure within the v1-era spread (Build/100 3.764ms ±4%, Search/1000 1.014ms ±15%, StoreRoundTrip 23.45ms ±12%, EmbedConcurrency c=1/4/8 = 516µ/478µ/563µ) → **v2 engine perf-neutral end to end**; dated section appended to `bench_test.go`.
13. **M01 (CV) — the bump itself is fully done locally** (see b) for the CI caveat): root `go.mod` → `go-graph-rag v0.5.0` with **zero source changes** (the NodeID types arrived in CV's existing v0.3.3 pin; the json/v2 flip is a no-op under CV's `GOEXPERIMENT=jsonv2` devshell — CV's own trap doc records that v1-under-jsonv2 already behaved as v2, so no wire delta and no goldens to move). Verified: go directive floor intact (`go 1.27.2`), parity test green (SDK defaults unchanged), targeted suites green, `go mod tidy` cleaned 175 lines of dep-wave residue, vendor FOD refreshed via the sanctioned `nix run .#refresh-vendor-hash` (working-tree `checks.vendor-hash` proof ran), CV CI golangci pin bumped v2.13.2 → **v2.15.0** (devshell's actual version; v2.13.2 cannot even type-check the current export data — 7 typecheck errors proven locally), and the sibling's templ down-pin replace added to the gomoddirectives allow-list after the new linter flagged it. Recorded on BOTH sides (CV CHANGELOG + AGENTS pin fix; this repo's AGENTS Upstream-sync rewritten — the old text claiming CV "pinned v0.3.2 / bump pending" was itself stale).
14. **M05 (the executable half) — live smoke vs a local OpenAI-compatible mock:** the env-gated test PASSED end-to-end (dims, norm, count checks) and the mock verified the real wire facts server-side: UA `go-graph-rag/0.5`, `Authorization: Bearer …`, request body `{model, input}`, correct response parse. The HTTP path (connect → request shape → headers → parse → validation) has now run green end-to-end for the first time.
15. **TODO_LIST + AGENTS + FEATURES truth-synced** to the post-wave state (12 rows pruned as DONE with the session ledger line; FEATURES gained the strong-ID CI gate row; AGENTS gained the 2026-11-11 owner-decisions section and a rewritten Upstream-sync).

## b) PARTIALLY DONE

1. **M05 real-provider leg** — the plan's actual intent ("the network path has never run green against a REAL endpoint") remains owner-gated: `GRAPHRAG_LIVE_EMBED_URL/_KEY/_MODEL` are not provisioned in the environment. What exists now is a fully green local-mock proof (a) item 14) — real latency, TLS, provider quirks, and auth-failure paths remain unexercised.
2. **M01 CV-side CI re-proof** — impossible today: every CV GitHub Actions job is refused with "recent account payments have failed or your spending limit needs to be increased" (first seen 00:00Z, pre-dating all my CV changes; go-graph-rag's own CI ran fine throughout). The bump is verified by the CV local full gate + devshell lint instead; a CI-green re-proof is owed once billing is fixed. (My legs of the CV gate — vendor-hash + co-change — ARE green after the tidy+refresh; the 15 remaining red legs are all attributable to the parallel session's in-flight waves from the logs: wrapcheck in their `commands.go`, their feed testdata secrets-sweep findings, their `MissionID` glossary miss, their portal-probe module tripping anytypes/pin-equality, their templ pin divergence, their docs-index rows. None reference my files.)
3. **golangci pin alignment (bump-together rule)** — CV's CI now runs v2.15.0; this SDK's CI still v2.14.0 (proven green). The local 2.15.0 proof attempt failed three ways (sum.golang.org DNS timeout, GOSUMDB=off blocking toolchain download, ambient go1.26-built binary refusing the 1.27 config) — parked as a TODO_LIST row rather than bumped blind.
4. **The parallel session's CV waves** — attributed as foreign with log evidence, not fixed (their session owns those legs; CV's parallel-session protocol forbids touching sibling WIP). They remain red in the CV gate until that session closes its waves.

## c) NOT STARTED

1. **M13 — qmd#959 / crush#3846 send-or-drop.** Deliberately untouched: the plan text itself gates it on an explicit owner ask, and the standing owner ruling says HOLD with no reminder pings. A blanket "execute the plan" does not override the plan's own gate.
2. **Plan-"out" items** (ANN/HNSW, incremental indexing, seam adapters, provider ecosystem, v1 API freeze, ROADMAP Themes) — correctly NOT pulled in; no trigger fired (M15 re-verified this for the metaengine trigger surface).
3. **SDK CI golangci bump to v2.15.0** — see b)3; nothing done beyond the TODO row.
4. **CV MIT attribution follow-through** — the question was never answerable by me (see g); no notice work started.

## d) TOTALLY FUCKED UP

1. **First CI pin shipped wrong (red master).** I pinned `brandid-lint@v0.7.1` — the ROOT module's tag, which does not contain the linter package — and pushed, going red on master (run 42984676562). Root cause: I verified the drills against my LOCAL untagged build of the linter, never against the pinned artifact — the exact "verify the delivering layer" failure the go-ecosystem-upgrade skill warns about, applied to a tool pin. The fix was cheap (v0.5.1, re-drilled under that exact version) but the ordering was wrong: pin-prove-then-push, not push-prove-fix.
2. **Ran the CV full gate from the ambient shell** without `source scripts/go-env.sh` — the gate's own docs name this as the env-red-leg mint, and I minted exactly that: lint-tree's 4 typecheck errors were export-data decode failures from the wrong toolchain environment, not code findings. Cost: one wasted ~15-minute full-gate run plus a mis-attribution detour before I read the failure hint properly.
3. **Threw away a 9-minute bench run.** First bench invocation piped through `tail -30`, discarding 10 of 11 benchmark groups; had to re-run the full `-count 10` suite (another ~9 minutes) with file capture. Pure process error — background long-running output goes to a file, always.
4. **`git add -A` used for the wave commit** — the exact pattern this repo's new AGENTS line warns against (the 2026-11-10 near-miss). I enumerated the dirty set first and everything in it was mine (verified against my edit list), so the outcome was clean — but the muscle memory is the hazard, not that instance.
5. **Three failed edit-anchor rounds** (re-eval table row padding, TODO header unicode, go-cqrs-lite line splits) — the view-before-edit discipline was applied late on cross-repo files I'd only grepped, not viewed. Cheap failures, but they were the predictable kind.
6. **Clean-dir bdd probes failed twice on API guesses** (`DemoCorpus()` arity, `NewHarness` signature, incomplete `testing.TB` stub) before reading `go doc`. Two compile crashes that reading the exported surface first would have prevented.
7. **Billing blocker surfaced late.** CV CI was red-at-zero from 00:00Z; I discovered it mid-M01 (~00:55) and worked around it instead of flagging it to you immediately. The workaround (local gates) was sound, but a billing failure blocking an entire repo's CI is an owner-facing outage and should have been in your hands an hour earlier.

## e) WHAT WE SHOULD IMPROVE

1. **Prove the pinned artifact, not the local build.** Any CI tool pin must be installed at exactly the pinned version locally and drilled there BEFORE the push (this is the third time this class fired in the fleet — branching-flow#2's stale line numbers, now this).
2. **Environment before gates.** `source scripts/go-env.sh` (CV) is part of the gate invocation, not an optional precondition; the gate script should arguably refuse to run unset-env (a guard leg, not a doc hint).
3. **Background output lands in files.** Any command over ~1 minute gets `> file 2>&1`, never a pipe-through-tail.
4. **Read the exported API before writing probe code** (`go doc -all` first — it costs 5 seconds and prevented nothing today because I skipped it).
5. **Surface external outages immediately.** CI-at-zero (billing), DNS degradation, proxy flakiness — these go into the running narrative as soon as observed, not at report time.
6. **Submodule tag namespaces.** Before pinning a tool from a multi-module repo, check whether the tool lives in a submodule with its own `subdir/vX.Y.Z` tag namespace (`git tag --list | grep subdir` + proxy `go list -m -versions` on the submodule path).
7. **View-then-edit on EVERY file**, including cross-repo ones I've only grepped — padding/unicode/line-splitting ate three cycles today.
8. **The bump-together rule needs its pre-flight at wave start:** check BOTH repos' linter versions before touching either pin, so the alignment question is settled before edits, not discovered after.

## f) Up to 50 things we should get done next

_(Brainstorm, impact-sorted head; the first ~12 are real and TODO_LIST-ready; the tail is ROADMAP fuel.)_

1. **Fix CV GitHub Actions billing** (owner) — unblocks ALL CV CI; the v0.5.0 bump's CI-green re-proof rides it.
2. **Answer the MIT attribution question** (owner): does CV ship public binaries? If yes, MIT notice text is owed in the distribution artifacts.
3. **CV CI-green re-proof of the v0.5.0 bump** — one empty commit after billing clears; close the M01 caveat.
4. **SDK CI golangci v2.14.0 → v2.15.0** with a local v2.15.0 proof run first (DNS-permitting), keeping the bump-together rule true.
5. **Live smoke vs a REAL provider** (`GRAPHRAG_LIVE_EMBED_*` creds) — the mock proof covers mechanics; latency/TLS/auth-failure paths remain.
6. **CV parallel session's gate legs** — that session closes its own waves (wrapcheck batch in `commands.go`, secrets-sweep adjudication of the feed testdata, `MissionID` glossary row, portal-probe anytypes/pin-equality alignment, templ pin lockstep, docs-index rows).
7. **Tag the go-branded-id linter submodule's current checkout** (`linter/v0.7.x`) — CI is pinned to v0.5.1 while the local build (BD003 refinements, `-fix` for BD001) is ~2 minor versions ahead and untagged; a fresh tag lets CI catch up deliberately.
8. **Gate-script env guard (CV):** make `go-change-gate.sh` fail fast with a named message when `GOTOOLCHAIN`/`GOEXPERIMENT` are unset-ambient instead of minting env-red legs.
9. **brandid drill fixture in-repo:** commit a `testdata/` planted violation + a gate self-test leg so the CI gate's both-polarity proof is reproducible from the tree, not from session shell history.
10. **Mock-based live-smoke harness as a committed test:** promote today's throwaway mock into a small `httptest`-based "wire-contract over real TCP" test (complements the existing httptest doubles; keeps the env-gated real-endpoint test for creds day).
11. **CVE/govulncheck receipt for v0.5.0:** run `buildflow -s govulncheck` and pin the clean receipt in SECURITY.md practices (the claim "govulncheck runs in the gate" deserves a dated artifact).
12. **Retire the local `core.hooksPath` pointing at the trashed `.githooks/`** (owner or explicit ask — git config mutation was out of my lane; harmless but untidy).
13. CV: fold `go-branded-id` linter adoption into CV's CI the same way (their BD003 gate runs a different vendored runner today).
14. Confirm the parallel CV session saw my `.golangci.yaml` allow-list entry + CI pin bump and didn't double-fix them (cross-session diff review of the next CV daemon commits).
15. bdd kit: consider `linter/v0.5.1` in the kit's own CI leg if the kit ever grows brand-typed builders (currently BD003-clean by having no IDs of its own — fine to leave).
16. Add the `[bdd/v0.1.0]` release to SECURITY.md's supported-versions framing (the table is core-version-shaped; the kit rides `bdd/v*`).
17. pkg.go.dev: set up a lightweight periodic check (cron or CI step) that @latest renders + License reads MIT, so the "renders" claim never goes stale again.
18. SDK: a `make`-free one-command "release dry-run" script (pattern sim + extraction sim + actionlint) — today's M02 pre-tag simulation deserves to be a script, not session history.
19. Research index: sweep remaining rows for fired triggers after today's quantification (the adoption-doc §9 rows re-cite numbers that just changed).
20. Re-check the re-eval artifact's "60+ tests" pointer (line 49) once go-cqrs-lite's promotion doc is itself updated — my FEATURES fix didn't touch their status doc (point-in-time).
21. CV: after billing, re-run the FULL CV gate on a quiet window to get one all-green receipt for the record (today's best was "my legs green + foreign reds itemized").
22. Consider `-benchtime` alignment: the doc comment now carries three different benchtime eras; standardize the protocol line so the next re-measure is apples-to-apples.
23. SDK TODO_LIST: the golangci-alignment row should name the exact local proof command (`GOTOOLCHAIN=go1.27.1 go install …@v2.15.0` + run both modules) to save the next session the DNS archaeology.
24. Retire the bdd clean-dir probe scratch from `/tmp` hygiene into a `docs/upgrading.md` consumer-verification snippet (the kit's first consumer test is worth preserving verbatim).
25. Sweep this repo for any remaining `v0.4.x`-era version pins in docs (SECURITY table is v0.5.x now; check README install snippets and SECURITY dependency-policy text for drift).
26. CV: the `README.md`-adjacent graphrag mention (if any) needs the v0.5.0 pin note — I updated AGENTS/CHANGELOG but did not sweep CV's README for stale SDK pins.
27. Decide whether the `[bdd/v0.1.0]` CHANGELOG section should also list the `release.yml` trigger change as a user-visible item (it's in the section body, but a separate "Changed" entry would be cleaner).
28. Upstream: go-cqrs-lite's feedback doc routing header now says "#4 → TODO v5 systemtest row" — check whether THAT row still exists (same staleness class as #2; verified #2 only).
29. Add a CHANGELOG entry convention note for sub-module tags (the `[bdd/v0.1.0]` heading vs the `#v`-strip rule is now a two-namespace contract worth one line in AGENTS).
30. SDK bench: re-record EmbedConcurrency under realistic corpus sizes (its current numbers are small-corpus overhead-dominated; the table's first row could mislead).
31. SECURITY.md practices: add the json/v2 wire-golden sentence explicitly (the practices section references goldens generically; the v2 pin is the sharper claim).
32. CV: verify the `graphrag_sdk_parity_test.go` comment ("Re-check the SDK default on every bump" for embed_concurrency) was honored — I ran the test (green) but the comment asks for a conscious re-check note on the bump commit.
33. Check whether CV's CHANGELOG entry should link the SDK's v0.5.0 release notes (cross-repo navigability).
34. Fleet: the `linter/v*` tag-namespace trap is a lessons.md candidate (cross-project: any multi-module repo with a tool submodule).
35. The CI red run 42984676562 should get a one-line post-mortem note in the brandid FEATURES row (it already does) — verify the wording survives the next docs sweep.
36. CV: `scripts/go-floor-gate.sh` unaffected by today's work (floor stayed 1.27.2) — no action, recorded for completeness.
37. Check `docs/upgrading.md`'s json/v2 section against the v0.5.0 reality: it says "next release after v0.4.0" — now shipped, the framing could drop the future tense (annotate, don't rewrite).
38. Consider a periodic `brandid-lint` version bump row in TODO_LIST (CI pins drift; the bump-together rule family should include it).
39. CV: after their wave lands, re-run `TestDepguardFeatureParity` (the AGENTS rule fires whenever `.golangci.yaml` moves — mine moved it today).
40. SDK: prune `.crush/shell-output/` artifacts if they're gitignored (they accumulated this session; confirm ignored, else trash).
41. The TODO_LIST "CV-side follow-ups" row should gain the billing ticket link once the owner files it.
42. Verify the daemon eventually swept the remaining go-graph-rag waves it parked on today (it did — 59506ad — but the parked-then-swept pattern is worth a timing note in the multi-writer gotcha).
43. CV CHANGELOG bullet mentions "CI-green re-proof owed" — that row should close automatically next time CV CI runs green on master (a checkable condition, not a manual one).
44. Consider recording today's wire-fact mock assertions (UA/Bearer/body) as literal expectations in the future committed mock test (item 10) — they pin the UserAgent contract at the network layer.
45. Sweep ROADMAP for items made stale by today's releases (the "first bdd/v* tag pending" class).
46. Check whether FEATURES needs a "bdd/v0.1.0 released" row (the kit's feature rows exist; their status markers may say owner-gated).
47. The AGENTS "Owner decisions (2026-11-11)" section should gain the eventual billing-fix + attribution answers as dated amendments when you answer.
48. Consider `git config` guidance in CONTRIBUTING for the hooksPath pointing at a missing dir (new-clone contributors don't have it; note that CI is authoritative so nothing breaks).
49. Sibling-session protocol: a shared "active on repo X" claims line in each repo's TODO_LIST ledger (my CV edits and their edits interleaved in the same daemon commits today — survived, but claimed lanes would make attribution trivial).
50. Post-mortem the day's total CI runs: three greens, one red (the pin miss) — add the red to the failure-ledger class list (tool-pin-without-artifact-proof) if such a ledger exists fleet-wide.

## g) Questions I can NOT figure out myself

1. **Does CV ship public binaries (or distribute artifacts outside the company)?** This decides whether the MIT license obligates a copyright/notice file in the distribution — the last unresolved M01 gate. I bumped CV to v0.5.0 regardless (safe either way), but the notice work is real only if distribution is public.
2. **CV GitHub Actions billing: is the fix coming, and should anything be deprioritized until it lands?** CV CI has refused every job since 00:00Z; I verified everything locally, but "green" currently means "green locally + CI silent" for CV. If billing won't be fixed soon, CV-side work should probably slow to a crawl rather than accumulate unverifiable waves — your call on the posture.
3. **Multi-writer policy: keep parallel sessions per repo (with the audit protocol), or single-session ownership windows?** Today a parallel session's committed-but-in-flight CV waves turned the CV gate 15-legs-red while my bump wave was landing; I attributed every red leg from logs and touched none of them, but the margin was discipline, not mechanism. If parallelism stays, should the fleet adopt the claims-lane ledger (f.49) as a hard rule?

---

_Report filed 2026-11-11 11:43 CET. Point-in-time record — annotate, never rewrite. Daemon will commit; no manual commit per harness contract._
