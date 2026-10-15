package bdd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/larsartmann/go-graph-rag"
	"github.com/onsi/gomega/gcustom"
	"github.com/onsi/gomega/types"
)

// HaveTopHit matches a *graphrag.SearchResult whose best hit is the node
// nodeID: the primary "did retrieval rank the right thing first" assertion.
func HaveTopHit(nodeID string) types.GomegaMatcher { //nolint:ireturn // matcher factories return the Gomega interface by design
	return gcustom.MakeMatcher(func(actual *graphrag.SearchResult) (bool, error) {
		ids := hitIDs(actual)

		return len(ids) > 0 && ids[0] == nodeID, nil
	}).
		WithFailMessage(func(actual *graphrag.SearchResult) string {
			return describeHits(actual, fmt.Sprintf("top hit to be %q", nodeID))
		}).
		WithNegatedFailureMessage(func(actual *graphrag.SearchResult) string {
			return describeHits(actual, fmt.Sprintf("top hit NOT to be %q", nodeID))
		})
}

// ContainHit matches a *graphrag.SearchResult with at least one hit for
// nodeID, at any rank.
func ContainHit(nodeID string) types.GomegaMatcher { //nolint:ireturn // matcher factories return the Gomega interface by design
	return gcustom.MakeMatcher(func(actual *graphrag.SearchResult) (bool, error) {
		return slices.Contains(hitIDs(actual), nodeID), nil
	}).
		WithFailMessage(func(actual *graphrag.SearchResult) string {
			return describeHits(actual, fmt.Sprintf("a hit for %q", nodeID))
		}).
		WithNegatedFailureMessage(func(actual *graphrag.SearchResult) string {
			return describeHits(actual, fmt.Sprintf("no hit for %q", nodeID))
		})
}

// HaveHitIDs matches a *graphrag.SearchResult whose hits are exactly the
// nodeIDs, in order: the pinned retrieval-inventory assertion.
func HaveHitIDs(nodeIDs ...string) types.GomegaMatcher { //nolint:ireturn // matcher factories return the Gomega interface by design
	return gcustom.MakeMatcher(func(actual *graphrag.SearchResult) (bool, error) {
		return slices.Equal(hitIDs(actual), nodeIDs), nil
	}).
		WithFailMessage(func(actual *graphrag.SearchResult) string {
			return describeHits(actual, fmt.Sprintf("hits %v in order", nodeIDs))
		}).
		WithNegatedFailureMessage(func(actual *graphrag.SearchResult) string {
			return describeHits(actual, fmt.Sprintf("hits other than %v", nodeIDs))
		})
}

// HaveNoHitsOfKind matches a *graphrag.SearchResult with zero hits of the
// given kind: the hub-exclusion assertion for SearcherOptions policies.
func HaveNoHitsOfKind(kind graphrag.NodeKind) types.GomegaMatcher { //nolint:ireturn // matcher factories return the Gomega interface by design
	return gcustom.MakeMatcher(func(actual *graphrag.SearchResult) (bool, error) {
		for _, hit := range actual.Hits {
			if hit.Node.Kind == kind {
				return false, nil
			}
		}

		return true, nil
	}).
		WithFailMessage(func(actual *graphrag.SearchResult) string {
			return describeHits(actual, fmt.Sprintf("no hits of kind %q", kind))
		}).
		WithNegatedFailureMessage(func(actual *graphrag.SearchResult) string {
			return describeHits(actual, fmt.Sprintf("at least one hit of kind %q", kind))
		})
}

// HaveSimilarEdge matches a []graphrag.SimilarPair scan (as returned by
// Searcher.SimilarPairs or Harness.SimilarPairs) that contains the pair
// aID/bID in either direction. Pair SimilarEdgeScore with a numerical
// assertion when the score itself matters.
func HaveSimilarEdge(aID string, bID string) types.GomegaMatcher { //nolint:ireturn // matcher factories return the Gomega interface by design
	return gcustom.MakeMatcher(func(actual []graphrag.SimilarPair) (bool, error) {
		_, ok := SimilarEdgeScore(actual, aID, bID)

		return ok, nil
	}).
		WithFailMessage(func(actual []graphrag.SimilarPair) string {
			return describePairs(actual, fmt.Sprintf("a similar edge between %q and %q", aID, bID))
		}).
		WithNegatedFailureMessage(func(actual []graphrag.SimilarPair) string {
			return describePairs(actual, fmt.Sprintf("no similar edge between %q and %q", aID, bID))
		})
}

// SimilarEdgeScore returns the score of the aID/bID pair from a SimilarPairs
// scan, in either direction, and whether the pair was present at all.
func SimilarEdgeScore(pairs []graphrag.SimilarPair, aID string, bID string) (float64, bool) {
	for _, pair := range pairs {
		if (pair.A.ID == aID && pair.B.ID == bID) || (pair.A.ID == bID && pair.B.ID == aID) {
			return pair.Score, true
		}
	}

	return 0, false
}

// hitIDs renders the hit node IDs of a search result in rank order; an empty
// result yields a nil slice so comparisons stay exact.
func hitIDs(result *graphrag.SearchResult) []string {
	if result == nil {
		return nil
	}

	ids := make([]string, 0, len(result.Hits))

	for _, hit := range result.Hits {
		ids = append(ids, hit.Node.ID)
	}

	return ids
}

// describeHits renders matcher failure context: expectation plus the actual
// ranked hits with kinds and scores.
func describeHits(result *graphrag.SearchResult, expectation string) string {
	if result == nil {
		return "Expected a search result, got nil (wanted " + expectation + ")"
	}

	parts := make([]string, 0, len(result.Hits))

	for _, hit := range result.Hits {
		parts = append(parts, fmt.Sprintf("%s(%s)=%.3f", hit.Node.ID, hit.Node.Kind, hit.Score))
	}

	return fmt.Sprintf("Expected %s, got hits [%s]", expectation, strings.Join(parts, ", "))
}

// describePairs renders matcher failure context over a similar-pairs scan.
func describePairs(pairs []graphrag.SimilarPair, expectation string) string {
	parts := make([]string, 0, len(pairs))

	for _, pair := range pairs {
		parts = append(parts, fmt.Sprintf("%s~%s=%.3f", pair.A.ID, pair.B.ID, pair.Score))
	}

	return fmt.Sprintf("Expected %s, got pairs [%s]", expectation, strings.Join(parts, ", "))
}
