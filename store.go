package graphrag

import (
	"context"
	"database/sql"
	"encoding/json/v2"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite" // SQLite driver registration
)

// storeOperationTimeout bounds every store statement. The store is an
// auxiliary, rebuildable index; no operation may hang a caller indefinitely.
const storeOperationTimeout = 5 * time.Second

// storeSchemaSQL creates the embedding cache and graph tables.
const storeSchemaSQL = `
PRAGMA busy_timeout = 5000;
PRAGMA synchronous = OFF;
CREATE TABLE IF NOT EXISTS graphrag_embeddings (
    content_hash TEXT    NOT NULL,
    provider     TEXT    NOT NULL,
    model        TEXT    NOT NULL,
    dims         INTEGER NOT NULL,
    vector       BLOB    NOT NULL,
    created_at   TEXT    NOT NULL,
    PRIMARY KEY (content_hash, provider, model)
);
CREATE TABLE IF NOT EXISTS graphrag_nodes (
    id             TEXT PRIMARY KEY,
    kind           TEXT NOT NULL,
    label          TEXT NOT NULL,
    attrs          TEXT NOT NULL,
    embedding_hash TEXT NOT NULL DEFAULT '',
    updated_at     TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS graphrag_edges (
    source   TEXT    NOT NULL,
    target   TEXT    NOT NULL,
    relation TEXT    NOT NULL,
    weight   REAL    NOT NULL DEFAULT 1.0,
    PRIMARY KEY (source, target, relation)
);
CREATE INDEX IF NOT EXISTS idx_graphrag_edges_target
    ON graphrag_edges(target);`

// ErrStoreClosed is returned when the store is used after Close.
var ErrStoreClosed = errors.New("graphrag: store is closed")

// EmbeddingEntry couples a content hash with its vector for cache writes.
type EmbeddingEntry struct {
	Hash   string
	Vector Vector
}

// Cache is the persistence seam for the embedding cache. Store implements
// it, and Build accepts it so tests can supply in-memory fakes.
type Cache interface {
	// CachedEmbedding returns the vector for a previously embedded text.
	CachedEmbedding(provider, model, contentHash string) (Vector, bool, error)

	// PutEmbeddings persists vectors under the provider/model namespace.
	PutEmbeddings(provider, model string, entries []EmbeddingEntry) error
}

// Store persists the embedding cache and the knowledge graph in SQLite. It
// is a derived, rebuildable index over the event store: losing it costs a
// rebuild (embeddings are cached by content hash, so re-indexing re-pays
// nothing for unchanged texts). One writer at a time is expected (CLI or
// server); readers serialize through the single connection.
type Store struct {
	db     *sql.DB
	dsn    string
	closed bool
}

// OpenStore opens (or creates) the SQLite database at dsn and migrates the
// schema. Use ":memory:" for a throwaway store.
func OpenStore(dsn string) (*Store, error) {
	// accepted duplication: this open/max-conns/schema-exec prologue is ~10
	// lines; extracting it would force a sqlite-carrying shared dependency,
	// against the dependency-minimal design.
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("graphrag: open sqlite %s: %w", dsn, err)
	}

	// PRAGMAs in storeSchemaSQL only stick on the connection that executes
	// them, so the pool is pinned to a single connection.
	db.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), storeOperationTimeout)
	defer cancel()

	if _, execErr := db.ExecContext(ctx, storeSchemaSQL); execErr != nil {
		go func() { _ = db.Close() }()

		return nil, fmt.Errorf("graphrag: migrate sqlite store: %w", execErr)
	}

	return &Store{db: db, dsn: dsn, closed: false}, nil
}

// Close closes the database.
func (s *Store) Close() error {
	if s.closed {
		return nil
	}

	s.closed = true

	if err := s.db.Close(); err != nil {
		return fmt.Errorf("graphrag: close sqlite store: %w", err)
	}

	return nil
}

