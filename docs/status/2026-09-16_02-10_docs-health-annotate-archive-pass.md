# Status: Docs-Health AUDIT — annotate/archive pass over all 2026-0* files

|             |                                                                                                                                                                                 |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Date/Time   | 2026-09-16 02:10 CEST (via `date`)                                                                                                                                              |
| Scope       | This session only: full docs-health AUDIT (VERIFY + HARVEST + BUILD/UPDATE + ANNOTATE + ARCHIVE) over all 8 `2026-0*` files and the 6 living docs, on explicit user instruction |
| Deliverable | Living-docs truth sync + inline annotations + first `docs/planning/archived/` entry + `docs/research/README.md` index                                                           |
| Format      | `.md` per explicit user instruction (skill default is styled HTML; override honored and flagged)                                                                                |
| Go code     | Untouched by design; all gates green before and after                                                                                                                           |

## Executive summary

One command-scope executed end to end: all 8 `2026-0*` files (2 planning, 4 status, 2
research) read in full, classified per the docs-health decision table, and
dispatched — the fully-executed pareto plan was inline-annotated (all 67 table
rows + header) and archived to `docs/planning/archived/`; three status reports
got 31 inline `~~…~~ done` markers with hashes/evidence; the owner-decision
package and both research reports were correctly left alone. The six living
docs were claim-verified against code (zero citation drift), harvested (seam
design + dep-quantification rows into TODO_LIST; two do-now rows executed and
deleted), and updated in place; `docs/research/README.md` (index) is new.
Gates: pristine build/vet/test, golangci `0 issues.`, tidy clean, dprint
clean. One real failure shipped: the inline health report's found-state
Accuracy arithmetic was wrong (8.5 printed; the table implies 8.75) — the
exact "garbled math" class the skill warns about. Corrected here; post-fix
scores (10/10) are unaffected.

## a) FULLY DONE

1. **Skill compliance first**: `docs-health` SKILL.md loaded before any action;
   annotate tooling (`annotate-rows.py`, `annotate-prose.py`) read before use;
   every batch dry-run before a real write; `status-report` loaded before this
   file.
2. **Full read-in, nothing from memory**: all 8 `2026-0*` files (866 lines) +
   all 6 living docs + targeted source excerpts (`example_test.go`,
   `bench_test.go` doc, `graph.go` KindUnknown block, `go.mod`).
3. **VERIFY against code — zero drift found**: ~60 `FEATURES.md` file:line
   citations swept against `build.go`/`graph.go`/`search.go`/`store.go`/
   `vector.go`/`config.go`/`embed*.go` (despite the session-start snapshot
   showing `graph.go`/`embed_openai_test.go` modified — all landed citations
   still accurate). Also confirmed: 4 example function names;
   `UserAgentVersion = "0.1"`; `DefaultSimilarThreshold = 0.75`;
   `SearchOptions.MinScore` at `search.go:40`; bench reference numbers
   (quadratic, ~373ms/1000 docs, ~15min extrapolated at 50k); `go.mod` 3
   direct deps.
4. **HARVEST**: `TODO_LIST.md` +High row "Design the pluggable store/search
   seam (scenario B)" (owner decision 2026-09-15: DESIGN NOW — three backend
   classes, `dgo` never in core `go.mod`) and +Medium row "Quantify metaengine
   dep-tree delta" (metaengine report f.2). Two do-now rows (README MinScore,
   ROADMAP cost cross-link) were **executed, then deleted** from TODO_LIST per
   its done-items-leave lifecycle.
5. **BUILD/UPDATE living docs**:
   - `README.md`: quick-start snippet now sets `MinScore: 0.25` (field
     verified at `search.go:40`) — kills the confusing 0.00-similarity hit.
   - `ROADMAP.md`: settled `KindUnknown` idea removed (it lives in
     `graph.go:12` doc + CHANGELOG); ~50k note upgraded to the measured bench
     numbers; Theme 1 +3 raw ideas (shared ANN benchmark harness,
     metadata-filtered ANN semantics, ANN-winner ADR); external-DB non-goal
     cross-linked to `docs/research/2026-09-15_dgraph-adoption.md` (dgraph
     item 18).
   - `CHANGELOG.md`: `[Unreleased]` Added entry for both research reports +
     owner-decision package + research index (dgraph item 39).
   - `FEATURES.md`: +PLANNED row "Pluggable store/search seam" (no code
     exists — honest status).
   - `AGENTS.md`: + "Owner decisions (2026-09-15)" section (corpus NO ⇒
     storage stays ROADMAP fuel; seam DESIGN NOW, not started;
     status/planning/research are point-in-time: annotate or archive, never
     rewrite).
   - NEW `docs/research/README.md`: index table (artifact / date / verdict /
     reopen-triggers) closing dgraph item 38 + metaengine f.3.
