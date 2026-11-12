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
> - 2026-11-12 ~13:00–onward Crush (glm-5.3): follow-up wave on owner GO "execute the
>   whole TODO list": golangci CI pin aligned v2.14.0→v2.15.0 (pinned artifact proven
>   0-issues on BOTH modules before the bump), govulncheck receipt pinned in SECURITY.md,
>   docs-truth sweep (README bdd install, upgrading headings, bench protocol line,
>   AGENTS gotcha corrections), status-report f-item harvest into the rows below.

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

| Task                                                                                                                                                                                                                                                        | Status       | Impact | Effort | Evidence                                                                                                                                                                       |
| ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| CV-side follow-ups: CI-green re-proof of the v0.5.0 bump (GitHub Actions billing refusal, owner-side) + the MIT attribution answer (public binaries ⇒ notice owed)                                                                                          | 🔵 `BLOCKED` | Low    | S      | bump verified locally 2026-11-11 (parity test green, zero source delta, devshell lint clean); CV CI could not arbitrate; AGENTS Upstream-sync carries the record (M01 residue) |
| Tag go-branded-id's linter submodule (`linter/v0.7.x`): CI pins `linter/v0.5.1` while the local build (BD003 refinements, `-fix` for BD001) is ~2 minors ahead and untagged; a fresh tag lets CI catch up deliberately                                      | 🔴 `TODO`    | Low    | M      | found 2026-11-11 (status report f.7); the release lives in the go-branded-id repo (go-release territory)                                                                       |
| Committed brandid drill fixture + gate self-test so the CI gate's both-polarity proof is reproducible from the tree, not session history; design note: the planted violation must live where the root gate does NOT scan but an explicit polarity step does | 🔴 `TODO`    | Low    | M      | found 2026-11-11 (status report f.9); drills were session-shell-only                                                                                                           |
| Release dry-run script (CHANGELOG pattern sim + extraction sim + actionlint) so the M02-style pre-tag simulation is one command                                                                                                                             | 🔴 `TODO`    | Low    | S      | found 2026-11-11 (status report f.18); the sims proved the bdd/v0.1.0 release chain                                                                                            |
| pkg.go.dev periodic render check (@latest renders + License reads MIT) as a cron or CI step, so the "renders" claim never goes stale again                                                                                                                  | 🔴 `TODO`    | Low    | S      | found 2026-11-11 (status report f.17); godoc-restriction history in AGENTS                                                                                                     |
| go-cqrs-lite: their feedback routing header cites a "TODO v5 systemtest row" (feedback #4) not found under that name in their TODO_LIST (their systemtest hits are DONE receipts) — flag for THEIR next docs-health pass                                    | 🔴 `TODO`    | Low    | S      | checked 2026-11-12 (status report f.28); #2 same class verified shipped 2026-11-11 (M11)                                                                                       |
| Retire the stale local `core.hooksPath` (points at the trashed `.githooks/`; harmless — git ignores missing hook dirs)                                                                                                                                      | 🔵 `BLOCKED` | Low    | S      | needs explicit owner ask (git config mutation out of sessions' lane); CONTRIBUTING now documents the one-line cleanup                                                          |
| Go floor/toolchain posture: the 2026-11-12 govulncheck receipt shows every advisory is Go stdlib fixed in a newer 1.27.x patch — consider pointing consumers at the newest patch toolchain or raising the floor                                             | 🔵 `BLOCKED` | Low    | S      | floor is owner-pinned (af3be2c precedent); receipt in SECURITY.md                                                                                                              |
