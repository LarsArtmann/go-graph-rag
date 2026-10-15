package graphrag_test

import (
	"os"
	"regexp"
	"testing"

	graphrag "github.com/larsartmann/go-graph-rag"
)

// changelogReleaseHeader matches a released CHANGELOG section header such
// as "## [0.3.2] - 2026-10-15". [Unreleased] deliberately does not match.
var changelogReleaseHeader = regexp.MustCompile(`(?m)^## \[(\d+)\.(\d+)\.(\d+)\]`)

// TestUserAgentVersionMatchesNewestRelease pins the release contract: the
// User-Agent stamp must always carry the major.minor of the newest released
// CHANGELOG section. Cutting a release means adding the section and bumping
// UserAgentVersion in the same change; this test fails when either half
// lands alone. It asserts consistency, never a hardcoded version value, so
// it survives every future release without edits.
func TestUserAgentVersionMatchesNewestRelease(t *testing.T) {
	data, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		t.Fatalf("read CHANGELOG.md: %v", err)
	}

	m := changelogReleaseHeader.FindStringSubmatch(string(data))

	if m == nil {
		t.Fatal("no released version header (## [x.y.z] - date) found in CHANGELOG.md")
	}

	newest := m[1] + "." + m[2]

	if graphrag.UserAgentVersion != newest {
		t.Errorf("UserAgentVersion %q does not match the newest released CHANGELOG version %q (header %q)"+
			" — bump both together when cutting a release",
			graphrag.UserAgentVersion, newest, m[0])
	}
}
