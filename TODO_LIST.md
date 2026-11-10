# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, use ROADMAP.md.
> Items are ranked by impact. Status is verified, not assumed.
>
> Multi-writer repo: the auto-commit daemon and parallel sessions commit here.
> Sessions touching shared living docs claim below (adopted 2026-11-11, plan
> `docs/planning/2026-11-11_00-13_SUPERB-pareto-execution-plan.md` M08).
>
> - 2026-11-11 00:13–00:30 Crush (glm-5.3): pareto plan + this harvest.

## Status legend

| Status           | Meaning                                                     |
| ---------------- | ----------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                                   |
| 🟡 `IN_PROGRESS` | Actively being worked on.                                   |
| 🔵 `BLOCKED`     | Cannot proceed, external dependency or decision needed.     |
| 🟢 `DONE`        | Completed. Remove from this list and log in `CHANGELOG.md`. |

> DONE rows are pruned at harvest time (all logged in CHANGELOG). The 2026-11-10
> harvest removed 12: license posture, v0.4.0/v0.5.0 releases, v0.4.1 supersession,
> ContentHash, errors.AsType, inbound-grant, v0.3.1 release backfill, ubuntu-26
> proof, version-test, ginkgo typed-table doc, kit coverage audit, upgrading doc,
> EmbedConcurrency doc, StoreRoundTrip bench.

## Open items

### High Impact

| Task                                                                                   | Status       | Impact | Effort | Evidence                                                                                                                                                                                                                                                                                                                                                                                 |
| -------------------------------------------------------------------------------------- | ------------ | ------ | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| CV-side adoption of v0.5.0 (bump + `NodeID`/`ContentHash` migration + json wire delta) | 🔵 `BLOCKED` | High   | M      | primary consumer still pins v0.3.x; every 2026-11-10 deliverance (MIT, breaking types, json/v2) reaches value only through this bump. go-ecosystem-upgrade flow in a CV session; `docs/upgrading.md` has both migration sections. Gates: owner GO + the attribution answer (public binaries ⇒ MIT notice owed downstream). Fold in: CV golangci pin v2.13.2→v2.14.0 (bump-together rule) |

### Medium Impact

| Task                                                                           | Status           | Impact | Effort | Evidence                                                                                                                                                                                                                                                                                            |
| ------------------------------------------------------------------------------ | ---------------- | ------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Tag `bdd/v0.1.0` + release automation for the kit module                       | 🔵 `BLOCKED`     | Med    | S      | kit green, `-race` clean, 47 specs; `Harness.Search` contract FROZEN for v0.1.0 in the bdd ADR (2026-10-15 — the old "API review" prerequisite is settled). Remaining: owner GO + `release.yml` `bdd/v*` trigger (two lines) + planted verify + tag at green SHA + proxy/clean-dir kit fetch verify |
| Docs-truth pass: annotate stale claims, archive resolved reports               | 🔴 `TODO`        | Med    | M      | corpus misleads after 2026-11-10: v1-era "policy" claims, "violation" framings, "Q1 open"/"NodeID unreleased" claims all falsified by events (MIT, v0.3.3, v0.5.0). ANNOTATE inline (never rewrite), `git mv` fully-resolved reports to `docs/status/archived/`; plan task M03 (F14-F19)            |
| Decide SECURITY.md contact channel + practices section                         | 🔵 `BLOCKED`     | Med    | owner  | buildflow securitymd (2026-10-15) wanted channel + practices; the file NOW names GitHub private vulnerability reporting (verified 2026-11-10) — re-run securitymd, owner confirms channel, add practices section if still missing (M04)                                                             |
| Run live smoke test against a real endpoint                                    | 🔵 `BLOCKED`     | Med    | S      | `embed_openai_live_test.go` asserts never executed; needs `GRAPHRAG_LIVE_EMBED_*` creds (M05)                                                                                                                                                                                                       |
| metaengine dep-tree quantification (owner "run soon" item, 2026-09-16)         | 🔵 `BLOCKED`     | Med    | owner  | never scheduled; seam ADR §6 carries the 2026-09 numbers; fresh scratch-module `go mod graph` measurement feeds any adapter-module decision (M06); the 21-07 re-eval re-cited, did not re-measure                                                                                                   |
| Multi-writer guardrails (claims ledger + playbook + CONTRIBUTING note)         | 🟡 `IN_PROGRESS` | Med    | S      | claims-ledger line ADDED to this file 2026-11-11 (F30-adjacent); remaining: AGENTS "check concurrent-session authorship before reverting" line + CONTRIBUTING daemon/sessions auto-commit note (M08)                                                                                                |
| Review the 2026-11-10 metaengine re-eval content claims                        | 🔴 `TODO`        | Med    | S      | mechanics verified (issue #2 closed, artifact committed) but the research's upstream claims (versions, test counts, ADR status) were not re-verified by a second session (M07/F28-29)                                                                                                               |
| Upstream ack: go-cqrs-lite #2 routing note is stale (shipped, says TODO)       | 🔴 `TODO`        | Low    | S      | one-line comment/PR upstream; verify-before-filing at THEIR source first (M11)                                                                                                                                                                                                                      |
| Watch tobi/qmd#959 + charmbracelet/crush#3846; patches live in the issue texts | 🔴 `TODO`        | Med    | M      | filed 2026-09-15; owner ruling 2026-09-16: HOLD, no reminder pings — send-or-drop only if the owner asks (M13)                                                                                                                                                                                      |

### Low Impact

| Task                                                   | Status       | Impact | Effort | Evidence                                                                                                                                                                             |
| ------------------------------------------------------ | ------------ | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Promote `brandid-lint -rules BD001,BD003` to CI        | 🔵 `BLOCKED` | Low    | owner  | currently local-only; CI step + planted-failure drills (BD001 stub, BD003 weak-ID) before enabling (M09)                                                                             |
| Bench re-record under the v2 engine                    | 🔴 `TODO`    | Low    | M      | `bench_test.go`'s newest reference row documents the v1-era tree (2026-11-10); re-run `-benchtime 10x -count 10` + benchstat when numbers next matter (M10)                          |
| pkg.go.dev v0.5.0 banner recheck + local cruft cleanup | 🔴 `TODO`    | Low    | XS     | version page showed the same "not latest" cache banner v0.4.0 had (settled there); also trash empty local `.githooks/` + `docs/drafts/` dirs and decide the drafts convention (M12)  |
| Cross-repo test-count reconcile (60+ vs 206)           | 🔴 `TODO`    | Low    | S      | go-cqrs-lite promotion doc says "60+ tests under -race"; 206 test funcs counted 2026-11-10 — agree one counting rule, fix the wrong doc (M14)                                        |
| Sibling research artifacts trigger re-check            | 🔴 `TODO`    | Low    | S      | system promoted out of EXPERIMENTAL + ADR-0123 Accepted (both 2026-11-10 findings) may fire the reopen-triggers of the two 2026-09-19 metaengine artifacts; re-check after M07 (M15) |
