# TODO List

> Short-term, actionable, bounded work items, verified against the actual code.
> For long-term vision and unrefined ideas, use ROADMAP.md.
> Items are ranked by impact. Status is verified, not assumed.

## Status legend

| Status           | Meaning                                                     |
| ---------------- | ------------------------------------------------------------ |
| 🔴 `TODO`        | Not started. Needs doing.                                   |
| 🟡 `IN_PROGRESS` | Actively being worked on.                                   |
| 🔵 `BLOCKED`     | Cannot proceed, external dependency or decision needed.     |
| 🟢 `DONE`        | Completed. Remove from this list and log in `CHANGELOG.md`. |

## Open items

| Task                                                             | Status       | Impact | Effort  | Evidence                                                                        |
| ---------------------------------------------------------------- | ------------ | ------ | ------- | ------------------------------------------------------------------------------- |
| Decide license posture (public + PROPRIETARY + PR flow)          | 🔵 `BLOCKED` | High   | owner   | package prepared: `docs/planning/2026-09-15_owner-decision-package.md` Q1       |
| Cut `v0.2.0` (godoc examples invisible on pkg.go.dev until tag)  | 🔵 `BLOCKED` | High   | owner   | package prepared: `docs/planning/2026-09-15_owner-decision-package.md` Q2       |
| QMD MCP `get`/`multi_get` returns serialized garbage             | 🔴 `TODO`    | Low    | 60min   | external: crush-config repo; repro captured in the 2026-09-15 pareto plan M14   |
| CV-repo extraction-story annotations                             | 🔴 `TODO`    | Low    | 15min   | external: CV status report item c7/c8 + TODO_LIST graphrag row (plan M12)       |