6. **ANNOTATE (inline, mandatory-placement compliant; 31 + 67 markers)**:
   - `18-19` audit report: items b1–b2 + f1–f16, f19–f20 struck with hashes
     (`41beb47`, `16fc16c7`) or status §a evidence; header's stale "remote CI
     still red" inline-corrected. **f17/f18 left untouched** (owner-gated =
     open).
   - metaengine status report: items 3, 13–16 struck (research index exists;
     qmd/crush issues filed; CV annotations `16fc16c7`; CI green since push;
     protection/topics done).
   - dgraph status report: items 18, 31, 36–39 struck (ROADMAP cross-link;
     linter policy in AGENTS; this harvest; the 18-19 annotate pass; research
     index; CHANGELOG entry).
   - pareto plan: header corrected + **all 15 M-rows and all 52 F-rows**
     struck with per-row evidence (`annotate-rows.py` for M; dotted-ID variant
     for F; F12.1/F12.2 hand-annotated where the script's literal-`~~` guard
     correctly refused). Shape checks passed on every script write.
7. **ARCHIVE**: `git mv docs/planning/2026-09-15_18-30_SUPERB-pareto-execution-plan.md
   → docs/planning/archived/`. Completeness gate `grep -rLn '~~'
   docs/planning/archived/` → prints nothing. No living doc references the old
   path (grep-verified).
8. **Quality gates (pristine, post-edit)**: `env -u GOEXPERIMENT
   GOTOOLCHAIN=go1.26.7 go build/vet/test` green; `golangci-lint run ./...` →
   `0 issues.`; `go mod tidy` + `git diff --exit-code go.mod go.sum` clean;
   `dprint fmt` (4 files realigned) then `dprint check` exit 0.
9. **Inline AUDIT health report** printed to chat with both scores and visible
   math (found-state had an arithmetic slip — see d1; post-fix state is
   10/10).

## b) PARTIALLY DONE

1. **`docs/DOMAIN_LANGUAGE.md` not re-audited line-by-line** — the only living
   doc without a fresh claim check this pass (it was built + verified same day
   in the 18-19 audit). Disclosed in the health report rather than hidden.
2. **Format gate run partially**: `dprint fmt/check` yes; `buildflow format`
   (which adds the lychee link check over the session's many new cross-file
   links) not run. All link targets were created/verified to exist, so links
   should resolve — but the link checker itself never said so.
3. **F12.1 rendering unverified**: the hand annotation strikes a cell that
   contains a literal `` `~~...~~` `` code span inside the strikethrough. I
   asserted safety from CommonMark tokenization order (code spans first) but
   never looked at a rendered output.
4. **Working tree at handoff**: 4 paths uncommitted (`FEATURES.md`,
   `TODO_LIST.md`, `docs/research/README.md`, the plan rename) — the
   auto-commit daemon will sweep them; no manual commit (harness forbids
   without explicit request).
5. **dgraph report f-list coverage**: only its 6 resolved items were
   annotated; the other 44 remain deliberately open (spike/owner/hygiene
   grade) — correct per skill ("absence of a marker IS the open signal"), but
   readers should expect a mixed open/resolved list there by design.
6. **Health-report found-state math**: printed 8.5, table implies 8.75 —
   corrected in d1; the post-fix 10/10 stands (zero findings remain).

## c) NOT STARTED

1. **The seam design itself** — owner decision #2 (DESIGN NOW) is encoded as
   the TODO_LIST High row; zero design work done. Main candidate for the next
   work session.
2. **Owner decisions** — license posture (Q1) and v0.2.0 cut (Q2) still
   await answers; package checkboxes unfilled.
3. **All pre-existing TODO_LIST rows** untouched this session: tag protection,
   `Build` identical-text doc comment, benchstat re-run, `ExampleNewStore`,
   `SearcherOptions` example, store round-trip benchmark, CONTRIBUTING
   inbound-grant (BLOCKED on Q1), CV-repo push (external), live smoke creds
   (BLOCKED), qmd#959/crush#3846 watch, GH release automation, `qmd embed`.
4. **ROADMAP-grade arcs** untouched: ANN/HNSW spike, incremental indexing,
   plus the three newly added ideas (harness, filtered-ANN semantics, ADR).
5. **No Go code, dependency, or CI changes** — by design for a docs-health
   pass; nothing to start there.

## d) TOTALLY FUCKED UP (honesty ledger)

1. **Garbled math in the inline health report.** I printed "Accuracy 8.5/10
   found (10 − 0.5·1 Medium − 0.25·3 Low)" but my own findings table contains
   FOUR Low findings (ROADMAP unmeasured-50k framing, AGENTS missing owner
   decisions, FEATURES missing seam row, CHANGELOG missing research entries)
   ⇒ 8.75. The score line was not a pure function of the table — the exact
   failure mode the skill's math-discipline section exists to prevent, and the
   third session in this fleet to hit a scoring/counting slip. Damage: the
   found-state number only (post-fix 10/10 has zero findings to miscount).
   Correct value: **8.75**.
