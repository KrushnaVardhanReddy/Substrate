# Substrate — `init` Command Specification (P1-T07)

> **Status:** APPROVED ✅
> **Spec-First Gate:** Jules MUST NOT implement P1-T07 until this document is approved.
> **Scope:** Adds a `substrate init` subcommand to the CLI that scaffolds a `substrate.yaml` config and a `.github/workflows/substrate.yml` GitHub Action workflow file in the user's repository.
> **Depends on:** `docs/specs/override-config.md` (substrate.yaml format), P1-T03 (CLI scaffold already exists)

---

## Problem Statement

New users landing on the GitHub Marketplace listing must manually write both the `substrate.yaml` config and the GitHub Actions workflow from scratch, referencing the README. This creates friction.

A single command — `substrate init` — eliminates that friction by generating ready-to-use template files in the user's project directory.

---

## Command Definition

```
substrate init [flags]
```

### Flags

| Flag | Type | Default | Description |
|---|---|---|---|
| `--service` | `string` | directory name of cwd | The service name written into `substrate.yaml` |
| `--spec` | `string` | auto-detected (see below) | Path to the OpenAPI spec file, relative to repo root |
| `--schema-type` | `string` | `openapi` | Schema type: `openapi`, `sql`, `graphql`, `protobuf` |
| `--branch` | `string` | `main` | The protected branch to check PRs against |
| `--force` | `bool` | `false` | Overwrite existing files without prompting |
| `--no-workflow` | `bool` | `false` | Skip generating `.github/workflows/substrate.yml` |
| `--no-config` | `bool` | `false` | Skip generating `substrate.yaml` |

---

## Auto-Detection of Spec Path

If `--spec` is not provided, the command searches for a spec file in the current directory and common subdirectories in this order:

1. `openapi.yaml`
2. `openapi.json`
3. `api/openapi.yaml`
4. `api/openapi.json`
5. `docs/openapi.yaml`
6. `docs/openapi.json`
7. `swagger.yaml`
8. `swagger.json`

If a file is found → use its path in the generated `substrate.yaml`.
If no file is found → use `openapi.yaml` as a placeholder and print a warning:

```
⚠️  No OpenAPI spec found. Set spec_path in substrate.yaml to the correct path.
```

---

## Output Files

### File 1: `substrate.yaml`

Generated at the **current working directory** (repo root).

**Template:**

```yaml
# substrate.yaml — Substrate configuration
# Docs: https://github.com/KrushnaVardhanReddy/Substrate

service: <service-name>
schema_type: <schema-type>
spec_path: <spec-path>

owners:
  - team: your-team-name
    contact: your-team@company.com

# Overrides — acknowledge intentional breaking changes
# overrides:
#   - rule_id: ENDPOINT_REMOVED
#     path: paths./your-endpoint
#     reason: "Reason this is safe (min 20 chars)"
#     approved_by: you@company.com
#     expires: YYYY-MM-DD
```

Placeholders `<service-name>`, `<schema-type>`, and `<spec-path>` are replaced with real values (from flags or auto-detection).

### File 2: `.github/workflows/substrate.yml`

Generated at `.github/workflows/substrate.yml` relative to cwd. The `.github/workflows/` directory is created if it does not exist.

**Template:**

```yaml
# Substrate API Contract Guard
# Docs: https://github.com/KrushnaVardhanReddy/Substrate

name: Substrate API Contract Guard

on:
  pull_request:
    branches: [<branch>]

jobs:
  check-api-contracts:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Check API Breaking Changes
        uses: KrushnaVardhanReddy/Substrate@v0.1.0
        with:
          base_schema: <spec-path>
          head_schema: <spec-path>
          config: substrate.yaml
```

Placeholder `<branch>` is replaced with the `--branch` flag value. `<spec-path>` is replaced with the detected or provided spec path.

---

## Behaviour

### If files already exist

If `substrate.yaml` or `.github/workflows/substrate.yml` already exist and `--force` is NOT set:

```
⚠️  substrate.yaml already exists. Use --force to overwrite.
⚠️  .github/workflows/substrate.yml already exists. Use --force to overwrite.
```

The command exits with code `0` (not an error — files may already be configured).

If `--force` IS set → overwrite without prompting.

### Success output

```
✅ Substrate initialized!

Created:
  substrate.yaml
  .github/workflows/substrate.yml

Next steps:
  1. Review substrate.yaml and update 'owners' with your team details.
  2. Commit both files and open a PR to see Substrate in action.
  3. Docs: https://github.com/KrushnaVardhanReddy/Substrate
```

### Dry run (no flag needed — future consideration)

Not in scope for P1-T07. Can be added in a follow-up.

---

## Exit Codes

| Code | Condition |
|---|---|
| `0` | Success (files created or already exist) |
| `1` | `--force` requested but write permission denied |
| `3` | Internal error (unexpected) |

---

## Implementation Notes

- Add `initCmd` as a new Cobra subcommand registered in `cmd/substrate/main.go`
- Keep template strings as Go string constants in a new file: `engine/internal/init/templates.go`
- The `init` package must NOT import the `diff` package — it is completely independent
- Service name default: `filepath.Base(cwd)` — the directory name of wherever the command is run
- All file writes use `os.WriteFile` — no external dependencies

---

## Files To Create / Modify

| File | Change |
|---|---|
| `engine/internal/init/init.go` | New — core logic: auto-detect, write files |
| `engine/internal/init/templates.go` | New — template strings as Go constants |
| `engine/internal/init/init_test.go` | New — unit tests (see Test Requirements) |
| `engine/cmd/substrate/main.go` | Modify — register `initCmd` with Cobra |

---

## Test Requirements

| Test Case | Input | Expected |
|---|---|---|
| No existing files, no flags | cwd with no spec | Both files created, warning about missing spec path |
| No existing files, `--spec api/openapi.yaml` | — | Both files created with correct spec path |
| Auto-detect spec | cwd with `openapi.yaml` present | Both files created with `openapi.yaml` as spec path |
| Files already exist, no `--force` | — | Warning printed, exit 0, files unchanged |
| Files already exist, `--force` | — | Files overwritten |
| `--no-workflow` flag | — | Only `substrate.yaml` created |
| `--no-config` flag | — | Only `.github/workflows/substrate.yml` created |
| `--service my-api` | — | `service: my-api` in generated `substrate.yaml` |
| `--branch develop` | — | `branches: [develop]` in generated workflow |

Tests MUST use a temporary directory (`t.TempDir()`) and MUST NOT write to the real filesystem outside of it.

---

## README Update (after P1-T07 merges)

Add to the Quick Start section — **before** Step 1:

```
## Step 0 — Run `substrate init` (fastest setup)

If you have the Substrate CLI installed:

    substrate init --spec path/to/openapi.yaml

This generates `substrate.yaml` and `.github/workflows/substrate.yml` automatically.
Skip to Step 3 — just commit and push.
```

---

## Next Step

> Once this spec is approved:
> - Create Jules prompt at `prompts/phase-1-diff-engine/t07_init_command.txt`
> - Submit via `python3 scripts/jules_submit.py --task 7`
