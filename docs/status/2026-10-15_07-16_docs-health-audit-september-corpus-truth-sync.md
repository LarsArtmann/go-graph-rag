# Status: Docs-health AUDIT — September corpus annotate/archive + living-docs truth sync

| Field   | Value                                                                                                                                                                                                                                                                                                                                                     |
| ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Date    | 2026-10-15 07:16 CEST                                                                                                                                                                                                                                                                                                                                     |
| Scope   | Single command-scope on explicit instruction: view ALL `**/2026-0*` files, execute the docs-health skill (AUDIT = BUILD + HARVEST + VERIFY + ANNOTATE + ARCHIVE), make the six living docs superb, archive fully-done files                                                                                                                               |
| Verdict | **All six living docs re-synced to the BDD/NodeID era, the SUPERB plan annotated + archived with every row verdict-carrying, three 09-19 status reports annotated inline (65 verdicts), one stale research claim corrected, all gates green. One mid-session file-wipe (Git Town WIP) root-caused, recovered in full, and recorded as an AGENTS gotcha.** |
| Format  | `.md` at the operator's explicit path demand (status-report skill's HTML default overridden per its escape clause; standing dispatch exception applies)                                                                                                                                                                                                   |
| Commits | Daemon-collected: `b3af4e7`, `c5e21a2` (plan archive), `075d849` (living docs), `a7e8ef6` (report annotations + research fix), `e7a6d8f`, `1d52a81` (final sweep + AGENTS gotcha). Tree clean at report time.                                                                                                                                             |

## The one-paragraph version

Docs-health AUDIT over all 21 `2026-0*` files plus the six living docs. VERIFY confirmed the BDD-era claims against code and the world (master CI green at `b3bf8fe`; `go-test.yml` pins 1.27.1 with the four bdd steps; `release.yml` still `v*`-only; dependabot has a `/bdd` group; `NodeID` at `graph.go:61`; both erraudit findings still present at the cited lines; benchstat installed again; wire goldens intact). ANNOTATE struck/routed 119 verdicts across four point-in-time files and archived the SUPERB plan; HARVEST rebuilt TODO_LIST with 18 open rows and zero historical material; BUILD refreshed README, AGENTS, FEATURES, ROADMAP, DOMAIN_LANGUAGE, SECURITY.md, CHANGELOG. The session's one real incident: Git Town stashed my TODO_LIST/ROADMAP/17-03 writes into a dangling WIP commit and reverted the files mid-session; root-caused via `git fsck`, recovered with `git restore --source=08f450e`, verified stable across two sweeps, and recorded as an AGENTS gotcha.

## Brutal self-review (folded in per the 06-14 precedent)

**What did I forget?** The 17-03 g-section got no annotations — I decided mid-flight it duplicated 14-04's g-verdicts and never stated the decision (defensible, but silent). I also nearly trusted the stale AGENTS.md snapshot from session context instead of the disk file, and my first dprint/DOMAIN_LANGUAGE edit used a from-memory `||`-prefixed old_string that didn't exist.

**What could I have done better?** Marker-count sweeps after every annotate batch instead of at the end — the Git Town wipe sat unnoticed for ~10 minutes because I verified late. On the `repair` tool failure I burned a cycle on a sed workaround before reading the tool source; `prose` mode was the right tool for §5 from the start.

**What could I still improve?** The `grep -L ... && FAIL || OK` gate produced an ambiguous verdict once (the exact 14-04 d3 lesson) — unambiguous verdict forms should be the default, not the recovery. Git Town's WIP-stash behavior should either be configured off for repos with an auto-commit daemon or become a standing post-write `git status` reflex (now an AGENTS gotcha).

---

## a) FULLY DONE

