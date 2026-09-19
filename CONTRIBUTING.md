# Contributing

Thanks for your interest in contributing!

Note: this repository is public but the code is PROPRIETARY (see LICENSE).
Pull requests are welcome; reuse beyond PR submission needs explicit
permission from the author.

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

Install the pre-push guard once per clone — it runs all three gates on
every push and blocks on any failure:

    git config core.hooksPath .githooks

1. Pristine build — the tree must build without `GOEXPERIMENT=jsonv2`.
2. Tidy drift — `go mod tidy` must leave go.mod/go.sum unchanged.
3. dprint check — markdown/json/yaml formatting must be clean.

Markdown/JSON/YAML formatting is enforced by dprint in CI and by the hook
locally; format with `nix run nixpkgs#dprint -- fmt` (or `dprint fmt`).

## Reporting Issues

Please use GitHub Issues to report bugs or request features.
