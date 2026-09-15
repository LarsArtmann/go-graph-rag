# Status: Pareto Plan Full Execution — M1–M15 complete, owner decisions pending

- **Date:** 2026-09-15 19:33 CEST
- **Session scope:** Executed the remaining backlog of
  `docs/planning/2026-09-15_18-30_SUPERB-pareto-execution-plan.md` (M2–M15, 45
  fine tasks) to completion in one run, plus final push/CI verification.
  This report covers that run and what was noticed during it.
- **Repo state at report time:** `master` = `origin/master` = `07b558d`, tree
  clean, CI green on HEAD (incl. the new dprint step), branch protection
  active. Two explicit commits this session (`289669b`, `07b558d`); the rest
  rode daemon heuristic commits (`ca389d1`…`8f24847`).

---

## a) FULLY DONE

All verifiable with the cited evidence:

1. **Branch protection + dependabot closure (M2)** — `master` requires the
   `build-test-lint` check, strict, no force-pushes/deletions, admins exempt
   (direct pushes keep working). Verified via GET echo. Dependabot PR #1
   (checkout v7.0.1, setup-go v7.0.0) squash-merged as `41beb47`; CI green on
   the merge commit; local fast-forwarded. PR queue now empty.
2. **Runnable godoc examples (M3)** — `example_test.go` with
   `ExampleBuild`, `ExampleSearcher_Search`, `ExampleSearcher_SimilarPairs`,
   `ExampleNewProvider`; `// Output:` blocks enforced by `go test`. README
   quick-start rewritten to mirror the examples and **compile-verified AND
   executed** in a throwaway module with a `replace` directive (output
   captured: context block renders correctly). The "snippet never compiled"
   self-critique from the previous session is closed.
3. **SECURITY.md (M4)** — reporting path, supported versions, scope notes
   (endpoint trust, key handling, unencrypted SQLite), dependency policy.
   README links it; all repo-relative links grep-verified. GitHub **private
   vulnerability reporting enabled** (API-verified `enabled: true`).
4. **Repo metadata (M5)** — 8 topics (`graphrag, rag, embeddings,
   vector-search, go, sqlite, knowledge-graph, semantic-search`) + homepage
   (pkg.go.dev module URL); both echoed back by the API.
5. **UA version stamp (M6)** — `UserAgentVersion` const wires the
   OpenAI-compat User-Agent (`embed_openai.go`); httptest asserts the exact
   header value (`embed_openai_test.go`). Lint clean.
6. **KindUnknown decision (M7)** — **kept exported** (unexporting breaks
   v0.1.0 consumers; the zero value deserves a name). Doc comment in
   `graph.go` now states the full contract: appears only on `SimilarPair`
   placeholders, treat as "data hole", never a matchable category.
7. **Benchmark suite (M8)** — `bench_test.go`: Build/Search at 100/1000
   docs + SimilarPairs. Reference numbers recorded in the file doc comment
   (Build/1000 ≈ 373ms ⇒ quadratic scaling measured, ~15min extrapolated at
   50k nodes — hard data behind the ROADMAP revisit trigger).
8. **dprint in CI (M9)** — SHA-pinned `dprint/check@7dc032d` (# v2.5, tag
   commit verified) as a CI step; local format pass applied (5 markdown
   files realigned); `dprint check` clean. Local runner documented:
   `nix run nixpkgs#dprint -- fmt`.
9. **Race run (M10)** — `go test ./... -race` clean (4.0s); CONTRIBUTING
   already documents the command; recorded in CHANGELOG.
10. **Pre-push pristine-build hook (M11)** — `.githooks/pre-push` committed,
    `core.hooksPath` wired locally, documented in CONTRIBUTING. **Both paths
    proven**: clean tree passes ("pristine build ok" on dry-run push); a
    planted `encoding/json/v2` canary was blocked with the exact diagnostic.
    The hook also fired on both real pushes this session.
