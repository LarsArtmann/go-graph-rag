package bdd

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	graphrag "github.com/larsartmann/go-graph-rag"
	"github.com/onsi/gomega/types"
)

// errWrongActual backs matcher type errors: every matcher accepts exactly
// one actual type and says so when handed another.
var errWrongActual = errors.New("bdd: matcher got an unexpected actual type")

// HaveTopHit matches a *graphrag.SearchResult whose best hit is the node
// nodeID: the primary "did retrieval rank the right thing first" assertion.
func HaveTopHit( //nolint:ireturn // matcher factories return the Gomega interface by design
	nodeID string,
) types.GomegaMatcher {
	return topHitMatcher{nodeID: nodeID}
}

type topHitMatcher struct {
	nodeID string
}

func (m topHitMatcher) Match(actual any) (bool, error) {
	result, err := asSearchResult(actual)
	if err != nil {
		return false, err
	}

	ids := hitIDs(result)

	return len(ids) > 0 && ids[0] == m.nodeID, nil
}

func (m topHitMatcher) FailureMessage(actual any) string {
	return describeHits(actual, fmt.Sprintf("top hit %q", m.nodeID))
}

func (m topHitMatcher) NegatedFailureMessage(actual any) string {
	return describeHits(actual, fmt.Sprintf("any top hit except %q", m.nodeID))
}

// ContainHit matches a *graphrag.SearchResult with at least one hit for
// nodeID, at any rank.
func ContainHit( //nolint:ireturn // matcher factories return the Gomega interface by design
	nodeID string,
) types.GomegaMatcher {
	return containHitMatcher{nodeID: nodeID}
}

type containHitMatcher struct {
	nodeID string
}

func (m containHitMatcher) Match(actual any) (bool, error) {
	result, err := asSearchResult(actual)
	if err != nil {
		return false, err
	}

	return slices.Contains(hitIDs(result), m.nodeID), nil
}

func (m containHitMatcher) FailureMessage(actual any) string {
	return describeHits(actual, fmt.Sprintf("a hit for %q at any rank", m.nodeID))
}

func (m containHitMatcher) NegatedFailureMessage(actual any) string {
	return describeHits(actual, fmt.Sprintf("no hit for %q", m.nodeID))
}

// HaveHitIDs matches a *graphrag.SearchResult whose hits are exactly the
// nodeIDs, in order: the pinned retrieval-inventory assertion.
func HaveHitIDs( //nolint:ireturn // matcher factories return the Gomega interface by design
	nodeIDs ...string,
) types.GomegaMatcher {
	return hitIDsMatcher{nodeIDs: nodeIDs}
}

type hitIDsMatcher struct {
	nodeIDs []string
}

func (m hitIDsMatcher) Match(actual any) (bool, error) {
	result, err := asSearchResult(actual)
	if err != nil {
		return false, err
	}

	return slices.Equal(hitIDs(result), m.nodeIDs), nil
}

func (m hitIDsMatcher) FailureMessage(actual any) string {
	return describeHits(actual, fmt.Sprintf("hits %v in rank order", m.nodeIDs))
}

func (m hitIDsMatcher) NegatedFailureMessage(actual any) string {
	return describeHits(actual, fmt.Sprintf("any hit list except %v", m.nodeIDs))
}

// HaveNoHitsOfKind matches a *graphrag.SearchResult with zero hits of the
// given kind: the hub-exclusion assertion for SearcherOptions policies.
func HaveNoHitsOfKind( //nolint:ireturn // matcher factories return the Gomega interface by design
	kind graphrag.NodeKind,
) types.GomegaMatcher {
	return noHitsOfKindMatcher{kind: kind}
}

type noHitsOfKindMatcher struct {
	kind graphrag.NodeKind
}

func (m noHitsOfKindMatcher) Match(actual any) (bool, error) {
	result, err := asSearchResult(actual)
	if err != nil {
		return false, err
	}

	for _, hit := range result.Hits {
		if hit.Node.Kind == m.kind {
			return false, nil
		}
	}

	return true, nil
}

