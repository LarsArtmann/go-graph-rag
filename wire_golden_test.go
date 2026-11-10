package graphrag_test

// Wire-contract goldens: the JSON shapes of Node, Edge, Hit, and
// SearchResult are frozen public surface (ADR §3: "never delegated").
// These tests pin the exact bytes encoding/json/v2 produces for a
// fully-populated fixture — v2 is the canonical engine since the
// 2026-11-10 owner ruling ("go v2"); before that the same shapes were
// pinned under the v1 engine, which omits zero numerics (e.g. Score)
// where v2 writes them. Any drift — a renamed tag, a changed field, an
// engine difference — fails a test instead of a consumer.
//
// Regenerating: `go test -run WireGolden -update .` rewrites the golden
// files. A regeneration is a WIRE-CONTRACT EVENT: it needs a CHANGELOG
// entry under [Unreleased] and, once v1 is declared, a major-version bump.
// Never re-pin silently.

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	graphrag "github.com/larsartmann/go-graph-rag"
	"github.com/stretchr/testify/require"
)

var updateGoldens = flag.Bool("update", false, "rewrite the wire-contract golden files")

func TestWireGolden(t *testing.T) {
	t.Parallel()

	posted := time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC)

	// hubNode deliberately leaves every NodeAttrs field zero: under
	// encoding/json/v2 the numeric omitempty is a no-op, so Score and
	// friends serialize as explicit zeros — that presence is part of the
	// pinned contract (v1 omitted them; see the v0.5.0 CHANGELOG entry).
	hubNode := graphrag.Node{ID: "hub:go", Kind: "hub", Label: "Go"}

	docNode := graphrag.Node{
		ID:    "doc:42",
		Kind:  "document",
		Label: "Wire contract fixture",
		Attrs: graphrag.NodeAttrs{
			URL:          "https://example.com/fixture",
			Source:       "fixture",
			Location:     "Remote (EU)",
			RemotePolicy: "remote",
			Status:       "open",
			Score:        0.8125,
			PostedAt:     &posted,
		},
	}

	edge := graphrag.Edge{Source: "doc:42", Target: "doc:7", Relation: "similar_to", Weight: 0.8125}

	hit := graphrag.Hit{
		Node:     docNode,
		Score:    0.8125,
		RefMatch: 0.5,
		Related: []graphrag.RelatedNode{
			{Node: hubNode, Relation: "mentions", Weight: 1},
		},
	}

	searchResult := &graphrag.SearchResult{
		Query:       "go graph",
		Provider:    "hash",
		Model:       "v1",
		Hits:        []graphrag.Hit{hit},
		ContextText: "deterministic render",
	}

	cases := []struct {
		name  string
		value any
	}{
		{name: "node", value: docNode},
		{name: "node_zero_attrs", value: hubNode},
		{name: "edge", value: edge},
		{name: "hit", value: hit},
		{name: "search_result", value: searchResult},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			blob, err := json.Marshal(testCase.value, jsontext.WithIndentPrefix(""), jsontext.WithIndent("  "))
			require.NoError(t, err)

			blob = append(blob, '\n')

			golden := filepath.Join("testdata", "wire", testCase.name+".golden")
			if *updateGoldens {
				require.NoError(t, os.MkdirAll(filepath.Dir(golden), 0o755))
				require.NoError(t, os.WriteFile(golden, blob, 0o644))

				return
			}

			want, err := os.ReadFile(golden)
			require.NoError(t, err, "golden missing; generate once with `go test -run WireGolden -update .`")
			require.Equal(t, string(want), string(blob),
				"wire contract drifted; if INTENDED, re-pin with -update and add a CHANGELOG entry")
		})
	}
}
