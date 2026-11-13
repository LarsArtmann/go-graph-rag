# Status Report — golangci-lint config-verify gate repair

- **Date:** 2026-11-13 15:39 CET
- **Repo:** go-graph-rag (module + `bdd/` sub-module)
- **Trigger:** buildflow gate failure (user-pasted run output): all three
  `golangci-lint` steps failed `config verify` (exit 3).
- **Outcome:** GATE GREEN. All three schema-invalid elements fixed in the one
  parent `.golangci.yaml`; behavior proven unchanged; daemon committed the fix
  (`266787e`).

## Scope of this report

**VERIFIED this session (live commands, timestamps from shell output):**

| Claim                                        | Proof                                                                                                                                                                                   |
| -------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Root config schema-valid                     | `golangci-lint config verify` exit 0 (15:02)                                                                                                                                            |
| bdd discovers same config; also schema-valid | `cd bdd && golangci-lint config verify` exit 0 (15:02)                                                                                                                                  |
| buildflow gate green                         | `buildflow -s golangci-lint` → "✓ BuildFlow passed 3/3" (15:03, real 8.3s lint); cached re-run confirmed                                                                                |
| Lint results unchanged by the fix            | `GOTOOLCHAIN=go1.27.1 golangci-lint run ./...` → "0 issues." exit 0, BOTH modules (15:04-15:07) — matches the recorded v2.15.0 zero-issue baseline                                      |
| Root module healthy                          | `go build` + `go vet` + `go test ./...` → ok 3.028s                                                                                                                                     |
| bdd module healthy                           | `env -u GOEXPERIMENT` build + vet + test → ok 0.021s                                                                                                                                    |
| Formatting green                             | `nix run nixpkgs#dprint -- check` exit 0 (15:39)                                                                                                                                        |
| Removed keys were dead config                | sourcegraph @golangci/golangci-lint@v2.15.0: `FunlenSettings` = lines/statements/ignore-comments ONLY; `VarnamelenSettings` has no `ignore-for-root-var`; mnd `IgnoredNumbers []string` |
| Daemon commit intact                         | `266787e` (15:03:59, `.golangci.yaml` only); grep re-check: 35 quoted numbers, 0 hits for dead keys                                                                                     |

**NOT verified (and why):**

- **Remote CI status / push state of `266787e`** — out of scope per instruction
  ("do not research other stuff"); the daemon's push behavior was not checked.
- **Full buildflow pipeline** — only the scoped `-s golangci-lint` step, the
  formatter check, and go tests ran; no whole-pipeline green proof this session.
- **`-race` locally** — CI runs it; local habit is plain `go test` (AGENTS.md).
- **Whether dprint actually covers `.golangci.yaml`** — check passed (exit 0);
  I did not confirm whether the file is IN the include list or silently
  excluded.
- **Anything unrelated to this incident** — per instruction.

## Incident anatomy

