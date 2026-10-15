package bdd

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	graphrag "github.com/larsartmann/go-graph-rag"
)

// errNoStore backs Reload when the harness was built without WithStore.
var errNoStore = errors.New("bdd harness: WithStore not configured")

// errNotBuilt backs Search and SimilarPairs when Build has not succeeded yet.
var errNotBuilt = errors.New("bdd harness: Build must succeed first")

// TB is the minimal test facade the harness needs: fatal setup errors, a
// per-test temp directory, and cleanup registration. Both testing.TB values
// (plain Test functions) and ginkgo.GinkgoT() satisfy it, so plain and BDD
// specs share one harness.
type TB interface {
	Fatalf(format string, args ...any)
	TempDir() string
	Cleanup(cleanup func())
}

// Harness wires the real SDK machinery a spec needs: provider, an optional
// SQLite store acting as Build cache and graph persistence, Build, and a
// Searcher over the built index. Construct it with NewHarness; the zero
// Harness is not usable.
type Harness struct {
	// TB receives fatal setup errors and owns cleanup of opened stores.
	TB TB

	// Provider embeds documents and queries: the deterministic offline
	// hash provider unless WithProvider says otherwise.
	Provider graphrag.Provider

	// Cache serves Build cache lookups; nil unless WithStore was used.
	Cache graphrag.Cache

	// Store persists graph and embeddings when WithStore is used; nil
	// otherwise.
	Store *graphrag.Store

	// Result holds the outcome of the latest successful Build.
	Result *graphrag.BuildResult

	// Searcher answers queries over the latest built or reloaded index.
	Searcher *graphrag.Searcher

	documents    []graphrag.Document
	edges        []graphrag.Edge
	dsn          string
	buildOptions graphrag.BuildOptions
	searchPolicy graphrag.SearcherOptions
}

// Option configures a Harness inside NewHarness.
type Option func(*Harness)

// WithProvider replaces the offline hash provider, e.g. with
// FailingProvider to exercise error paths.
func WithProvider(provider graphrag.Provider) Option {
	return func(harness *Harness) {
		harness.Provider = provider
	}
}

// WithBuildOptions sets the options passed to graphrag.Build.
func WithBuildOptions(options graphrag.BuildOptions) Option {
	return func(harness *Harness) {
		harness.buildOptions = options
	}
}

// WithSearcherOptions sets the kind policy of the harness Searcher.
func WithSearcherOptions(options graphrag.SearcherOptions) Option {
	return func(harness *Harness) {
		harness.searchPolicy = options
	}
}

// WithStore opens a real SQLite store in a per-test temp directory, wires it
// as the Build cache, and persists the graph on every Build so Reload specs
// can rehydrate from disk. The store closes automatically with the test.
func WithStore() Option {
	return func(harness *Harness) {
		dsn := filepath.Join(harness.TB.TempDir(), "graphrag-bdd.db")

		store, err := graphrag.OpenStore(dsn)
		if err != nil {
			harness.TB.Fatalf("bdd harness: open store %s: %v", dsn, err)
		}

		harness.TB.Cleanup(func() { _ = store.Close() })

		harness.Store = store
		harness.Cache = store
		harness.dsn = dsn
	}
}

// NewHarness prepares a harness over the given documents and edges: the
// deterministic offline hash provider, no cache, no store, neutral search
// policy. Call Build to produce the index; Setup failures are fatal to tb.
func NewHarness(
	tb TB,
	documents []graphrag.Document,
	edges []graphrag.Edge,
	options ...Option,
) *Harness {
	harness := &Harness{
		TB:           tb,
		Provider:     graphrag.NewHashProvider(),
		Cache:        nil,
		Store:        nil,
		Result:       nil,
		Searcher:     nil,
		documents:    documents,
		edges:        edges,
		dsn:          "",
		buildOptions: graphrag.BuildOptions{},    //nolint:exhaustruct_v5 // zero options select the SDK defaults
		searchPolicy: graphrag.SearcherOptions{}, //nolint:exhaustruct_v5 // the neutral policy IS the zero value
	}

	for _, option := range options {
		option(harness)
	}

	return harness
}

