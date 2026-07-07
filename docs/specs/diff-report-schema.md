# Substrate — DiffReport JSON Schema Specification

> **Status:** DRAFT — Awaiting sign-off before any Go implementation begins.
> **Spec-First Gate:** The `engine/` Go code MUST NOT be written until this document is approved.

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
  "safe_changes": []
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

## CLI Output Modes

The engine CLI must support two output modes controlled by a `--format` flag:

### `--format json` (default)
Output the raw `DiffReport` JSON to stdout. Useful for piping into CI tools.

### `--format text`
Output a human-readable summary to stdout. Example:

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

> With this contract agreed upon, the next document to write is:
> `docs/specs/breaking-change-rules.md` — the rule table that defines exactly which change patterns trigger `FIELD_REMOVED`, `ENUM_VALUE_ADDED`, etc.
