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

    GOTOOLCHAIN=go1.26.7 go build ./... && GOTOOLCHAIN=go1.26.7 go vet ./...
    GOTOOLCHAIN=go1.26.7 go test ./... -race
    golangci-lint run ./...

The module must build without `GOEXPERIMENT=jsonv2`; see AGENTS.md for the
pristine-toolchain check and all canonical commands.

Install the pre-push guard once per clone (blocks pushes of trees that only
build under the dev shell's `GOEXPERIMENT=jsonv2`):

    git config core.hooksPath .githooks

Markdown/JSON/YAML formatting is enforced by dprint in CI; format locally
with `nix run nixpkgs#dprint -- fmt` (or `dprint fmt`).

## Reporting Issues

Please use GitHub Issues to report bugs or request features.
