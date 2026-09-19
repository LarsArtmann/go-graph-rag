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

| Task                                                     | Status       | Impact | Effort | Evidence                                                                                                        |
| -------------------------------------------------------- | ------------ | ------ | ------ | --------------------------------------------------------------------------------------------------------------- |
| Decide license posture (public + PROPRIETARY + PR flow)  | 🔵 `BLOCKED` | High   | owner  | package: `docs/planning/2026-09-15_owner-decision-package.md` Q1; pkg.go.dev hides godoc for EVERY version ("Documentation not displayed due to license restrictions", re-verified live 2026-09-19 on v0.2.0) until a recognized LICENSE exists — Q1 now gates the entire public-SDK face |

### Medium Impact

| Task                                                                     | Status       | Impact | Effort | Evidence                                                                                                                                                                                                |
| ------------------------------------------------------------------------ | ------------ | ------ | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Run live smoke test against a real endpoint                              | 🔵 `BLOCKED` | Med    | S      | `embed_openai_live_test.go` asserts never executed; needs `GRAPHRAG_LIVE_EMBED_*` creds                                                                                                                 |
| Watch tobi/qmd#959 + charmbracelet/crush#3846; patches live in the issue texts | 🔴 `TODO`    | Med    | M      | filed 2026-09-15; re-checked 2026-09-19: both OPEN, 0 comments; owner ruling 2026-09-16: HOLD past the 2026-09-22 window, no reminder pings — send-or-drop only if the owner asks                                        |
| Apply CONTRIBUTING inbound-grant line                                    | 🔵 `BLOCKED` | Med    | S      | drafted in the owner package; presumes license answer (Q1); current wording verified unchanged 2026-09-19                                                                                                               |
| Extend `.githooks/pre-push` with tidy-drift + dprint checks              | 🔴 `TODO`    | Med    | S      | the hook guards only the pristine build (`.githooks/pre-push`); the json-v2 regression recurred three times via daemon commits (pre-v0.1.0, `e67bd9b`, `e4a9145`) — each caught pristine, but two of three gates are unhooked |
| Guard the seam ADR against core drift                                    | 🔴 `TODO`    | Med    | S      | compile-only test for the ADR §5 interface sketch + a fake `GraphStore` fixture, so the sketch cannot rot against core types (ADR follow-up; `docs/status/2026-09-16_14-14_seam-plan-execution-status.md` f26/f48)       |

### Low Impact

| Task                                                                     | Status       | Impact | Effort | Evidence                                                                                                                                                                     |
| ------------------------------------------------------------------------ | ------------ | ------ | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Benchmark hygiene batch                                                  | 🔴 `TODO`    | Low    | M      | benchstat-grade `BenchmarkStoreRoundTrip` (`-count 10`; currently count=3), document the ~37s 10k fixture-build cost in `bench_test.go`, add `BenchmarkSimilarPairs/docs=100` for size symmetry |
| Document `check-rows.py` in the docs-health `SKILL.md` body              | 🔴 `TODO`    | Low    | S      | tool exists + regression-tested (SKILLS repo, `0d1aca6`); the skill body doesn't mention it (carried since 2026-09-16)                                                        |
| Resolve the lychee-reported redirecting URL to its final target          | 🔴 `TODO`    | Low    | S      | one URL from the 2026-09-16 lychee pass (`docs/status/2026-09-16_14-14_seam-plan-execution-status.md` f17)                                                                    |

## Closed outside this repo (tracked upstream)

- QMD MCP `get`/`multi_get` garbage output — root-caused (embedded-resource
  content items), filed as tobi/qmd#959 and charmbracelet/crush#3846 (the
  Medium watch row owns the follow-up).
- CV-repo extraction-story annotations — committed in the CV repo
  (`16fc16c7`, daemon-committed; awaiting the CV repo's next push).
- CV consumer bump to graph-rag v0.2.0 — pending in the owner package
  Phase 8 (external repo, separate session).
