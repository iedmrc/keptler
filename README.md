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
