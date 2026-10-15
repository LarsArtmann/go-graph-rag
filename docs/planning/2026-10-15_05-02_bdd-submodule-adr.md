# ADR — The BDD test-kit sub-module (`bdd/`)

|          |                                                                                                                                                                                                                                     |
| -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Date     | 2026-10-15                                                                                                                                                                                                                          |
| Status   | Accepted and implemented (kit code + 47-spec suite in tree; first `bdd/v*` tag pending owner decision)                                                                                                                              |
| Question | How do consumers test go-graph-rag in BDD style (Ginkgo/Gomega) over the REAL SDK machinery, without the core module's `go.mod` ever growing test-only dependencies?                                                                |
| Decision | A separate, composable Go module `github.com/larsartmann/go-graph-rag/bdd` owning ginkgo+gomega, exposing builders, a demo corpus, a real-machinery `Harness`, Gomega matchers, and a `FailingProvider`. Core stays byte-identical. |
| Method   | Three consumers studied file-by-file (CV, knowledge-graph, legal-graph-ai-system, all pinning v0.3.2); kit scope fixed by the fixtures every consumer hand-rolls; dogfooded by the module's own black-box spec suite.               |

## 1. Context

All three known consumers test against the real SDK surface and hand-roll the same fixtures:

- **CV** (`/home/lars/projects/CV`): testify-based unit tests plus `platform/bddkit` with a
  `CreateBDDSuite` bootstrap; ginkgo v2.33.0 / gomega v1.44.0 already in its dependency tree.
- **knowledge-graph**: stdlib-only table tests (testify banned there); no ginkgo. A BDD kit must
  be opt-in, never transitive.
- **legal-graph-ai-system**: testify + a mini-corpus fixture; its `TODO_LIST.md:57` plans a
  testify→ginkgo migration.

Meanwhile AGENTS.md non-negotiable #3 pins the core module budget at stdlib + `samber/lo` +
`modernc.org/sqlite` (+testify in tests). A test-only dependency (ginkgo, gomega) in core would
break that budget for every consumer, and an in-repo `internal/testutil` would force the
dependency question anyway the moment consumers import it. The go-modularize failure mode this
avoids: test deps leaking into a module's public dependency graph.

## 2. Decision

1. **Separate module** `bdd/` with module path `github.com/larsartmann/go-graph-rag/bdd`. It
   requires the SDK by published version (`v0.3.2` today) and carries a `replace
   github.com/larsartmann/go-graph-rag => ../` for local/CI development. Replace directives
   apply only inside the main module, so consumers always resolve the require from the proxy —
   the committed replace is invisible to them.
2. **Kit owns exactly what consumers hand-roll** (evidence-based scope):
   - `Doc`/`Edge`/`WeightedEdge` builders (`builders.go`) — the `edge()` helper shape from
     legal-graph-ai-system, generalized.
   - `DemoCorpus` + demo vocabulary + `DemoSearcherOptions` (`corpus.go`) — a six-document,
     five-edge article/topic corpus calibrated so hash-provider behaviors (near-duplicate
     `RelationSimilar`, hub expansion, cache hits) are deterministic.
   - `Harness` (`harness.go`) — provider + optional real SQLite store + `Build` + `Searcher` +
     `Reload` (rehydrating from disk via `LoadGraph`/`LoadEmbeddings` and the
     `snapshot.EmbeddingHash` join). Accepts a minimal `TB` interface satisfied by both
     `testing.TB` and `GinkgoT()`.
   - Gomega matchers (`matchers.go`): `HaveTopHit`, `ContainHit`, `HaveHitIDs`,
     `HaveNoHitsOfKind`, `HaveSimilarEdge`, plus `SimilarEdgeScore`.
   - `FailingProvider` (`providers.go`) for error-path specs.
3. **The module's own suite is the acceptance suite**: 47 black-box specs
   (`package bdd_test`) exercising build/search/store/provider/matcher behavior against the
   real SDK — dogfooding from the exact consumer perspective the kit exists for.
4. **Explicit YAGNI drops** (revisit only on consumer demand): `CreateBDDSuite` bootstrap helper
   (raw `RegisterFailHandler`+`RunSpecs` is three lines), `CountingProvider`, `WithCache`
   without store, `WithMemoryStore`, kit-local runnable godoc examples (the spec suite is the
   example).

## 3. Interface sketch (binding, implemented)

