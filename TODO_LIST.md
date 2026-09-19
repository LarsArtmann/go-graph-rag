# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, use ROADMAP.md.
> Items are ranked by impact. Status is verified, not assumed.

## Status legend

| Status           | Meaning                                                     |
| ---------------- | ----------------------------------------------------------- |
| 🔴 `TODO`        | Not started. Needs doing.                                   |
| 🟡 `IN_PROGRESS` | Actively being worked on.                                   |
| 🔵 `BLOCKED`     | Cannot proceed, external dependency or decision needed.     |
| 🟢 `DONE`        | Completed. Remove from this list and log in `CHANGELOG.md`. |

## Open items

### High Impact

| Task                                                    | Status       | Impact | Effort | Evidence                                                                                                                                                                                                                                                                                  |
| ------------------------------------------------------- | ------------ | ------ | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Decide license posture (public + PROPRIETARY + PR flow) | 🔵 `BLOCKED` | High   | owner  | package: `docs/planning/2026-09-15_owner-decision-package.md` Q1; pkg.go.dev hides godoc for EVERY version ("Documentation not displayed due to license restrictions", re-verified live 2026-09-19 on v0.2.0) until a recognized LICENSE exists — Q1 now gates the entire public-SDK face |

### Medium Impact

| Task                                                                           | Status       | Impact | Effort | Evidence                                                                                                                                                                          |
| ------------------------------------------------------------------------------ | ------------ | ------ | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Run live smoke test against a real endpoint                                    | 🔵 `BLOCKED` | Med    | S      | `embed_openai_live_test.go` asserts never executed; needs `GRAPHRAG_LIVE_EMBED_*` creds                                                                                           |
| Watch tobi/qmd#959 + charmbracelet/crush#3846; patches live in the issue texts | 🔴 `TODO`    | Med    | M      | filed 2026-09-15; re-checked 2026-09-19: both OPEN, 0 comments; owner ruling 2026-09-16: HOLD past the 2026-09-22 window, no reminder pings — send-or-drop only if the owner asks |
| Apply CONTRIBUTING inbound-grant line                                          | 🔵 `BLOCKED` | Med    | S      | drafted in the owner package; presumes license answer (Q1); current wording verified unchanged 2026-09-19                                                                         |

### Low Impact

| Task                                                        | Status       | Impact | Effort | Evidence                                                                                                                                                                                                                   |
| ----------------------------------------------------------- | ------------ | ------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Commit the `check-rows.py` SKILL.md edit in the SKILLS repo | 🔵 `BLOCKED` | Low    | S      | the edit landed 2026-09-19 (Tooling paragraph now documents check-rows.py: per-row COMPLETE/UNTOUCHED/PARTIAL classification); external commit/push awaits the owner ruling (plan guard: external repos edit-don't-commit) |

## Completed 2026-09-19 (SUPERB hardening + speed plan, `536f7c6`)

Removed per the TODO lifecycle (done → CHANGELOG): pre-push three-gate
extension (all blocking paths proven), ADR §5 compile guard
(`adrsketch_test.go`), `EmbeddingConfig.EmbedConcurrency` worker pool,
SDK hardening batch (cache fail-fast + Searcher deep-clone + single-hash),
benchmark hygiene (StoreRoundTrip `-count 10`, fixture-cost warning,
`SimilarPairs/docs=100`), lychee redirect URL (SECURITY.md → final target,
re-run 15/15 OK 0 redirects).

## Closed outside this repo (tracked upstream)

- QMD MCP `get`/`multi_get` garbage output — root-caused (embedded-resource
  content items), filed as tobi/qmd#959 and charmbracelet/crush#3846 (the
  Medium watch row owns the follow-up).
- CV-repo extraction-story annotations — committed in the CV repo
  (`16fc16c7`, daemon-committed; awaiting the CV repo's next push).
- CV consumer bump to graph-rag v0.2.0 — pending in the owner package
  Phase 8 (external repo, separate session).
