# Research index

Point-in-time adoption/decision research. These files are historical
snapshots: never rewrite them — annotate inline or archive. Re-run an
evaluation on the revisit triggers inside each artifact, not on a hunch.

| Artifact                                           | Date       | Question                                                                 | Verdict                                                                                                                  | Reopen when          |
| -------------------------------------------------- | ---------- | ------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------ | -------------------- |
| `2026-09-15_dgraph-adoption.md`                    | 2026-09-15 | Adopt Dgraph as store, backend, or at all?                               | Do not adopt into the SDK; revisit as a consumer-owned backend at the ~50k-node trigger (scenario B/C)                   | §7 decision triggers |
| `2026-09-15_metaengine-system-adoption.md`         | 2026-09-15 | Adopt `go-cqrs-lite/metaengine` and/or `system`?                         | Adopt neither into the SDK; CV app-layer adoption per SUPERB T01/T29                                                     | §9 triggers          |
| `2026-09-19_metaengine-graph-vector-comparison.md` | 2026-09-19 | How do this SDK's capabilities relate to metaengine's graph/vector ones? | Complementary layers: storage ADTs vs GraphRAG vertical; overlap point superseded by this SDK; no adoption change        | §7 triggers          |
| `2026-09-19_metaengine-go-graph-rag-merge.md`      | 2026-09-19 | Smart-merge the two (move logic into the SDK; metaengine depends on it)? | No merge in either direction: ~80 net-shared lines that are deliberately divergent and test-pinned; leaf module deferred | §9 triggers          |