1. **Skill compliance first**: docs-health SKILL.md + all six references (verify-checklist, health-report-format, harvest-guide, doc-ownership, build-guide, agents-quality-guide, resolving-items) loaded before acting; status-report skill loaded before this file; every annotate batch dry-run before the real write.
2. **Full read/verify of all 21 `2026-0*` files**: 19 non-archived (status ×11, planning ×3, research ×4 + index) read or verification-inspected, plus the 2 already-archived planning files (marker-verified: 68/57 strikes). Nothing from memory.
3. **VERIFY against code and the world, findings acted on**: master CI **green at HEAD** (`b3bf8fe` run success; the red 00:11 run was the v0.3.2 tag commit — matches the 05-18 report); `go-test.yml` pins `go-version: "1.27.1"` + `GOTOOLCHAIN` with the four bdd steps; `release.yml` trigger confirmed `v*`-only; `.github/dependabot.yml` **does** cover `/bdd` (closes 05-18 c3); `NodeID` confirmed at `graph.go:61`; erraudit findings confirmed live at `embed_openai.go:334` (`errors.As`) and `embed_openai_test.go:249` (`errors.Is`); benchstat **installed** at `~/go/bin/benchstat` (unblocks 17-03 b2/f15); wire goldens = 5 files, untyped-literal compile proof in tests; `UserAgentVersion = "0.3"` (minor-stamp convention held through v0.3.1/v0.3.2); bench_test.go doc comment current through the 09-19 re-measure.
4. **ANNOTATE (119 verdicts) + ARCHIVE**:
   - `docs/planning/2026-09-19_15-39_SUPERB-hardening-speed-plan.md`: 8 M-rows + 33 F-rows struck `done at 7f0f0ed`, §5's 12 items routed/won't-implemented with destinations, EXECUTED header note with three plan deviations → **`git mv` to `docs/planning/archived/`**; `check-rows` COMPLETE; completeness gate OK (44 marker lines).
   - `docs/status/2026-09-19_14-04`: 27 verdicts (b1-b7, c2, c5, f3/f7-f17/f34/f46-f48, g1, g3).
   - `docs/status/2026-09-19_17-03`: 28 verdicts (f×19, b×4, c-bullets×5 incl. `done at fca9280` for the v0.3.0 cut).
   - `docs/status/2026-09-19_18-00`: 10 verdicts (b-table ×2, c-table ×2, f×6).
   - Open items (Q1, live-smoke creds, owner gates, trigger-gated arcs) left deliberately bare per the strict archive line; none of the three reports archived (Q1 unresolved).
5. **Stale research claim corrected**: `2026-09-19_metaengine-go-graph-rag-merge.md` verdict's "json/v2 (66 files; CI-red)" reason annotated with the 2026-10-15 graduation correction (verdict stands on layering/divergence/API-freeze grounds).
6. **BUILD/HARVEST living docs (all six + 2)**: README (install `@v0.3.2`, json/v1-as-policy wording, `NodeID` design constraint); AGENTS (Upstream sync v0.3.2 + CV-pinned fact; Git Town gotcha); FEATURES (+3 rows: `NodeID` strong identity, ADR seam compile guard, wire-contract goldens); TODO_LIST (full rewrite: trophy section + "Closed outside" history deleted, closed check-rows row removed, 18 open rows incl. the full harvest of the three 10-15 reports — release decision, `ContentHash`, golangci bump, metaengine quantification, UA test, kit follow-ups); ROADMAP (+4 raw ideas: DocumentID split trigger, `NodeAttrs.Source`/`Edge.Source` naming, v1-vs-v2 JSON test, placeholder surfacing); DOMAIN_LANGUAGE (+`NodeID` glossary row); SECURITY.md (supported-versions `v0.3.x`); CHANGELOG ([Unreleased] Fixed entry for the doc re-sync).
7. **Upstream skill fix**: `annotate.go repair` crashed on single-line specs (`spec.hi = spec.lo` parsed before `lo`); fixed in the SKILLS repo source, verified by `annotate.go self-test` (all passed) + the 31-assertion fixture suite.
8. **All quality gates green**: `dprint check` clean (after fmt), pristine `env -u GOEXPERIMENT GOTOOLCHAIN=go1.27.1` build + vet, core `go test ./...` ok, bdd suite ok, tidy drift clean both modules, `golangci-lint` 0 issues both modules.
9. **Wipe recovery + memory**: Git Town incident root-caused (dangling commit `08f450e` found via `git fsck --lost-found`), content recovered byte-complete, stability verified across two post-recovery sweeps, AGENTS gotcha written so the next session checks `git status` after write batches.

## b) PARTIALLY DONE

