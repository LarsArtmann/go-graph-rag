# Owner Decision Package — license posture & v0.1.x release cadence

- **Date:** 2026-09-15
- **Prepared by:** session executing `docs/planning/2026-09-15_18-30_SUPERB-pareto-execution-plan.md` (M13)
- **Status:** ~~AWAITING OWNER DECISION on both questions. Nothing below is
  executed;~~ Q2 was executed 2026-09-16 (see the inline notes in the
  checklist below); Q1 (license posture) is still awaiting the owner.

---

## Question 1: License posture — public repo + PROPRIETARY + PR flow. Intended?

Current state: the repository is **public**, the LICENSE is **PROPRIETARY**
("All rights reserved"), README advertises the pkg.go.dev reference, and
CONTRIBUTING says "pull requests are welcome" while also saying reuse beyond
PR submission needs explicit permission.

### Options

| # | Option                                      | Pros                                                                            | Cons                                                                                                              |
| - | ------------------------------------------- | ------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| 1 | **Keep as is** (public + proprietary)       | Zero action; code visible as portfolio/reference; no legal reuse by others      | pkg.go.dev hosts docs nobody may legally use; "PRs welcome" reads contradictory; adoption ~0 by design            |
| 2 | **Open source permissive** (MIT/Apache-2.0) | Real adoption possible; standard PR flow; ecosystem goodwill                    | Anyone (including competing CV-matching products) may reuse the retrieval engine; license decision is one-way-ish |
| 3 | **Source-available middle** (BUSL/PolyForm) | Grants usage on your terms (e.g. non-commercial)                                | Legally murky for random consumers; Go ecosystem frowns on non-OSI licenses; complexity for zero audience         |
| 4 | **Private repo**                            | Strongest protection; GOPRIVATE+nix pattern already proven in LarsArtmann repos | Kills discoverability, pkg.go.dev, portfolio value; CI/proxy machinery becomes private-path                       |

### Recommendation (one pick)

**Option 1 — keep public + PROPRIETARY until v1, revisit open-sourcing at
v1.** Rationale:

- The primary consumer is your own CV repo (via the module proxy); external
  adoption is not the current goal, and the v0.x API is explicitly unstable.
- Open-sourcing later is easy; un-open-sourcing after others build on it is
  not. Option 2's cost (competitor reuse) is real today, its benefit
  (community) is speculative at v0.x.
- The contradiction in CONTRIBUTING is cheap to soften without a license
  change (see below).

If you keep option 1, apply this one-line softening (ready to commit on your
word): change CONTRIBUTING's "Pull requests are welcome" to "Issues and PRs
are welcome; by submitting you grant the right to use your contribution in
this project" — the standard inbound-grant clause that makes the PR flow
legally coherent with the proprietary license.

## Question 2: Cut v0.1.1 now, or ride [Unreleased]?

### What is unreleased (since tag `v0.1.0`)

- Fixed: the json/v2 regression restore (master-only; **v0.1.0 was never
  affected** — consumers on the proxy are fine).
- Added: runnable godoc examples (`example_test.go`), `UserAgentVersion`
  const, benchmark suite, SECURITY.md, full living-doc set, dprint CI step,
  branch protection, repo topics/homepage, dependabot bump.

### The deciding fact

**pkg.go.dev renders documentation for the latest tagged version.** The
entire public-SDK-face work (examples, UA const docs, tightened godoc) is
invisible on pkg.go.dev until a new tag exists. Tagging is the only way the
work becomes the public face.

### Recommendation (one pick)

**Cut `v0.2.0`, not `v0.1.1`.** SemVer: `UserAgentVersion` is a new exported
symbol → minor bump in 0.x. v0.1.1 would mis-signals "bugfix only" while the
real story is "public face complete".

### v0.2.0 go-release checklist (execute on your word)

```text
Executed 2026-09-16 (release date); notes below recorded inline.

[x] Phase 0: confirm master green (CI) and tree clean
[x] Phase 1: version = v0.2.0 (minor: new exported const + docs face)
[x] Phase 2: CHANGELOG — moved [Unreleased] into [0.2.0] - 2026-09-16, empty placeholders left
[x] Phase 3: go.mod clean — no replace directives (verified 2026-09-15), go mod tidy + verify
[x] Phase 4: env -u GOEXPERIMENT GOTOOLCHAIN=go1.26.7 go build ./... && go vet ./... && go test ./... -race -count=1
[x] Phase 4: golangci-lint run ./... (v2.13.2) — 0 issues
[x] Phase 5: git tag -a v0.2.0 -m "..." at the verified commit; git show v0.2.0 --stat sanity check
[x] Phase 5: push master + tag together
[x] Phase 6: go list -m -versions github.com/larsartmann/go-graph-rag shows v0.2.0
[x] Phase 6: go get in a clean /tmp module resolves + builds (trash the dir after)
[ ] Phase 6: pkg.go.dev shows the examples under Build/Search/SimilarPairs/NewProvider
    (2026-09-16: proxy indexed v0.2.0; pkg.go.dev doc processing still pending hours
    later. Deeper finding: v0.1.0's page says "Documentation not displayed due to
    license restrictions" — no LICENSE file pkg.go.dev recognizes, so godoc is
    hidden for every version. The public-face payoff of releases stays blocked on
    the Q1 license decision; see TODO_LIST High row.)
[x] Phase 7: GitHub Release created for v0.2.0 — NOT via workflow: the tag-push
    release.yml failed on the awk -v escape bug (fixed on master, 350b815);
    release created manually with the same extraction logic and notes
[ ] Phase 8: CV repo: go get github.com/larsartmann/go-graph-rag@v0.2.0 (go-ecosystem-upgrade flow)
    (pending — consumer-side bump, separate session)
```

Tags are immutable once the proxy caches them: if anything is wrong after
tagging, the fix is v0.2.1 — never re-tag.

---

## Owner answers (fill in)

1. License posture: ☐ keep proprietary to v1 (recommended) ☐ Apache-2.0/MIT now ☐ other: ____
2. ~~Release: ☐ cut v0.2.0 now (recommended) ☐ ride [Unreleased] until ____ ☐ v0.1.1 anyway~~ → EXECUTED 2026-09-16: v0.2.0 cut (tag `805aeba`); the tag-push release workflow failed on the awk extraction bug (fixed `350b815`) and the release was created manually; pkg.go.dev indexes v0.2.0 but renders no godoc (license restriction — blocked on Q1).
