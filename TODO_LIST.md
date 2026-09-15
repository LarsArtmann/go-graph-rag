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

| Task                                                            | Status       | Impact | Effort | Evidence                                                                  |
| --------------------------------------------------------------- | ------------ | ------ | ------ | ------------------------------------------------------------------------- |
| Decide license posture (public + PROPRIETARY + PR flow)         | 🔵 `BLOCKED` | High   | owner  | package prepared: `docs/planning/2026-09-15_owner-decision-package.md` Q1 |
| Cut `v0.2.0` (godoc examples invisible on pkg.go.dev until tag) | 🔵 `BLOCKED` | High   | owner  | package prepared: `docs/planning/2026-09-15_owner-decision-package.md` Q2 |

## Closed outside this repo (2026-09-15, tracked upstream)

- QMD MCP `get`/`multi_get` garbage output — root-caused (embedded-resource
  content items), filed as tobi/qmd#959 and charmbracelet/crush#3846.
- CV-repo extraction-story annotations — committed in the CV repo
  (`16fc16c7`, daemon-committed; awaiting the CV repo's next push).
