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

## High Impact

| Task                                                             | Status         | Impact | Effort | Evidence                                                                                          |
| ---------------------------------------------------------------- | -------------- | ------ | ------ | ------------------------------------------------------------------------------------------------- |
| Turn CI green again and push the fix (json-v1 restore is local)   | 🔴 `TODO`      | High   | 5min   | local gates green pristine; master CI still red at run `34971957570` until pushed                 |
| Branch protection on `master` (require `go-test`)                 | 🔴 `TODO`      | High   | 10min  | `gh api .../branches/master/protection` → 404 "Branch not protected"                              |
| Godoc examples for `Build` / `Search` / `SimilarPairs`            | 🔴 `TODO`      | High   | 1h     | zero `func Example*` in tree (grep-verified); also gives the neutral `NewSearcher` direct coverage |

## Medium Impact

| Task                                                        | Status    | Impact | Effort | Evidence                                                        |
| ----------------------------------------------------------- | --------- | ------ | ------ | --------------------------------------------------------------- |
| `SECURITY.md` (go-cqrs-lite parity)                          | 🔴 `TODO` | Med    | 20min  | absent from repo root (ls-verified)                             |
| Repo topics + homepage                                       | 🔴 `TODO` | Med    | 5min   | `gh api` shows `topics: []`, `homepage: null`                   |
| Version-stamp the OpenAI user agent                          | 🔴 `TODO` | Med    | 15min  | hardcoded `"go-graph-rag/0.1"` at `embed_openai.go:239`         |
| Decide `KindUnknown`'s public fate                           | 🔴 `TODO` | Med    | 30min  | exported zero-value placeholder, `graph.go:16`; keep or unexport + accessor |
| Benchmark suite (Build + Search on a synthetic corpus)        | 🔴 `TODO` | Med    | 2h     | zero `func Benchmark*` in tree (grep-verified)                   |
| Wire `dprint.json` into CI or drop it                         | 🔴 `TODO` | Med    | 15min  | `dprint.json` exists but nothing installs or runs it (not in `go-test.yml`) |

## Low Impact

| Task                                                          | Status         | Impact | Effort | Evidence                                                       |
| ------------------------------------------------------------- | -------------- | ------ | ------ | -------------------------------------------------------------- |
| Decide repo posture: public + PROPRIETARY license vs CONTRIBUTING "fork" flow | 🔵 `BLOCKED` | Low | owner decision | `README.md:66` PROPRIETARY; `CONTRIBUTING.md:7` invites forks |
