package bdd_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	graphrag "github.com/larsartmann/go-graph-rag"
	graphragbdd "github.com/larsartmann/go-graph-rag/bdd"
)

var _ = Describe("Searching the built index", func() {
	var (
		ctx     context.Context
		harness *graphragbdd.Harness
	)

	BeforeEach(func() {
		ctx = context.Background()
		docs, edges := graphragbdd.DemoCorpus()
		harness = graphragbdd.NewHarness(
			GinkgoT(), docs, edges, graphragbdd.WithSearcherOptions(graphragbdd.DemoSearcherOptions()))

		_, err := harness.Build(ctx)
		Expect(err).NotTo(HaveOccurred())
	})

	It("ranks the article closest to the query first", func() {
		result, err := harness.Search(ctx, "hybrid search combines vector similarity with graph expansion")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(graphragbdd.HaveTopHit("article:retrieval"))
	})

	It("expands hits through caller edges into topic hubs", func() {
		result, err := harness.Search(ctx, "hybrid search combines vector similarity")
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Hits).NotTo(BeEmpty())

		relatedIDs := make([]string, 0)
		for _, hit := range result.Hits {
			for _, related := range hit.Related {
				relatedIDs = append(relatedIDs, related.Node.ID)
			}
		}

		Expect(relatedIDs).To(ContainElement("topic:search"))
	})

	It("caps the hit list at TopK", func() {
		result, err := harness.Search(ctx, "documents", graphrag.SearchOptions{TopK: 2})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Hits).To(HaveLen(2))
	})

	It("keeps hub kinds out of hits while documents fill the slots", func() {
		result, err := harness.Search(ctx, "documents embedding", graphrag.SearchOptions{TopK: 3})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Hits).To(HaveLen(3))
		Expect(result).To(graphragbdd.HaveNoHitsOfKind(graphragbdd.KindTopic))
	})

	It("drops weak matches below MinScore", func() {
		result, err := harness.Search(ctx, "storage cache sqlite", graphrag.SearchOptions{MinScore: 0.99})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Hits).To(BeEmpty())
	})

	It("keeps exact-text matches above a near-one MinScore", func() {
		result, err := harness.Search(
			ctx,
			"The embedding cache persists vectors in SQLite keyed by content hash.",
			graphrag.SearchOptions{MinScore: 0.99},
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(graphragbdd.ContainHit("article:storage"))
	})

	It("renders a non-empty context for the LLM", func() {
		result, err := harness.Search(ctx, "hybrid search combines vector similarity")
		Expect(err).NotTo(HaveOccurred())
		Expect(result.ContextText).NotTo(BeEmpty())
		Expect(result.Provider).To(Equal(graphrag.ProviderHash))
	})

	It("propagates query embedding failures", func() {
		docs, edges := graphragbdd.DemoCorpus()
		failing := graphragbdd.NewHarness(
			GinkgoT(), docs, edges,
			graphragbdd.WithProvider(graphragbdd.FailingProvider{Err: errProviderBoom}),
			graphragbdd.WithSearcherOptions(graphragbdd.DemoSearcherOptions()),
		)

		_, buildErr := failing.Build(ctx)
		Expect(buildErr).NotTo(HaveOccurred())

		_, searchErr := failing.Search(ctx, "any query")
		Expect(searchErr).To(MatchError(errProviderBoom))
	})
})

var _ = Describe("Searching under the neutral policy", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("ranks every embedded node, topic hubs included", func() {
		docs, edges := graphragbdd.DemoCorpus()
		harness := graphragbdd.NewHarness(GinkgoT(), docs, edges)

		_, err := harness.Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		result, searchErr := harness.Search(ctx, "storage and caching")
		Expect(searchErr).NotTo(HaveOccurred())
		Expect(result).To(graphragbdd.ContainHit("topic:storage"))
	})
})
