# Status: Golangci Pin Alignment + Docs-Truth Follow-up Wave

- **Date:** 2026-11-12, 14:07 CET
- **Session:** Crush (glm-5.3), resumed one day after the SUPERB wave (whose report is
  `2026-11-11_11-43_superb-plan-execution-wave-status.md`), on owner command "execute the
  whole TODO list".
- **Scope — what this report VERIFIED:** every claim below was re-verified this session with a
  live command or artifact: the pinned golangci-lint v2.15.0 binary installed at the exact tag
  (`go install …@v2.15.0`, reports "2.15.0 built with go1.27.1", mod-sum confirmed) and proven
  0-issues on BOTH modules BEFORE the CI bump; the BuildFlow govulncheck step (exit 0, all 8
  findings inspected line-by-line from `/tmp/govuln_receipt.log`); the full local gate suite on
  the edited tree (core build/vet/`test -race -count=1` 4.3s, bdd build/vet/`test -race -count=1`
  1.3s, tidy drift clean both modules, dprint fmt+check clean, gofmt clean, brandid BD001,BD003
  exit 0); and CI green on the pushed head `9e4f41e` (run 43217553936, watched to completion
  with `gh run watch --exit-status`, `build-test-lint` 1m42s, master == origin, tree clean).
- **Scope — what this report did NOT verify:** no real-provider network call (still no
  `GRAPHRAG_LIVE_EMBED_*` creds); no CV-side work at all (billing still owner-blocked; the
  instruction for this session was this repo only); the go-cqrs-lite "#4 → TODO v5 systemtest
  row" question was checked read-only but their ledger was NOT reorganized by me; M13
  (qmd#959/crush#3846) untouched per the standing HOLD; the `.crush/shell-output` log still
  exists (trash blocked, see d).

## a) FULLY DONE

1. **golangci pin alignment — the one actionable TODO_LIST row.** Installed the exact pinned
   artifact to an isolated GOBIN (`GOBIN=/tmp/golangci-2150 go install …@v2.15.0` with
   `GOTOOLCHAIN=go1.27.1`), verified the binary self-reports v2.15.0/go1.27.1, ran it on core
   AND bdd — 0 issues each — and ONLY THEN bumped `.github/workflows/go-test.yml` to
   `@v2.15.0` with the provenance lesson in the comment ("prove the pinned artifact, never a
   local build" — the 2026-11-11 brandid red-run class). Pushed with the daemon's commits;
   CI green with the Lint step installing `@v2.15.0` fresh. The bump-together rule is TRUE
   again: CI = dev-shell = CV, all v2.15.0. TODO_LIST row pruned, logged in CHANGELOG
   `[Unreleased]`, AGENTS gotcha rewritten to the aligned state.
2. **govulncheck receipt (f.11).** `buildflow -s govulncheck` exit 0. Characterized ALL 8
   findings from the log, not summarized: every one is Go standard library (`net/http`,
   `crypto/x509`, `crypto/tls`, `mime/multipart`, `net/http/internal/http2`, all `@go1.27`),
   ZERO third-party modules; several have reachable call paths from this SDK's own code
   (`embed_openai.go` http calls, `search.go` `fmt.Fprintf`); all are fixed in newer 1.27.x
   patch toolchains (1.27.3/1.27.4), not by any code or dependency change. The SECURITY.md
   bullet now carries the dated precise receipt — deliberately NOT a blanket "clean" claim.
3. **Docs-truth sweep over live docs** (all view-before-edit, all gates re-run after):
   - README: BDD install line reflects the released kit (`bdd/v0.1.0`, `bdd/v*` namespace) —
     the "first bdd/v* tag pending (owner-gated)" fossil is gone (f.25/f.45).
   - `docs/upgrading.md`: headings now name actual releases — "To v0.5.0 (encoding/json/v2)"
     and "To v0.3.3 (`NodeID`)" with the unlisted-shipping honesty note (f.37).
   - AGENTS ×6: the json-gotcha's catcher sentence corrected (it cited the v1-era grep
     direction as what CI enforces — CI greps the bare v1 import; nothing is exempt, including
     `wire_golden_test.go` which imports v2); the golangci gotcha rewritten to the aligned
     state; the release.yml bullet now covers `bdd/v*` + documents the CHANGELOG
     section-heading contract (f.29); the bdd/ bullet reflects the released kit; the
     multi-writer gotcha gained the park-then-sweep pattern (f.42).
   - `bench_test.go`: canonical re-measure protocol line (`-benchtime 10x -count 10` +
     benchstat, "compare re-measures to re-measures") so the next bench run is apples-to-apples
     (f.22); AGENTS protocol sentence aligned to match.
   - CONTRIBUTING: explains the dropped local hooks and the stale `core.hooksPath` one-line
     cleanup for new-clone contributors (f.48).
   - SECURITY ×3: `bdd/v0.1.x` support line (f.16), the dated govulncheck receipt (f.11), the
     wire-golden bullet naming `encoding/json/v2` explicitly (f.31).
4. **Checks that correctly ended in "no change needed":** f.10/f.44 — the wire-contract test
   over real TCP ALREADY EXISTS: `embed_openai_test.go` spins `httptest.NewServer` and asserts
   all three wire facts verbatim (`User-Agent: go-graph-rag/<version>`, `Bearer` auth, model
   in body); the M05 python mock added nothing that a committed test should duplicate. f.46 —
   FEATURES bdd rows already current (no owner-gated markers left). f.19/f.20 — the re-eval
   artifact's "60+ tests per FEATURES" citation now carries a dated annotation (their FEATURES
   changed under the citation; point-in-time text preserved); the research index itself is
   current. f.35 — the strong-ID FEATURES row wording survived intact.
