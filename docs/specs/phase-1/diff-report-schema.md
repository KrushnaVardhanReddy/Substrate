# Substrate — DiffReport JSON Schema Specification

> **Status:** APPROVED ✅ — Updated after competitive scan. Implementation may proceed.
> **Spec-First Gate:** Any structural change to this schema requires updating this document first.

---

## Overview

The `DiffReport` is the **single output contract** of the Substrate Diff Engine. Every consumer of the engine — the GitHub App, the SvelteKit Dashboard, the CLI, and future integrations (Slack, Jira, PagerDuty) — parses this exact JSON structure.

**Nothing about this schema changes without updating this spec first.**

---

## Top-Level Structure

```json
{
  "substrate_version": "0.1.0",
  "schema_type": "openapi",
  "compared_at": "2026-07-07T01:00:00Z",
  "summary": {
    "total_changes": 3,
    "breaking_count": 1,
    "warning_count": 1,
    "safe_count": 1,
    "overall_severity": "BREAKING"
  },
  "breaking_changes": [],
  "warnings": [],
  "safe_changes": [],
  "compliance_alerts": []
}
```

### Top-Level Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `substrate_version` | `string` | ✅ | Semver of the Substrate engine that produced this report |
| `schema_type` | `SchemaType` | ✅ | The format of the schema that was analyzed |
| `compared_at` | `string` (ISO 8601) | ✅ | Timestamp when the diff was generated |
| `summary` | `Summary` | ✅ | Aggregate counts and overall severity |
| `breaking_changes` | `Change[]` | ✅ | Array of breaking changes (may be empty) |
| `warnings` | `Change[]` | ✅ | Array of potentially breaking changes (may be empty) |
| `safe_changes` | `Change[]` | ✅ | Array of confirmed non-breaking changes (may be empty) |
| `compliance_alerts` | `ComplianceAlert[]` | ❌ | Array of compliance warnings (e.g. PII, HIPAA) |

---

## SchemaType Enum

```json
"schema_type": "openapi"
```

| Value | Description |
|---|---|
| `openapi` | OpenAPI 3.x YAML or JSON |
| `sql` | SQL migration file (Postgres DDL) |
| `graphql` | GraphQL SDL schema |
| `protobuf` | Protocol Buffers `.proto` file |
| `avro` | Apache Avro schema |

**Phase 1 implements `openapi` only. All others are reserved.**

---

## Summary Object

```json
{
  "total_changes": 3,
  "breaking_count": 1,
  "warning_count": 1,
  "safe_count": 1,
  "overall_severity": "BREAKING"
}
```

| Field | Type | Description |
|---|---|---|
| `total_changes` | `integer` | Sum of breaking + warning + safe counts |
| `breaking_count` | `integer` | Number of entries in `breaking_changes` |
| `warning_count` | `integer` | Number of entries in `warnings` |
| `safe_count` | `integer` | Number of entries in `safe_changes` |
| `overall_severity` | `Severity` | Highest severity level present in the report |

### Severity Enum (overall_severity)

| Value | When Used |
|---|---|
| `BREAKING` | At least one breaking change exists |
| `WARNING` | No breaking changes, but at least one warning exists |
| `SAFE` | All changes are safe (no breaking, no warnings) |
| `NO_CHANGES` | The two schema files are identical |

---

## Change Object

Each entry in `breaking_changes`, `warnings`, and `safe_changes` is a `Change` object:

```json
{
  "id": "chg_001",
  "rule_id": "FIELD_REMOVED",
  "severity": "BREAKING",
  "path": "components.schemas.Customer.properties.email",
  "description": "Field 'email' was removed from schema 'Customer'.",
  "before": {
    "type": "string",
    "format": "email",
    "example": "user@example.com"
  },
  "after": null,
  "recommendation": "Add a deprecation period before removing this field. Use 'deprecated: true' and notify consumers."
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `id` | `string` | ✅ | Unique identifier for this change within the report (`chg_001`, `chg_002`, etc.) |
| `rule_id` | `RuleID` | ✅ | The rule that was triggered (see Breaking Change Rules Spec) |
| `severity` | `ChangeSeverity` | ✅ | Severity of this specific change |
| `path` | `string` | ✅ | Dot-notation path to the changed element in the schema tree |
| `description` | `string` | ✅ | Human-readable description of what changed |
| `before` | `any \| null` | ✅ | The value before the change. `null` if the element was newly added. |
| `after` | `any \| null` | ✅ | The value after the change. `null` if the element was removed. |
| `recommendation` | `string \| null` | ✅ | Actionable advice for the developer. `null` for safe changes. |

### ChangeSeverity Enum

| Value | Placed In Array | Meaning |
|---|---|---|
| `BREAKING` | `breaking_changes` | Will break downstream consumers |
| `WARNING` | `warnings` | May break downstream consumers depending on their implementation |
| `SAFE` | `safe_changes` | Confirmed to not break any downstream consumers |

---

## ComplianceAlert Object

Each entry in `compliance_alerts` is a `ComplianceAlert` object:

```json
{
  "path": "components.schemas.Customer.properties.ssn",
  "compliance_type": "PII:SSN",
  "message": "Detected field matching PII:SSN pattern"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `path` | `string` | ✅ | Dot-notation path to the matched element in the schema tree |
| `compliance_type` | `string` | ✅ | The matched compliance type (e.g., `PII:SSN`, `PCI:PAYMENT`) |
| `message` | `string` | ✅ | Human-readable explanation of the alert |

---

## Path Convention

The `path` field uses **dot-notation** relative to the schema document root:

| Schema Type | Example Path |
|---|---|
| OpenAPI — removed field | `components.schemas.Customer.properties.email` |
| OpenAPI — changed response code | `paths./customers/{id}.get.responses.200` |
| OpenAPI — required field added | `components.schemas.CreateOrderRequest.required` |
| OpenAPI — endpoint removed | `paths./customers/{id}` |
| SQL — dropped column | `tables.customers.columns.email` |
| SQL — changed column type | `tables.orders.columns.total_amount.type` |

---

## Complete Example

```json
{
  "substrate_version": "0.1.0",
  "schema_type": "openapi",
  "compared_at": "2026-07-07T01:00:00Z",
  "summary": {
    "total_changes": 3,
    "breaking_count": 1,
    "warning_count": 1,
    "safe_count": 1,
    "overall_severity": "BREAKING"
  },
  "breaking_changes": [
    {
      "id": "chg_001",
      "rule_id": "FIELD_REMOVED",
      "severity": "BREAKING",
      "path": "components.schemas.Customer.properties.email",
      "description": "Field 'email' was removed from schema 'Customer'.",
      "before": { "type": "string", "format": "email" },
      "after": null,
      "recommendation": "Add 'deprecated: true' to the field for one release cycle before removing it. Coordinate with consumers: billing-service, marketing-service."
    }
  ],
  "warnings": [
    {
      "id": "chg_002",
      "rule_id": "ENUM_VALUE_ADDED",
      "severity": "WARNING",
      "path": "components.schemas.OrderStatus.enum",
      "description": "New enum value 'DISPUTED' was added to 'OrderStatus'.",
      "before": ["PENDING", "COMPLETED", "CANCELLED"],
      "after": ["PENDING", "COMPLETED", "CANCELLED", "DISPUTED"],
      "recommendation": "Consumers using exhaustive switch/match statements on this enum may fail on the new value. Verify downstream handlers."
    }
  ],
  "safe_changes": [
    {
      "id": "chg_003",
      "rule_id": "OPTIONAL_FIELD_ADDED",
      "severity": "SAFE",
      "path": "components.schemas.Customer.properties.phone",
      "description": "Optional field 'phone' was added to schema 'Customer'.",
      "before": null,
      "after": { "type": "string", "nullable": true },
      "recommendation": null
    }
  ]
}
```

---

## CLI Command Syntax

The engine CLI binary is named `substrate`. All commands follow this structure:

```
substrate <command> [arguments] [flags]
```

### Commands

| Command | Arguments | Description |
|---|---|---|
| `diff` | `<base-spec> <revision-spec>` | Compare two schema files and output a DiffReport |
| `validate` | `<spec-file>` | Validate a single spec file for errors (no diff) |

### Flags (apply to `diff`)

| Flag | Values | Default | Description |
|---|---|---|---|
| `--format` | `json`, `text`, `changelog` | `json` | Output format |
| `--config` | file path | `./substrate.yaml` | Path to override config file |
| `--flatten-allof` | boolean | `true` | Merge `allOf` schemas before diffing |
| `--schema-type` | `openapi`, `sql`, `graphql` | auto-detect | Force schema type (skips auto-detection) |

### Examples

```bash
# Standard CI usage — outputs JSON, exits 2 if breaking
substrate diff base.yaml pr.yaml

# Human-readable output for local dev
substrate diff base.yaml pr.yaml --format text

# Changelog format for PR comment generation
substrate diff base.yaml pr.yaml --format changelog

# Validate a single spec before diffing
substrate validate openapi.yaml

# Use a custom config file location
substrate diff base.yaml pr.yaml --config ./config/substrate.yaml
```

---

## CLI Output Modes

The engine CLI supports three output modes controlled by `--format`:

### `--format json` (default)
Output the raw `DiffReport` JSON to stdout. Used by the GitHub App and CI pipelines.

### `--format text`
Human-readable summary for local development. Example:

```
Substrate Diff Report
─────────────────────
Schema Type:  openapi
Analyzed At:  2026-07-07T01:00:00Z

Summary: 3 changes — 1 BREAKING, 1 WARNING, 1 SAFE

❌ BREAKING CHANGES (1)
  [chg_001] FIELD_REMOVED
  Path: components.schemas.Customer.properties.email
  Description: Field 'email' was removed from schema 'Customer'.
  Recommendation: Add 'deprecated: true' before removing.

⚠️  WARNINGS (1)
  [chg_002] ENUM_VALUE_ADDED
  Path: components.schemas.OrderStatus.enum
  Description: New enum value 'DISPUTED' was added.
  Recommendation: Check exhaustive switch statements in consumers.

✅ SAFE CHANGES (1)
  [chg_003] OPTIONAL_FIELD_ADDED
  Path: components.schemas.Customer.properties.phone
  Description: Optional field 'phone' was added.
```

### `--format changelog`
Markdown-formatted output designed to be pasted directly into a GitHub PR comment. Example:

```markdown
## 🔍 Substrate Schema Report

**Schema:** `openapi` | **Changes:** 3 (1 breaking, 1 warning, 1 safe)

### ❌ Breaking Changes
| ID | Rule | Path | Description |
|---|---|---|---|
| chg_001 | `FIELD_REMOVED` | `schemas.Customer.email` | Field 'email' was removed |

> **Recommendation:** Add `deprecated: true` for one release cycle before removing.

### ⚠️ Warnings
| ID | Rule | Path | Description |
|---|---|---|---|
| chg_002 | `ENUM_VALUE_ADDED` | `schemas.OrderStatus.enum` | New value 'DISPUTED' added |

### ✅ Safe Changes
| ID | Rule | Path | Description |
|---|---|---|---|
| chg_003 | `FIELD_ADDED_OPTIONAL` | `schemas.Customer.phone` | Optional field 'phone' added |

---
*Generated by [Substrate](https://github.com/KrushnaVardhanReddy/substrate) — schema contract protection*
```

---

## Exit Codes

| Code | Condition |
|---|---|
| `0` | No changes, or all changes are SAFE |
| `1` | At least one WARNING exists (no breaking changes) |
| `2` | At least one BREAKING change exists |
| `3` | Error — invalid input file, unsupported schema type, or parse failure |

Exit code `2` is what CI pipelines use to **block the merge**.

---

## Next Step

> Both downstream specs are now complete:
> - `docs/specs/breaking-change-rules.md` — Rule table ✅ (37 rules)
> - `docs/specs/override-config.md` — Override + repo registration config (P1-T00, in progress)
>
> Jules can be submitted with `t01_go_scaffold.txt` once `override-config.md` is written.
