# Keptler CLI

Keptler is a generation-first secrets CLI. It turns annotated `.env.example` templates into concrete secret values for local development and CI pipelines. Secrets are generated only when missing and cached in an Age-encrypted state file.

## Installation

This project requires Go 1.23 or newer. To build the `keptler` binary, run:

```bash
go build ./cmd/keptler
```

## Usage

The `generate` command materialises secrets declared in your template:

```bash
./keptler generate -f env.example -o secret.env
```

`secret.env` and the Age-encrypted `.keptler.state.age` file are created in the working directory. Subsequent runs reuse existing values when possible.

See [docs/Functional Specification.md](docs/Functional%20Specification.md) for the detailed specification and [docs/TASKS.md](docs/TASKS.md) for the development roadmap.

## Automated Releases

Tagged commits are built and published by GitHub Actions using
[GoReleaser](https://goreleaser.com/). The workflow cross‑compiles
`keptler` for Linux, macOS and Windows on both `amd64` and `arm64`
architectures and uploads the archives to the corresponding GitHub Release.
