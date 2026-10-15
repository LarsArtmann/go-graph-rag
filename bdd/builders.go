package bdd

import graphrag "github.com/larsartmann/go-graph-rag"

// DocOption tunes the Document built by Doc.
type DocOption func(*graphrag.Document)

// WithKind sets the node kind of the built document.
func WithKind(kind graphrag.NodeKind) DocOption {
	return func(doc *graphrag.Document) {
		doc.Kind = kind
	}
}

// WithLabel sets the display label of the built document; Doc defaults it to
// the document ID.
func WithLabel(label string) DocOption {
	return func(doc *graphrag.Document) {
		doc.Label = label
	}
}

// Doc builds one indexing document: id and text are required, the kind
// defaults to the kit's demo article kind, and the label defaults to the ID.
// Options override the defaults in order.
func Doc(id graphrag.NodeID, text string, options ...DocOption) graphrag.Document {
	document := graphrag.Document{
		ID:    id,
		Kind:  KindArticle,
		Label: string(id),
		Text:  text,
		Attrs: graphrag.NodeAttrs{}, //nolint:exhaustruct_v5 // demo docs carry no attrs; callers add their own
	}

	for _, option := range options {
		option(&document)
	}

	return document
}

// Edge builds one weighted graph edge with weight 1, the shape callers
// hand-roll in every consumer test suite.
func Edge(source graphrag.NodeID, target graphrag.NodeID, relation graphrag.Relation) graphrag.Edge {
	return graphrag.Edge{
		Source:   source,
		Target:   target,
		Relation: relation,
		Weight:   1,
	}
}

// WeightedEdge builds one graph edge with an explicit weight.
func WeightedEdge(
	source graphrag.NodeID,
	target graphrag.NodeID,
	relation graphrag.Relation,
	weight float64,
) graphrag.Edge {
	return graphrag.Edge{
		Source:   source,
		Target:   target,
		Relation: relation,
		Weight:   weight,
	}
}