1. **Ubuntu 26 runner migration test (deadline 2026-10-19)** — CI is green today on the current runner, but the forced-migration workflow test (17-03 f14) has not been done; 4 days left.
2. **Self-review depth** — folded into this report per the `.md` instruction; the canonical styled self-review at `docs/reviews/` was not produced (same deviation class as the 06-14 session, same justification).
3. **Benchstat re-measure** — benchstat is installed again, but the manual 2026-09-19 StoreRoundTrip summary in `bench_test.go` was NOT re-run through it (TODO_LIST Low row now owns this).
4. **Daemon collection tail** — at report time all session work is committed except whatever lands after this file; final `git status` was clean pre-report.
5. **The annotate.go fix is edit-only in the SKILLS repo** — committed by its daemon per the external-repo rule; the fan-out link was verified to resolve to the repo checkout.

## c) NOT STARTED (owner-gated or trigger-gated, now tracked)

1. **Next core release decision (v0.3.3 vs v0.4.0)** — v0.3.2 shipped from red-CI code with the dead-allocation defect (fixed post-tag in `[Unreleased]`); the `NodeID` change is also unreleased and breaking for `string`-variable callers; `bdd/go.mod` require-sync + first `bdd/v*` tag ride the same decision. TODO_LIST High row (questions g1-g3 of the 06-14 report).
2. **Q1 license posture** — unchanged; still gates the entire public godoc face (TODO_LIST High row).
3. **`bdd/v0.1.0` tag + `release.yml` `bdd/v*` trigger** — owner-gated; API review of `Harness.Search` variadic options is the recorded prerequisite (TODO_LIST Medium row).
4. **`ContentHash` strong type, golangci bump with CV parity, metaengine dep-tree quantification, UA↔CHANGELOG consistency test, Ginkgo typed-entries doc, kit spec-coverage audit, NodeID consumer note, EmbedConcurrency README mention, benchstat StoreRoundTrip re-run, brandid-lint gate promotion** — all TODO_LIST rows as of this pass.
5. **Ubuntu 26 runner test** — see b1.
6. Everything else in the 14-04/17-03/18-00 f-lists that is trigger-gated (ANN at ~50k, R5/R6/R7/R8, corpus re-ask ~2027-09, second-consumer watch) — deliberately bare in the reports, ROADMAP-indexed.

## d) TOTALLY FUCKED UP

1. **Files vanished mid-session and I noticed ~10 minutes late.** Git Town stashed my TODO_LIST/ROADMAP/17-03 writes into dangling WIP commit `08f450e` and reverted the working tree — AFTER the annotate tool had reported success. Caught only by the end-of-pass marker sweep (17-03 showed 0 `~~` markers against an expected ~33); TODO_LIST/ROADMAP showed 0 harvest content. Recovery was clean (`git restore --source=08f450e -- <3 files>`, markers verified), but the detection should have been immediate: the daemon's 06:58 commit included 5 files while my 06:52-06:57 writes were absent from it — that anomaly was visible earlier. Mitigation recorded: AGENTS gotcha (post-write `git status` check; recovery via `git fsck --lost-found`).
2. **An ambiguous gate verdict, again.** My first completeness check ran `grep -L '~~' <file> && echo FAIL || echo OK` and reported "FAIL: no markers" for the archived plan whose direct `grep -c` showed 44 marker lines — the chain's semantics, not the file, was wrong. The 14-04 report's d3 lesson ("a gate that cannot prove it measured must never be reported as a pass") was violated once before being re-run unambiguously.
3. **Two wasted cycles on edit failures that were my own doing**: a from-memory old_string for DOMAIN_LANGUAGE (`||` table prefix that doesn't exist), and a multiedit on 17-03 attempted after the annotate tool had bumped the file's mtime (the documented re-view-after-tool-write reflex, paid for twice).
4. **`repair`-mode detour**: hit the upstream single-line linespec crash, worked around it by sed-ing `N-N` ranges into the spec file before reading the tool source — then discovered `repair` was the wrong tool for multi-line prose items anyway (it only annotates item-start lines). The `prose` mode handled §5 correctly on the first dry-run.
5. **Nothing shipped broken**: the tree at report time is gate-green, marker-complete, and CI-green; the recovery was byte-complete.

## e) WHAT WE SHOULD IMPROVE

1. **Marker-count sweep after EVERY annotate batch**, not at pass end — in Git Town territory the window between write and wipe is minutes.
2. **After any batch of file writes, check `git status` still shows them** — now encoded in AGENTS.md; make it reflexive whenever the reflog shows WIP stash commits.
3. **Unambiguous gate verdicts by default**: `if grep -q …; then OK else MISSING; fi`, never `grep -L` chains.
4. **On unexpected tool failures, read the tool source before workarounds** — one command (`grep -n "out of range" annotate.go`) would have found the bug immediately.
5. **Exact-match old_strings only from a fresh view** — from-memory table prefixes and post-write edits both cost round trips this session.
6. **Section-skip decisions get stated** — skipping 17-03's g-section was reasonable (duplication of 14-04's answered g1/g3) but should have been recorded in the pass notes, not discovered in self-review.
7. **`repair` vs `prose` vs `rows` mode selection**: prose for numbered lists (including wrapped items), rows for tables, repair only for whole-corpus TSV batches of item-start lines — worth a line in the skill body the next time it's touched.

