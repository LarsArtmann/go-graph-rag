# Metaengine / System Adoption Re-Evaluation: post-v4.9.0 decision-matrix re-run

|          |                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Date     | 2026-11-10                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| Question | The 2026-09-15 evaluation adopted NEITHER `metaengine` nor `system` and fed 8 findings back to go-cqrs-lite. Two correctness blockers shipped in `system/v4.9.0` (2026-09-21) — does the verdict survive the re-run? (go-graph-rag issue #2)                                                                                                                                                                                                                      |
| Method   | Primary sources re-verified file-by-file and at-tag in the go-cqrs-lite checkout, plus the module proxy (`go list -m -versions`, 2026-11-10). Every claim carries a file:line or command citation. The 2026-09-15 artifact is untouched (point-in-time record); this file is the re-run.                                                                                                                                                                          |
| Verdict  | **Verdict HELD: adopt neither. The re-run strengthens it.** All 8 feedback items are now shipped, substantially closed, or confirmed open-and-unscheduled — and every item that shipped was a library-quality fix that touches none of the four SDK-level disqualifiers. The two decisive capability gaps (ANN vectors #1, rich graph edges #7) remain open on go-cqrs-lite's ROADMAP, unscheduled. Next full re-run at metaengine graduation or v5 landing (§5). |

## 1. TL;DR

go-cqrs-lite moved v4.9.0 → `system/v4.23.0` + `metaengine/v4.23.0` (proxy-verified 2026-11-10) since
the 2026-09-15 evaluation. The movement is real and mostly in the direction the feedback asked for:
three of the eight items shipped verbatim, two substantially closed, and the ecosystem promoted
`system` out of EXPERIMENTAL with ADR-0123 now Accepted (2026-11-03). But the re-run confirms the
issue's own suspicion: the "wrong-shaped for GraphRAG" verdict stands, because none of the shipped
work touches the four findings that actually decided the original question —

1. the write-model inversion (`Build(docs, edges)` bulk vs `Apply` events),
2. no ANN vector engine anywhere (no `hnsw`/`sqlite-vec`/`hannoy` in any `go.mod`),
3. label/weight-less `Edge{From, To any}` (`metaengine/v4/types.go:60-62`), and
4. the dependency-policy break (system still requires the koanf/watermill/otel chain; both
   modules still PROPRIETARY — now an asymmetry, since this SDK relicensed to MIT on 2026-11-10).

## 2. Feedback scorecard (8 items, re-verified)

| # | Item (2026-09-15 feedback)                         | Status now                      | Evidence (go-cqrs-lite, 2026-11-10)                                                                                                                                                      |
| - | -------------------------------------------------- | ------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | Ship an ANN vector engine                          | OPEN — unscheduled              | No `hnsw`/`sqlite-vec`/`hannoy` in any `go.mod`; ROADMAP.md:672-675 keeps it trigger-gated ("first candidate: a `sqliteengine` vector32() option")                                       |
| 2 | Fail closed on grouped materialized views          | **SHIPPED** (post-issue)        | `metaengine/tursoengine/v4/matview.go:16-39`: `WithKnownGroupedViewBug()` opt-in gate; grouped specs refused with a fix-forwarding error message                                         |
| 3 | Fail closed on `EventAdapter.Save` racy fallback   | **SHIPPED** in system/v4.9.0    | Tag tree `system/doc.go` documents the REFUSED ladder; `system/v4/errors.go` + `roles.go` + `adapter_event.go` implement it; 3 pinning tests (`adapter_racy_save_test.go:42,79,118`)     |
| 4 | Cut `system`'s engine requires from consumer graph | SUBSTANTIALLY CLOSED            | `system/v4/go.mod`: engine requires slimmed 4 → 1 (`sqliteengine` only); `system/integration/v4` module now exists. koanf/watermill/otter/otel chain remains — full v5 cut still pending |
| 5 | Stamp experimental status in each module `doc.go`  | **SHIPPED** (2026-09-20)        | Exactly 17 metaengine-family `doc.go` files carry `# Experimental` (repo grep); `system` correctly carries none — it was promoted (see §3)                                               |
| 6 | Close the test-mass gap on `system`                | SUBSTANTIALLY CLOSED            | 206 test funcs in `system/v4/*_test.go` (was 3 core tests at evaluation); promotion doc records the suite green under `-race` and 38 direct consumers (census 2026-10-24)                |
| 7 | Richer graph edges (label, weight) — v5 material   | OPEN — v5 material, unscheduled | `Edge{From, To any}` unchanged (`metaengine/v4/types.go:60-62`); ROADMAP.md:725-727 raw idea; `graphadapter/v4/doc.go` even documents "flat node identity" as a limitation               |
| 8 | Smaller DX batch (fold errors, modules ref, …)     | OPEN — ROADMAP raw idea         | ROADMAP.md:729-731 (Consumer-DX batch)                                                                                                                                                   |

Score: 3 shipped verbatim, 2 substantially closed, 3 open and honestly routed. Every routed-open
item matches what the issue predicted for #1/#7; #2 shipped after the issue was filed — the issue
undercounts the shipped set.

## 3. What else moved since 2026-09-15 (context the matrix re-run needs)

- **ADR-0123 is Accepted** (was Proposed): `docs/adr/0123-v5-unification-single-composition-root.md:4`
  — status synced 2026-11-03; 38 direct `system/v4` consumers (census 2026-10-24); `stack` presets
  deprecated toward v5; v5 itself has NOT landed (ADR-0152 dual-support rollout builds on it).
- **`system` was promoted out of EXPERIMENTAL** (2026-11-03,
  `docs/status/2026-11-03_18-42_system-promotion-out-of-experimental.md`): module suite green
  (10.0s run; 60+ tests under `-race` per FEATURES), no experimental stamp in its `doc.go`.
- **metaengine family stays 🧪 Experimental** (all 17 `doc.go` stamps; FEATURES.md engine rows).
- **License asymmetry is new**: go-cqrs-lite is still PROPRIETARY (`LICENSE`), while this SDK went
  MIT on 2026-11-10 (CHANGELOG `[0.4.0]`). The original CONTRA #7 ("third parties inherit a
  proprietary transitive") is now stronger: an MIT SDK carrying a proprietary dependency would
  undermine its own relicensing rationale for any future adapter module.
- **`go-graph-rag` is unchanged on every dimension the matrix scores**: `Build` bulk write path
  (`build.go:94`), weighted/typed `Edge{Source, Target, Relation, Weight}` (`graph.go:72`), single
  SQLite file with `synchronous=OFF` (`store.go:21`), in-memory linear cosine (`vector.go`), zero
  go-cqrs-lite imports (repo-wide grep). No §9 trigger fired from the SDK side.

## 4. Decision matrix re-run (against system v4.18.0 / metaengine v4.23.0 state)

| Option                      | 2026-09-15 verdict | Re-run verdict         | Why unchanged / changed                                                                                                                                                                                                            |
| --------------------------- | ------------------ | ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Neither (status quo)**    | ✅ KEEP            | ✅ KEEP (strengthened) | All four SDK-level disqualifiers verified unchanged (§1); the shipped correctness work makes the modules BETTER libraries without making them a better fit for this SDK                                                            |
| **metaengine only**         | ✗ REJECT           | ✗ REJECT               | Write model, no-ANN, poor edges, dep policy: all stand. #2/#5 shipped raise general trust; #1/#7 — the capability overlap that motivated the look — remain unscheduled                                                             |
| **system only**             | ✗ REJECT           | ✗ REJECT               | Promotion + 206 tests + ADR-0123 Accepted make it a mature APPLICATION composition root; still a category error for a library (owns consumer topology), nothing retrieval-relevant, still the largest dep tree                     |
| **Both**                    | ✗ REJECT           | ✗ REJECT               | Superset of the two rejections                                                                                                                                                                                                     |
| **App-layer adoption (CV)** | ✅ PROCEED         | ✅ PROCEED (de-risked) | The venue the evaluation recommended is exactly where go-cqrs-lite invested: fail-closed saves, slimmed engine requires, promotion, accepted unification ADR. CV's SUPERB T01/T29 path is on firmer ground than at evaluation time |

## 5. Revisit-trigger scoreboard (original §9, re-scored)

| Trigger (original §9)                                       | Status     | Note                                                                                                                                           |
| ----------------------------------------------------------- | ---------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| metaengine ships an ANN-capable engine                      | NOT fired  | ROADMAP trigger-gated, no candidate scheduled (ROADMAP.md:672-675)                                                                             |
| multi-engine persistence need with one API                  | NOT fired  | SDK still one rebuildable SQLite file (store.go)                                                                                               |
| consumer demands live SSE search results from the SDK       | NOT fired  | No such demand recorded; adapter-module path (§8.3 of the original) remains the answer if it fires                                             |
| metaengine + system graduate from Experimental AND v5 lands | HALF fired | `system` graduated 2026-11-03 + ADR-0123 Accepted; metaengine family (the module with the Vector/Graph ADTs) is still 🧪 and v5 has not landed |
| SDK becomes event-fed                                       | NOT fired  | `Build` bulk path unchanged                                                                                                                    |

**Next re-run condition:** metaengine family graduates from Experimental or the v5 execution lands
(ADR-0123 fold churn settled). A graduation event would retire the "experimental churn" CONTRA and
deserves a fresh matrix; until then this verdict is stable.

## 6. Verdict

KEEP the status quo. The correctness feedback loop worked — go-cqrs-lite shipped #2/#3/#5 and
substantially closed #4/#6 — and that investment strengthens the CV application-layer adoption the
evaluation recommended. It changes nothing about the SDK: the decisive gaps (#1 ANN, #7 rich edges)
are the ones still open, unscheduled, and in the v5 lane. No go-graph-rag work item results from
this re-run; the tracking surface is the research index's reopen-triggers.

## 7. Evidence appendix

- Issue: `LarsArtmann/go-graph-rag#2` (filed post-v4.9.0-wave; claims re-verified here).
- Original evaluation: `docs/research/2026-09-15_metaengine-system-adoption.md` (untouched).
- Feedback doc: go-cqrs-lite
  `docs/feedback/reviewed/archived/2026-09-15_go-graph-rag_metaengine-system-evaluation-feedback.md`
  (8 items; routing note records #2 → TODO, #1/#7/#8 → ROADMAP).
- v4.9.0 wave closeout: go-cqrs-lite
  `docs/status/archived/2026-09-21_23-24_go-graph-rag-wave-closeout-system-v4.9.0-tag-wave.md`.
- Versions (proxy, 2026-11-10): `go list -m -versions github.com/larsartmann/go-cqrs-lite/system/v4`
  → v4.18.0 latest (v4.9.0 present); `…/metaengine/v4` → v4.23.0 latest.
- #3: `git show system/v4.9.0:system/doc.go` (ladder: AtomicAppender → Transactional → REFUSED);
  current `system/v4/errors.go`, `system/v4/roles.go`, `system/v4/adapter_event.go`;
  `system/v4/adapter_racy_save_test.go:42,79,118`.
- #5: repo grep `# Experimental` over `**/doc.go` → exactly 17 metaengine-family modules.
- #2: `metaengine/tursoengine/v4/matview.go:16-39` (`WithKnownGroupedViewBug`, grouped refusal).
- #4/#6: `system/v4/go.mod` (require block: one engine, koanf/watermill/otter/otel chain);
  `system/integration/v4/`; `rg -c "^func Test" system/v4/*_test.go` → 206.
- Promotion/ADR: go-cqrs-lite `docs/adr/0123-v5-unification-single-composition-root.md:4`;
  `docs/status/2026-11-03_18-42_system-promotion-out-of-experimental.md`.
- Open items: go-cqrs-lite `ROADMAP.md:672-675` (#1), `:725-727` (#7), `:729-731` (#8);
  `metaengine/v4/types.go:60-62`; `metaengine/graphadapter/v4/doc.go` (flat node identity).
- License: go-cqrs-lite `LICENSE` (PROPRIETARY); this repo CHANGELOG `[0.4.0]` (MIT, 2026-11-10).
- SDK unchanged: `build.go:94`, `graph.go:72`, `store.go:21`, `vector.go`; repo-wide grep: no
  `go-cqrs-lite` imports.

---

_Re-run of `2026-09-15_metaengine-system-adoption.md` on its §9 trigger surface, requested by issue #2. Next re-run at metaengine graduation or v5 landing._