2. **Trusted the document's self-count instead of counting.** The plan's prose
   says "47 fine tasks"; its table has 52 rows. My first F-spec batch was
   built from the prose number and additionally dropped F12.3; only the
   script's loud "expected 1 match, found 0 / already annotated" failures
   revealed both. Worse: the official completeness gate (`grep -rLn '~~'`)
   is per-FILE — it would NOT have caught unstruck rows. I archived a plan
   whose every row is struck only because a tool refused to lie, not because
   my own gate could prove it.
3. **Weak evidence marker shipped.** F12.2 was first annotated "done — see
   F12.1" instead of citing `16fc16c7`; fixed one edit later. Evidence-first
   is the core ANNOTATE rule; a placeholder is exactly what the "so what?"
   test rejects, and it only stayed out of the archive because I re-read my
   own output.
4. **Two wasted edit round trips.** After the annotate scripts wrote the plan
   file, I re-read it via `bash sed` and attempted edits — "file modified
   since read" twice. The rule is re-read with the read TOOL (view registers
   edit-eligibility); I knew it and reached for sed anyway.
5. **Daemon surprise treated as a puzzle.** Post-gate `git status` showed only
   4 modified paths and I briefly read that as lost edits before realizing
   the auto-commit daemon had already committed the rest (535dcbb…1f72ae0,
   8e6b731). Zero damage, but the fleet lesson "expect the daemon to have
   committed mid-session" is documented and I still stumbled before
   confirming via `git log`/`git show`.
6. **Tooling gap worked around, not upstreamed (in-session).** F12.1's cell
   legitimately contains `` `~~...~~` `` and the annotate script's
   already-annotated guard false-positives on tildes inside code spans; I
   hand-annotated around it. The fix belongs in `annotate-rows.py` (skip the
   guard when the `~~` is inside backticks) — queued as f-item 12, not done.

## e) WHAT WE SHOULD IMPROVE

1. **Count rows, never trust self-counts**: one `grep -c '^| F'` before bulk
   annotation prevents d2 entirely. Documents' own "N tasks" prose rots.
2. **Per-row completeness assertion after bulk sweeps** (struck-row count ==
   table-row count); the per-file grep gate proves ≥1 strikethrough, not
   coverage. Candidate strengthening for the docs-health skill assets.
3. **Re-read with the read tool after any script write**, before any edit —
   mechanical, no exceptions (d4).
4. **Every marker cites its hash at first write**; "see X" is a failure even
   when corrected within a minute (d3).
5. **Run the full format gate after adding cross-file links** — `buildflow
   format` (lychee) rather than dprint alone (b2).
6. **Verify rendered markdown for strike-marked code spans** before archiving;
   spec-knowledge rendering claims are unverified claims (b3).
7. **Fix the annotate tooling's literal-tilde false positive** rather than
   hand-annotating around it each time (d6/f12).
8. **Scoring discipline**: substitute table counts into the formula
   mechanically and re-count before printing (d1) — same class as the
   pipeline-masking and bench-magnitude lessons already in the fleet ledger.

## f) Up to 50 things we should get done next

