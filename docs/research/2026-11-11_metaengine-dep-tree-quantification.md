# Metaengine/system dependency-tree quantification

- **Date:** 2026-11-11
- **Question:** what does adopting `go-cqrs-lite/system` + `metaengine` actually cost a consumer's dependency graph? (The standing owner item "metaengine dep-tree quantification — run soon" since 2026-09-16; feeds any adapter-cost question in the seam ADR.)
- **Method:** scratch module (`/tmp/mq`) importing exactly two modules —
  `github.com/larsartmann/go-cqrs-lite/metaengine/v4 v4.23.0` and
  `github.com/larsartmann/go-cqrs-lite/system/v4 v4.18.0` (latest published on
  the proxy at measure time), toolchain `go1.27.2`, no other source. Measured
  with `go list -m all`, `go list -deps .`, `go mod graph`, and the built
  binary size of the trivial importer. Scratch discarded after capture; the
  repo's own `go.mod` untouched.

## Numbers

| Metric                                          | Value                                                                                                                                                   |
| ----------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Modules in the resolved graph                   | **184**                                                                                                                                                 |
| — of which `go-cqrs-lite` submodules            | **32**                                                                                                                                                  |
| External packages pulled (non-stdlib)           | **138**                                                                                                                                                 |
| Packages including stdlib                       | 336                                                                                                                                                     |
| Module-graph edges                              | 962                                                                                                                                                     |
| Trivial-importer binary (8-dim reference `fmt`) | 8.19 MB                                                                                                                                                 |
| Top fan-in modules                              | `go-branded-id` v0.7.0, `go-logr` v1.4.4 (18 requirers each), `go.yaml.in/yaml/v3`, `x448/float16`, `oklog/ulid/v2` (17), `xxhash/v2` (16), `otel` (14) |
| Heaviest non-Lars third-party families          | OpenTelemetry (6+ modules: otel, sdk, sdk/metric, trace, exporters, log, auto/sdk), koanf (5), watermill, otter cache, gobreaker                        |

## Findings

1. **`system/v4` is the fat entry point, exactly as ADR-0123 assumed:**
   requiring only `system/v4` transitively pulls `metaengine/v4`,
   `metaengine/sqliteengine/v4`, and `metaengine/projectionadapter/v4`
   (plus `command`, `event`, `query`, `decider`, `dispatcher`, `projection*`,
   `snapshot`, `watermill`, `otel`, `record`, `id`, `dedup`, `schema`,
   `metadata`, `commandlifecycle*`, `scheduling/engine`, `go-error-family`,
   `go-flightrecorder`, `go-sse`, `go-branded-id`, `go-codec`). There is no
   thin-`system` adoption shape: the composition root IS the whole stack.
2. **OTel arrives unconditionally** through `otel/v4` + `watermill/v4` — a
   consumer that wants zero observability dependencies still carries the
   otel module family in its graph (binary impact deferred; the trivial
   binary links only reachable code).
3. **Engine modules are pay-per-use but not opt-out-able:** `sqliteengine/v4`
   (plus adapter) ride in with `system`; the other engines (badger, bbolt,
   duckdb, bigtable, dgraph, iroh, mysql — separate modules in the monorepo)
   enter only when imported.
4. **Contrast with this SDK:** a go-graph-rag consumer resolves a graph of
   roughly 16 modules (samber/lo, modernc.org/sqlite family, stdlib) and a
   trivial importer of ~2 MB — about one order of magnitude smaller in
   modules and roughly 4× smaller in trivial-binary size than the
   `system`+`metaengine` shape. This is the quantified cost side of the
   seam ADR's "core stays lean; adapters own their dependencies" rule.

## Verdict

Measurement only — **does not change the standing verdict** (adopt neither;
2026-09-15, re-verified 2026-11-10). It quantifies what "CV-side app-layer
adoption" would add to a consumer graph (~184 modules, ~8 MB trivial-binary
ceiling) and confirms there is no slimmer `system`-only path. If the owner
ever wants the adapter route, the cost is now a number, not a guess.

## Reopen when

- metaengine v5 lands or the module is graduated (re-run — the tree churns).
- `system/v4` gains a slim composition mode (would invalidate finding 1).
- The seam ADR's adapter trigger fires (§9) — reuse these numbers as the
  baseline.