## f) Up to 50 things to get done next (ranked; 🔵 = owner-gated; items already in TODO_LIST marked ✓)

**Releases / external**

1. 🔵 ✓ Decide + cut the next core release: **v0.3.3 (fixes-only) vs v0.4.0 (bundle `NodeID`)**; `UserAgentVersion` bump; `bdd/go.mod` require-sync rides along (TODO_LIST High; questions g1-g3 of the 06-14 report).
2. 🔵 ✓ Q1 license posture — still gates the entire public godoc face (TODO_LIST High).
3. 🔵 ✓ Tag `bdd/v0.1.0` after the `Harness.Search` API review; extend `release.yml` to `bdd/v*` or keep kit releases manual (TODO_LIST Medium).
4. 🔵 ✓ metaengine dep-tree quantification (owner "run soon", 2026-09-16; now a TODO_LIST Medium row).
5. Push-watch: confirm the daemon-collected docs commits stay green (tip was green at `b3bf8fe`; this report adds more).

**Engineering, ready now (TODO_LIST rows)**

6. ✓ Introduce `ContentHash` strong type (BD003 blind spot: cache keys still plain strings).
7. ✓ Modernize `errors.As` → `errors.AsType` at `embed_openai.go:334`; re-check `embed_openai_test.go:249`.
8. ✓ Bump golangci-lint v2.13.2 → v2.14.x WITH CV parity (closes the local/CI skew risk on the bdd lint steps).
9. ✓ Decide SECURITY.md contact channel + practices section (buildflow securitymd 2 findings).
10. ✓ Run the live smoke test (needs `GRAPHRAG_LIVE_EMBED_*` creds).
11. ✓ Watch tobi/qmd#959 + charmbracelet/crush#3846 (held past 09-22 by ruling; no pings).
12. ✓ Apply CONTRIBUTING inbound-grant line (queues behind Q1).
13. ✓ UA-version ↔ newest-CHANGELOG-version consistency test.
14. ✓ Document the Ginkgo typed-`DescribeTable` gotcha in `bdd/doc.go`.
15. ✓ Kit spec-coverage audit (`WeightedEdge`, `WithKind`, `SimilarEdgeScore`).
16. ✓ "Upgrading to NodeID" consumer note (README or `docs/`).
17. ✓ README: mention `EmbedConcurrency` for network-bound cold builds.
18. ✓ Re-run StoreRoundTrip through benchstat (installed again; replaces the manual 09-19 summary).
19. ✓ Promote `brandid-lint -rules BD001,BD003` to pre-push/CI (owner decision + planted-failure drills).

**Session-specific follow-ups (not yet rows)**

