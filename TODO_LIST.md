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
> - 2026-11-11 00:47–onward Crush (glm-5.3): plan execution wave M01–M15 (M13 stays HOLD):
>   CV bump done, bdd/v0.1.0 released + chain-verified, docs-truth pass done (6 archived),
>   SECURITY 0 findings, live smoke green vs local mock, metaengine quantified (184 modules),
>   re-eval verified, guardrails written, brandid in CI (drilled), upstream acks landed,
>   counts reconciled (205, rule: func Test decls), triggers re-checked.

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

| Task | Status | Impact | Effort | Evidence |
| ---- | ------ | ------ | ------ | -------- |

### Medium Impact

| Task                                                                           | Status       | Impact | Effort | Evidence                                                                                                                                                                                                                                                                                  |
| ------------------------------------------------------------------------------ | ------------ | ------ | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Run live smoke test against a REAL provider endpoint                           | 🔵 `BLOCKED` | Med    | S      | the env-gated test PASSED end-to-end 2026-11-11 against a local OpenAI-compatible mock (UA `go-graph-rag/0.5`, Bearer auth, request body, response parse, dims/norm checks — wire facts verified server-side); the real-provider run still needs `GRAPHRAG_LIVE_EMBED_*` creds (M05; F23) |
| Watch tobi/qmd#959 + charmbracelet/crush#3846; patches live in the issue texts | 🔴 `TODO`    | Med    | M      | filed 2026-09-15; owner ruling 2026-09-16: HOLD, no reminder pings — send-or-drop only if the owner asks (M13)                                                                                                                                                                            |

### Low Impact

| Task                                                                                                                                                               | Status       | Impact | Effort | Evidence                                                                                                                                                                                       |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------ | ------ | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| CV-side follow-ups: CI-green re-proof of the v0.5.0 bump (GitHub Actions billing refusal, owner-side) + the MIT attribution answer (public binaries ⇒ notice owed) | 🔵 `BLOCKED` | Low    | S      | bump verified locally 2026-11-11 (parity test green, zero source delta, devshell lint clean); CV CI could not arbitrate; AGENTS Upstream-sync carries the record (M01 residue)                 |
| Align SDK CI golangci pin with CV's v2.15.0 (bump-together rule)                                                                                                   | 🔴 `TODO`    | Low    | S      | CV CI bumped to v2.15.0 2026-11-11 (its devshell version); SDK CI still v2.14.0 (proven green); local 2.15.0 proof blocked by DNS/toolchain-download flake today — bump with a local proof run |
