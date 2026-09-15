// Package graphrag provides semantic retrieval primitives: embedding
// providers (an OpenAI-compatible HTTP client and a deterministic offline
// hasher), a typed knowledge graph, a SQLite-backed embedding cache and
// graph store, and hybrid retrieval that blends vector similarity with
// graph expansion (GraphRAG).
//
// The package is deliberately dependency-light: everything except the SQLite
// store is stdlib-only, and every persisted artifact is derived data that can
// be rebuilt from source data at any time.
package graphrag

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"slices"
)

// float32Bytes is the encoded size of one vector dimension.
const float32Bytes = 4

// ErrVectorDecode is returned when a stored vector blob is truncated or its
// length is not a multiple of four bytes.
var ErrVectorDecode = errors.New("graphrag: vector blob is truncated or misaligned")

// Vector is a dense embedding vector of float32 components.
type Vector []float32

// Dims returns the number of dimensions of the vector.
func (v Vector) Dims() int {
	return len(v)
}

// Normalize returns the L2-normalized copy of v. The zero vector is returned
// unchanged because it has no direction to preserve.
func (v Vector) Normalize() Vector {
	var sumSquares float64

	for _, component := range v {
		sumSquares += float64(component) * float64(component)
	}

	if sumSquares == 0 {
		return slices.Clone(v)
	}

	norm := math.Sqrt(sumSquares)
	// The length is len(v) which is non-zero here, but makezero cannot prove
	// that through the zero-magnitude early return above.
	normalized := make(Vector, len(v)) //nolint:makezero // length mirrors v, non-zero by construction

	for i, component := range v {
		normalized[i] = float32(float64(component) / norm)
	}

	return normalized
}

// Encode serializes the vector as a little-endian float32 blob, the on-disk
// format of the SQLite vector cache.
func (v Vector) Encode() []byte {
	// Length is derived from v and therefore non-zero whenever there is
	// anything to encode; makezero cannot see that through len(v).
	blob := make([]byte, len(v)*float32Bytes) //nolint:makezero // length mirrors v, non-zero by construction

	for i, component := range v {
		binary.LittleEndian.PutUint32(blob[i*float32Bytes:], math.Float32bits(component))
	}

	return blob
}

// DecodeVector deserializes a little-endian float32 blob produced by Encode.
func DecodeVector(blob []byte) (Vector, error) {
	if len(blob)%float32Bytes != 0 {
		return nil, fmt.Errorf("%w: %d bytes", ErrVectorDecode, len(blob))
	}

	// Length is blob/4, non-zero for any decodable input (0-length blobs
	// still round-trip as empty vectors).
	vector := make(Vector, len(blob)/float32Bytes) //nolint:makezero // length mirrors blob, non-zero by construction

	for i := range vector {
		bits := binary.LittleEndian.Uint32(blob[i*float32Bytes:])
		vector[i] = math.Float32frombits(bits)
	}

	return vector, nil
}

// Cosine returns the cosine similarity between a and b. Vectors with
// mismatched dimensions or zero magnitude yield 0 so that malformed inputs
// degrade to "no signal" instead of poisoning rankings with NaNs.
func Cosine(a, b Vector) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}

	var dot, normA, normB float64

	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// HashText returns the hex-encoded SHA-256 of text. It is the cache key for
// embeddings: identical text never re-pays the embedding provider.
func HashText(text string) string {
	sum := sha256.Sum256([]byte(text))

	return hex.EncodeToString(sum[:])
}
