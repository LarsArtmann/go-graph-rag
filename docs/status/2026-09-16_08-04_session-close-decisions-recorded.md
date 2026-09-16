# Status: Session close — docs-health pass + owner decisions recorded

|             |                                                                                                                                                                                                                                                                |
| ----------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Date/Time   | 2026-09-16 08:04 CEST (via `date`)                                                                                                                                                                                                                             |
| Scope       | Closes the session that ran the full docs-health AUDIT (annotate/archive pass) and then recorded the owner's three answers to its open questions                                                                                                               |
| Deliverable | Decision recording: `AGENTS.md` (Owner decisions 2026-09-16), `TODO_LIST.md` seam-row deliverable, ANSWERED block in `docs/status/2026-09-16_02-10_docs-health-annotate-archive-pass.md`                                                                       |
| Format      | `.md` per explicit user instruction (skill default is styled HTML; override honored and flagged)                                                                                                                                                               |
| Companion   | `docs/status/2026-09-16_02-10_docs-health-annotate-archive-pass.md` is the detailed record of the audit phase (a–f there remain the authoritative per-item ledger); this report closes the session and carries the forward list with the new decisions applied |
| Go code     | Untouched all session; gates green throughout                                                                                                                                                                                                                  |

## Executive summary

The session's audit phase (all 8 `2026-0*` files verified/annotated, the
fully-executed pareto plan archived with all 67 rows struck, six living docs
truth-synced, gates green) completed at 02:10 and is fully reported in the
companion report. Since then: the owner answered all three open questions
(archive policy, seam deliverable shape, ROADMAP settled-idea policy), and all
three answers are recorded where future sessions will see them. Nothing new
was started; nothing regressed. Two small edit-race round trips occurred in
the decision turn — the exact class flagged in the companion report's
improvement list, recurring within the hour. Repo state: `master` clean of
local work except two doc paths pending the daemon sweep; CI green on HEAD;
all quality gates green.

## a) FULLY DONE

**Audit phase (02:10 report — summary; full per-item evidence there):**

1. All 8 `2026-0*` files read and classified; pareto plan inline-annotated
   (header + 15 M-rows + 52 F-rows) and archived to `docs/planning/archived/`
   (completeness gate clean); three status reports got 31 inline resolution
   markers; owner-decision package and both research reports correctly left
   alone.
2. Living docs claim-verified against code (~60 `FEATURES.md` citations, zero
   drift), harvested (seam-design + dep-quantification rows into TODO_LIST),
   and updated (README MinScore fix, ROADMAP de-staled + cross-linked +
   3 new raw ideas, CHANGELOG research entry, FEATURES PLANNED row, AGENTS
   owner-decisions section); `docs/research/README.md` index created.
3. Gates: pristine `env -u GOEXPERIMENT` build/vet/test, golangci `0 issues.`,
   tidy drift clean, dprint fmt+check clean.

**Decision turn (this delta, 02:10 → 08:04):**

4. Owner answered all three g) questions via the question tool:
   1. Status reports **archive** to `docs/status/archived/` once fully
      resolved (same as planning docs).
   2. Seam deliverable: **ADR + Go interface sketch** — binding before code.
   3. ROADMAP settled ideas: **remove** (no struck-through zombies).
5. All three recorded durably: new "Owner decisions (2026-09-16)" section in
   `AGENTS.md`; `TODO_LIST.md` High seam row now states the deliverable
   ("ADR + Go interface sketch, binding before code"); ANSWERED block appended
   to the 02-10 report's g) section (same pattern as the dgraph report).
6. Verified no work was lost to the daemon: `git log`/`git status` checked
   before and after the edits; daemon commits `0d451ff`/`8f9b2ed` captured the
   earlier audit edits, the two newest doc paths are pending sweep (b1).
7. New report dprint-formatted and check-clean.

## b) PARTIALLY DONE

1. **Daemon sweep pending**: `TODO_LIST.md` and
   `docs/status/2026-09-16_02-10_….md` show as modified at report time
   (edits landed 02:15; daemon had not committed them by 08:04). Content is
   correct and dprint-clean; only the commit is outstanding.
2. **The carried-forward f) list**: re-validated against current state, but
   none of its items were executed in the decision turn (by design — waiting
   for instructions).
3. **Loose ends inherited from the audit phase** (unchanged, all listed in the
   02-10 report b/c): DOMAIN_LANGUAGE line audit, buildflow/lychee format run
   over the new cross-links, F12.1 render check.

## c) NOT STARTED

1. **The seam design itself** — owner-approved (DESIGN NOW), deliverable now
   pinned (ADR + Go interface sketch, three backend classes, `dgo`/metaengine
   never in core `go.mod`). Zero design work done. Next session's main
   candidate.