`config verify` rejected the parent `.golangci.yaml` with **three** invalid
elements (the pasted buildflow output showed only two — one per step; jsonschema
reports map-order-dependent first errors, so bdd's step surfaced a different
one than root's):

1. `funlen.ignore-strings: true` — property does not exist in golangci-lint
   (any version checked, v1 or v2). Dead config since the extraction.
2. `varnamelen.ignore-for-root-var: true` — same. Found ONLY by running
   `config verify` locally; invisible in the pasted summary. Fixing just the
   two visible errors would have left the gate red.
3. `mnd.ignored-numbers` — 35 bare numbers; schema requires strings.

**Why CI never caught it:** CI (`go-test.yml`) runs `golangci-lint run`, and
`run` does not schema-validate. Only `config verify` does. The dead keys were
behaviorally inert, so every green CI run since `1e65a63` (v0.1.0 extraction,
config copied from CV) proved nothing about schema validity.

**CV sync:** CV's current `.golangci.yaml` already carries quoted numbers and
neither dead key — CV self-fixed at some point after the extraction copy; this
repo was the stale one. No CV-side change needed; this repo now matches CV's
shape for these blocks (keep-in-sync rule satisfied for this incident).

## a) FULLY DONE

1. **Gate repaired end-to-end** — all three invalid elements fixed; `config
   verify` exit 0 from root and from `bdd/`.
2. **Hidden third error found and fixed** (`varnamelen.ignore-for-root-var`) —
   local verify ran BEFORE any edit; no fix-now-repair-later loop.
3. **Behavior-neutrality proven two independent ways** — source-level (struct
   definitions at the exact pinned tag v2.15.0) and empirical (`0 issues.`
   parity with the recorded baseline on both modules).
4. **mnd fix behavior-preserving by construction** — `IgnoredNumbers` is
   `[]string`; quoting `"0".."420"` feeds the linter identical values.
5. **buildflow golangci-lint gate green** — real (uncached) lint 8.3s/step,
   "✓ BuildFlow passed 3/3"; cached re-run + uncached re-invocation confirm.
6. **Both modules healthy** — build/vet/test green (root 3.028s, bdd 0.021s
   under pristine `env -u GOEXPERIMENT`).
7. **dprint check green** (exit 0).
8. **Provenance traced and CV sync checked** — drift origin `1e65a63`; CV
   already clean; zero CV-side work required.
9. **Daemon commit integrity verified** — `266787e` contains exactly the
   intended config change; content re-checked, not just `git status`.
10. **Memory updated** — AGENTS.md gotcha: run-vs-verify validation gap, drift
    history, jsonschema map-order explanation (+10 lines, uncommitted).

## b) PARTIALLY DONE

1. **Upstream proof of the fix** — daemon committed `266787e`, but push state
   and GitHub Actions greenness were NOT checked (out of scope this session).
2. **Full-pipeline confirmation** — scoped gate + formatter + tests green; a
   whole `buildflow` pass was not run.
3. **Close-the-loop on two observations** — (a) scoped `-s golangci-lint` run
   printed a "regression verdict" comparing steps that did not run in that
   invocation (cosmetic noise); (b) `-s dprint` is an unknown tool name in this
   buildflow build. Both worked around; neither filed upstream nor recorded as
   deliberate non-fix yet.
4. **AGENTS.md gotcha** — written, uncommitted (harness forbids commit without
   explicit request; daemon will sweep it, contents verified present at report
   time).

## c) NOT STARTED

1. **`golangci-lint config verify` as a CI step** (go-test.yml, both modules) —
   the mechanization for this exact drift class; CI is structurally blind to
   schema-invalid configs until this lands.
2. **CV-only dead-entry cleanup in this repo's config** — noticed during the
   read, deliberately untouched: CV package refs in `exhaustruct_v5`
   ignore-patterns, `ireturn` allow-list, `wrapcheck` globs, `varnamelen`
   ignore-decls (gin/httpx/sql), CV-only `exclusions.paths`
   (career-pipeline/internal/database/schema/_templ), `gomoddirectives`
   replace-allow-list (cellbuf), `swaggo` formatter. Owner decision needed:
   strip vs keep byte-sync.
3. **TODO_LIST harvest** of this report's section (f).
4. **Drift-prevention mechanism beyond one-off sync** — CI step (above) and/or
   a fleet-level BuildFlow guard.

## d) TOTALLY FUCKED UP

Nothing was destroyed this session — no reverts, no data loss, no lost writes
(Git Town checked: daemon commit content verified intact). Honest misses:

1. **The pasted summary under-reported the failure and I nearly trusted it** —
   two errors visible, three real. Caught only because I ran `config verify`
   as reproduction BEFORE editing. Had I edited straight from the paste, the
   gate stays red after "fixing" it. Lesson recorded in AGENTS.md.
2. **One wasted call** — `buildflow -s dprint` (unknown tool) instead of
   checking `buildflow list providers` first.
3. **AGENTS.md grew +10 lines past the already-flagged 220-line threshold** —
   the standing go-structure-linter `agent-config` error in the buildflow mix
   gets worse; trimming is queued, not done.
4. **Repo-level (pre-existing, surfaced this session):** the gate carried dead
   config for ~2 months because the only schema validator (config verify) runs
   nowhere in CI. CI greenness was false comfort for exactly the class of
   failure the owner dropped the pre-push hook to avoid duplicating.

## e) WHAT WE SHOULD IMPROVE

1. **Always run `golangci-lint config verify` locally before ANY config edit
   batch** — it lists ALL errors; buildflow summaries show one-per-step in
   map order.
2. **Distrust single-error failure reports from schema validators** — iterate
   locally until the error list is empty, then edit.
3. **Source-verify "dead config" claims at the exact pinned tag before
   removing keys** — sourcegraph `@v2.15.0` struct lookup took one query and
   converted "probably unused" into proof. Repeatable pattern for any linter
   config surgery.
4. **Check `buildflow list providers` before scoped `-s` runs** — step names
   are not guessable.
5. **Mechanize drift prevention** — a "keep in sync" line in AGENTS.md did not
   prevent two months of drift; a CI verify step would have caught it at the
   next push.
6. **After daemon commits, verify content not just status** (done here; keep
   doing it).
7. **Keep AGENTS.md additions tight** — the 220-line gate error is standing;
   every gotcha must earn its lines.
8. **Prefer the authoritative AGENTS.md command loop (direct commands + exit
   codes) as final proof** over tool summaries — the two agreed here, but the
   habit is what catches disagreements.

## f) Things to get done next (brainstorm, impact-ranked; not a commitment list)

Session-derived (this incident):

1. Add `golangci-lint config verify` step to `.github/workflows/go-test.yml`
   for BOTH modules — mechanizes the gate CI cannot see today. Highest impact.
2. Commit the AGENTS.md gotcha bullet (needs explicit authorization).
3. Confirm GitHub Actions green on `266787e` (and that the daemon pushed it).
4. Decide CV-only dead-entry policy for `.golangci.yaml` (strip vs byte-sync;
   see question 3).
5. Harvest this report into `TODO_LIST.md` (docs-health HARVEST).
6. Identify this repo's buildflow formatter step name (`buildflow list
   providers`); stop hand-running dprint for gate purposes.
7. File upstream (BuildFlow repo) or record as deliberate non-fix: scoped
   `-s` runs printing "regression verdict" for steps that did not run.
8. Consider a fleet-level linter-config drift guard in BuildFlow (per-repo
   verify exists; CV-vs-repo drift is the fleet gap).
9. Confirm whether dprint's include list actually covers `.golangci.yaml`
   (check passed; coverage unproven).
10. Reconcile `run.go: "1.27"` in the config with the go.mod 1.27.1 floor.

Documented repo debt (known from AGENTS.md, untouched this session, listed for
completeness of the queue):

11. Manual GitHub Releases backfill for v0.3.1 and v0.3.3 (standing TODO row).
12. CV-side GitHub Actions billing fix (blocks the CV CI re-proof of v0.5.0).
13. MIT attribution answer for CV public binaries.
14. metaengine dep-tree quantification (owner "run soon" since 2026-09-16).
15. pkg.go.dev re-verification rides the next tagged release.
16. Trim AGENTS.md below 220 lines (condense owner-decision history into
    archived summaries) — clears the standing buildflow mix error.
17. Pristine-build habit (`env -u GOEXPERIMENT`) for root-module runs too
    (this session applied it only to bdd).
18. Local `-race` runs as an occasional habit (CI is authoritative).
19. CHANGELOG: considered and REJECTED for this fix — repo tooling, not module
    API surface; no consumer-visible change.
20. Full `buildflow` pipeline pass once, to prove fleet-green beyond the
    scoped step.
21. Consider whether `bdd/` deserves its own `.golangci.yaml` (e.g. prod-only
    linters that make no sense for a test kit) — brainstorm, likely "no".
22. Next golangci tightening: sync all three configs (root, bdd discovery,
    CV) in one wave, per the standing keep-in-sync rule.

(Stopped at 22 solid items rather than padding to 50 — every line above is
either session-observed or already documented in AGENTS.md/TODO_LIST; the
remainder would be invented filler.)

## g) Questions I cannot answer myself

1. **Commit authorization:** shall I commit `AGENTS.md` + this report myself
   (enumerated paths, detailed message), or leave both for the auto-commit
   daemon?
2. **CI gate change:** add `golangci-lint config verify` for both modules to
   `go-test.yml`? It changes the CI contract (new failure mode: schema-invalid
   config blocks merge) and would have caught this drift at the first push
   after `1e65a63`.
3. **Config policy:** strip this repo's CV-only golangci leftovers (honest
   divergence where this repo genuinely differs from CV) or keep byte-mirroring
   CV and carry inert entries?

— WAITING FOR INSTRUCTIONS —