Brainstorm ranked by impact; TODO_LIST-grade items are already routed (this
session's harvest), items 18+ are spike/ROADMAP/hygiene fuel, not commitments.

**P0 — owner-gated (everything else queues behind these)**

1. License posture decision (owner package Q1) — BLOCKED.
2. Approve the v0.2.0 cut (owner package Q2) — BLOCKED.
3. Execute the v0.2.0 go-release checklist on approval (tag → proxy →
   pkg.go.dev → CV bump).
4. Apply the CONTRIBUTING inbound-grant line (after Q1).
5. Push the CV repo (external; carries the `16fc16c7` extraction annotations).

**P1 — the chosen engineering work**

6. Design the pluggable store/search seam (owner: DESIGN NOW; TODO_LIST High
   row): ADR under `docs/planning/`, three backend classes (embedded ANN libs
   / metaengine engines / raw server DBs) with per-class vector semantics —
   ANN vs brute-force, metric coupling, distance availability, filtered-ANN
   behavior; `dgo`/metaengine never enter core `go.mod`.
7. Add the tag-protection rule before v0.2.0 (`gh api …/tags/protection`).
8. Document the identical-text → single-vector rule on `Build` (doc comment;
   status 19-33 e.5).

**P1 — this session's loose ends**

9. Run `buildflow format` (lychee) over the session's new cross-file links.
10. Verify the F12.1 nested-tilde cell renders correctly on GitHub.
11. Add a per-row completeness assertion to the annotate workflow
    (skill-maintenance surface).
12. Fix `annotate-rows.py`'s already-annotated false positive on tildes inside
    code spans (skill-maintenance surface).
13. Line-by-line re-audit of `docs/DOMAIN_LANGUAGE.md` (only living doc not
    re-verified this pass).
14. CV-side cross-link decision for the metaengine report (owner taste;
    metaengine g2).
15. Fix the go-cqrs-lite skill's `modules.md` "system EXPERIMENTAL" wording
    (skill-creator; 10 min; metaengine f.4).
16. Quantify the metaengine dep-tree delta — `go mod graph` before/after
    (TODO_LIST Medium row; metaengine f.2).
17. Add the "new research artifacts get an index row" rule to AGENTS.md
    (one line, keeps `docs/research/README.md` alive).

**P2 — Dgraph research follow-ups (gated on triggers; ROADMAP fuel)**

18. Run the §8 spike for real if scenario B/C ever activates
    (`dgraph/standalone` + `similar_to` + filter composition).
19. Decide `dgo/v240` vs `dgo/v250` (only on an adoption path).
20. Close the 2021–2025 Dgraph license-history gap.
21. Add a second, independent community source (HN quoted-phrase / Reddit).
22. Set release-cadence + second-steward tripwires (issue or calendar).
23. Dgraph security-posture review (CVE history, audits).
24. Re-check the pkg.go.dev importer count (23) on any revisit.
25. Archive raw research notes (thread IDs, endpoints) into the dgraph report
    appendix (dgraph item 50).
26. Confirm lychee tolerates the forum JSON links on future runs (item 46).
27. HTML renders of the research reports if ever presented to humans.
28. metaengine report hardening: read `readmodels.md` + `advanced.md`, verify
    pebble "~7x" from bench source, CHANGELOG release cadence (f.6–8).
29. Worked "search mapped onto metaengine" sketch (f.9).
30. A dependency-proposal template generalizing both research reports (f.21).

**P3 — SDK engineering (existing TODO_LIST rows, untouched this session)**

31. Benchstat-protocol benchmark re-run; update reference numbers.
32. `ExampleNewStore` godoc example.
33. `SearcherOptions` (ReferenceKind/DocumentKinds) example variant.
34. Store round-trip benchmark (persist + snapshot load, 1k/10k nodes).
35. Run the live smoke test against a real endpoint (BLOCKED on
    `GRAPHRAG_LIVE_EMBED_*` creds).
36. Watch tobi/qmd#959 + charmbracelet/crush#3846; send the offered PR if
    silent ~1 week.
37. GitHub Release automation workflow (v0.2.0 follow-up).
38. Run `qmd embed` (173 CV docs unembedded; external tooling).
39. `lsp_restart golangci_lint_ls` after daemon commits (habit).

**P4 — ROADMAP-scale arcs**

40. ANN/HNSW spike (`sqlite-vec` / `hannoy` / in-process port) behind the
    `Searcher` seam.
41. Incremental indexing design (upsert/delete vs `ReplaceGraph` swap).
42. Shared ANN benchmark harness (recall + p50/p95 vs brute-force).
43. Metadata-filtered vector-search semantics (kind filters + `MinScore` with
    ANN).
44. ANN-winner ADR incl. `RelationSimilar` derivation at ANN scale.
45. Freeze the wire contract (`Hit`/`SearchResult` JSON keys) with golden
    tests.
46. Metrics/tracing seams for Build/Search decoration.
47. Store compaction + embedding-cache eviction policy.
48. Backup/export story beyond "delete and rebuild".
49. Re-ask the corpus question at the 12-month horizon (owner said NO until
    ~2027-09; storage work stays ROADMAP fuel until then).
50. Watch for a second SDK consumer (trigger for the isolated
    `go-graph-rag/metaengine` adapter module, metaengine report §8.3).

## g) QUESTIONS (cannot answer myself)

1. **Archive policy for status reports**: today I archived the fully-executed
   planning doc under `docs/planning/archived/`. Should fully-resolved STATUS
   reports likewise move to `docs/status/archived/` once every item is
   resolved (none qualify today), or do status reports stay in place forever
   with inline markers as the only resolution signal?
2. **Seam-design deliverable shape**: for the owner-approved seam design
   (f.6), do you want a prose-only ADR decision record under
   `docs/planning/`, or an ADR that includes a concrete Go interface sketch
   (seam signatures) so the design is binding before any code?
3. **Settled roadmap ideas**: I removed the settled `KindUnknown` idea from
   ROADMAP (it lives in the `graph.go` doc comment and CHANGELOG). Keep
   removal as the pattern, or keep settled ideas in ROADMAP struck through as
   a visible decision trail?

Waiting for instructions.
