package bdd_test

import (
	graphrag "github.com/larsartmann/go-graph-rag"
	graphragbdd "github.com/larsartmann/go-graph-rag/bdd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func searchResultWithHits(ids ...graphrag.NodeID) *graphrag.SearchResult {
	hits := make([]graphrag.Hit, 0, len(ids))

	for index, id := range ids {
		hits = append(hits, graphrag.Hit{
			Node: graphrag.Node{
				ID:   id,
				Kind: graphragbdd.KindArticle,
			},
			Score: float64(len(ids)-index) / 10,
		})
	}

	return &graphrag.SearchResult{Hits: hits}
}

func similarPairs() []graphrag.SimilarPair {
	return []graphrag.SimilarPair{
		{A: graphrag.Node{ID: "a"}, B: graphrag.Node{ID: "b"}, Score: 0.9},
		{A: graphrag.Node{ID: "c"}, B: graphrag.Node{ID: "d"}, Score: 0.8},
	}
}

var _ = Describe("Kit matchers", func() {
	DescribeTable("search-result matchers",
		func(matcher types.GomegaMatcher, actual any, shouldMatch bool) {
			matched, err := matcher.Match(actual)
			Expect(err).NotTo(HaveOccurred())
			Expect(matched).To(Equal(shouldMatch))
		},
		Entry("HaveTopHit matches the ranked first hit",
			graphragbdd.HaveTopHit("a"), searchResultWithHits("a", "b"), true),
		Entry("HaveTopHit rejects a lower-ranked hit",
			graphragbdd.HaveTopHit("b"), searchResultWithHits("a", "b"), false),
		Entry("HaveTopHit rejects an empty hit list",
			graphragbdd.HaveTopHit("a"), searchResultWithHits(), false),
		Entry("ContainHit matches at any rank",
			graphragbdd.ContainHit("b"), searchResultWithHits("a", "b", "c"), true),
		Entry("ContainHit rejects absent nodes",
			graphragbdd.ContainHit("z"), searchResultWithHits("a", "b"), false),
		Entry("HaveHitIDs pins the exact rank order",
			graphragbdd.HaveHitIDs("a", "b"), searchResultWithHits("a", "b"), true),
		Entry("HaveHitIDs rejects reordered hits",
			graphragbdd.HaveHitIDs("b", "a"), searchResultWithHits("a", "b"), false),
		Entry("HaveHitIDs rejects shorter hit lists",
			graphragbdd.HaveHitIDs("a"), searchResultWithHits("a", "b"), false),
		Entry("HaveNoHitsOfKind matches a kind-free hit list",
			graphragbdd.HaveNoHitsOfKind(graphragbdd.KindTopic), searchResultWithHits("a", "b"), true),
		Entry("HaveSimilarEdge matches in either direction",
			graphragbdd.HaveSimilarEdge("b", "a"), similarPairs(), true),
		Entry("HaveSimilarEdge rejects absent pairs",
			graphragbdd.HaveSimilarEdge("a", "c"), similarPairs(), false),
	)

	It("reports an error when the actual value has the wrong type", func() {
		_, err := graphragbdd.HaveTopHit("a").Match("not a search result")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unexpected actual type"))
	})

	It("describes the ranked hits in failure messages", func() {
		message := graphragbdd.HaveTopHit("z").FailureMessage(searchResultWithHits("a", "b"))
		Expect(message).To(ContainSubstring(`top hit "z"`))
		Expect(message).To(ContainSubstring("a(article)=0.200"))
	})

	It("exposes the score of a looked-up similar edge", func() {
		score, ok := graphragbdd.SimilarEdgeScore(similarPairs(), "d", "c")
		Expect(ok).To(BeTrue())
		Expect(score).To(BeNumerically("~", 0.8, 0.0001))

		missingScore, missing := graphragbdd.SimilarEdgeScore(similarPairs(), "a", "d")
		Expect(missing).To(BeFalse())
		Expect(missingScore).To(BeZero())
	})
})
