# Status: BDD sub-module implementation (2026-10-15 04:21)

> RESOLVED 2026-10-15 ~05:00 (follow-up session): the "package-classification
> anomaly" was no toolchain bug — the spec files were named `*_spec.go`, and
> only the `_test.go` FILENAME suffix makes a file a test file (the
> `package bdd_test` clause does not). After `git mv` to `*_spec_test.go`, one
> real bug surfaced (a spec building with FailingProvider) and was fixed; the
> suite is 43/43 green, order-independent, `-race` clean, lint 0 issues. Kit
> delivered: CI + pre-push gates extended and proven by planted failures, ADR
> at `docs/planning/2026-10-15_05-02_bdd-submodule-adr.md`, docs updated.
> Open items moved to TODO_LIST (tag owner-gated, securitymd, core erraudit).

Task: provide an independent/composable `go.mod` sub-module for BDD-style
testing in go-graph-rag, informed by studying how CV, knowledge-graph, and
legal-graph-ai-system consume the SDK.

Session state: **blocked mid-implementation** on a Go toolchain
package-classification anomaly (details below). All research and design work
is done; kit source compiles pristine; the spec suite is written but cannot
run yet. No core files were touched.

---

## a) FULLY DONE

1. **Skills loaded** (go-modularize, bdd-testing, buildflow) before action.
2. **Consumer research (3 repos, via sub-agents)** — all findings concrete with file:line:
   - All three pin `go-graph-rag v0.3.2` from the proxy (no replaces).
   - All three test against REAL SDK machinery (hash provider, real Build /
     Searcher, real SQLite) — no mocks of graphrag interfaces anywhere.
   - All three hand-roll the same fixtures: `edge()` builders, `doc()`
     closures, temp-store setup, top-hit / similar-edge / hub-not-ranking
     assertions.
   - CV: ginkgo v2.33.0 + gomega v1.44.0 already in root go.mod, plus
     `platform/bddkit` bootstrap helper; graphrag-adjacent tests stayed
     testify by an old sibling-consistency decision.
   - knowledge-graph: stdlib-only tests (testify banned repo-wide); a BDD
     review recommendation was closed "Won't implement" for that repo.
   - legal-graph-ai-system: testify today, explicit "testify → ginkgo/gomega"
     migration item in its TODO_LIST.
3. **Repo state studied**: `.golangci.yaml` (v2 format, ginkgolinter enabled,
   `_test.go` excluded from most structural linters), CI go-test.yml gates,
   pre-push 3-gate hook, release.yml (`v*` tag trigger — `bdd/v*` will NOT
   match), `.gitignore` ignores `go.work`/`go.work.sum` inside the
   buildflow-managed block (→ no committed workspace; replace-directive
   strategy chosen instead), no flake.nix, no `.buildflow.yml`.
4. **Design fixed and self-reviewed** against go-modularize litmus +
   AGENTS non-negotiables:
   - Module `github.com/larsartmann/go-graph-rag/bdd`, package `bdd`, own
     go.mod owning ginkgo/gomega (core go.mod untouched — non-negotiable #3
     preserved via the seam-ADR "separate module owns its deps" pattern).
   - `require go-graph-rag v0.3.2` (published version → proxy-resolvable for
     consumers) + `replace github.com/larsartmann/go-graph-rag => ../`
     (inert for consumers, pins local/CI to sibling checkout).
   - Kit API: `Doc`/`Edge`/`WeightedEdge` builders, `DemoCorpus` +
     demo vocabulary (article/topic/cites/mentions — domain-neutral,
     non-negotiable #1 respected), `Harness` (NewHarness + WithProvider /
     WithBuildOptions / WithSearcherOptions / WithStore; Build / Search /
     SimilarPairs / Reload with hash-join of reloaded embeddings),
     `FailingProvider`, Gomega matchers (`HaveTopHit`, `ContainHit`,
     `HaveHitIDs`, `HaveNoHitsOfKind`, `HaveSimilarEdge`) + `SimilarEdgeScore`.
   - Dogfood: the module's own Ginkgo suite is a black-box consumer-perspective
     spec suite over the SDK (build/search/store/provider/matcher specs).
5. **Kit source implemented and compiles**: `bdd/go.mod`, `doc.go`,
   `builders.go`, `corpus.go`, `providers.go`, `harness.go`, `matchers.go`.
   `go mod tidy` + `go build ./...` + `go vet ./...` all pass under
   `env -u GOEXPERIMENT GOTOOLCHAIN=go1.27.1` (pristine).
6. **Spec suite written** (not yet runnable): `bdd_suite_test.go`,
   `build_spec.go`, `search_spec.go`, `store_spec.go`, `provider_spec.go`,
   `matchers_spec.go` — ~30 specs in `package bdd_test` (black-box).