// Shutdown satisfies the samber/do lifecycle duck type.
func (s *Store) Shutdown(_ context.Context) error {
	return s.Close()
}

// HealthCheck satisfies the samber/do HealthcheckerWithContext duck type:
// it pings the EXISTING pool (never a fresh connection), so a health
// dashboard row backed by it is a real, fail-capable check instead of
// green-by-default. A closed store reports unhealthy instead of silently
// passing.
func (s *Store) HealthCheck(ctx context.Context) error {
	if s.closed {
		return ErrStoreClosed
	}

	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("graphrag: store ping: %w", err)
	}

	return nil
}

// DSN reports the data source name the store was opened with.
func (s *Store) DSN() string {
	return s.dsn
}

// storeContext returns the per-operation bounded context. Every store
// method opens this bounded context and wraps errors the same way — that
// IS the store's documented discipline; extracting further would obscure
// per-method error context (accepted duplication by design).
func (s *Store) storeContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), storeOperationTimeout)
}

// CachedEmbedding looks up a previously stored vector.
func (s *Store) CachedEmbedding(provider, model, contentHash string) (Vector, bool, error) {
	ctx, cancel := s.storeContext()
	defer cancel()

	var blob []byte

	err := s.db.QueryRowContext(ctx,
		`SELECT vector FROM graphrag_embeddings WHERE content_hash = ? AND provider = ? AND model = ?`,
		contentHash, provider, model,
	).Scan(&blob)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, fmt.Errorf("graphrag: query embedding cache: %w", err)
	}

	vector, err := DecodeVector(blob)
	if err != nil {
		return nil, false, fmt.Errorf("graphrag: decode cached vector: %w", err)
	}

	return vector, true, nil
}

// statement opens a bounded statement context (the store's documented
// discipline: every method gets a timeout, caller cancellation is
// deliberately not threaded) and returns a commit step that commits on
// success and rolls back if invoked with a non-nil error.
// accepted duplication: the begin/commit envelope is shared by PutEmbeddings
// and ReplaceGraph; statement bodies and labels differ entirely.
func (s *Store) statement(errPrefix string) (context.Context, *sql.Tx, func(error) error, error) {
	ctx, cancel := s.storeContext()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		cancel()

		return nil, nil, nil, fmt.Errorf("graphrag: begin %s: %w", errPrefix, err)
	}

	return ctx, tx, func(stepErr error) error {
		defer cancel()

		if stepErr != nil {
			_ = tx.Rollback()

			return stepErr
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("graphrag: commit %s: %w", errPrefix, err)
		}

		return nil
	}, nil
}

