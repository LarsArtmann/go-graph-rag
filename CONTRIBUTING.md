# Contributing

Thanks for your interest in contributing!

This project is MIT-licensed (see LICENSE). Issues and pull requests are
welcome; by submitting one you grant the right to use your contribution in
this project under the same license.

## How to Contribute

1. Open an issue describing the bug or feature
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Development Setup

Run the following commands to set up your development environment:

    GOTOOLCHAIN=go1.27.1 go build ./... && GOTOOLCHAIN=go1.27.1 go vet ./...
    GOTOOLCHAIN=go1.27.1 go test ./... -race
    golangci-lint run ./...

The module must build without `GOEXPERIMENT=jsonv2`; see AGENTS.md for the
pristine-toolchain check and all canonical commands.

CI is the authoritative gate: every push and pull request runs, for both
modules, a json/v1 policy check (no `encoding/json/v2` outside the
sanctioned golden test), build + vet, tests with `-race`, golangci-lint,
tidy-drift, and dprint format checks. The equivalent local commands are
documented in AGENTS.md; run them when you want feedback before pushing.

Markdown/JSON/YAML formatting is enforced by dprint in CI and by the hook
locally; format with `nix run nixpkgs#dprint -- fmt` (or `dprint fmt`).

## Reporting Issues

Please use GitHub Issues to report bugs or request features.
