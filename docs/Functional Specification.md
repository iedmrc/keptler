# Keptler CLI — Functional Specification (v0.1)

---

## 1 Purpose

Keptler is a **generation‑first secrets CLI** that turns declarative templates into concrete, high‑entropy secret values for **both** local development **and** CI pipelines:

1. **Local run (default)** – only the working directory is touched: existing output/state are reused, missing or rule‑violating values are generated, **no network calls**.
2. **Remote run** – explicitly enabled (`--remote github` or `KEPTLER_REMOTE=github`, or auto‑enabled inside GitHub Actions with `CI=true` and `GITHUB_TOKEN`). Keys are first looked up in **GitHub Repository Secrets**; new values are written back to GitHub *and* to local files.

Keptler never stores secrets centrally; the only on‑disk persistence is an **age‑encrypted cache** that prevents unwanted regeneration.

---

## 2 Problem Statement

Teams still create secrets by hand: copying passwords from chat, committing `.env` files, generating JWT keys ad hoc. This yields weak entropy, leaks, and drift between developers and pipelines. Treating secrets as *code* eliminates manual error:

- **Templates** declare *what* is required (and how it should be formed).
- **Keptler** guarantees compliant, synchronised values—everywhere.

---

## 3 Goals & Non‑Goals (v0.1)

| Category              | In Scope                                                        | Out of Scope                              |
| --------------------- | --------------------------------------------------------------- | ----------------------------------------- |
| **Secret generation** | Random strings, RSA (PKCS1/PKCS8), Ed25519, derived public keys | Provider‑fetched values, P‑256/ECC curves |
| **Remote back‑end**   | GitHub Repo Secrets create/lookup                               | GitHub Environment/Org, Vault, AWS SM     |
| **CLI commands**      | `generate`, `rotate`, `list`, `validate`                        | GUI/TUI, standalone `sync`                |
| **Persistence**       | `.keptler.state.age` (X25519‑Age)                               | Central vault                             |
| **OS support**        | Linux & macOS                                                   | Windows (post‑v0.1, best‑effort)          |

---

## 4 Personas

| Persona              | Key need                                                                             |
| -------------------- | ------------------------------------------------------------------------------------ |
| **Backend engineer** | One‑command `.env` creation, strong keys, easy rotation                              |
| **CI engineer**      | Deterministic secrets; respect existing GitHub values; commit new ones automatically |
| **Security auditor** | Entropy policy evidence, drift detection                                             |
| **Product owner**    | Clear acceptance criteria and predictable release scope                              |

---

## 5 CLI Surface

```text
keptler [global flags] <command> [command flags]

Global flags
  --quiet        Suppress non‑error output
  --json         Machine‑readable output
  --no-color     Disable ANSI colours
```

### 5.1 Commands

