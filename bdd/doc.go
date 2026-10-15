// Package bdd is the BDD-style test kit for github.com/larsartmann/go-graph-rag.
//
// It lives in its own module so the core SDK keeps its stdlib-only dependency
// budget while consumers get Ginkgo/Gomega support as an opt-in, test-only
// dependency. Import it from test files as:
//
//	import graphragbdd "github.com/larsartmann/go-graph-rag/bdd"
//
// The kit covers the three things every consumer of the SDK rebuilds by hand
// today: fixture builders for documents and edges, an offline harness that
// wires provider, optional SQLite store, Build, and Searcher together, and
// Gomega matchers for the recurring search-result assertions (top hit, hit
// inventory, near-duplicate edges, hub exclusion).
//
// A typical spec:
//
//	var _ = Describe("matching", func() {
//	    var harness *graphragbdd.Harness
//
//	    BeforeEach(func() {
//	        harness = graphragbdd.NewHarness(
//	            GinkgoT(),
//	            []graphrag.Document{graphragbdd.Doc("cv:senior", "senior go engineer")},
//	            []graphrag.Edge{graphragbdd.Edge("cv:senior", "skill:go", "mentions")},
//	        )
//	        _, err := harness.Build(ctx)
//	        Expect(err).NotTo(HaveOccurred())
//	    })
//
//	    It("ranks the closest document first", func() {
//	        result, err := harness.Search(ctx, "go engineer")
//	        Expect(err).NotTo(HaveOccurred())
//	        Expect(result).To(graphragbdd.HaveTopHit("cv:senior"))
//	    })
//	})
//
// Specs run against real SDK machinery, never mocks: the default provider is
// the deterministic offline hash embedder and WithStore persists through a
// real SQLite file, matching how the SDK is used in production.
//
// # Typed DescribeTable entries
//
// Ginkgo resolves Entry arguments through reflection. Once a table function's
// parameters are named types (graphrag.NodeID, graphrag.NodeKind, ...), a
// plain string literal inside Entry panics at suite build time with
// "reflect: Call using string as type graphrag.NodeID". Convert literals in
// the Entry — Entry("ranks first", graphrag.NodeID("doc:1")) — and keep
// untyped literals only for parameters that are still plain string.
package bdd