5. **TODO_LIST harvest + truth-sync.** Golangci row pruned (DONE, logged in CHANGELOG); 7 rows
   added from the status-report (f) sweep — 5 TODO (go-branded-id `linter/v0.7.x` tag f.7;
   brandid drill fixture + gate self-test f.9 with a design caveat; release dry-run script
   f.18; pkg.go.dev periodic render check f.17; go-cqrs-lite f.28 flag for THEIR docs-health
   pass) and 2 BLOCKED (`core.hooksPath` retirement f.12 needs an explicit owner ask; Go
   floor/toolchain posture — floor is owner-pinned, af3be2c precedent). Session ledger line
   added (2026-11-12). CHANGELOG `[Unreleased]` gained three Changed bullets.
6. **Full local gate suite green on the edited tree:** core build/vet/test `-race -count=1`,
   bdd build/vet/test `-race -count=1` (`env -u GOEXPERIMENT`), pinned-lint 0 issues both,
   tidy drift clean both, dprint fmt + check clean, gofmt clean, brandid exit 0. Then pushed
   (`c0dc47e..9e4f41e`) and watched CI to green (run 43217553936, `build-test-lint` 1m42s,
   master == origin, tree clean, 0 modified files).

## b) PARTIALLY DONE

1. **f.9 brandid drill fixture — designed, not built.** The design hazard is real: a planted
   violation must live where the root gate does NOT scan while an explicit polarity CI step
   scans it — and whether brandid-lint skips `testdata/` is UNVERIFIED. Rather than
   half-build a fixture that could either permanently redden the gate or prove nothing, it is
   a TODO row with the design note. Effort M, deliberate.
2. **f.28 (go-cqrs-lite routing header)** — checked read-only: their "#4 → TODO v5 systemtest
   row" target is not findable under that name in their TODO_LIST (the systemtest hits there
   are DONE/struck-through W3 receipts). I did NOT edit their point-in-time ledger: my
   uncertainty is about THEIR row taxonomy, and annotating a sibling's archived doc on an
   incomplete read is the cross-repo overreach the multi-writer protocol warns about. Recorded
   as a TODO row flagging their next docs-health pass.
3. **f.40 (`.crush/shell-output`)** — confirmed gitignored (`!!` under
   `status --porcelain --ignored`), stale 39 KB log present since 01:50; `trash` FAILED
   (cross-volume, see d). Residue stays, documented in this report.
4. **CV CI-green re-proof of the v0.5.0 bump** — still impossible (GitHub Actions billing
   refusal, owner-side); untouched this session, out of the session's scope.
5. **M05 real-provider live smoke** — still creds-blocked; the local-mock proof from yesterday
   stands unchanged.

## c) NOT STARTED (deliberately)

1. **M13 — qmd#959 / crush#3846 send-or-drop.** Standing owner ruling: HOLD, no reminder
   pings; a blanket "execute the TODO list" does not override it (consistent with yesterday).
2. **go-branded-id `linter/v0.7.x` tag** (f.7) — different repo, go-release territory;
   TODO row only.
3. **Release dry-run script** (f.18) and **pkg.go.dev periodic render check** (f.17) —
   TODO rows only.
4. **Go floor/toolchain posture decision** — BLOCKED on the owner (floor pins are
   owner-requested historically); the receipt gives the owner the data to decide.
5. **CV MIT attribution follow-through** — owner question unanswered; nothing started.

## d) TOTALLY FUCKED UP

1. **Dated my own actions with the wrong day — six times.** The session resumed a day after
   the handoff and I wrote "2026-11-11" stamps into five files (go-test.yml, AGENTS, SECURITY,
   bench_test.go, the re-eval annotation) before running `date`; today is 2026-11-12. Caught
   by the /tmp file timestamps ("Nov 12"), fixed all stamps pre-push — but the root failure is
   that I trusted session continuity over the clock for DATED CLAIMS in a repo that treats
   dates as evidence. The rule is now obvious: `date` first in any resumed session, before
   any dated write.
