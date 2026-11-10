# Upgrading

## To the json/v2 release (next release after v0.4.0)

The SDK now marshals with `encoding/json/v2` end to end (owner ruling
2026-11-10; on Go 1.27 the v1 package is a shim over the v2 engine).
Consumer-visible delta: zero numerics under `omitempty` serialize
explicitly — `NodeAttrs.Score` appears as `"score": 0` where the v1
engine omitted the key. Both engines decode to identical values, so
stored graphs, embedding caches, and JSON consumers survive untouched;
only exact-byte comparisons (checksums over serialized JSON) notice.

## To the `NodeID` release (next release after v0.3.2)

Graph identity moved from `string` to the defined type `graphrag.NodeID`, and
every ID-taking API follows: `Node.ID`, `Edge.Source`/`Edge.Target`,
`Document.ID`, `Graph.Node`/`OutEdges`/`InEdges`, `BuildResult.Vectors`,
`NewSearcher`/`NewSearcherWithOptions`, `Store.ReplaceGraph`, and the BDD
kit's builders and matchers.

**Most code needs no change.** Untyped string literals still assign directly:

```go
graphrag.Node{ID: "job:1"} // still compiles
```

Only identifiers held in `string` variables need an explicit conversion at
the call site — one line per call:

```go
id := "job:1" // string from your own code

node, ok := g.Node(graphrag.NodeID(id))
```

The JSON wire shape is byte-identical: node IDs encode and decode as plain
strings, so stored snapshots, embedding caches, and cross-service payloads
survive the upgrade untouched. Documents and nodes share one ID namespace: a
document's ID becomes its node's ID at index time.