> Amended 2026-10-15 (BD003 strong-ID hardening): core node IDs became the strong type
> `graphrag.NodeID`, so the kit's ID-taking surfaces below carry it. Untyped string literals still
> compile at every call site; only variable-typed callers add a conversion.
>
> Amended 2026-10-15 (API review, pre-freeze): `Harness.Search`'s variadic options keep the
> signature, but the contract is hardened — at most one `SearchOptions` is accepted; passing more
> fails the search instead of silently dropping options (a variadic cannot merge them, and a kit
> must surface spec bugs loudly). The sketch above is otherwise FROZEN for `bdd/v0.1.0`.

```go
package bdd // import github.com/larsartmann/go-graph-rag/bdd

// Minimal test facade: satisfied by testing.TB values AND ginkgo.GinkgoT().
type TB interface {
    Fatalf(format string, args ...any)
    TempDir() string
    Cleanup(cleanup func())
}

// builders.go
func Doc(id graphrag.NodeID, text string, options ...DocOption) graphrag.Document // WithKind, WithLabel
func Edge(source, target graphrag.NodeID, relation graphrag.Relation) graphrag.Edge
func WeightedEdge(source, target graphrag.NodeID, relation graphrag.Relation, weight float64) graphrag.Edge

// corpus.go — demo vocabulary: KindArticle, KindTopic, RelationCites, RelationMentions
func DemoCorpus() ([]graphrag.Document, []graphrag.Edge)
func DemoSearcherOptions() graphrag.SearcherOptions

// providers.go
type FailingProvider struct{ Err error } // implements graphrag.Provider; Embed always fails

// harness.go
type Harness struct { TB; Provider; Cache; Store; Result; Searcher /*; private corpus+policy*/ }
type Option func(*Harness) // WithProvider, WithBuildOptions, WithSearcherOptions, WithStore
func NewHarness(tb TB, documents []graphrag.Document, edges []graphrag.Edge, options ...Option) *Harness
func (h *Harness) Build(ctx context.Context) (*graphrag.BuildResult, error)
func (h *Harness) Search(ctx context.Context, query string, opts ...graphrag.SearchOptions) (*graphrag.SearchResult, error)
func (h *Harness) SimilarPairs(minScore float64) []graphrag.SimilarPair
func (h *Harness) Reload() (*Harness, error) // reopen store, LoadGraph+LoadEmbeddings, hash-join vectors

// matchers.go — on *graphrag.SearchResult (HaveSimilarEdge on []graphrag.SimilarPair)
func HaveTopHit(nodeID graphrag.NodeID) types.GomegaMatcher
func ContainHit(nodeID graphrag.NodeID) types.GomegaMatcher
func HaveHitIDs(nodeIDs ...graphrag.NodeID) types.GomegaMatcher
func HaveNoHitsOfKind(kind graphrag.NodeKind) types.GomegaMatcher
func HaveSimilarEdge(aID, bID graphrag.NodeID) types.GomegaMatcher
func SimilarEdgeScore(pairs []graphrag.SimilarPair, aID, bID graphrag.NodeID) (float64, bool)
```

## 4. Dependency rules

- Kit budget: `ginkgo/v2 v2.33.1` + `gomega v1.44.0` + the SDK. Nothing else without an ADR
  amendment. (v2.33.1 over CV's v2.33.0: patch bump, tidy's resolution, no API delta.)
- The kit imports no `encoding/json` at all, keeping the v1-only policy trivially true.
- Core `go.mod`/`go.sum` must stay byte-identical through kit work (CI tidy-drift gate covers
  both modules separately).

## 5. Release and tagging

- Consumers cannot fetch the module until a `bdd/v*` tag exists. Tagging is owner-gated (see
  TODO_LIST); until then the sibling-checkout replace is the only way to use it.
- `.github/workflows/release.yml` triggers on `v*` only, which does NOT match `bdd/v*`: cutting a
  kit release needs either a trigger addition or a manual release. Decide at first tag time.
- CI (`go-test.yml`) and the pre-push hook run build+vet+test+lint+tidy for BOTH modules; the
  bdd pristine gate includes `go vet` because `go build` skips `_test.go` files and the spec
  suite is most of that module.

## 6. Triggers for revisiting

- A second consumer adopts the kit → reconsider `CreateBDDSuite`-style bootstrap parity.
- Kit needs fake/mocked SDK seams → those belong behind core interfaces (seam ADR), not in the
  kit; the kit wires REAL machinery only.
- Core adds a `Provider`/`Cache`-adjacent interface the kit should harness → extend `Harness`
  options, keep the zero-value defaults neutral.