// PutEmbeddings persists vectors in one transaction.
func (s *Store) PutEmbeddings(provider, model string, entries []EmbeddingEntry) error {
	if len(entries) == 0 {
		return nil
	}

	ctx, tx, finish, err := s.statement("embedding cache tx")
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx,
		`INSERT OR REPLACE INTO graphrag_embeddings (content_hash, provider, model, dims, vector, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return finish(fmt.Errorf("graphrag: prepare embedding insert: %w", err))
	}

	defer func() { _ = stmt.Close() }()

	now := time.Now().UTC().Format(time.RFC3339)

	for _, entry := range entries {
		if _, execErr := stmt.ExecContext(ctx,
			entry.Hash, provider, model, entry.Vector.Dims(), entry.Vector.Encode(), now,
		); execErr != nil {
			return finish(fmt.Errorf("graphrag: insert embedding: %w", execErr))
		}
	}

	return finish(nil)
}

// LoadGraphSnapshot is the full graph read model: nodes, edges, and the
// node-to-embedding-hash linkage needed to join vectors onto nodes.
type LoadGraphSnapshot struct {
	Nodes         []Node
	Edges         []Edge
	EmbeddingHash map[string]string
}

// ReplaceGraph atomically swaps the persisted graph. Rebuilds are full
// replaces: the graph is derived data, and delete-then-insert inside one
// transaction keeps the stored view consistent.
func (s *Store) ReplaceGraph(nodes []Node, embeddingHashes map[string]string, edges []Edge) error {
	ctx, tx, finish, err := s.statement("graph tx")
	if err != nil {
		return err
	}

	for _, statement := range []string{`DELETE FROM graphrag_nodes`, `DELETE FROM graphrag_edges`} {
		if _, execErr := tx.ExecContext(ctx, statement); execErr != nil {
			return finish(fmt.Errorf("graphrag: clear graph: %w", execErr))
		}
	}

	nodeStmt, err := tx.PrepareContext(
		ctx,
		`INSERT OR REPLACE INTO graphrag_nodes (id, kind, label, attrs, embedding_hash, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return finish(fmt.Errorf("graphrag: prepare node insert: %w", err))
	}

	defer func() { _ = nodeStmt.Close() }()

	now := time.Now().UTC().Format(time.RFC3339)

	for _, node := range nodes {
		encodedAttrs, marshalErr := json.Marshal(node.Attrs)
		if marshalErr != nil {
			return finish(fmt.Errorf("graphrag: marshal attrs for %s: %w", node.ID, marshalErr))
		}

		hash := embeddingHashes[node.ID]

		if _, execErr := nodeStmt.ExecContext(ctx,
			node.ID, string(node.Kind), node.Label, string(encodedAttrs), hash, now,
		); execErr != nil {
			return finish(fmt.Errorf("graphrag: insert node %s: %w", node.ID, execErr))
		}
	}

	edgeStmt, err := tx.PrepareContext(ctx,
		`INSERT OR REPLACE INTO graphrag_edges (source, target, relation, weight) VALUES (?, ?, ?, ?)`,
	)
	if err != nil {
		return finish(fmt.Errorf("graphrag: prepare edge insert: %w", err))
	}

	defer func() { _ = edgeStmt.Close() }()

	for _, edge := range edges {
		if _, execErr := edgeStmt.ExecContext(ctx,
			edge.Source, edge.Target, string(edge.Relation), edge.Weight,
		); execErr != nil {
			return finish(fmt.Errorf("graphrag: insert edge %s->%s: %w", edge.Source, edge.Target, execErr))
		}
	}

	return finish(nil)
}

// LoadGraph reads nodes, edges, and embedding hashes back.
func (s *Store) LoadGraph() (*LoadGraphSnapshot, error) {
	ctx, cancel := s.storeContext()
	defer cancel()

	nodes, hashes, err := s.loadNodes(ctx)
	if err != nil {
		return nil, err
	}

	edges, err := s.loadEdges(ctx)
	if err != nil {
		return nil, err
	}

	return &LoadGraphSnapshot{Nodes: nodes, Edges: edges, EmbeddingHash: hashes}, nil
}

// loadNodes reads all nodes plus their embedding hashes.
func (s *Store) loadNodes(ctx context.Context) ([]Node, map[string]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, kind, label, attrs, embedding_hash FROM graphrag_nodes ORDER BY id`,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("graphrag: query nodes: %w", err)
	}

	defer func() { _ = rows.Close() }()

	nodes := make([]Node, 0)
	hashes := make(map[string]string)

	for rows.Next() {
		var (
			node            Node
			attrsJSON, hash string
		)

		if scanErr := rows.Scan(&node.ID, &node.Kind, &node.Label, &attrsJSON, &hash); scanErr != nil {
			return nil, nil, fmt.Errorf("graphrag: scan node: %w", scanErr)
		}

		if unmarshalErr := json.Unmarshal([]byte(attrsJSON), &node.Attrs); unmarshalErr != nil {
			return nil, nil, fmt.Errorf("graphrag: decode attrs for %s: %w", node.ID, unmarshalErr)
		}

		nodes = append(nodes, node)
		hashes[node.ID] = hash
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, nil, fmt.Errorf("graphrag: iterate nodes: %w", rowsErr)
	}

	return nodes, hashes, nil
}

// loadEdges reads all edges.
func (s *Store) loadEdges(ctx context.Context) ([]Edge, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT source, target, relation, weight FROM graphrag_edges ORDER BY source, relation, target`,
	)
	if err != nil {
		return nil, fmt.Errorf("graphrag: query edges: %w", err)
	}

	defer func() { _ = rows.Close() }()

	edges := make([]Edge, 0)

	for rows.Next() {
		var edge Edge

		if scanErr := rows.Scan(&edge.Source, &edge.Target, &edge.Relation, &edge.Weight); scanErr != nil {
			return nil, fmt.Errorf("graphrag: scan edge: %w", scanErr)
		}

		edges = append(edges, edge)
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("graphrag: iterate edges: %w", rowsErr)
	}

	return edges, nil
}