2. **Owner-gated release work** — license posture (Q1) and v0.2.0 cut (Q2)
   remain unanswered; the v0.2.0 checklist, CONTRIBUTING inbound-grant line,
   and tag-protection rule all queue behind them.
3. **All other TODO_LIST rows and ROADMAP arcs** — untouched (see f).

## d) TOTALLY FUCKED UP (honesty ledger — delta turn + session-wide score)

1. **The improvement lesson recurred within the hour.** The 02-10 report's
   e-list says "re-read with the read tool after any script/daemon write
   before editing." In the decision turn I hit the same race twice anyway:
   `TODO_LIST.md` (daemon touched it between my read and edit) and the 02-10
   report (dprint touched it after my write). Recovery was immediate both
   times (view → retry), zero damage — but prevention remained a rule I
   recalled only after paying the cost, not a habit I executed. Two wasted
   round trips in a four-edit turn is a poor ratio.
2. **ANSWERED block appended appendix-style.** The docs-health skill's #1
   failure mode is appendix-only annotation. The g) ANSWERED block matches
   fleet precedent (dgraph report) and the questions are not numbered work
   items, so nothing is masked — but the deliberate, correct move would have
   been inline markers under each question with the block as supplement. I
   reused the pattern without re-deciding it.
3. **Session-wide, the audit phase's real failure stands corrected but its
   root causes are still open tooling/process debt**: the health-report
   arithmetic slip (8.5 printed vs 8.75 implied — corrected in writing), the
   self-count trust failure (plan prose "47 fine tasks" vs 52 table rows),
   and the per-file-only completeness gate. None of the three has a
   mechanical guard yet (f.11/f.12).
4. Nothing catastrophic: no code touched, no data lost, no gate red, no
   rewrites of historical docs.

## e) WHAT WE SHOULD IMPROVE

1. **Make re-read-before-edit mechanical, not remembered**: after ANY daemon
   commit, dprint run, or script write, the next edit on that file starts
   with a fresh `view`. Two same-session recurrences say "rule" where I need
   "reflex."
2. **Inline-first even for fresh files**: when answering questions or
   resolving items in a doc I just wrote, prefer inline markers under the
   item; appendix blocks only as supplement.
3. **Batch decision-recording into one atomic pass**: the three recording
   edits (AGENTS, TODO_LIST, report) went serially with a daemon race
   landing between them; one multiedit-style sweep per file immediately
   after the answers would shrink the race window to zero.
4. **Build the missing guards** (carried from 02-10): per-row annotate
   completeness assertion; fix `annotate-rows.py`'s literal-tilde false
   positive; never trust a document's self-count.
5. **Scoring discipline** (carried): substitute table counts mechanically and
   re-count before printing any score line.

## f) Up to 50 things we should get done next

Carried forward from the 02-10 report, re-validated, with the three owner
decisions applied (marked ▸ where changed). Brainstorm, ranked; not
commitments. TODO_LIST-grade rows are already routed.

**P0 — owner-gated**

1. License posture decision (owner package Q1) — BLOCKED.
2. Approve the v0.2.0 cut (owner package Q2) — BLOCKED.
3. Execute the v0.2.0 go-release checklist on approval (tag → proxy →
   pkg.go.dev → CV bump).
4. Apply the CONTRIBUTING inbound-grant line (after Q1).
5. Push the CV repo (external; carries the `16fc16c7` annotations).

**P1 — the chosen engineering work**

6. ▸ Design the pluggable store/search seam: ADR under `docs/planning/` WITH
   Go interface sketch (owner 2026-09-16), three backend classes (embedded
   ANN libs / metaengine engines / raw server DBs) with per-class vector
   semantics (ANN vs brute-force, metric coupling, distance availability,
   filtered-ANN behavior); `dgo`/metaengine never enter core `go.mod`.
7. Add the tag-protection rule before v0.2.0.
8. Document the identical-text → single-vector rule on `Build` (doc comment).

**P1 — session loose ends**

9. Run `buildflow format` (lychee) over the session's new cross-file links.
10. Verify the archived plan's F12.1 nested-tilde cell renders on GitHub.
11. Add a per-row completeness assertion to the annotate workflow
    (skill-maintenance surface).
12. Fix `annotate-rows.py`'s already-annotated false positive on tildes inside
    code spans (skill-maintenance surface).
13. Line-by-line re-audit of `docs/DOMAIN_LANGUAGE.md`.
14. CV-side cross-link decision for the metaengine report (owner taste).
15. Fix the go-cqrs-lite skill's `modules.md` "system EXPERIMENTAL" wording
    (skill-creator; 10 min).
