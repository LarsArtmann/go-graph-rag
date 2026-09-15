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

| Task                                                            | Status       | Impact | Effort | Evidence                                                                                                                                     |
| --------------------------------------------------------------- | ------------ | ------ | ------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| Decide license posture (public + PROPRIETARY + PR flow)         | 🔵 `BLOCKED` | High   | owner  | package prepared: `docs/planning/2026-09-15_owner-decision-package.md` Q1                                                                    |
| Cut `v0.2.0` (godoc examples invisible on pkg.go.dev until tag) | 🔵 `BLOCKED` | High   | owner  | package prepared: `docs/planning/2026-09-15_owner-decision-package.md` Q2                                                                    |
| Document identical-text → single-vector rule on `Build`         | 🔴 `TODO`    | High   | S      | `resolveVectors` dedups by content hash; first same-text doc gets the vector, siblings get none — undocumented (status 2026-09-15 19:33 e.5) |
| Add tag-protection rule before v0.2.0                           | 🔴 `TODO`    | High   | S      | branch protection exists; tags unprotected (default state)                                                                                   |
| Design the pluggable store/search seam (scenario B)             | 🔴 `TODO`    | High   | M      | owner decision 2026-09-15: DESIGN NOW (dgraph status §g ANSWERED); `docs/research/2026-09-15_dgraph-adoption.md` §6/§8; encode three backend classes (embedded ANN libs / metaengine engines / server DBs) so adapters never force a breaking SDK change and `dgo` never enters core `go.mod` |

### Medium Impact

| Task                                                                     | Status       | Impact | Effort | Evidence                                                                                |
| ------------------------------------------------------------------------ | ------------ | ------ | ------ | --------------------------------------------------------------------------------------- |
| Re-run benchmarks with benchstat protocol; update reference numbers      | 🔴 `TODO`    | Med    | M      | current numbers are a single `-benchtime=100ms` run (`bench_test.go` doc)               |
| Run live smoke test against a real endpoint                              | 🔵 `BLOCKED` | Med    | S      | `embed_openai_live_test.go` asserts never executed; needs `GRAPHRAG_LIVE_EMBED_*` creds |
| Watch tobi/qmd#959 + charmbracelet/crush#3846; send offered PR if silent | 🔴 `TODO`    | Med    | M      | both filed 2026-09-15 with suggested patches                                            |
| `ExampleNewStore` godoc example                                          | 🔴 `TODO`    | Med    | S      | store.go is the only major file without one                                             |
| `SearcherOptions` (ReferenceKind/DocumentKinds) example variant          | 🔴 `TODO`    | Med    | S      | examples only cover the neutral policy                                                  |
| Store round-trip benchmark (persist + snapshot load, 1k/10k nodes)       | 🔴 `TODO`    | Med    | M      | only Build/Search/SimilarPairs are benchmarked                                          |
| Quantify metaengine dep-tree delta (`go mod graph` before/after)         | 🔴 `TODO`    | Med    | M      | hardens CONTRA #1 of `docs/research/2026-09-15_metaengine-system-adoption.md` (report f.2) into numbers |
| Apply CONTRIBUTING inbound-grant line                                    | 🔵 `BLOCKED` | Med    | S      | drafted in the owner package; presumes license answer (Q1)                              |

### Low Impact

| Task                                                           | Status    | Impact | Effort | Evidence                                                        |
| -------------------------------------------------------------- | --------- | ------ | ------ | --------------------------------------------------------------- |
| GitHub Release automation as v0.2.0 follow-up                  | 🔴 `TODO` | Low    | M      | release is manual `gh release` today                            |
| Run `qmd embed` (173 CV docs unembedded)                       | 🔴 `TODO` | Low    | S      | external: MCP banner; CV-side tooling                           |

## Closed outside this repo (2026-09-15, tracked upstream)

- QMD MCP `get`/`multi_get` garbage output — root-caused (embedded-resource
  content items), filed as tobi/qmd#959 and charmbracelet/crush#3846.
- CV-repo extraction-story annotations — committed in the CV repo
  (`16fc16c7`, daemon-committed; awaiting the CV repo's next push).