2. **Stale-read edit failure on bench_test.go** — my "read" was a bash `head`, which does not
   register with the edit tool's freshness check; the daemon had swept the file at 09:15 and
   the edit bounced ("modified since read"). Zero damage (the tool caught it — that is what it
   is for), but it was the same lesson AGENTS already records: view-before-edit is not
   ceremonial, it is the mtime race contract.
3. **Self-inflicted anchor miss** on the go-test.yml date fix: I anchored my own freshly
   written text as if it started at line-begin when it sat mid-line after a parenthetical.
   One retry. Cheap, and entirely avoidable by anchoring on the exact rendered line.
4. **The working-tree "shrank" mid-session** — 7 modified files I was tracking became
   daemon-committed under me while the gates ran. I reacted correctly (authorship check per
   the multi-writer gotcha — every file was mine), but for one moment the tree disagreed with
   my mental model, and that moment is exactly where a reflexive `git restore`/`checkout`
   would have destroyed committed work. The discipline held; the hazard was real.
5. **trash(1) cross-volume dead end**: `.crush/shell-output` lives on `/mnt/hot`; trash's
   fallback trash dirs cannot be created there (permissions). I attempted the trash before
   understanding the constraint. The safety rule (never `rm`) held, so the outcome is honest
   documented residue — but the attempt order was wrong: check the volume layout of a path
   before promising its deletion.
6. **Scope interpretation honesty**: "execute the WHOLE TODO list" — the TODO_LIST had exactly
   ONE actionable row; the rest of this wave came from the prior report's section (f), which
   yesterday's session pre-declared "TODO_LIST-ready". Executing them was my interpretation of
   the owner's intent (backed by this repo's proactive-maintenance culture), not list
   membership. If the owner meant strictly TODO_LIST rows, I over-delivered; if they meant
   everything finishable, f.9/f.17/f.18 remain rows, not work. Named here so the boundary is
   visible, not silent.

## e) WHAT WE SHOULD IMPROVE

1. **Clock before claims.** Any resumed session runs `date` before its first dated write.
   Handoff dates describe yesterday; only the clock describes now.
2. **View-tool reads are the freshness contract.** bash `head`/`cat` do not register; on a
   daemon-shared tree every edit must ride a view-tool read seconds old.
3. **Prove the pinned artifact, then bump** — executed correctly this time (install exact tag
   → run both modules → only then touch CI). Keep it as the standing ritual for every CI
   tool-pin change; the comment in go-test.yml now teaches it to the next session.
4. **Receipts characterize, they don't cheerlead.** "govulncheck clean" would have been a
   lie; "8 stdlib findings, zero third-party, fixed in newer toolchains" is a receipt. The
   same standard applies to any future scanner output written into SECURITY/docs.
5. **Flag-into-a-row beats annotate-their-doc** when a sibling repo's ledger looks wrong and
   authorship/understanding of their taxonomy is incomplete. The f.28 restraint cost nothing
   and avoided a wrong edit in someone else's point-in-time record.
6. **Check volume layout before promising deletions** (`trash` cannot cross volumes here);
   when a cleanup is environment-blocked, leave documented residue rather than escalating to
   destructive tools.
7. **Bake proof commands into TODO rows at creation** (the f.23 pattern): the golangci row
   eventually carried its exact proof command; rows born from harvests should ship with theirs
   (`linter/v0.7.x`: `git tag linter/v0.7.x` + proxy check + CI bump drill; fixture row: the
   testdata-skip experiment command).

## f) Top 25 things we should get done next

_Impact-sorted. Items 1–3 are owner-owned; 4–9 are this repo's executable rows; the tail is
context/fuel. The 2026-11-11 report's (f) items are consolidated here with their outcomes._

1. **Fix CV GitHub Actions billing** (owner) — still the sole blocker riding on it: the v0.5.0
   CI-green re-proof (yesterday f.3), the quiet-window all-green receipt (f.21), the billing
   ticket link (f.41), and the auto-closing "re-proof owed" bullet (f.43).
2. **Answer the MIT attribution question** (owner): does CV ship public binaries? Notice work
   is real only if distribution is public.
3. **Multi-writer policy ruling** (owner): parallel sessions + claims-lane ledger vs
   single-window ownership (yesterday g.3 + f.49) — today's daemon-interleaved sweeps were
   again survived by discipline, not mechanism.
4. **Tag go-branded-id's linter submodule `linter/v0.7.x`** (f.7) — CI pins `linter/v0.5.1`;
   the local build is ~2 minors ahead (BD003 refinements, `-fix` for BD001). Cross-repo,
   go-release flow, then a deliberate CI pin bump with the drill.
