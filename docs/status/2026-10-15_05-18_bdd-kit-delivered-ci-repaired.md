# Status: BDD kit delivered, CI repaired — session self-review (2026-10-15 05:18)

Scope: the 2026-10-15 ~04:25–05:15 session (continuation of the 04:21 blocked
session). Everything below is from this session's run and what it directly
exposed. Format: `.md` at the operator's explicit request (status-report skill's
HTML default overridden per its escape clause).

---

## a) FULLY DONE

All items verifiably complete (tests/lint green, in HEAD via daemon commits,
artifacts listed by `git ls-files`).

1. **Blocker root-caused and killed.** The "go tool package-classification
   anomaly" was the spec FILENAMES: `*_spec.go` instead of `*_spec_test.go` —
   only the `_test.go` suffix makes a file a test file; the `package bdd_test`
   clause does not. Fixed with `git mv` (5 files). Evidence: suite compiles and
   runs immediately after.
2. **GinkgoT()/testing.TB incompatibility fixed.** `FullGinkgoTInterface`
   lacks `private()`; the kit now defines a minimal `TB` interface
   (`Fatalf`/`TempDir`/`Cleanup`) satisfied by both. Evidence: ginkgo
   `GinkgoTInterface` signatures verified in module cache before writing.
3. **Broken spec fixed.** "propagates query embedding failures" built with an
   always-failing provider (Build could never succeed); now swaps the provider
   post-build so only the query embed fails. Evidence: spec passes.
4. **Spec suite fully green and hardened.** 43/43 specs; order-independence
   verified across 3 fresh random seeds; `-race` clean; zero committed focus
   (grep-proven). Ginkgo's `-count>1` ban discovered and worked around
   (repeated `-count=1` runs).
5. **golangci-lint bdd: 17 findings → 0.** Auto-fixed 10 format findings
   (gci/goimports/golines); hand-fixed wsl_v5 x3, exhaustruct x3, contextcheck
   x1, and realigned 5 `//nolint:ireturn` directives the formatter had
   decoupled from their findings (surfacing 5 nolintlint staleness errors).
6. **erraudit (bdd) fixed.** `errors.As` → `errors.AsType` in
   `build_spec_test.go` per the go-error-modernization skill. Evidence:
   `buildflow -s erraudit` no longer reports bdd.
7. **Pre-push hook extended to both modules and PROVEN.** Gate 1 gains the bdd
   build+vet arm (`go vet` included because `go build` skips `_test.go` files —
   the spec suite IS most of that module); gate 2 covers both tidy drifts.
   Proven by planted failures: vet-detectable spec bug blocks (exit 1), planted
   require blocks (exit 1), clean tree passes (exit 0).
8. **CI workflow extended.** Four new bdd steps (build+vet, `test -race`, lint,
   tidy drift), `-race` added to the root Test step (closes the
   `race-condition-test` buildflow error).
9. **Red master CI root-caused and fixed in-tree.** Cause: `setup-go`'s
   `go-version: "1.27"` resolved to Go 1.27.2 on the runner, whose export-data
   format (v5) the pinned golangci-lint v2.13.2 cannot decode — every typecheck
   import failed (9 errors), master red 2026-10-11..15 (3 runs verified via
   `gh`). Fix: pin `go-version: "1.27.1"` (restores the last-green combo) with
   an explanatory comment.
10. **Core defect fixed (one line).** Dead allocation at `embed_openai.go:270`
    introduced by daemon commit `2435271a` (2026-10-11 refactor left
    `make([]Vector, …)` immediately overwritten by `lo.Flatten`). Semantics
    preserved; ineffassign gone; CHANGELOG entry written. This is the session's
    ONLY core change (constraint deviation — see d5).
11. **json/v2 graduation discovered and documented.** Empirically verified
    (core plant, bdd plant, /tmp module, `go env` config check): on go1.27.1,
    `encoding/json/v2` imports compile WITHOUT `GOEXPERIMENT`. The pristine
    build gate can no longer discriminate json/v2; the v1-only rule is now
    policy, not mechanics. Hook header, AGENTS non-negotiable #2/#5, and the
    recurring-regression gotcha all corrected; gate-1 proof re-done with a real
    vet failure.