## b) PARTIALLY DONE

- **Spec suite**: fully authored but blocked from running (see d). None of
  its behavioral assertions have been calibrated against the real hash
  provider yet (top-hit queries, hub exclusion at TopK 3, MinScore 0.99
  exact-match, reload equality, cache-hit counting).
- **ginkgo pin**: tidy selected v2.33.1; CV pins v2.33.0. Divergence noted,
  decision pending.

## c) NOT STARTED

- CI workflow changes (bdd build/vet/test/lint/tidy steps in go-test.yml).
- Pre-push hook extension (pristine build + tidy drift inside bdd/).
- golangci-lint run over the bdd module (incl. parent-config discovery check)
  and resulting fix loop; gofumpt/golines/gci formatting pass.
- Docs: planning ADR for the sub-module decision, README section, CHANGELOG
  Unreleased entry, FEATURES.md, TODO_LIST.md row, AGENTS.md updates.
- dprint fmt/check of new markdown; buildflow run; `-race` runs; planted-failure
  dry-run push proof of the extended gates.

## d) TOTALLY FUCKED UP (the blocker)

`go test ./...` in `bdd/` fails at setup:

```
found packages bdd (bdd_suite_test.go) and bdd_test (build_spec.go) in .../bdd
```

The tool classifies the spec files (which declare `package bdd_test`, byte-verified
via `cat -A`) into the regular package (`GoFiles`), and only `bdd_suite_test.go`
into `XTestGoFiles`. `gofmt -e` parses every file correctly. Facts isolated so far:

- Reproduces in `/tmp` copies (`/tmp/bddprobe`, `/tmp/grcopy`) — not caching,
  not git filters, not the real directory specifically.
- Minimal repro (`module example.com/x/bdd` + a.go + two plain `_test.go`
  files with `package bdd_test`) classifies CORRECTLY.
- Content bisect so far: a spec file reduced to imports + `var _ = Describe(...)`
  is still misclassified when importing (context+ginkgo) or (ginkgo+gomega);
  `bdd_suite_test.go` — same dot-imports but WITH a `func TestXxx(t *testing.T)`
  — classifies correctly. Leading hypothesis: misclassification correlates
  with `_test.go` files that have ginkgo containers but no Test function, or
  with import resolution under `go list -e`; NOT yet confirmed.
- Session mistakes that cost cycles (self-critique):
  1. Wrote two files with placeholder/broken content mid-write (providers.go,
     matchers_spec.go) — caught on run, rewrote. Write-then-think waste.
  2. Guessed the gomega custom-matcher API twice (wrong import path
     `types/gcustom`, wrong `WithFailMessage` method) instead of reading the
     module cache first — two failed build cycles, then switched to
     hand-rolled `types.GomegaMatcher` structs (correct, verified).
  3. Missed that `graphrag.Build` returns `*BuildResult` (pointer) — the
     grep of signatures I did showed the line but not the return type.
  4. Sloppy first draft in store_spec.go compared `int` to a slice.
  5. First `go mod tidy` dropped my ginkgo require ("all matched no packages"
     warning — noticed but not re-pinned), which let tidy pick v2.33.1 later.

## e) WHAT WE SHOULD IMPROVE

- Read exact API signatures (return types included) from source BEFORE
  writing dependent code; read locally-available library source (module
  cache) before authoring against it.
- On toolchain weirdness: go straight to content bisect in a /tmp copy; I
  burned several rounds on theories before bisecting.
- Stop writing placeholder content with `write` — finish the thought first.
- Run buildflow/format earlier in the flow to keep the tree continuously
  clean instead of a big normalization pass at the end.
- Update the todo list more often mid-flow (it drifted stale once).

## f) NEXT (ranked, up to 50)

**Blocker investigation**

1. In /tmp repro: add `func TestZ(t *testing.T)` to the misclassified spec file → recheck.
2. In /tmp repro: swap dot-imports for named ginkgo/gomega imports → recheck.
3. Run `go list` (no `-e`) in /tmp/grcopy/bdd after a clean `go mod tidy` there.
4. `go clean -cache` + retry real dir.
5. Force `GOTOOLCHAIN=go1.27.1` on the repro (some runs used plain go1.27.0).
6. Check whether classification differs under `go test` vs `go list`.
7. If confirmed toolchain-quirk: minimize to a bug-reportable case + pick a
   workaround (e.g. Test function per spec file is ginkgo-idiomatic anyway).
8. Fallback of last resort: in-package tests (`package bdd`) — violates the
   black-box rule; needs a conscious rule-break note if ever chosen.

