# Substrate — Override & Repository Config Specification

> **Status:** APPROVED ✅ — Required before Phase 2 (GitHub App) is built.
> **Spec-First Gate:** Jules MUST NOT implement the override parser (P1-T03) until this document is approved.
> **Scope:** Defines the `substrate.yaml` file format. This file lives in the root of every repository that uses Substrate.

---

## Overview

`substrate.yaml` is the **single per-repository configuration file** for Substrate. It serves two distinct purposes in one file:

1. **Repo Registration** — tells the GitHub App where the schema file lives and what service this repo represents
2. **Override Config** — lets DevOps/teams acknowledge and skip specific detections with an audit trail

Both sections are optional independently but the file must be valid YAML.

---

## Full Example

```yaml
# substrate.yaml — place in the root of your repository

# ── Section 1: Repo Registration ──────────────────────────────────────────────
service: customer-service
schema_type: openapi
spec_path: openapi/api.yaml
owners:
  - team: platform-team
    contact: platform@company.com

# ── Section 2: Override Config ────────────────────────────────────────────────
overrides:
  - rule_id: FIELD_REMOVED
    path: components.schemas.Customer.properties.legacy_email
    reason: "Deprecated 6 months ago. All 3 consumers migrated. See RFC-2041."
    approved_by: jane@company.com
    expires: 2026-09-01

  - rule_id: ENUM_VALUE_ADDED
    path: components.schemas.OrderStatus.enum
    reason: "All consumers use default handling for unknown enum values. Verified with all teams."
    approved_by: mike@company.com
    expires: 2026-08-15
```

---

## Section 1: Repo Registration Fields

These fields are read by the **GitHub App** (Phase 2) to locate the schema file in a repository.

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `service` | `string` | ✅ | — | Human-readable name for this service. Shown in PR comments and the dashboard. |
| `schema_type` | `string` | ❌ | auto-detect | Schema format. One of: `openapi`, `sql`, `graphql`, `protobuf`. If omitted, auto-detected from `spec_path` file extension. |
| `spec_path` | `string` | ✅ | — | Relative path to the schema file from the repo root. E.g. `openapi/api.yaml`, `schema.graphql`, `migrations/`. |
| `owners` | `Owner[]` | ❌ | `[]` | List of owning teams. Used for dashboard ownership display and impact notifications. |

### Owner Object

| Field | Type | Required | Description |
|---|---|---|---|
| `team` | `string` | ✅ | Team name (e.g., `platform-team`) |
| `contact` | `string` | ❌ | Email or Slack handle for notifications |

### Auto-detection Rules for `schema_type`

If `schema_type` is not set, Substrate infers it from `spec_path`:

| File extension / pattern | Inferred type |
|---|---|
| `.yaml`, `.json` (containing `openapi:` key) | `openapi` |
| `.graphql`, `.gql` | `graphql` |
| `.proto` | `protobuf` |
| `.sql`, directory named `migrations/` | `sql` |

If auto-detection fails → engine exits with code `3` and prints an error.

---

## Section 2: Override Config Fields

Overrides let teams acknowledge a breaking change that has already been safely coordinated. The diff engine marks it as **"✅ Acknowledged"** in the PR comment instead of blocking.

### Override Object

| Field | Type | Required | Description |
|---|---|---|---|
| `rule_id` | `string` | ✅ | The Rule ID to suppress (e.g., `FIELD_REMOVED`). Must match a valid rule from `breaking-change-rules.md`. |
| `path` | `string` | ✅ | Dot-notation path to the specific change being acknowledged (e.g., `components.schemas.Customer.properties.legacy_email`). Must match the `path` field in the DiffReport `Change` object exactly. |
| `reason` | `string` | ✅ | Human-readable justification. Stored in audit log. Minimum 20 characters. |
| `approved_by` | `string` | ✅ | Email or GitHub username of the person who approved this override. |
| `expires` | `string` (ISO date) | ✅ | Date when this override expires (`YYYY-MM-DD`). After this date, the rule fires again automatically. Maximum: 1 year from creation. |

### Expiry Behaviour

- If today's date is **before** `expires` → change is marked **"✅ Acknowledged (override active)"**
- If today's date is **on or after** `expires` → override is ignored. The change is treated as a new blocking detection
- Expired overrides are listed separately in the PR comment: **"⚠️ Expired Override — please renew or fix"**

### Matching Logic

The engine matches an override to a detected change when **both** conditions are true:
1. `override.rule_id` == `change.rule_id`
2. `override.path` == `change.path` (exact string match)

If a detected breaking change has no matching override → it blocks the merge as normal.

---

## PR Comment Behaviour

The GitHub App (Phase 2) uses the override status to annotate the PR comment:

| Scenario | PR Comment Shows |
|---|---|
| Breaking change, no override | `❌ BREAKING — blocks merge` |
| Breaking change, valid override | `✅ Acknowledged — override active until 2026-09-01 (jane@company.com)` |
| Breaking change, expired override | `⚠️ Override expired 2026-07-01 — please renew or fix the issue` |
| Warning, no override | `⚠️ WARNING — merge allowed` |
| Safe change | `✅ SAFE` |

---

## Validation Rules

The engine validates `substrate.yaml` on startup. Any validation failure exits with code `3`:

| Rule | Error |
|---|---|
| `service` is empty or missing | `substrate.yaml: 'service' is required` |
| `spec_path` is empty or missing | `substrate.yaml: 'spec_path' is required` |
| `spec_path` file does not exist in repo | `substrate.yaml: spec_path 'openapi/api.yaml' not found` *(Note: This check is bypassed in HTTP/Cloud execution mode since the spec file is not physically on disk alongside the config)* |
| `override.rule_id` is not a known rule ID | `substrate.yaml: unknown rule_id 'MY_CUSTOM_RULE'` |
| `override.expires` is in the past by more than 30 days | `substrate.yaml: override for 'FIELD_REMOVED' expired on 2026-01-01 — remove or renew` |
| `override.reason` is fewer than 20 characters | `substrate.yaml: override reason too short (min 20 chars)` |
| `schema_type` is an unknown value | `substrate.yaml: unknown schema_type 'avro' (supported: openapi, sql, graphql, protobuf)` |

---

## CLI Integration

The engine reads `substrate.yaml` automatically from `./substrate.yaml` (repo root).
A custom path can be specified with `--config`:

```bash
substrate diff base.yaml pr.yaml --config ./config/substrate.yaml
```

If no `substrate.yaml` is present, the engine runs with defaults (no overrides, schema type auto-detected).

---

## Schema Version

`substrate.yaml` includes an optional `version` field for forward compatibility:

```yaml
version: "1"   # optional, defaults to "1"
service: my-service
spec_path: openapi.yaml
```

If a future Substrate version introduces breaking changes to the config format, the `version` field allows backward compatibility detection.

---

## Next Step

> This spec is complete. The following tasks can now proceed:
> - **P1-T01** (Jules): Go module scaffold — can start immediately
> - **P1-T03** (Jules): Override config parser — implement the `substrate.yaml` parser based on this spec
> - **P2-T01** (Jules): GitHub App — uses `service`, `spec_path`, `schema_type` from this spec to locate schema files in repos
