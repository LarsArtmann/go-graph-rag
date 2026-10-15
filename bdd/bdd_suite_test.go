package bdd_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBDDBehaviors(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "go-graph-rag BDD Kit Suite")
}