| Command    | Behaviour                                                                                                                                                                                                                                                                                                                                                      | Key flags                                                                                                                                  |                       |
| ---------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ | --------------------- |
| `generate` | Materialise all secrets.**Local mode (default)** – read template → reuse values from `secret.env` + state if rule‑compliant → generate the rest → write `secret.env` + state. 0 network calls.**Remote mode** – remote lookup **first** → fallback to local reuse → generate remaining → always write local files, **also** create/update GitHub Repo Secrets. | `-f, --file <path>` (repeatable)`-o, --out <path>` (default `secret.env`)`--merge` (don’t overwrite extra keys in output)\`--remote github | none`(default`none\`) |
| `rotate`   | Re‑generate a subset of **materialised** secrets and propagate changes (remote if enabled).                                                                                                                                                                                                                                                                    | Positional `KEY …` or `--filter '<expr>'` (e.g. `age>90d`)`--output-update`                                                                |                       |
| `list`     | Show *generated* and *pending* secrets. Columns: `STATE` (generated/pending), `RULE`, `AGE`, `REMOTE` (in‑sync/missing/diverged), `VALUE` (masked).                                                                                                                                                                                                            | \`--format table                                                                                                                           | json`, `--plain\`     |
| `validate` | Lint templates + ensure entropy policy; if `--remote` is set, verifies connectivity.                                                                                                                                                                                                                                                                           | `-f, --file <path>` (repeatable), `--strict`                                                                                               |                       |

### 5.2 Remote secret logic

| Aspect                            | Behaviour                                                                                                                                                         |
| --------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Activation**                    | Explicit `--remote github` or `KEPTLER_REMOTE=github`. Auto‑enabled inside GitHub Actions only when `CI=true` **and** `GITHUB_TOKEN` present. Otherwise disabled. |
| **Scope**                         | Repository‑level secrets (key = secret name).                                                                                                                     |
| **Auth**                          | Workflow `GITHUB_TOKEN` or PAT with `repo → secrets`.                                                                                                             |
| **Lookup sequence (remote mode)** | 1 GitHub Repo Secret → 2 `secret.env` value (if rule‑compliant) → 3 encrypted state → 4 generate.                                                                 |
| **Write‑back**                    | Generated or rotated values are pushed to GitHub only in remote mode. Local mode writes only output + state.                                                      |
| **Drift detection**               | `list` computes hash of remote value; marks `diverged` if differs from state.                                                                                     |

### 5.3 Rotation filter DSL

Initial operator: `age>Nd|Nw|Nmo` (e.g. `age>90d`).

---

## 6 Template formats

### 6.1 YAML (`secrets.template.yml`)

```yaml
version: 1
defaults:
  length: 32
  charset: alnum
secrets:
  DB_PASSWORD: { rule: random, length: 48 }
  JWT_PRIVATE_KEY: { rule: rsa-private-key, bits: 4096, format: pkcs8 }
  JWT_PUBLIC_KEY: { rule: derive, source: JWT_PRIVATE_KEY }
```

### 6.2 Annotated `.env.example`

```dotenv
# keptler: random length=48 charset=alnum
DB_PASSWORD=
JWT_PRIVATE_KEY= # keptler: rsa-private-key bits=4096 format=pkcs8
JWT_PUBLIC_KEY=  # keptler: derive source=JWT_PRIVATE_KEY
```

Either **comment‑above** or **inline** styles are accepted; first `# keptler:` tag wins.

---

## 7 Attribute reference

| Attribute  | Type   | Applies to          | Default   | Notes                                                                |
| ---------- | ------ | ------------------- | --------- | -------------------------------------------------------------------- |
| `rule`     | enum   | all                 | —         | `random`, `rsa-private-key`, `ed25519-private-key`, `uuid`, `derive` |
| `length`   | int    | random              | 32        | Characters before encoding                                           |
| `charset`  | enum   | random              | `alnum`   | `alnum`, `hex`, `base64`, `urlsafe`                                  |
| `encoding` | enum   | random              | none      | `base64`, `hex`, `urlsafe`                                           |
| `bits`     | int    | rsa-private-key     | 2048      | 4096 recommended                                                     |
| `format`   | enum   | rsa-private-key     | `pkcs1`   | `pkcs1`, `pkcs8`                                                     |
| `curve`    | enum   | ed25519-private-key | `ed25519` | future: `p256`                                                       |
| `source`   | string | derive              | —         | Key to derive from                                                   |

### 7.1 Entropy policy

Any `random` secret must reach **≥128 bits** (length × bits/char). Table: `alnum`=5.95, `hex`=4.0, `base64`=6.0, `urlsafe`=5.94.

---

## 8 Files produced

| File                          | Purpose                                 | Git‑ignored              |
| ----------------------------- | --------------------------------------- | ------------------------ |
| `secret.env`                  | Plaintext secrets for local dev/Compose | Recommended yes          |
| `.keptler.state.age`          | age‑encrypted cache + metadata          | Auto add to `.gitignore` |
| Permissions forced to `0600`. |                                         |                          |

---

## 9 Security controls

1 Masked output by default (`--plain` opt‑in). 2 Zero‑wipe sensitive buffers. 3 State encrypted with X25519‑Age (ChaCha20‑Poly1305). 4 Exponential back‑off on GitHub 429.

---

## 10 Acceptance criteria

### Functional

1 **Local run** `keptler generate` (no `--remote`) hits no network, reuses valid values, generates the rest, writes output + state. 2 **CI run** `keptler generate --remote github` looks up GitHub secrets first, generates missing/invalid ones, updates GitHub and local files, and is idempotent on re‑run. 3 `list --format table` shows correct `REMOTE` status (`in‑sync`, `missing`, `diverged`). 4 `rotate DB_PASSWORD --output-update` regenerates key, updates GitHub (remote mode), rewrites files, updates state. 5 `validate` fails on entropy <128 bits or remote connectivity error when `--remote github` active.

### Non‑functional

6 Network calls: timeout 5 s, 3 retries. 7 Stripped binary ≤10 MB; typical run <300 ms (no network). 8 Unit test coverage ≥80 %. 9 State file broader than `0600` → error.

---

## 11 Roadmap (post‑v0.1)

- GitHub Environment/Org scopes
- AWS Secrets Manager, Vault targets
- Provider‑fetched secrets (`fetch` rule)
- Auto‑rotation reminders
- Windows support

---

## 12 Glossary

| Term               | Definition                                              |
| ------------------ | ------------------------------------------------------- |
| **Materialise**    | Generate & write concrete secret value from template    |
| **Pending secret** | Declared but not yet materialised                       |
| **Remote status**  | Comparison between cached value hash and external store |
| **Entropy**        | Measure of randomness (bits) – higher is stronger       |

---

*Revision 2025‑06‑16*

