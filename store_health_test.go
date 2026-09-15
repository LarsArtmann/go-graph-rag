package graphrag

import (
	"context"
	"testing"
)

// The DI scope sweep (go-health dashboard) asserts the registered *Store
// against the structural HealthcheckerWithContext duck type; this pins the
// mechanism that makes the dashboard row fail-capable.
func TestStoreHealthCheck(t *testing.T) {
	t.Parallel()

	store, err := OpenStore(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	defer func() { _ = store.Close() }()

	checker := interface {
		HealthCheck(ctx context.Context) error
	}(store)

	if err := checker.HealthCheck(context.Background()); err != nil {
		t.Fatalf("open store must be healthy, got %v", err)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}

	if err := checker.HealthCheck(context.Background()); err == nil {
		t.Fatal("a closed store must fail its health check, not report healthy")
	}
}
