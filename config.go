package graphrag

import "time"

// Config is the canonical graphrag configuration (koanf-tagged so a
// consumer's root config can alias it). It selects the embedding
// provider, the SQLite index location, and the graph-tuning knobs.
type Config struct {
	// Enabled lets the consumer gate the feature; a false value is a
	// consumer-side decision (the SDK itself defines no disabled state).
	Enabled bool `koanf:"enabled"`

	// StoreDSN is the SQLite file backing the index (":memory:" for
	// throwaway stores).
	StoreDSN string `koanf:"store_dsn"`

	// Embedding selects and parameterizes the provider.
	Embedding EmbeddingConfig `koanf:"embedding"`

	// SimilarTopK is the per-node nearest-neighbor count for
	// similar_to edges (0 selects DefaultSimilarTopK).
	SimilarTopK int `koanf:"similar_top_k"`

	// SimilarThreshold is the minimum cosine for a similar_to edge
	// (0 selects DefaultSimilarThreshold).
	SimilarThreshold float64 `koanf:"similar_threshold"`
}

// EmbeddingConfig selects and parameterizes a Provider. It doubles as the
// koanf-tagged configuration section (camel-cased keys mirror the struct
// names used by the root config).
type EmbeddingConfig struct {
	// Provider selects the implementation: "" or "hash" (offline default),
	// "openai-compat" for any OpenAI-compatible /embeddings endpoint.
	Provider string `koanf:"provider"`

	// BaseURL is the API root for openai-compat (e.g.
	// https://api.openai.com/v1).
	BaseURL string `koanf:"base_url"`

	// APIKey is the bearer token for openai-compat (env-injected).
	APIKey string `koanf:"api_key"`

	// Model names the embedding model; switching it invalidates the cache.
	Model string `koanf:"model"`

	// Timeout bounds one embeddings request (0 selects the default).
	Timeout time.Duration `koanf:"timeout"`

	// MaxRetries is the retry budget for transient failures.
	MaxRetries int `koanf:"max_retries"`

	// MaxChars bounds each embedded text (0 selects the default).
	MaxChars int `koanf:"max_chars"`

	// EmbedConcurrency is the maximum number of /embeddings batch requests
	// sent in parallel (0 or 1 selects the serial path). Values above 1
	// help network-bound cold builds: batches are independent HTTP calls,
	// so they overlap safely, and results keep input order regardless of
	// completion order. EmbedConcurrency never changes vectors, only how
	// fast they arrive.
	EmbedConcurrency int `koanf:"embed_concurrency"`
}
