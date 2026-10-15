package bdd_test

import (
	graphrag "github.com/larsartmann/go-graph-rag"
	graphragbdd "github.com/larsartmann/go-graph-rag/bdd"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Kit builders", func() {
	Describe("Doc", func() {
		It("defaults kind to the demo article kind and label to the ID", func() {
			document := graphragbdd.Doc("article:1", "some text")

			Expect(document.ID).To(Equal(graphrag.NodeID("article:1")))
			Expect(document.Kind).To(Equal(graphragbdd.KindArticle))
			Expect(document.Label).To(Equal("article:1"))
			Expect(document.Text).To(Equal("some text"))
		})

		It("applies options in order over the defaults", func() {
			document := graphragbdd.Doc(
				"topic:1",
				"some text",
				graphragbdd.WithKind(graphragbdd.KindTopic),
				graphragbdd.WithLabel("Custom label"),
			)

			Expect(document.Kind).To(Equal(graphragbdd.KindTopic))
			Expect(document.Label).To(Equal("Custom label"))
		})
	})

	Describe("Edge", func() {
		It("defaults the weight to 1", func() {
			edge := graphragbdd.Edge("article:1", "topic:1", graphragbdd.RelationMentions)

			Expect(edge.Source).To(Equal(graphrag.NodeID("article:1")))
			Expect(edge.Target).To(Equal(graphrag.NodeID("topic:1")))
			Expect(edge.Relation).To(Equal(graphragbdd.RelationMentions))
			Expect(edge.Weight).To(Equal(1.0))
		})
	})

	Describe("WeightedEdge", func() {
		It("carries the explicit weight through to the graph edge", func() {
			edge := graphragbdd.WeightedEdge("topic:1", "topic:2", graphragbdd.RelationCites, 0.42)

			Expect(edge.Source).To(Equal(graphrag.NodeID("topic:1")))
			Expect(edge.Target).To(Equal(graphrag.NodeID("topic:2")))
			Expect(edge.Relation).To(Equal(graphragbdd.RelationCites))
			Expect(edge.Weight).To(Equal(0.42))
		})
	})
})