// Build runs graphrag.Build over the harness corpus, persists the graph when
// WithStore is used, and wires the Searcher. It can be called again (e.g. to
// assert cache-hit behavior); on error the previous index stays untouched.
func (h *Harness) Build(ctx context.Context) (*graphrag.BuildResult, error) {
	result, err := graphrag.Build(ctx, h.Provider, h.Cache, h.documents, h.edges, h.buildOptions)
	if err != nil {
		return result, fmt.Errorf("bdd harness: build: %w", err)
	}

	if h.Store != nil {
		if replaceErr := h.Store.ReplaceGraph( //nolint:contextcheck // core Store.ReplaceGraph takes no context parameter
			result.Nodes,
			result.EmbeddingHash,
			result.Edges,
		); replaceErr != nil {
			return result, fmt.Errorf("bdd harness: persist graph: %w", replaceErr)
		}
	}

	h.Result = result
	h.Searcher = graphrag.NewSearcherWithOptions(result.Nodes, result.Edges, result.Vectors, h.searchPolicy)

	return result, nil
}

// Search queries the built index through the harness provider. The optional
// first SearchOptions element overrides the zero options.
func (h *Harness) Search(
	ctx context.Context,
	query string,
	searchOptions ...graphrag.SearchOptions,
) (*graphrag.SearchResult, error) {
	if h.Searcher == nil {
		return nil, fmt.Errorf("bdd harness: search: %w", errNotBuilt)
	}

	var options graphrag.SearchOptions

	if len(searchOptions) > 0 {
		options = searchOptions[0]
	}

	result, err := h.Searcher.Search(ctx, h.Provider, query, options)
	if err != nil {
		return nil, fmt.Errorf("bdd harness: search: %w", err)
	}

	return result, nil
}

// SimilarPairs scans the built index for node pairs at least minScore apart.
func (h *Harness) SimilarPairs(minScore float64) []graphrag.SimilarPair {
	if h.Searcher == nil {
		h.TB.Fatalf("bdd harness: similar pairs: %v", errNotBuilt)
	}

	return h.Searcher.SimilarPairs(minScore)
}

// Reload reopens the store file built by WithStore, loads the persisted
// graph and cached embeddings, and returns a second harness whose Searcher
// answers from the reloaded snapshot. The original harness stays usable.
func (h *Harness) Reload() (*Harness, error) {
	if h.Store == nil {
		return nil, fmt.Errorf("bdd harness: reload: %w", errNoStore)
	}

	store, err := graphrag.OpenStore(h.dsn)
	if err != nil {
		return nil, fmt.Errorf("bdd harness: reopen store %s: %w", h.dsn, err)
	}

	h.TB.Cleanup(func() { _ = store.Close() })

	snapshot, err := store.LoadGraph()
	if err != nil {
		return nil, fmt.Errorf("bdd harness: load graph: %w", err)
	}

	vectorsByHash, err := store.LoadEmbeddings(h.Provider.Name(), h.Provider.Model())
	if err != nil {
		return nil, fmt.Errorf("bdd harness: load embeddings: %w", err)
	}

	reloaded := &Harness{
		TB:           h.TB,
		Provider:     h.Provider,
		Cache:        store,
		Store:        store,
		Result:       h.Result,
		Searcher:     nil,
		documents:    h.documents,
		edges:        h.edges,
		dsn:          h.dsn,
		buildOptions: h.buildOptions,
		searchPolicy: h.searchPolicy,
	}
	reloaded.Searcher = graphrag.NewSearcherWithOptions(
		snapshot.Nodes, snapshot.Edges, joinVectors(snapshot, vectorsByHash), h.searchPolicy)

	return reloaded, nil
}

// joinVectors maps hash-keyed cached vectors onto node IDs through the
// snapshot's node-to-hash linkage; nodes without a cached vector are skipped.
func joinVectors(
	snapshot *graphrag.LoadGraphSnapshot,
	vectorsByHash map[string]graphrag.Vector,
) map[string]graphrag.Vector {
	joined := make(map[string]graphrag.Vector, len(snapshot.Nodes))

	for _, node := range snapshot.Nodes {
		hash, ok := snapshot.EmbeddingHash[node.ID]
		if !ok {
			continue
		}

		if vector, cached := vectorsByHash[hash]; cached {
			joined[node.ID] = vector
		}
	}

	return joined
}
