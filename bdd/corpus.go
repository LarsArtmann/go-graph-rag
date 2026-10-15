package bdd

import graphrag "github.com/larsartmann/go-graph-rag"

// The demo vocabulary is the kit's own neutral, domain-agnostic kind and
// relation set for DemoCorpus; consumers define their real vocabulary
// caller-side, exactly as with the core SDK.
const (
	// KindArticle marks demo documents that carry embeddable text.
	KindArticle graphrag.NodeKind = "article"

	// KindTopic marks demo hub nodes: short label texts meant for graph
	// expansion rather than primary ranking.
	KindTopic graphrag.NodeKind = "topic"

	// RelationCites links one demo article to another.
	RelationCites graphrag.Relation = "cites"

	// RelationMentions links a demo article to a topic hub.
	RelationMentions graphrag.Relation = "mentions"
)

// DemoSearcherOptions is the demo search policy: articles are the ranked
// documents, topics render as detail labels.
func DemoSearcherOptions() graphrag.SearcherOptions {
	return graphrag.SearcherOptions{
		ReferenceKind: "",
		DocumentKinds: []graphrag.NodeKind{KindArticle},
		DetailKind:    KindTopic,
	}
}

// DemoCorpus returns a small, deterministic, domain-neutral corpus whose
// behaviors are pinned by this kit's own spec suite:
//
//   - article:retrieval and article:retrieval-dup are near-duplicates and
//     always connect through a RelationSimilar edge at the default
//     threshold with the offline hash provider;
//   - the other two articles are lexically distinct and never produce
//     similar edges against each other or the retrieval pair;
//   - topic:search and topic:storage are hubs reached through
//     RelationMentions edges; article:cites chains the texts.
//
// Use it for smoke specs or as the starting point for a domain corpus.
func DemoCorpus() ([]graphrag.Document, []graphrag.Edge) {
	documents := []graphrag.Document{
		Doc("article:retrieval",
			"Hybrid search combines vector similarity with graph expansion to rank documents."),
		Doc("article:retrieval-dup",
			"Hybrid search combines vector similarity with graph expansion to rank related documents."),
		Doc("article:storage",
			"The embedding cache persists vectors in SQLite keyed by content hash."),
		Doc("article:chunking",
			"Documents are chunked before embedding so each vector covers one passage."),
		Doc("topic:search", "Search and retrieval", WithKind(KindTopic)),
		Doc("topic:storage", "Storage and caching", WithKind(KindTopic)),
	}

	edges := []graphrag.Edge{
		Edge("article:retrieval", "topic:search", RelationMentions),
		Edge("article:retrieval-dup", "topic:search", RelationMentions),
		Edge("article:storage", "topic:storage", RelationMentions),
		Edge("article:retrieval", "article:chunking", RelationCites),
		Edge("article:chunking", "article:storage", RelationCites),
	}

	return documents, edges
}