11. **CV-repo annotations (M12)** — status-report items c7/c8 struck through
    with fixed-at hashes (`40fa417` fix, `289669b` follow-ups); CV TODO_LIST
    graphrag replace-part annotated resolved. Committed in the CV repo
    (`16fc16c7`) — see (d)4 for the race story.
12. **Owner-decision package (M13)** —
    `docs/planning/2026-09-15_owner-decision-package.md`: license options
    (keep-proprietary-to-v1 recommended) + full v0.2.0 go-release checklist
    (version rationale: `UserAgentVersion` is a new exported symbol ⇒ minor;
    deciding fact: pkg.go.dev only renders tagged versions).
13. **QMD MCP bug (M14)** — root-caused on BOTH sides at source level:
    qmd's `get` returns `type: "resource"` embedded-resource content
    (`src/mcp/server.ts:462`); crush stringifies unknown content types with
    `fmt.Sprintf("%v", v)` (`internal/agent/tools/mcp/tools.go:78` on main,
    still unhandled). Reproduced twice by docid (`&{0x1399… map[] <nil>}`).
    Filed **tobi/qmd#959** and **charmbracelet/crush#3846**, cross-linked,
    both through the voice checker (0 FAIL). No existing issue in either
    repo (searched open+closed).
14. **Live-endpoint smoke test (M15)** — `embed_openai_live_test.go`,
    env-gated (`GRAPHRAG_LIVE_EMBED_URL/_KEY/_MODEL`), skips cleanly offline
    (verified), how-to-run documented in the test comment.
15. **Living-docs truth sync** — TODO_LIST reduced to the 2 owner-blocked
    rows + upstream-tracked closures; CHANGELOG `[Unreleased]` expanded
    (12 entries); FEATURES.md gained a Developer-experience section and the
    examples/benchmarks PLANNED rows removed; AGENTS.md updated (new files,
    dprint command, branch protection, UA bump rule, hook install).

## b) PARTIALLY DONE

1. **CV-repo push** — annotations are committed locally (`16fc16c7`) but CV
   `master` sits ≥1 commit ahead of `origin` (daemon commit incl. my
   annotations). Blocker: no push authorization for that repo. Effort: S
   (one command, owner's call).
2. **Upstream QMD/crush fixes** — issues filed with suggested patches
   (offered a PR to qmd); not fixed upstream yet, so `mcp_qmd_get` STILL
   returns garbage in this very session's tooling. Workaround: read files
   from disk (used throughout). Effort: M each, blocked on maintainers.
3. **Live smoke test execution** — written and offline-skip-verified, but
   its real assertions (dims, norm, wire shape) have NEVER run against a
   live endpoint: no `GRAPHRAG_LIVE_EMBED_*` in the environment. Effort: S
   once creds exist. (Deviation noted: plan said hard unit-norm assert;
   loosened to 0.1 delta because legitimately unnormalized models exist.)
4. **Benchmark rigor** — numbers recorded from a single `-benchtime=100ms`
   run on one machine. Directionally solid (quadratic scaling is
   unambiguous) but not a benchstat-grade baseline. Effort: S/M.

## c) NOT STARTED

1. **v0.2.0 release** — owner-gated; checklist fully prepared (M13).
2. **License posture decision** — owner-gated; recommendation written.
3. **CONTRIBUTING inbound-grant softening** — one line drafted in the
   package, not applied (presumes the license answer).
4. **ANN/HNSW backend, incremental indexing** — ROADMAP-scale, deliberately
   not started (design decisions, not chores).
5. **CV-side consumer bump to v0.2.0** — depends on (1).
6. **Session status report at natural close** — I finished the last turn
   with a summary table but wrote no `docs/status` artifact; only produced
   this one on request. The workspace convention is one per session; I
   skipped it unprompted. (Self-caught; remedied now.)

## d) TOTALLY FUCKED UP (honesty ledger)

1. **Nearly recorded fake benchmark numbers.** The first corpus was
   degenerate: 1000 docs shared ~15 distinct texts, and `resolveVectors`
   dedups by content hash — so only ~15 vectors existed and Build/1000
   "measured" 1.17ms (1000× too fast). I had the wrong `// Output:`-style
   numbers in hand and was one edit away from committing them as reference
   truth. Caught ONLY because the magnitude failed a sanity check
   (500k×1024-dim cosines cannot fit in 1.17ms). Root cause: designed the
   corpus without accounting for hash dedup. Fix: unique per-doc token
   (`note%d`) + in-file comment explaining why. Lesson: **sanity-check
   benchmark magnitudes against a back-of-envelope estimate before
   recording.**
2. **Guessed example outputs first.** First draft of `example_test.go`
   shipped invented cosines (0.85/0.87/0.83); every one was wrong (real:
   0.73/0.54). One wasted test cycle. Correct workflow is run-then-paste;
   I pasted-then-ran.
3. **Avoidable lint/test churn on the live-test file** — trailing-comma
   syntax sloppiness + two `testifylint` findings (`Greater`→`Positive`)
   each cost a gate cycle; both would have been caught by writing more
   carefully against the repo's established testify style.
4. **Daemon won the CV commit race (again).** My detailed F12.3 commit was
   mid-pre-commit-hook when the hook failed on SIBLING test-count drift
   (`internal/httpx/forwarded_test.go` etc. — parallel session's tests, no
   baseline update). While my commit failed, the daemon committed my
   annotated files mixed with a sibling's `nix/packages.nix` (`16fc16c7`).
   Net: content landed (verified by grep), the planned detailed message
   did not. I did NOT update the sibling's baseline (correctly — not mine
   to judge) and did NOT `--no-verify` after the race (moot). Lesson
   confirmed: **explicit commits must be add+commit in ONE invocation, and
   the narrative should live in a report file, not depend on the commit
   message.**