// LoadEmbeddings returns every cached vector of one provider/model namespace
// keyed by content hash, ready to be joined onto nodes via their hashes.
func (s *Store) LoadEmbeddings(provider, model string) (map[string]Vector, error) {
	ctx, cancel := s.storeContext()
	defer cancel()

	rows, err := s.db.QueryContext(ctx,
		`SELECT content_hash, vector FROM graphrag_embeddings WHERE provider = ? AND model = ?`,
		provider, model,
	)
	if err != nil {
		return nil, fmt.Errorf("graphrag: query embeddings: %w", err)
	}

	defer func() { _ = rows.Close() }()

	vectors := make(map[string]Vector)

	for rows.Next() {
		var (
			hash string
			blob []byte
		)

		if scanErr := rows.Scan(&hash, &blob); scanErr != nil {
			return nil, fmt.Errorf("graphrag: scan embedding: %w", scanErr)
		}

		vector, decodeErr := DecodeVector(blob)
		if decodeErr != nil {
			return nil, fmt.Errorf("graphrag: decode embedding %s: %w", hash, decodeErr)
		}

		vectors[hash] = vector
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("graphrag: iterate embeddings: %w", rowsErr)
	}

	return vectors, nil
}

// StoreStats summarizes the persisted index.
type StoreStats struct {
	Nodes      int            `json:"nodes"`
	Edges      int            `json:"edges"`
	Embeddings int            `json:"embeddings"`
	ByKind     map[string]int `json:"byKind"`
}

// Stats reads index statistics for humans and endpoints.
func (s *Store) Stats() (StoreStats, error) {
	ctx, cancel := s.storeContext()
	defer cancel()

	stats := StoreStats{Nodes: 0, Edges: 0, Embeddings: 0, ByKind: make(map[string]int)}

	counters := []struct {
		query string
		dest  *int
	}{
		{`SELECT COUNT(*) FROM graphrag_nodes`, &stats.Nodes},
		{`SELECT COUNT(*) FROM graphrag_edges`, &stats.Edges},
		{`SELECT COUNT(*) FROM graphrag_embeddings`, &stats.Embeddings},
	}

	for _, counter := range counters {
		if err := s.db.QueryRowContext(ctx, counter.query).Scan(counter.dest); err != nil {
			return StoreStats{}, fmt.Errorf("graphrag: stats count: %w", err)
		}
	}

	rows, err := s.db.QueryContext(ctx, `SELECT kind, COUNT(*) FROM graphrag_nodes GROUP BY kind ORDER BY kind`)
	if err != nil {
		return StoreStats{}, fmt.Errorf("graphrag: stats by kind: %w", err)
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var (
			kind  string
			count int
		)

		if scanErr := rows.Scan(&kind, &count); scanErr != nil {
			return StoreStats{}, fmt.Errorf("graphrag: scan kind count: %w", scanErr)
		}

		stats.ByKind[kind] = count
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return StoreStats{}, fmt.Errorf("graphrag: iterate kind counts: %w", rowsErr)
	}

	return stats, nil
}