5. **Brandid drill fixture + gate polarity self-test** (f.9) — design caveat in the row:
   verify whether brandid-lint scans `testdata/` BEFORE choosing the fixture location.
6. **Release dry-run script** (f.18) — one command for the CHANGELOG pattern sim + extraction
   sim + actionlint that proved the bdd/v0.1.0 chain.
7. **pkg.go.dev periodic render check** (f.17) — @latest renders + License reads MIT, cron or
   CI step, so the "renders" claim never goes stale again.
8. **Go floor/toolchain posture** (owner decision, new): the receipt shows every current
   advisory is stdlib fixed in 1.27.3/1.27.4 patches while CI compiles with go1.27.1 — decide
   whether consumers/CI should build on the newest 1.27.x patch toolchain independent of the
   language floor, or accept the frozen-toolchain exposure until the next floor raise.
9. **go-cqrs-lite docs-health pass** (their repo): the f.28 routing-header row plus their own
   "205 tests" recount sync — their ledger, their pass; this repo only flagged it.
10. **Live smoke vs a REAL provider** — `GRAPHRAG_LIVE_EMBED_*` creds remain unprovisioned;
    the mock proof covers mechanics only (TLS, latency, provider quirks, auth-failure paths
    unexercised).
11. **M13 send-or-drop** (qmd#959 / crush#3846) — explicit owner ask only; patches live in
    the issue texts.
12. **CI tool-pin drift policy** (f.38): brandid and golangci are now both pinned with proven
    artifacts; consider one policy row covering periodic re-pin + bump-together families
    instead of per-tool rows.
13. **bdd/v0.1.0 consumer proof in the wild** — the kit's only consumer so far is our own
    clean-dir test; a real external consumer via the proxy is unproven.
14. **Wire-golden regeneration drill** — run `go test -run WireGolden -update` once on a
    throwaway branch to prove the documented regeneration path works before a real
    wire-contract event depends on it.
15. **CV parallel session's gate legs** (f.6) — their session closes them (wrapcheck batch,
    secrets-sweep adjudication, `MissionID` glossary row, portal-probe alignment, templ pin
    lockstep, docs-index rows).
16. **Retire the local `core.hooksPath`** — needs the owner's explicit ask (git config
    mutation); CONTRIBUTING now documents the one-liner.
17. **`.crush/shell-output` manual cleanup** — trash-blocked (cross-volume); owner may `rm`
    it directly.
18. **lessons.md candidate** (f.34, fleet-wide): "prove the pinned artifact" + multi-module
    tag namespaces (`linter/v*` trap) — belongs in the crush-config repo by commit, not an
    in-session write to the read-only install.
19. **CHANGELOG `[Unreleased]` bullets ride the next release cut** — no release pending;
    verify at cut time that the extraction picks the right section (core strip-`v` rule).
20. **f.27 (bdd section "Changed" entry) stays skipped deliberately** — the `[bdd/v0.1.0]`
    section is a published release record; annotate-not-rewrite wins.
21. **EmbedConcurrency bench under realistic corpus sizes** (f.30) — ROADMAP fuel; current
    numbers are small-corpus overhead-dominated.
22. **pkg.go.dev MIT/godoc re-verification rides the next tagged release** — v0.5.0 verified
    2026-11-11; nothing owed until the next tag.
23. **CV README stale-pin sweep** (f.26) — their next CV wave; I did not touch CV files.
24. **Fleet sweep**: copy the "prove the pinned artifact" CI-comment pattern to other repos'
    tool pins (branching-flow et al.) — fleet-wide consistency candidate.
25. **TODO_LIST re-rank at next harvest** — the list now carries 2 Medium + 9 Low rows; the
    CI-tool-pin-family rows (12 above) could merge into one policy row to keep the list
    honest.

## g) Top #1 question I can NOT figure out myself

**The Go floor/toolchain posture:** govulncheck proved that every current advisory affecting
this SDK is a standard-library one fixed in 1.27.3/1.27.4 patch toolchains — while CI's
`setup-go` pins go1.27.1 and the language floor is owner-pinned at 1.27.1 (af3be2c). Should
CI (and the documented consumer guidance) build on the newest 1.27.x patch toolchain,
decoupled from the language floor, or is freezing the build toolchain at the floor an
accepted exposure until the next owner-requested floor raise? This decides whether the
SECURITY.md receipt's advice ("build consumers with the newest 1.27.x toolchain") is the
whole answer or whether the repo itself should lead by example — and floor changes have
always been owner decisions, so I did not touch it.

---

_Report filed 2026-11-12 14:07 CET. Point-in-time record — annotate, never rewrite. Daemon
will commit; no manual commit per harness contract._