5. **Pipeline-masked exit code in the hook test.** My F11.2 blocking-path
   verification echoed `exit=0` — the exit code of `head`, not the failed
   push (no `pipefail`). The block was actually proven by git's "failed to
   push some refs" text. Same masking class AGENTS.md already warns about;
   I re-hit it in a verification context where the exit code was the point.
6. **Two edit-tool round trips lost to races** — `embed_openai_live_test.go`
   (my own `gofmt -w` between write and edit) and `TODO_LIST.md` (daemon
   commit between write and rewrite). Both re-read and recovered; both were
   preventable by re-reading immediately before editing after any
   formatting/daemon-triggering step.
7. **Plan-order deviation, unlogged at the time.** I ran M13/M15 before M12
   (annotations must cite pushed hashes, so they must follow the final
   push). Correct call — but a plan deviation made silently is how
   VERSCHLIMMBESSER guard drift starts; it belongs in the running log the
   moment it's decided.

## e) WHAT WE SHOULD IMPROVE

1. **Benchmark protocol**: adopt benchstat + repeated runs + recorded env
   before any number drives a design decision (the 50k trigger currently
   rests on one short run).
2. **Example workflow**: always run-then-paste outputs; never predict
   floats.
3. **Race-resistant commits**: single-invocation add+commit, verify with
   `git log` immediately, treat the daemon as the likely winner and put the
   story in docs/status files.
4. **`set -o pipefail` reflex** whenever an echoed exit code is the
   evidence (third time this class appears in session lessons).
5. **Surprising API behavior discovered, undocumented**: `Build` assigns a
   vector only to the FIRST document of each identical text; same-text
   siblings become vector-less nodes (expansion-only). Found during
   benchmark design; nothing user-facing documents it. Should be a doc
   comment on `Build` at minimum (behavior change would be API-visible ⇒
   owner decision).
6. **README snippet shows a 0.00-similarity hit** (Rust article for query
   "go channels") because `MinScore` defaults to 0 — honest but confusing
   to readers; setting `MinScore` in the snippet would read better.
7. **Stale LSP diagnostics vs CLI**: golangci_lint_ls kept showing warnings
   after CLI lint passed clean (trust-the-CLI lesson re-confirmed); a
   `lsp_restart` after daemon commits would reduce noise.
