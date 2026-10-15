package bdd

import (
	"context"
	"errors"

	graphrag "github.com/larsartmann/go-graph-rag"
)

// errNoFailureConfigured backs FailingProvider.Embed when the zero value is
// used, so specs always observe a non-nil error.
var errNoFailureConfigured = errors.New("bdd: FailingProvider has no error configured")

// FailingProvider is a graphrag.Provider whose Embed always fails. Use it in
// specs that exercise provider failure propagation (Build and Search error
// paths). The zero value fails with errNoFailureConfigured.
type FailingProvider struct {
	// Err is returned by Embed; nil selects a generic failure.
	Err error
}

// Name identifies the provider in cache namespaces as "failing".
func (p FailingProvider) Name() string {
	return "failing"
}

// Model identifies the provider's vector-space version as "v1".
func (p FailingProvider) Model() string {
	return "v1"
}

// Embed returns the configured error instead of vectors.
func (p FailingProvider) Embed(_ context.Context, _ []string) ([]graphrag.Vector, error) {
	if p.Err != nil {
		return nil, p.Err
	}

	return nil, errNoFailureConfigured
}
