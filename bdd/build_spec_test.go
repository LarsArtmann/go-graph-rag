package bdd_test

import (
	"context"
	"errors"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	graphrag "github.com/larsartmann/go-graph-rag"
	graphragbdd "github.com/larsartmann/go-graph-rag/bdd"
)

var errProviderBoom = errors.New("spec: provider exploded")

type flakyDimsProvider struct{}

func (flakyDimsProvider) Name() string  { return "flaky-dims" }
func (flakyDimsProvider) Model() string { return "v1" }

func (flakyDimsProvider) Embed(_ context.Context, _ []string) ([]graphrag.Vector, error) {
	return []graphrag.Vector{
		make(graphrag.Vector, graphrag.HashProviderDims),
		make(graphrag.Vector, graphrag.HashProviderDims+1),
	}, nil
}

var _ = Describe("Building an index", func() {
	var (
		ctx     context.Context
		harness *graphragbdd.Harness
		result  *graphrag.BuildResult
	)

	BeforeEach(func() {
		ctx = context.Background()
		docs, edges := graphragbdd.DemoCorpus()
		harness = graphragbdd.NewHarness(GinkgoT(), docs, edges)
	})

	JustBeforeEach(func() {
		var buildErr error
		result, buildErr = harness.Build(ctx)
		Expect(buildErr).NotTo(HaveOccurred())
	})

	It("embeds every text-carrying document", func() {
		Expect(result.Vectors).To(HaveLen(6))
		Expect(result.EmbeddingHash).To(HaveLen(6))
	})

	It("keeps every caller edge verbatim", func() {
		_, edges := graphragbdd.DemoCorpus()

		for _, edge := range edges {
			Expect(result.Edges).To(ContainElement(edge))
		}
	})

	It("derives a similar edge between the near-duplicate articles", func() {
		pairs := harness.SimilarPairs(graphrag.DefaultSimilarThreshold)

		Expect(pairs).To(graphragbdd.HaveSimilarEdge("article:retrieval", "article:retrieval-dup"))

		score, ok := graphragbdd.SimilarEdgeScore(pairs, "article:retrieval", "article:retrieval-dup")
		Expect(ok).To(BeTrue())
		Expect(score).To(BeNumerically(">=", graphrag.DefaultSimilarThreshold))
	})

	DescribeTable("connects lexically distinct articles through nothing",
		func(aID string, bID string) {
			pairs := harness.SimilarPairs(graphrag.DefaultSimilarThreshold)
			Expect(pairs).NotTo(graphragbdd.HaveSimilarEdge(aID, bID))
		},
		Entry("retrieval vs storage", "article:retrieval", "article:storage"),
		Entry("retrieval vs chunking", "article:retrieval", "article:chunking"),
		Entry("storage vs chunking", "article:storage", "article:chunking"),
	)

	It("produces the identical graph on a second cold build", func() {
		docs, edges := graphragbdd.DemoCorpus()
		second := graphragbdd.NewHarness(GinkgoT(), docs, edges)

		secondResult, err := second.Build(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(secondResult.Nodes).To(Equal(result.Nodes))
		Expect(secondResult.Edges).To(Equal(result.Edges))
	})

	It("embeds identical texts once and records the shared hash", func() {
		same := "Repeated indexing text hashed into one vector"
		docs := []graphrag.Document{
			graphragbdd.Doc("doc:first", same),
			graphragbdd.Doc("doc:second", same),
		}

		dedup := graphragbdd.NewHarness(GinkgoT(), docs, nil)
		dedupResult, err := dedup.Build(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(dedupResult.Vectors).To(HaveLen(1))
		Expect(dedupResult.EmbeddingHash).To(HaveLen(2))
		Expect(dedupResult.Embedded).To(Equal(1))
	})

	It("propagates provider failures", func() {
		docs, edges := graphragbdd.DemoCorpus()
		failing := graphragbdd.NewHarness(
			GinkgoT(), docs, edges, graphragbdd.WithProvider(graphragbdd.FailingProvider{Err: errProviderBoom}))

		_, err := failing.Build(ctx)
		Expect(err).To(MatchError(errProviderBoom))
	})

	It("rejects providers that mix vector dimensionalities", func() {
		docs := []graphrag.Document{
			graphragbdd.Doc("doc:a", "first text"),
			graphragbdd.Doc("doc:b", "second text"),
		}

		mixed := graphragbdd.NewHarness(
			GinkgoT(), docs, nil, graphragbdd.WithProvider(flakyDimsProvider{}))

		_, err := mixed.Build(ctx)
		Expect(err).To(HaveOccurred())

		var dimsErr *graphrag.ErrDimsInconsistent
		Expect(errors.As(err, &dimsErr)).To(BeTrue())
	})

	It("collapses duplicate document IDs onto the first occurrence", func() {
		docs := []graphrag.Document{
			graphragbdd.Doc("hub:skill", "go and graphql", graphragbdd.WithLabel("first")),
			graphragbdd.Doc("hub:skill", "different text entirely", graphragbdd.WithLabel("second")),
		}

		dedup := graphragbdd.NewHarness(GinkgoT(), docs, nil)
		dedupResult, err := dedup.Build(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(dedupResult.Nodes).To(HaveLen(1))
		Expect(slices.ContainsFunc(dedupResult.Nodes, func(node graphrag.Node) bool {
			return node.ID == "hub:skill" && node.Label == "first"
		})).To(BeTrue())
	})
})
