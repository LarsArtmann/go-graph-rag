# Security Policy

## Supported Versions

This is a library/SDK. Only the latest release line receives security fixes.

| Version | Supported          |
| ------- | ------------------ |
| v0.1.x  | :white_check_mark: |
| < v0.1  | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in go-graph-rag, please report it
responsibly:

1. **Do NOT open a public GitHub issue.**
2. Use [GitHub private vulnerability reporting](https://docs.github.com/en/code-security/how-tos/report-and-fix-vulnerabilities/report-privately)
   on this repository (Security tab → "Report a vulnerability").
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

## Dependency Policy

- Production dependencies are kept minimal: `samber/lo` and
  `modernc.org/sqlite` (CGo-free SQLite). Everything else is stdlib.
- `testify` is test-only and excluded from production builds.
- Dependabot keeps GitHub Actions and Go dependencies current; CI must stay
  green on `master` (branch protection enforces the `build-test-lint` check).