**Spec calibration (once runnable)**

9. Run suite green; fix corpus query texts until top-hit assertions hold.
10. Verify `Embedded` counting semantics for the dedup spec (docs vs texts).
11. Verify hub-exclusion spec (DocumentKinds policy, TopK 3).
12. Verify MinScore 0.99 exact-text-match spec.
13. Verify neutral-policy `ContainHit("topic:storage")` spec.
14. Verify reload spec (NodeCount/VectorCount/SimilarPairs equality).
15. Verify warm-rebuild spec (Embedded 0, CacheHits == len(Vectors)).
16. Verify RenderContext non-empty + provider fields spec.
17. `grep -E '^F(It|Describe|Context)'` (no committed focus) before finishing.

**Lint + format**

18. `golangci-lint run ./...` inside bdd/ — confirm parent-config discovery.
19. Fix findings in kit prod files (exhaustruct/nlreturn/wsl_v5/wrapcheck/mnd).
20. `golangci-lint fmt` (or buildflow format) for gofumpt/golines/gci pass.
21. `dprint fmt` + `dprint check` after docs land.

**Gates + wiring**

22. Extend `.githooks/pre-push`: pristine build + tidy drift inside bdd/.
23. Extend go-test.yml with bdd steps (build/vet/test/lint/tidy drift).
24. Confirm root `go build ./...` still skips the nested module (no go.work).
25. Run `buildflow`; expect the 8 accepted root-package-files findings; verify
    NO new findings originate from bdd/.
26. `go test -race ./...` in both modules.
27. Planted-failure dry-run push to prove the extended pre-push gates.
28. Verify core go.mod/go.sum byte-identical (non-negotiable #3 audit).

**Docs**

29. docs/planning/2026-10-15_...-bdd-submodule.md — ADR-style decision record
    with the kit interface sketch (repo convention).
30. README.md: "BDD test kit" section (install as test-only dep + spec snippet).
31. CHANGELOG.md Unreleased/Added entry.
32. FEATURES.md row/section for the kit.
33. TODO_LIST.md: release-tagging task (bdd/v0.1.0 + core tag + require bump).
34. AGENTS.md: bdd layout, commands (bdd gates), gotchas (replace semantics,
    go.work gitignored rationale, tag scheme `bdd/v*`, ginkgo pin rationale,
    consumer adoption path for CV/knowledge-graph/legal-graph).
35. Note in ADR: release.yml `v*` trigger does not match `bdd/v*` tags — a
    separate trigger or manual release is needed for bdd tags.

**Alignment + final verification**

36. Decide ginkgo pin (v2.33.0 CV-parity vs v2.33.1) and record rationale.
37. Document (ADR §consumer path) how CV/legal-graph/knowledge-graph could adopt.
38. Full matrix: pristine build both modules, vet, test, lint, tidy drift,
    dprint check — all green.
39. Review final diff: zero core-module file changes.
40. Clean up /tmp artifacts (/tmp/bddprobe, /tmp/minirepro, /tmp/grcopy,
    /tmp/keep_*.go).

**Optional / deferred (explicit YAGNI records)**

41. `CreateBDDSuite` bootstrap helper (CV platform/bddkit parity) — skipped,
    pending owner preference.
42. `CountingProvider` — dropped; BuildResult counters suffice.
43. `WithCache(cache)` option — dropped until a consumer needs a custom cache.
44. bdd/README.md — root README + godoc deemed sufficient.
45. Runnable godoc examples with `// Output:` — deferred (testableexamples
    enforcement makes churn costlier than value now).
46. Live-provider specs — out of scope (env-gated live test already exists in core).
47. Version-drift detection script for the bdd→core require — note in ADR
    release checklist instead of a script for now.
48. Consider `HaveSimilarEdge` score-aware variant if consumers ask.
49. Consider `WithMemoryStore` option only if a cheap cache-wiring spec demand appears.
50. Benchmark protocol untouched — verify bench_test.go numbers unaffected.

## g) QUESTIONS FOR THE OWNER (cannot be decided from the repo)

1. **Release posture**: once verified, should I cut `bdd/v0.1.0` (plus
   whatever core tag/require bump it needs), or is tagging strictly
   owner-gated? Until a tag or SHA-pin exists, consumers cannot fetch the
   module from the proxy.
2. **Bootstrap helper**: should the kit also ship a CV-`platform/bddkit`-style
   `CreateBDDSuite(t, name)` helper, or keep the raw Ginkgo bootstrap
   (I skipped it as YAGNI)?
3. **ginkgo pin**: v2.33.1 (what tidy selected) or v2.33.0 (exact CV parity)?
