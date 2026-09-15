package graphrag

import (
	"context"
	"hash/fnv"
	"strings"

	"github.com/samber/lo"
)

// HashProvider tuning knobs. Dimensions are fixed (not configurable) so the
// cache namespace "hash/v1" always means the same vector space.
const (
	// HashProviderDims is the dimensionality of hash embeddings.
	HashProviderDims = 1024

	// hashUnigramWeight is the contribution of a single token occurrence.
	hashUnigramWeight = 1.0

	// hashBigramWeight is the contribution of an adjacent token pair
	// occurrence; bigrams sharpen matching beyond bag-of-words.
	hashBigramWeight = 0.7

	// hashMinTokenLen drops tokens shorter than this (noise like "a", "x").
	hashMinTokenLen = 2
)

// HashProvider is a deterministic, offline embedding Provider built on the
// hashing trick: unigrams and bigrams are hashed into a fixed-width vector
// which is then L2-normalized. It exists so that semantic search works with
// zero configuration (and so tests never touch the network). Its quality is
// far below a real embedding model; point graphrag.embedding.provider at an
// OpenAI-compatible endpoint for production retrieval.
type HashProvider struct{}

// NewHashProvider returns the stateless hash embedder.
func NewHashProvider() *HashProvider {
	return &HashProvider{}
}

// Name identifies the provider in the embedding cache namespace.
func (p *HashProvider) Name() string {
	return ProviderHash
}

// Model identifies the vector-space version; bumping it invalidates caches.
func (p *HashProvider) Model() string {
	return "v1"
}

// Embed maps every text to its hashed vector. It never fails and ignores the
// context because it performs no I/O.
func (p *HashProvider) Embed(_ context.Context, texts []string) ([]Vector, error) {
	vectors := lo.Map(texts, func(text string, _ int) Vector { return hashEmbed(text) })

	return vectors, nil
}

// hashEmbed computes the normalized hashing-trick vector of text.
func hashEmbed(text string) Vector {
	vector := make(Vector, HashProviderDims) //nolint:makezero // fixed dimensionality, index-assigned
	tokens := hashTokenize(text)

	for i, token := range tokens {
		addHashed(vector, token, hashUnigramWeight)

		if i > 0 {
			addHashed(vector, tokens[i-1]+" "+token, hashBigramWeight)
		}
	}

	return vector.Normalize()
}

// addHashed accumulates weight at the token's hashed index.
func addHashed(vector Vector, token string, weight float64) {
	index := fnv.New32a()
	_, _ = index.Write([]byte(token)) // fnv never fails writing to its own buffer

	vector[index.Sum32()%HashProviderDims] += float32(weight)
}

// hashTokenize lowercases and splits text on any non-alphanumeric boundary,
// dropping tokens shorter than hashMinTokenLen.
func hashTokenize(text string) []string {
	raw := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !isWordRune(r)
	})

	tokens := make([]string, 0, len(raw))

	for _, token := range raw {
		if len([]rune(token)) >= hashMinTokenLen {
			tokens = append(tokens, token)
		}
	}

	return tokens
}

// isWordRune reports whether r can be part of a token.
func isWordRune(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z',
		r >= 'A' && r <= 'Z',
		r >= '0' && r <= '9',
		r == '+', r == '#', r == '.':
		return true
	default:
		return false
	}
}