8. **Homepage points at docs that don't show the new work yet** — pkg.go.dev
   renders the latest TAG (v0.1.0) until v0.2.0 is cut. Acceptable, but
   worth remembering when judging "the public face is done".

## f) Next tasks (ranked, feeds HARVEST)

| #  | Task                                                                                            | Impact   | Effort | Category      |
| -- | ----------------------------------------------------------------------------------------------- | -------- | ------ | ------------- |
| 1  | Owner: decide license posture (package Q1)                                                      | Critical | S      | Decision      |
| 2  | Owner: approve v0.2.0 cut (package Q2)                                                          | Critical | S      | Decision      |
| 3  | Execute the v0.2.0 go-release checklist on approval (tag, proxy, pkg.go.dev, CV bump)           | Critical | M      | Release       |
| 4  | Document the identical-text → single-vector rule on `Build` (doc comment)                       | High     | S      | Documentation |
| 5  | Add tag-protection rule before v0.2.0 (`gh api …/tags/protection`)                              | High     | S      | Security      |
| 6  | Apply CONTRIBUTING inbound-grant line (after Q1)                                                | Medium   | S      | Documentation |
| 7  | Push the CV repo (owner; carries the extraction-story annotations)                              | Medium   | S      | External      |
| 8  | Re-run benchmarks with benchstat protocol; update reference numbers                             | Medium   | M      | Quality       |
| 9  | Set `MinScore` in README snippet / examples to avoid 0.00 hits                                  | Low      | S      | Documentation |
| 10 | Run the live smoke test against a real endpoint (needs env creds)                               | Medium   | S      | Testing       |
| 11 | Watch tobi/qmd#959 + crush#3846; send offered PR if silent ~1 week                              | Medium   | M      | External      |
| 12 | Add `ExampleNewStore` — store.go is the only major file without an example                      | Medium   | S      | Documentation |
| 13 | Add a `SearcherOptions` (ReferenceKind/DocumentKinds) example variant                           | Medium   | S      | Documentation |
| 14 | Store round-trip benchmark (persist + snapshot load at 1k/10k nodes)                            | Medium   | M      | Quality       |
| 15 | Cross-link the measured quadratic cost from ROADMAP's 50k note                                  | Low      | S      | Documentation |
| 16 | ANN/HNSW spike (ROADMAP; benchmark data now exists to justify)                                  | High     | L      | Feature       |
| 17 | Incremental indexing design sketch (ROADMAP)                                                    | High     | L      | Feature       |
| 18 | CV-side open extraction tails: ADR, DOMAIN_LANGUAGE terms, AGENTS sweep (external, CV c1/c4/c5) | Medium   | M      | External      |
| 19 | Run `qmd embed` (173 CV docs need embedding — noticed in the MCP banner)                        | Low      | S      | Tooling       |
| 20 | `lsp_restart golangci_lint_ls` after daemon commits (habit)                                     | Low      | S      | Tooling       |
| 21 | GitHub Release automation (workflow) as v0.2.0 follow-up                                        | Low      | M      | Release       |

## g) Questions I cannot answer myself

1. **License posture:** keep public+PROPRIETARY until v1 (my recommendation),
   or open-source (Apache-2.0/MIT) now? I cannot know your intent for the
   public face, and it gates the CONTRIBUTING softening and PR-flow story.
2. **Release cadence:** cut **v0.2.0** now (godoc examples + UA const +
   public face are invisible on pkg.go.dev until a tag exists; SemVer says
   minor for the new exported const), or ride `[Unreleased]`? Only you can
   weigh timing.
3. **Live-endpoint credentials:** is there an OpenAI-compatible endpoint +
   API key I may use (via `GRAPHRAG_LIVE_EMBED_*` env vars) to actually run
   the M15 smoke test — and is spending its tokens acceptable? I cannot
   conjure credentials or authorize cost.

---

_Report written per status-report skill; `.md` format per explicit user
instruction (skill default is HTML). HARVEST note: items 4, 5, 8–15, 19–21
are new actionable rows not yet in TODO_LIST — they were routed there after
this report was written._
