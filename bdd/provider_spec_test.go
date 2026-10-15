package bdd_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	graphrag "github.com/larsartmann/go-graph-rag"
)

var _ = Describe("Embedding providers", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("selects the offline hash provider for an empty name", func() {
		provider, err := graphrag.NewProvider(graphrag.EmbeddingConfig{})
		Expect(err).NotTo(HaveOccurred())
		Expect(provider.Name()).To(Equal(graphrag.ProviderHash))
		Expect(provider.Model()).To(Equal("v1"))
	})

	It("rejects unknown provider names", func() {
		_, err := graphrag.NewProvider(graphrag.EmbeddingConfig{Provider: "no-such-provider"})
		Expect(err).To(MatchError(graphrag.ErrUnknownProvider))
	})

	It("embeds the same text into the same vector every time", func() {
		provider := graphrag.NewHashProvider()

		first, err := provider.Embed(ctx, []string{"deterministic offline embedding"})
		Expect(err).NotTo(HaveOccurred())

		second, err := provider.Embed(ctx, []string{"deterministic offline embedding"})
		Expect(err).NotTo(HaveOccurred())

		Expect(first).To(HaveLen(1))
		Expect(first[0]).To(Equal(second[0]))
		Expect(first[0].Dims()).To(Equal(graphrag.HashProviderDims))
	})

	DescribeTable("truncating on rune boundaries",
		func(text string, limit int, expected string) {
			Expect(graphrag.TruncateRunes(text, limit)).To(Equal(expected))
		},
		Entry("keeps short texts untouched", "short text", 10, "short text"),
		Entry("cuts long ASCII texts", "abcdefghij", 4, "abcd"),
		Entry("never splits a multi-byte rune", "héllo wörld", 5, "héllo"),
	)
})
