package graphrag_test

import graphrag "github.com/larsartmann/go-graph-rag"

// Test-only fixture vocabulary. The SDK ships domain-neutral kind and
// relation TYPES; these constants are the vocabulary the tests graph with.
const (
	kindJob     = graphrag.NodeKind("job")
	kindCompany = graphrag.NodeKind("company")
	kindSkill   = graphrag.NodeKind("skill")
	kindProject = graphrag.NodeKind("project")
	kindCV      = graphrag.NodeKind("cv")

	relAt       = graphrag.Relation("at")
	relRequires = graphrag.Relation("requires")
	relPrefers  = graphrag.Relation("prefers")
	relUses     = graphrag.Relation("uses")
)