func (m noHitsOfKindMatcher) FailureMessage(actual any) string {
	return describeHits(actual, fmt.Sprintf("no hits of kind %q", m.kind))
}

func (m noHitsOfKindMatcher) NegatedFailureMessage(actual any) string {
	return describeHits(actual, fmt.Sprintf("at least one hit of kind %q", m.kind))
}

// HaveSimilarEdge matches a []graphrag.SimilarPair scan (as returned by
// Searcher.SimilarPairs or Harness.SimilarPairs) that contains the pair
// aID/bID in either direction. Pair SimilarEdgeScore with a numerical
// assertion when the score itself matters.
func HaveSimilarEdge( //nolint:ireturn // matcher factories return the Gomega interface by design
	aID string,
	bID string,
) types.GomegaMatcher {
	return similarEdgeMatcher{aID: aID, bID: bID}
}

type similarEdgeMatcher struct {
	aID string
	bID string
}

func (m similarEdgeMatcher) Match(actual any) (bool, error) {
	pairs, err := asSimilarPairs(actual)
	if err != nil {
		return false, err
	}

	_, ok := SimilarEdgeScore(pairs, m.aID, m.bID)

	return ok, nil
}

func (m similarEdgeMatcher) FailureMessage(actual any) string {
	return describePairs(actual, fmt.Sprintf("a similar edge between %q and %q", m.aID, m.bID))
}

func (m similarEdgeMatcher) NegatedFailureMessage(actual any) string {
	return describePairs(actual, fmt.Sprintf("no similar edge between %q and %q", m.aID, m.bID))
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

// asSearchResult narrows a matcher actual to a search result.
func asSearchResult(actual any) (*graphrag.SearchResult, error) {
	result, ok := actual.(*graphrag.SearchResult)
	if !ok {
		return nil, fmt.Errorf("%w: want *graphrag.SearchResult, got %T", errWrongActual, actual)
	}

	return result, nil
}

// asSimilarPairs narrows a matcher actual to a similar-pairs scan.
func asSimilarPairs(actual any) ([]graphrag.SimilarPair, error) {
	pairs, ok := actual.([]graphrag.SimilarPair)
	if !ok {
		return nil, fmt.Errorf("%w: want []graphrag.SimilarPair, got %T", errWrongActual, actual)
	}

	return pairs, nil
}

// hitIDs renders the hit node IDs of a search result in rank order.
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

// describeHits renders matcher failure context: the expectation plus the
// actual ranked hits with kinds and scores.
func describeHits(actual any, expectation string) string {
	result, ok := actual.(*graphrag.SearchResult)
	if !ok {
		return fmt.Sprintf("Expected %s, but the actual value is a %T, not a search result", expectation, actual)
	}

	if result == nil {
		return fmt.Sprintf("Expected %s, but the search result is nil", expectation)
	}

	parts := make([]string, 0, len(result.Hits))

	for _, hit := range result.Hits {
		parts = append(parts, fmt.Sprintf("%s(%s)=%.3f", hit.Node.ID, hit.Node.Kind, hit.Score))
	}

	return fmt.Sprintf("Expected %s, got hits [%s]", expectation, strings.Join(parts, ", "))
}

// describePairs renders matcher failure context over a similar-pairs scan.
func describePairs(actual any, expectation string) string {
	pairs, ok := actual.([]graphrag.SimilarPair)
	if !ok {
		return fmt.Sprintf("Expected %s, but the actual value is a %T, not a similar-pairs scan", expectation, actual)
	}

	parts := make([]string, 0, len(pairs))

	for _, pair := range pairs {
		parts = append(parts, fmt.Sprintf("%s~%s=%.3f", pair.A.ID, pair.B.ID, pair.Score))
	}

	return fmt.Sprintf("Expected %s, got pairs [%s]", expectation, strings.Join(parts, ", "))
}