16. Quantify the metaengine dep-tree delta — `go mod graph` before/after
    (TODO_LIST Medium row).
17. Add the "new research artifacts get an index row" rule to AGENTS.md.
18. ▸ Apply the new archive policy going forward: fully-resolved status
    reports → `git mv docs/status/archived/` (none qualify today; policy now
    in AGENTS.md).

**P2 — Dgraph research follow-ups (gated on triggers)**

19. Run the §8 spike if scenario B/C activates (`dgraph/standalone`,
    `similar_to` + filter composition).
20. Decide `dgo/v240` vs `dgo/v250` (adoption path only).
21. Close the 2021–2025 Dgraph license-history gap.
22. Add a second, independent community source (HN quoted-phrase / Reddit).
23. Set release-cadence + second-steward tripwires.
24. Dgraph security-posture review (CVE history, audits).
25. Re-check the pkg.go.dev importer count (23) on any revisit.
26. Archive raw research notes into the dgraph report appendix.
27. Confirm lychee tolerates the forum JSON links on future runs.
28. HTML renders of research reports if presented to humans.
29. metaengine report hardening (readmodels/advanced refs, pebble 7x from
    bench source, release cadence).
30. Worked "search mapped onto metaengine" sketch.
31. A dependency-proposal template generalizing both research reports.

**P3 — SDK engineering (existing TODO_LIST rows)**

32. Benchstat-protocol benchmark re-run; update reference numbers.
33. `ExampleNewStore` godoc example.
34. `SearcherOptions` (ReferenceKind/DocumentKinds) example variant.
35. Store round-trip benchmark (persist + snapshot load, 1k/10k nodes).
36. Live smoke test against a real endpoint (BLOCKED on creds).
37. Watch tobi/qmd#959 + charmbracelet/crush#3846; PR if silent ~1 week.
38. GitHub Release automation workflow (v0.2.0 follow-up).
39. Run `qmd embed` (173 CV docs; external tooling).
40. `lsp_restart golangci_lint_ls` after daemon commits (habit).

**P4 — ROADMAP-scale arcs**

41. ANN/HNSW spike (`sqlite-vec` / `hannoy` / in-process port) behind the
    `Searcher` seam.
42. Incremental indexing design (upsert/delete vs `ReplaceGraph` swap).
43. Shared ANN benchmark harness (recall + p50/p95 vs brute-force).
44. Metadata-filtered vector-search semantics (kind filters + `MinScore`).
45. ANN-winner ADR incl. `RelationSimilar` derivation at ANN scale.
46. Freeze the wire contract (`Hit`/`SearchResult` JSON keys) with golden
    tests.
47. Metrics/tracing seams for Build/Search decoration.
48. Store compaction + embedding-cache eviction policy.
49. Backup/export story beyond "delete and rebuild".
50. Watch for a second SDK consumer (trigger for the isolated
    `go-graph-rag/metaengine` adapter module); re-ask the corpus question at
    the ~2027-09 horizon.

## g) QUESTIONS (cannot answer myself)

1. **Next session's priority**: the seam design (f.6) is the standing
   candidate, but Q1/Q2 (license, v0.2.0) gate all release-facing work and
   are one-word answers. Which do you want next: (a) seam ADR + interface
   sketch, (b) you answer Q1/Q2 now and I execute the release checklist, or
   (c) something else entirely?
2. **Dep-quantification timing** (f.16): the metaengine verdict (REJECT) is
   already decided, so the 30-minute `go mod graph` experiment only hardens
   the report's #1 contra for future readers. Run it soon, or only when a
   metaengine revisit trigger actually fires?
3. **CGO stance for adapter backends**: core stays CGO-free
   (`modernc.org/sqlite`). If the seam's adapter class ever wants
   `sqlite-vec` (a C extension), is CGO acceptable inside an optional
   adapter — or must the whole module family stay CGO-free, making pure-Go
   ANN (e.g. hannoy-style) the only in-repo path? The ADR needs this
   constraint written down; only you can set it.

Waiting for instructions.

## g) ANSWERED (same session, 08:10 CEST)

1. Next session: **SEAM DESIGN** (owner) — f.6 is the confirmed next work
   item: ADR + Go interface sketch, three backend classes.
2. Dep quantification: **RUN SOON** (owner) — not trigger-gated; TODO_LIST
   Medium row updated.
3. CGO stance: **CGO IN OPTIONAL ADAPTERS ONLY** (owner) — core stays
   CGO-free; the seam ADR writes this constraint down.

All three recorded in `AGENTS.md` (Owner decisions 2026-09-16) and
`TODO_LIST.md` (seam row + dep-quant row).

Waiting for instructions.
