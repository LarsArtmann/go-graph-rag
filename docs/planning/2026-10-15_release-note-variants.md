# Release-note variants for the pending cut (staged 2026-10-15)

Owner decides the version word; the matching block below is pasted into
`CHANGELOG.md` as the renamed section (the other blocks move under
`[Unreleased]` for the following cut). `release.yml` extracts everything
between `## [x.y.z]` and the next `## ` header, so keep `###` subsections.

## Variant A — `v0.3.3` (fixes only; `NodeID` + BDD kit held for v0.4.0)

### Fixed

- Dead allocation in the concurrent-embed collector (`embed_openai.go`): a
  2026-10-11 refactor switched slot-flattening to `lo.Flatten` but left the
  preceding `make([]Vector, 0, len(texts))` allocation in place, immediately
  overwritten — flagged by `ineffassign` and caught while re-baselining CI.
- CI Lint step green again: `setup-go`'s `go-version: "1.27"` resolved to Go
  1.27.2 on the runner, whose export-data format the pinned golangci-lint
  v2.13.2 cannot decode (every typecheck import failed; master red
  2026-10-11..15). The workflow now pins `go-version: "1.27.1"` to match the
  `GOTOOLCHAIN` pin; the two pins must move together with the golangci-lint
  pin (CV parity rule).
- Stale public-face docs re-synced (2026-10-15 docs-health pass): README's
  install command now pins `v0.3.2`, its dependency-light constraint states
  the json/v1 rule as policy (json/v2 graduated — the compiler no longer
  discriminates), and the `NodeID` strong-ID constraint is surfaced;
  `SECURITY.md`'s supported-versions table now says `v0.3.x` (was still
  `v0.1.x`, contradicting the file's own latest-line policy);
  `docs/DOMAIN_LANGUAGE.md` gains the `NodeID` term.

## Variant B — `v0.4.0` (everything currently under `[Unreleased]`)

### Added

- BDD test-kit sub-module `bdd/` (`github.com/larsartmann/go-graph-rag/bdd`,
  own go.mod): Ginkgo/Gomega test kit with `Doc`/`Edge` builders, a
  deterministic demo corpus + kind policy, a `Harness` over the real SDK
  machinery (hash provider by default, optional real SQLite store with
  `Reload` from disk), Gomega matchers (`HaveTopHit`, `ContainHit`,
  `HaveHitIDs`, `HaveNoHitsOfKind`, `HaveSimilarEdge`), and a
  `FailingProvider` for error paths. The kit keeps ginkgo/gomega out of the
  core dependency budget (AGENTS non-negotiable #3) and dogfoods itself via
  its own 43-spec black-box suite. Scope and decisions:
  `docs/planning/2026-10-15_05-02_bdd-submodule-adr.md`. Consumers cannot
  fetch it from the proxy until the first `bdd/v*` tag lands (owner-gated).

### Changed

- Node identity is now the strong type `NodeID` (brandid-lint BD003
  type-safety sweep): `Node.ID`, `Edge.Source`/`Edge.Target`, and
  `Document.ID` carry `graphrag.NodeID` (documents and nodes share one ID
  namespace: a document's ID becomes its node's ID), and every ID-taking
  surface follows — `Graph.Node`/`OutEdges`/`InEdges`, `BuildResult.Vectors`
  (`map[NodeID]Vector`) and `.EmbeddingHash`, `NewSearcher`/
  `NewSearcherWithOptions`, `Store.ReplaceGraph`,
  `LoadGraphSnapshot.EmbeddingHash`, plus the BDD kit's `Doc`/`Edge`/
  `WeightedEdge` builders and `HaveTopHit`/`ContainHit`/`HaveHitIDs`/
  `HaveSimilarEdge`/`SimilarEdgeScore`. JSON wire shapes are byte-identical
  (wire goldens unchanged), untyped string literals still assign directly
  (`Node{ID: "job:1"}` keeps compiling); only callers passing `string`
  variables add an explicit `graphrag.NodeID(...)` conversion. Consumer
  migration guide: `docs/upgrading.md`. The seam ADR §5 sketch was amended
  in step (ADR-first), so `VectorIndex.Replace` and `VectorMatch.NodeID` are
  `NodeID`-typed before any adapter exists.

### Fixed

(same three Fixed entries as Variant A)

## Consequences of each word

- **v0.3.3**: consumers get the two behavior fixes now; the `NodeID` sweep
  and the BDD kit stay unreleased; CV's pending bump should then skip
  straight to the later v0.4.0.
- **v0.4.0**: one cut ships all four sessions of work; CV bumps once;
  `UserAgentVersion` must move `0.3` → `0.4` in the same change (the UA↔
  CHANGELOG consistency test enforces it).
