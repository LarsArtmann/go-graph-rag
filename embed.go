package graphrag

import (
	"context"
	"errors"
	"fmt"
)

// Provider names accepted by NewProvider.
const (
	// ProviderHash is the deterministic offline embedder. It needs no
	// network, no API key, and produces stable vectors for tests and
	// cold-start environments, at the cost of retrieval quality.
	ProviderHash = "hash"

	// ProviderOpenAICompat targets any OpenAI-compatible /embeddings
	// endpoint (OpenAI, Mistral, Ollama, vLLM, LM Studio, ...).
	ProviderOpenAICompat = "openai-compat"
)

// DefaultEmbeddingMaxChars bounds each text sent to an embedding provider.
// Most embedding models cap input around 8k tokens; truncating here keeps
// requests inside model limits without callers having to know them.
const DefaultEmbeddingMaxChars = 8000

// ErrUnknownProvider is returned by NewProvider for a provider name that no
// implementation is registered under.
var ErrUnknownProvider = errors.New("graphrag: unknown embedding provider")

// Provider converts texts into dense vectors. Implementations must be safe
// for concurrent use.
type Provider interface {
	// Name identifies the implementation in the embedding cache key
	// namespace so vectors from different providers never mix.
	Name() string

	// Model identifies the model (or version) whose vectors are being
	// cached; switching models must invalidate the cache.
	Model() string

	// Embed returns exactly one vector per input text, in input order.
	Embed(ctx context.Context, texts []string) ([]Vector, error)
}

// NewProvider builds the Provider named by cfg.Provider. An empty name
// selects the offline hash provider so that a minimal configuration still
// works end to end.
func NewProvider( //nolint:ireturn // factory by design; callers consume the Provider seam
	cfg EmbeddingConfig,
) (Provider, error) {
	switch cfg.Provider {
	case "", ProviderHash:
		return NewHashProvider(), nil
	case ProviderOpenAICompat:
		return NewOpenAICompatProvider(cfg)
	default:
		return nil, fmt.Errorf(
			"%w: %q (want %q or %q)",
			ErrUnknownProvider,
			cfg.Provider,
			ProviderHash,
			ProviderOpenAICompat,
		)
	}
}

// truncateRunes bounds text to at most max runes. A non-positive max falls
// back to the package default.
func truncateRunes(text string, limit int) string {
	if limit <= 0 {
		limit = DefaultEmbeddingMaxChars
	}

	return TruncateRunes(text, limit)
}

// TruncateRunes bounds text to at most limit runes, cutting on a rune boundary
// so multi-byte characters are never split.
func TruncateRunes(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}

	return string(runes[:limit])
}
