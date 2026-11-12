# Security Policy

## Supported Versions

This is a library/SDK. Only the latest release line receives security fixes.

| Version | Supported          |
| ------- | ------------------ |
| v0.5.x  | :white_check_mark: |
| < v0.5  | :x:                |

The BDD test kit releases independently under `bdd/v*` tags; the current kit
line is `bdd/v0.1.x` (same latest-line-only policy).

## Reporting a Vulnerability

If you discover a security vulnerability in go-graph-rag, please report it
responsibly:

1. **Do NOT open a public GitHub issue.**
2. Use [GitHub private vulnerability reporting](https://docs.github.com/en/code-security/how-tos/report-and-fix-vulnerabilities/report-privately)
   on this repository (Security tab → "Report a vulnerability", or directly:
   <https://github.com/LarsArtmann/go-graph-rag/security/advisories/new>).
3. Include a proof of concept or steps to reproduce if possible.
4. You will receive an acknowledgment within 48 hours.

## Scope Notes

- **`OpenAICompatProvider` sends document and query text to the configured
  endpoint.** The SDK never chooses the endpoint; the operator does. If the
  texts you index are sensitive, point `BaseURL` at a endpoint you trust (or
  run the offline `NewHashProvider`, which performs no I/O).
- **API keys are passed through configuration** (`EmbeddingConfig.APIKey`) and
  sent only as a bearer token to the configured endpoint. They are never
  logged by this SDK — but they live in the caller's process memory and
  config files, which is the caller's threat model.
- **The SQLite store is a local file with no authentication.** Protect the
  file system path (`Config.StoreDSN`) like any other data-at-rest concern;
  the store applies no encryption.

## Security Practices

- **Minimal dependency surface:** production code depends only on
  `samber/lo` and `modernc.org/sqlite` (CGo-free SQLite); everything else is
  the Go standard library. No dependency is added without a hard reason
  (repo non-negotiable #3).
- **CGo-free core:** the core module builds without CGo, keeping the
  dependency/attack surface auditable as pure Go.
- **No outbound I/O by default:** the offline hash embedder (`NewHashProvider`)
  performs no network or file I/O; network calls exist only in the
  OpenAI-compatible provider, against the operator-configured endpoint.
- **Credentials are never logged:** API keys live in the caller's config
  (`EmbeddingConfig.APIKey`) and are sent only as bearer tokens to the
  configured endpoint; the SDK never logs them.
- **Local-data posture:** the SQLite store is an unauthenticated local file;
  callers protect the DSN path like any data-at-rest concern (see Scope
  Notes).
- **Continuous verification:** CI runs build, vet, race-enabled tests, lint,
  and tidy-drift on every push for both modules (branch-protected `master`);
  `govulncheck` runs in the BuildFlow gate (receipt 2026-11-11: zero
  third-party findings — every advisory is a Go standard-library one,
  fixed in 1.27.x patch releases newer than the go1.27 floor, so build
  consumers with the newest 1.27.x toolchain); Dependabot keeps Go
  dependencies and GitHub Actions current.
- **Wire compatibility is pinned:** byte-level goldens (`wire_golden_test.go`)
  cover the serialized graph/cache shapes under the canonical
  `encoding/json/v2` engine, so a serialization regression cannot ship
  silently.

## Dependency Policy

- Production dependencies are kept minimal: `samber/lo` and
  `modernc.org/sqlite` (CGo-free SQLite). Everything else is stdlib.
- `testify` is test-only and excluded from production builds.
- Dependabot keeps GitHub Actions and Go dependencies current; CI must stay
  green on `master` (branch protection enforces the `build-test-lint` check).