12. **Docs delivered.** ADR `docs/planning/2026-10-15_05-02_bdd-submodule-adr.md`
    (decision + binding interface sketch + tagging/release notes + YAGNI drops);
    README "BDD test kit" section; CHANGELOG Unreleased (1 Added, 2 Fixed);
    FEATURES "Testing kit" section (7 rows); TODO_LIST (3 new rows); AGENTS.md
    (9 edits: layout, commands, non-negotiables, gotchas incl. the
    `_test.go`-suffix trap, owner-decisions section); 04:21 status report
    annotated with the resolution (annotate-don't-rewrite rule).
13. **Hygiene.** dprint fmt+check clean repo-wide (the check caught my
    unformatted workflow edit); /tmp probe/repro/copy artifacts trashed;
    buildflow noise re-baselined (agent-config + race-condition errors cleared;
    remaining = the two documented owner-gated rows); all artifacts verified in
    HEAD.

## b) PARTIALLY DONE

1. **CI green on master — fix committed, run UNVERIFIED.** I never push
   (harness rule), so the workflow fix has not been proven by an actual
   Actions run. Confidence is high (restores the exact last-green pin combo),
   but the red badge persists until the next push. Secondary risk: lint-parity
   between local golangci-lint v2.14.0 and CI's pinned v2.13.2 for the NEW bdd
   module steps (v2.13.2 might flag something v2.14.0 does not). Effort to
   close: owner push + one CI run (S).
2. **Kit adoption: zero external consumers.** The kit is dogfooded only by its
   own suite. CV / knowledge-graph / legal-graph-ai-system have not adopted it
   (adoption is deliberately unscheduled; ADR §6 sets the revisit trigger).
   Effort: one pilot PR in a consumer (M).
3. **v0.3.2 shipped from red-CI code.** Tagged today 00:11 (daemon) with the
   dead allocation and the CI-breaking workflow. The tag itself is sound
   (tests pass); the fixes are post-tag and need a v0.3.3. Not yet scheduled
   (see f4).

## c) NOT STARTED

1. `bdd/v0.1.0` tag + release.yml `bdd/v*` trigger decision (owner-gated;
   release.yml's `v*` trigger will NOT fire for kit tags).
2. Consumer pilots (CV `platform/bddkit` comparison; legal-graph-ai-system's
   planned testify→ginkgo migration is the natural first user).
3. Dependabot/renovate coverage for the `bdd/` module (config existence
   unverified this session).
4. Kit spec-coverage audit (are `WeightedEdge`, `WithKind`, and
   `SimilarEdgeScore` each exercised by a spec? Not verified).
5. AGENTS "Upstream sync" refresh — still says "latest v0.3.0, CV bump
   pending"; v0.3.2 exists and earlier research observed CV at v0.3.2.
6. buildflow binary rebuild (freshness warning: built 518739c, repo HEAD
   d114f48) and the go-env legacy-DB migration (both advisory; noted, not
   acted on — outside this session's scope).

## d) TOTALLY FUCKED UP

Nothing irreversible this session. The honest fuckups, worst first:

1. **(Inherited, then repeated once by me) wrong debugging order on the
   blocker.** The prior session byte-verified file CONTENTS (`cat -A`) but
   never re-checked file NAMES; I nearly repeated the pattern by reading
   harness.go before `ls`-ing. The filename was visible in the first directory
   listing. Cost: ~a full prior session + this session's first hour.
2. **YAML colon bug — my own edit — then blamed the formatter.** Step names
   like `BDD module: build + vet` contain `:`, an ambiguous YAML plain
   scalar. When dprint choked, I bisected the PLUGIN (probe files, variant
   A/B, c1/c2/c3 — ~7 tool calls) before running a plain `yaml.safe_load`.
   The parse error (`expected '...'`) was a syntax error pointing at my line
   the whole time.
3. **Edited stale files three times.** `golangci-lint fmt` invalidated my
   earlier views; three edit attempts failed on "modified since last read"
   before I learned to re-view after ANY formatter run.
4. **Gate-1 "proof" initially rested on an unverified premise.** I planted
   json/v2 expecting the historical failure; exit 0 exposed that the premise
   (json/v2 gating) was 26 days stale. The exit-code discipline caught it
   immediately — but I should have empirically verified the plant class BEFORE
   building the proof on documented memory.
5. **Constraint deviation reported late.** I declared "core byte-identical" a
   hard constraint, then changed `embed_openai.go` (justified: daemon defect +
   red CI) and surfaced the deviation only in the final summary instead of at
   the moment of decision.
6. **Stale LSP diagnostic noise all session.** The ineffassign warning kept
   rendering at the OLD line number after the fix; I never restarted the LSP
   (binary lint confirmed 0 issues twice). Minor, but I let noise sit.

## e) WHAT WE SHOULD IMPROVE

1. **Cheapest-empirical-check-first debugging.** Before any deep bisect:
   filenames, `yaml.safe_load`, exit codes, `--version`. Two of this session's
   three time-sinks were answerable in one command.
2. **Trust docs less, verify-load-bearing claims more.** AGENTS.md's json/v2
   claim was authoritative AND stale. Documented claims get a 5-second
   empirical check before a gate is built on them.
3. **Re-view after formatters.** Any `fmt` run invalidates prior reads; make
   re-viewing reflexive.
4. **State constraint deviations when making them**, not in the final report.
5. **Pre-push hook lacks a lint gate.** Lint regressions now surface only in
   CI (which nobody watched for 4 days — see b1). A fast gate-4 (golangci
   changed-packages only) would close that gap.
6. **API freeze discipline before tagging.** `Harness.Search`'s
   variadic-first-wins options (`...SearchOptions`, only [0] used) is
   unidiomatic; decide polish-vs-freeze BEFORE `bdd/v0.1.0` makes it public.
7. **Watch CI after pushes.** Master sat red for 4 days unnoticed. The daemon
   pushes but nobody observes. Consider a notification or a post-push
   `gh run watch` habit in the owner loop.

## f) Up to 50 things to do next (ranked)

| #  | Task                                                                                                                         | Impact   | Effort | Category      |
| -- | ---------------------------------------------------------------------------------------------------------------------------- | -------- | ------ | ------------- |
| 1  | Owner: push master and verify the go-test run goes green                                                                     | Critical | S      | Bug           |
| 2  | After green: tag `bdd/v0.1.0` (consumers cannot fetch the kit until then)                                                    | Critical | S      | Release       |
| 3  | Decide release automation for the kit: extend release.yml trigger to `bdd/v*` or manual releases                             | High     | S      | Release       |
| 4  | Cut core v0.3.3 (dead-allocation fix; bundle erraudit sweep) — v0.3.2 shipped with it                                        | High     | S      | Release       |
| 5  | Core erraudit modernization: `errors.As`→`AsType` at `embed_openai.go:334`, re-check `embed_openai_test.go:249`              | High     | XS     | Quality       |
| 6  | SECURITY.md: owner picks contact channel (GitHub advisory vs email), then `buildflow -s securitymd --fix`                    | High     | S      | Security      |
| 7  | Kit API review BEFORE the tag: `Harness.Search` options shape (struct or explicit option vs variadic-first-wins)             | High     | S      | Quality       |
| 8  | Verify `.github/dependabot.yml` covers the `bdd/` module; add if missing                                                     | Med      | XS     | Quality       |
| 9  | Kit spec-coverage audit (`WeightedEdge`, `WithKind`, `SimilarEdgeScore` direct use)                                          | Med      | S      | Quality       |
| 10 | CV pilot: evaluate the kit against CV's `platform/bddkit` for graphrag specs                                                 | Med      | M      | Feature       |
| 11 | legal-graph-ai-system pilot: its planned testify→ginkgo migration on this kit                                                | Med      | M      | Feature       |
| 12 | Bump golangci-lint v2.13.2 → v2.14.x WITH CV parity (both repos together; removes the version-skew risk from b1)             | Med      | S      | Quality       |
| 13 | Add pre-push gate 4: golangci-lint on changed packages (both modules)                                                        | Med      | S      | Quality       |
| 14 | Refresh AGENTS "Upstream sync" (v0.3.2 exists; CV observed at v0.3.2)                                                        | Med      | XS     | Documentation |
| 15 | Rebuild the buildflow binary (freshness warning) + run the legacy-DB migration                                               | Low      | XS     | Tooling       |
| 16 | Record kit coverage baseline (`go test -cover ./...` in bdd/)                                                                | Low      | XS     | Quality       |
| 17 | Post-push: confirm CI's v2.13.2 lint agrees with local v2.14.0 on the bdd module                                             | Med      | XS     | Verification  |
| 18 | Document a local-only `go.work` convention (gitignore policy forbids committing one)                                         | Low      | XS     | Documentation |
| 19 | Add an ADR §6 trigger check when the second consumer adopts (CreateBDDSuite parity question)                                 | Low      | XS     | Documentation |
| 20 | Q1 license decision (pre-existing; gates pkg.go.dev for core AND kit)                                                        | High     | owner  | Documentation |
| 21 | Live smoke test (`GRAPHRAG_LIVE_EMBED_*` creds; pre-existing BLOCKED row)                                                    | Med      | S      | Verification  |
| 22 | Consider CI lint-parity guard: run CI's exact pinned version locally in a pre-tag script                                     | Low      | S      | Quality       |
| 23 | Sweep the 04:21 status report's 50-item list against current TODO_LIST during next docs-health HARVEST (partially stale now) | Low      | S      | Documentation |
| 24 | If kit adoption happens: add a consumer-compat spec module pattern note to the ADR (how to pin kit + SDK together)           | Low      | S      | Documentation |
| 25 | After bdd/v0.1.0: update README's install snippet (remove "requires a tag" caveat)                                           | Low      | XS     | Documentation |

## g) Questions I cannot answer myself

1. **Push + tag sequencing.** You push (I must not): should `bdd/v0.1.0` be
   tagged immediately after the CI run goes green, or only after a consumer
   pilot (CV or legal-graph) has actually exercised the kit — i.e., is v0.1.0
   a "usable now" tag or a "consumers-first" tag?
2. **Kit release automation.** Extend `release.yml` to trigger on `bdd/v*`
   tags with CHANGELOG extraction for the kit, or keep kit releases manual
   until v1? (The current `v*` trigger will silently never fire for the kit.)
3. **SECURITY.md contact channel.** GitHub private vulnerability reporting or
   a security email? This is public-facing policy text; buildflow can draft
   the section once the channel is chosen.

---

Point-in-time snapshot. Annotate, never rewrite. Section (f) feeds the next
docs-health HARVEST pass.
