package bdd_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	graphrag "github.com/larsartmann/go-graph-rag"
	graphragbdd "github.com/larsartmann/go-graph-rag/bdd"
)

var _ = Describe("Persisting through the store", func() {
	var (
		ctx     context.Context
		harness *graphragbdd.Harness
		result  *graphrag.BuildResult
	)

	BeforeEach(func() {
		ctx = context.Background()
		docs, edges := graphragbdd.DemoCorpus()
		harness = graphragbdd.NewHarness(
			GinkgoT(), docs, edges,
			graphragbdd.WithStore(),
			graphragbdd.WithSearcherOptions(graphragbdd.DemoSearcherOptions()),
		)

		var buildErr error
		result, buildErr = harness.Build(ctx)
		Expect(buildErr).NotTo(HaveOccurred())
	})

	It("reloads the persisted graph and cached embeddings", func() {
		reloaded, err := harness.Reload()
		Expect(err).NotTo(HaveOccurred())

		Expect(reloaded.Searcher.NodeCount()).To(Equal(len(result.Nodes)))
		Expect(reloaded.Searcher.VectorCount()).To(Equal(len(result.Vectors)))

		freshResult, searchErr := reloaded.Search(ctx, "hybrid search combines vector similarity")
		Expect(searchErr).NotTo(HaveOccurred())
		Expect(freshResult).To(graphragbdd.HaveTopHit("article:retrieval"))

		Expect(reloaded.SimilarPairs(graphrag.DefaultSimilarThreshold)).
			To(Equal(harness.SimilarPairs(graphrag.DefaultSimilarThreshold)))
	})

	It("serves every vector of a rebuild from the cache", func() {
		coldResult := result

		warmResult, err := harness.Build(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(warmResult.Embedded).To(BeZero())
		Expect(warmResult.CacheHits).To(Equal(len(coldResult.Vectors)))
		Expect(warmResult.Vectors).To(HaveLen(len(coldResult.Vectors)))
	})

	It("refuses to reload a harness without a store", func() {
		docs, edges := graphragbdd.DemoCorpus()
		bare := graphragbdd.NewHarness(GinkgoT(), docs, edges)

		_, buildErr := bare.Build(ctx)
		Expect(buildErr).NotTo(HaveOccurred())

		_, reloadErr := bare.Reload()
		Expect(reloadErr).To(MatchError(ContainSubstring("WithStore not configured")))
	})
})