20. Test the go-test workflow on the Ubuntu 26 runner before the 2026-10-19 forced migration (17-03 f14; 4 days left).
21. Bench sanity pass (`-benchtime 100ms`) on Build/SimilarPairs/StoreRoundTrip to confirm `NodeID` map-key typing is perf-neutral (06-14 forget #1 — still unmeasured).
22. Full benchstat re-record (`-count 10`) of the whole table under 1.27.1 if the owner wants post-NodeID numbers (06-14 g1).
23. Re-baseline buildflow after this AGENTS.md update (`agent-config` self-resolves when AGENTS is fresh).
24. Verify the archived plan's new location breaks no cross-file links (grep for the old `planning/2026-09-19_15-39` path repo-wide).
25. Visual render check of `docs/architecture-understanding/2026-09-19_13-30_architecture-review.html` was done 09-19; re-check only if R1-R4 changes warrant a supersede.
26. Consider `git town` config: disable the WIP-stash behavior for daemon-committed repos, or accept the AGENTS gotcha as the whole mitigation (owner taste).
27. After the golangci bump: re-verify CI's pinned version agrees with local on the bdd module (closes 05-18 f17).
28. Add a pre-push gate 4 (golangci on changed packages) if lint regressions keep surfacing only in CI (05-18 e5).

**Docs / docs-health hygiene**

29. ✓ (done this pass) TODO_LIST harvest of the three 10-15 reports — future passes: keep the ledger discipline.
30. Annotate this report's own resolved items in a future pass (standing rule; nothing to do now).
31. When Q1 lands: apply CONTRIBUTING grant, re-cut release, re-verify pkg.go.dev renders, then the archive sweep of the Q1-gated status reports becomes unblocked (14-04 f36/f33).
32. `docs/research/README.md` stays research-only; no row was needed this pass (correction was an annotation, not a new artifact) — keep the discipline.
33. If the SKILLS-repo annotate.go fix propagates via flake, re-verify the fan-out copy reads the fixed version (`bash scripts/check-skill-fanout.sh` from the crush-config repo).

**Trigger-gated arcs (ROADMAP-indexed, listed for completeness)**

34. R7: land ADR §5 interfaces + reference `VectorIndex` (on implementation approval; compile guard already speaks it).
35. R5: parallel pairwise scan (byte-identical output; ROADMAP Theme 1).
36. g2: dot-of-normalized fast path (owner-visible score drift).
37. ANN spike + harness + filtered-ANN ADR at the ~50k trigger.
38. Incremental indexing design.
39. Wire-contract v1 freeze decisions (incl. the new v1-vs-v2 decode-equals test idea).
40. Metrics/tracing seams; store compaction/eviction; backup/export.
41. Corpus re-ask ~2027-09; second-consumer watch; Dgraph/metaengine sweeps on their research-index triggers.
42. Distinct `DocumentID` revisit if documents ever split from nodes (new ROADMAP raw idea).
43. `NodeAttrs.Source` vs `Edge.Source` naming review (new ROADMAP raw idea).
44. Surface `SimilarPairs` vector-without-node placeholders louder (new ROADMAP raw idea).

**Standing hygiene**

45. Re-verify daemon commits landing on top of session work (standing AGENTS rule; held perfectly this pass).
46. Keep `dprint fmt` reflexive after every markdown table edit (held this pass; caught my own tables twice).
47. Keep `GOTOOLCHAIN=go1.27.1` + golangci pin moving together with CV (AGENTS rule).
48. Install `git config core.hooksPath .githooks` per clone before pushing (verified installed here).
49. Keep research/planning annotate-only discipline (held; the merge-research correction was additive).
50. Next docs-health pass: sweep the then-current TODO_LIST against the newest reports, archive any report that reaches every-item-resolved (14-14 became provably closer this pass).

## g) Questions I cannot figure out myself

1. **Release shape**: should the next cut be **v0.3.3** (dead-allocation + CI fixes only, keep `NodeID` unreleased) or **v0.4.0** (bundle the strong-ID change with its consumer migration note)? And do you want the bench table re-measured (benchstat, `-count 10`) as part of that release, or is the pre-`NodeID` table fine to keep?
2. **`bdd/v0.1.0` timing**: tag now that master CI is green (kit is dogfooded by its own 43-spec suite), or only after a consumer pilot (CV `platform/bddkit` comparison, or legal-graph-ai-system's planned testify→ginkgo migration)? Related: is the `Harness.Search` variadic-options shape acceptable to freeze, or should the API review block the tag?
3. **Git Town policy**: is Git Town's WIP-stash behavior intentionally active on this machine (in which case the AGENTS gotcha + post-write `git status` reflex is the mitigation), or should it be disabled for repos with an auto-commit daemon so mid-session wipes like today's can't recur?

---

_Point-in-time snapshot, 2026-10-15 07:16 CEST. Section (f) is docs-health HARVEST fuel; items marked ✓ already landed in TODO_LIST.md this pass. Annotate, never rewrite._
