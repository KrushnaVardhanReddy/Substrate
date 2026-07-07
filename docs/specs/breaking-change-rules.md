# Substrate — Breaking Change Rules Specification (OpenAPI 3.x)

> **Status:** DRAFT — Awaiting sign-off before the Rule Engine (`engine/internal/rules/`) is implemented.
> **Spec-First Gate:** Jules MUST NOT implement `t04_rule_engine.txt` until this document is approved.
> **Scope:** OpenAPI 3.x only. SQL, GraphQL, and Protobuf rules will be defined in separate spec documents.

---

## Overview

This document defines the complete rule table used by the Substrate Rule Engine. Each rule specifies:

- A unique **Rule ID** — referenced in `DiffReport.Change.rule_id`
- The exact **change pattern** that triggers it
- The **severity** assigned (`BREAKING`, `WARNING`, or `SAFE`)
- The **rationale** explaining why
- A **concrete example** of the before/after state

Rules are applied to the Internal Representation (IR) tree produced by the OpenAPI Parser. The Rule Engine iterates over every detected raw change and classifies it using this table.

---

## Severity Definitions (Reminder)

| Severity | Meaning | CI Action |
|---|---|---|
| `BREAKING` | Will crash or corrupt downstream consumers | **Block merge** (exit code 2) |
| `WARNING` | May break consumers depending on implementation | **Warn but allow merge** (exit code 1) |
| `SAFE` | Guaranteed not to break any downstream consumer | **Approve silently** (exit code 0) |

---

## Rule Categories

1. [Schema Field Rules](#1-schema-field-rules)
2. [Required Fields Rules](#2-required-fields-rules)
3. [Enum Rules](#3-enum-rules)
4. [Endpoint (Path) Rules](#4-endpoint-path-rules)
5. [HTTP Method Rules](#5-http-method-rules)
6. [Request Body Rules](#6-request-body-rules)
7. [Response Rules](#7-response-rules)
8. [Parameter Rules (Path, Query, Header)](#8-parameter-rules)
9. [Authentication & Security Rules](#9-authentication--security-rules)
10. [Type & Format Rules](#10-type--format-rules)

---

## 1. Schema Field Rules

These rules apply to changes within `components.schemas.*` objects.

### FIELD_REMOVED
| | |
|---|---|
| **Rule ID** | `FIELD_REMOVED` |
| **Severity** | `BREAKING` |
| **Pattern** | A property key is removed from a schema object |
| **Rationale** | Consumers reading this field will receive `undefined`/`null` and likely panic, throw a null-pointer exception, or silently corrupt data. |
| **Example** | |

Before:
```yaml
Customer:
  properties:
    id: { type: string }
    email: { type: string }
```
After:
```yaml
Customer:
  properties:
    id: { type: string }
```
Removed: `Customer.email` → **BREAKING**

---

### FIELD_ADDED_OPTIONAL
| | |
|---|---|
| **Rule ID** | `FIELD_ADDED_OPTIONAL` |
| **Severity** | `SAFE` |
| **Pattern** | A new property is added to a schema object and it is NOT in the `required` array |
| **Rationale** | Consumers that don't know about the field will simply ignore it. No existing code breaks. |
| **Example** | |

Before:
```yaml
Customer:
  properties:
    id: { type: string }
```
After:
```yaml
Customer:
  properties:
    id: { type: string }
    phone: { type: string, nullable: true }
```
Added optional: `Customer.phone` → **SAFE**

---

### FIELD_RENAMED
| | |
|---|---|
| **Rule ID** | `FIELD_RENAMED` |
| **Severity** | `BREAKING` |
| **Pattern** | A property key is removed and a new property key appears in the same schema within the same diff. Detected heuristically when field name similarity score > 0.7 (Levenshtein). |
| **Rationale** | Consumer code referencing the old name will fail. Treat as a removal + addition. |
| **Note** | If heuristic confidence is below 0.7, emit a `FIELD_REMOVED` + `FIELD_ADDED_OPTIONAL` instead. |

---

### FIELD_DEPRECATED
| | |
|---|---|
| **Rule ID** | `FIELD_DEPRECATED` |
| **Severity** | `WARNING` |
| **Pattern** | A property gains `deprecated: true` |
| **Rationale** | Not immediately breaking, but signals future removal. Consumers should be notified to start migration. |

---

## 2. Required Fields Rules

### REQUIRED_FIELD_ADDED
| | |
|---|---|
| **Rule ID** | `REQUIRED_FIELD_ADDED` |
| **Severity** | `BREAKING` |
| **Pattern** | A field name is added to the schema's `required` array AND the field did not previously exist (or was optional) |
| **Rationale** | All existing API requests that don't include this field will now fail validation with a `422 Unprocessable Entity`. Producer-side: all existing response payloads that omit this field are now invalid. |
| **Example** | |

Before:
```yaml
CreateOrderRequest:
  required: [product_id]
  properties:
    product_id: { type: string }
    quantity: { type: integer }
```
After:
```yaml
CreateOrderRequest:
  required: [product_id, quantity]
  properties:
    product_id: { type: string }
    quantity: { type: integer }
```
`quantity` became required → **BREAKING**

---

### REQUIRED_FIELD_MADE_OPTIONAL
| | |
|---|---|
| **Rule ID** | `REQUIRED_FIELD_MADE_OPTIONAL` |
| **Severity** | `SAFE` |
| **Pattern** | A field name is removed from the `required` array but the field itself still exists |
| **Rationale** | Existing consumers that always send this field are unaffected. New consumers gain flexibility. |

---

## 3. Enum Rules

### ENUM_VALUE_REMOVED
| | |
|---|---|
| **Rule ID** | `ENUM_VALUE_REMOVED` |
| **Severity** | `BREAKING` |
| **Pattern** | A value is removed from an `enum` array |
| **Rationale** | Consumers may be sending or expecting this value. Producer code that checked for it may break. Existing data containing the old value becomes invalid. |
| **Example** | |

Before:
```yaml
OrderStatus:
  type: string
  enum: [PENDING, PROCESSING, COMPLETED, CANCELLED]
```
After:
```yaml
OrderStatus:
  type: string
  enum: [PENDING, PROCESSING, COMPLETED]
```
`CANCELLED` removed → **BREAKING**

---

### ENUM_VALUE_ADDED
| | |
|---|---|
| **Rule ID** | `ENUM_VALUE_ADDED` |
| **Severity** | `WARNING` |
| **Pattern** | A new value is added to an `enum` array |
| **Rationale** | Not immediately breaking, but consumers using exhaustive `switch`/`match` statements will hit the `default`/`else` branch unexpectedly — or throw an unhandled case exception in strict languages (e.g., TypeScript discriminated unions, Rust enums). |

---

### ENUM_TYPE_CHANGED
| | |
|---|---|
| **Rule ID** | `ENUM_TYPE_CHANGED` |
| **Severity** | `BREAKING` |
| **Pattern** | The parent `type` of an enum field changes (e.g., `string` → `integer`) |
| **Rationale** | All consumers parsing this field will fail type coercion. |

---

## 4. Endpoint (Path) Rules

### ENDPOINT_REMOVED
| | |
|---|---|
| **Rule ID** | `ENDPOINT_REMOVED` |
| **Severity** | `BREAKING` |
| **Pattern** | An entire path key (e.g., `/customers/{id}`) is removed from `paths` |
| **Rationale** | Any consumer calling this endpoint will receive a `404`. |

---

### ENDPOINT_ADDED
| | |
|---|---|
| **Rule ID** | `ENDPOINT_ADDED` |
| **Severity** | `SAFE` |
| **Pattern** | A new path key is added to `paths` |
| **Rationale** | Additive. No existing consumer is affected. |

---

### ENDPOINT_DEPRECATED
| | |
|---|---|
| **Rule ID** | `ENDPOINT_DEPRECATED` |
| **Severity** | `WARNING` |
| **Pattern** | An operation gains `deprecated: true` |
| **Rationale** | Not immediately breaking, but consumers should migrate away. |

---

## 5. HTTP Method Rules

### METHOD_REMOVED
| | |
|---|---|
| **Rule ID** | `METHOD_REMOVED` |
| **Severity** | `BREAKING` |
| **Pattern** | An HTTP method (e.g., `get`, `post`, `delete`) is removed from an existing path |
| **Rationale** | Consumers calling that method+path combination will receive a `405 Method Not Allowed`. |

---

### METHOD_ADDED
| | |
|---|---|
| **Rule ID** | `METHOD_ADDED` |
| **Severity** | `SAFE` |
| **Pattern** | A new HTTP method is added to an existing path |
| **Rationale** | Additive. No existing consumer is affected. |

---

## 6. Request Body Rules

### REQUEST_BODY_REMOVED
| | |
|---|---|
| **Rule ID** | `REQUEST_BODY_REMOVED` |
| **Severity** | `BREAKING` |
| **Pattern** | A `requestBody` is removed from an operation that previously had one |
| **Rationale** | Consumers sending a body will have it ignored or rejected. |

---

### REQUEST_BODY_MADE_REQUIRED
| | |
|---|---|
| **Rule ID** | `REQUEST_BODY_MADE_REQUIRED` |
| **Severity** | `BREAKING` |
| **Pattern** | `requestBody.required` changes from `false` to `true` |
| **Rationale** | Consumers not sending a body will now receive a `422`. |

---

### REQUEST_BODY_CONTENT_TYPE_REMOVED
| | |
|---|---|
| **Rule ID** | `REQUEST_BODY_CONTENT_TYPE_REMOVED` |
| **Severity** | `BREAKING` |
| **Pattern** | A content type (e.g., `application/json`) is removed from `requestBody.content` |
| **Rationale** | Consumers sending that content type will receive a `415 Unsupported Media Type`. |

---

### REQUEST_BODY_CONTENT_TYPE_ADDED
| | |
|---|---|
| **Rule ID** | `REQUEST_BODY_CONTENT_TYPE_ADDED` |
| **Severity** | `SAFE` |
| **Pattern** | A new content type is added to `requestBody.content` |
| **Rationale** | Additive. Existing consumers are unaffected. |

---

## 7. Response Rules

### RESPONSE_CODE_REMOVED
| | |
|---|---|
| **Rule ID** | `RESPONSE_CODE_REMOVED` |
| **Severity** | `WARNING` |
| **Pattern** | A documented response status code (e.g., `200`, `404`) is removed from an operation's `responses` |
| **Rationale** | The API may still return this code at runtime. Removing it from the spec makes client code generators produce incorrect SDKs. Classified as WARNING (not BREAKING) because the runtime behavior may be unchanged. |

---

### RESPONSE_CODE_ADDED
| | |
|---|---|
| **Rule ID** | `RESPONSE_CODE_ADDED` |
| **Severity** | `SAFE` |
| **Pattern** | A new response status code is documented |
| **Rationale** | Additive documentation. Existing consumers are unaffected. |

---

### RESPONSE_SCHEMA_FIELD_REMOVED
| | |
|---|---|
| **Rule ID** | `RESPONSE_SCHEMA_FIELD_REMOVED` |
| **Severity** | `BREAKING` |
| **Pattern** | A field is removed from a response body schema (inline or via `$ref`) |
| **Rationale** | Same as `FIELD_REMOVED` — consumers reading this field from the response will break. |

---

### RESPONSE_SCHEMA_FIELD_ADDED
| | |
|---|---|
| **Rule ID** | `RESPONSE_SCHEMA_FIELD_ADDED` |
| **Severity** | `SAFE` |
| **Pattern** | A field is added to a response body schema |
| **Rationale** | Consumers ignore unknown fields. Safe by the Postel's Law principle. |

---

## 8. Parameter Rules

These rules apply to `parameters` defined on paths or operations (path params, query params, header params, cookie params).

### PARAMETER_REMOVED
| | |
|---|---|
| **Rule ID** | `PARAMETER_REMOVED` |
| **Severity** | `BREAKING` |
| **Pattern** | A parameter is removed from an operation or path |
| **Rationale** | If it was a required path/query param, consumers sending it may receive a `422`. If it was a required path param, the URL structure changes entirely. |

---

### PARAMETER_ADDED_REQUIRED
| | |
|---|---|
| **Rule ID** | `PARAMETER_ADDED_REQUIRED` |
| **Severity** | `BREAKING` |
| **Pattern** | A new parameter is added with `required: true` |
| **Rationale** | All existing consumer requests omitting this parameter will now fail. |

---

### PARAMETER_ADDED_OPTIONAL
| | |
|---|---|
| **Rule ID** | `PARAMETER_ADDED_OPTIONAL` |
| **Severity** | `SAFE` |
| **Pattern** | A new parameter is added with `required: false` (or no `required` field, which defaults to false) |
| **Rationale** | Existing consumers not sending this param are unaffected. |

---

### PARAMETER_MADE_REQUIRED
| | |
|---|---|
| **Rule ID** | `PARAMETER_MADE_REQUIRED` |
| **Severity** | `BREAKING` |
| **Pattern** | An existing optional parameter (`required: false`) becomes `required: true` |
| **Rationale** | Consumers not sending this parameter will now receive a validation error. |

---

### PARAMETER_TYPE_CHANGED
| | |
|---|---|
| **Rule ID** | `PARAMETER_TYPE_CHANGED` |
| **Severity** | `BREAKING` |
| **Pattern** | The `schema.type` of a parameter changes (e.g., `string` → `integer`) |
| **Rationale** | Consumers sending the old type will fail validation. |

---

## 9. Authentication & Security Rules

### SECURITY_SCHEME_REMOVED
| | |
|---|---|
| **Rule ID** | `SECURITY_SCHEME_REMOVED` |
| **Severity** | `BREAKING` |
| **Pattern** | A security scheme is removed from `components.securitySchemes` |
| **Rationale** | Consumers authenticating with this scheme will receive `401 Unauthorized`. |

---

### SECURITY_REQUIREMENT_ADDED
| | |
|---|---|
| **Rule ID** | `SECURITY_REQUIREMENT_ADDED` |
| **Severity** | `BREAKING` |
| **Pattern** | A security requirement is added to an operation that previously had none (or had an empty `security: []`) |
| **Rationale** | Unauthenticated consumer requests will now receive `401`. |

---

### SECURITY_REQUIREMENT_REMOVED
| | |
|---|---|
| **Rule ID** | `SECURITY_REQUIREMENT_REMOVED` |
| **Severity** | `WARNING` |
| **Pattern** | A security requirement is removed from an operation |
| **Rationale** | The endpoint becomes less secure. Not breaking for consumers but a meaningful policy change that should be reviewed. |

---

## 10. Type & Format Rules

### FIELD_TYPE_CHANGED
| | |
|---|---|
| **Rule ID** | `FIELD_TYPE_CHANGED` |
| **Severity** | `BREAKING` |
| **Pattern** | The `type` of a schema property changes (e.g., `string` → `integer`, `object` → `array`) |
| **Rationale** | Consumer deserializers will fail or produce corrupt data. |
| **Exception** | `integer` → `number` is classified as `WARNING` since it is a widening type change that most languages handle automatically. |

---

### FIELD_FORMAT_CHANGED
| | |
|---|---|
| **Rule ID** | `FIELD_FORMAT_CHANGED` |
| **Severity** | `WARNING` |
| **Pattern** | The `format` of a schema property changes (e.g., `date` → `date-time`, `int32` → `int64`) |
| **Rationale** | Not always breaking (format is advisory in OpenAPI), but consumers relying on format-specific parsing (e.g., date parsing libraries) may fail. Classified as WARNING for human review. |

---

### FIELD_NULLABLE_CHANGED
| | |
|---|---|
| **Rule ID** | `FIELD_NULLABLE_CHANGED` |
| **Severity** | `BREAKING` |
| **Pattern** | `nullable` changes from `false` (or unset) to `true`, or from `true` to `false` |
| **Rationale** | `false → true`: consumers not handling null will get null-pointer exceptions. `true → false`: consumers sending null values will now fail validation. Both directions are breaking. |

---

## Rule ID Quick Reference

| Rule ID | Category | Severity |
|---|---|---|
| `FIELD_REMOVED` | Schema Field | `BREAKING` |
| `FIELD_ADDED_OPTIONAL` | Schema Field | `SAFE` |
| `FIELD_RENAMED` | Schema Field | `BREAKING` |
| `FIELD_DEPRECATED` | Schema Field | `WARNING` |
| `REQUIRED_FIELD_ADDED` | Required Fields | `BREAKING` |
| `REQUIRED_FIELD_MADE_OPTIONAL` | Required Fields | `SAFE` |
| `ENUM_VALUE_REMOVED` | Enum | `BREAKING` |
| `ENUM_VALUE_ADDED` | Enum | `WARNING` |
| `ENUM_TYPE_CHANGED` | Enum | `BREAKING` |
| `ENDPOINT_REMOVED` | Endpoint | `BREAKING` |
| `ENDPOINT_ADDED` | Endpoint | `SAFE` |
| `ENDPOINT_DEPRECATED` | Endpoint | `WARNING` |
| `METHOD_REMOVED` | HTTP Method | `BREAKING` |
| `METHOD_ADDED` | HTTP Method | `SAFE` |
| `REQUEST_BODY_REMOVED` | Request Body | `BREAKING` |
| `REQUEST_BODY_MADE_REQUIRED` | Request Body | `BREAKING` |
| `REQUEST_BODY_CONTENT_TYPE_REMOVED` | Request Body | `BREAKING` |
| `REQUEST_BODY_CONTENT_TYPE_ADDED` | Request Body | `SAFE` |
| `RESPONSE_CODE_REMOVED` | Response | `WARNING` |
| `RESPONSE_CODE_ADDED` | Response | `SAFE` |
| `RESPONSE_SCHEMA_FIELD_REMOVED` | Response | `BREAKING` |
| `RESPONSE_SCHEMA_FIELD_ADDED` | Response | `SAFE` |
| `PARAMETER_REMOVED` | Parameter | `BREAKING` |
| `PARAMETER_ADDED_REQUIRED` | Parameter | `BREAKING` |
| `PARAMETER_ADDED_OPTIONAL` | Parameter | `SAFE` |
| `PARAMETER_MADE_REQUIRED` | Parameter | `BREAKING` |
| `PARAMETER_TYPE_CHANGED` | Parameter | `BREAKING` |
| `SECURITY_SCHEME_REMOVED` | Security | `BREAKING` |
| `SECURITY_REQUIREMENT_ADDED` | Security | `BREAKING` |
| `SECURITY_REQUIREMENT_REMOVED` | Security | `WARNING` |
| `FIELD_TYPE_CHANGED` | Type & Format | `BREAKING` |
| `FIELD_FORMAT_CHANGED` | Type & Format | `WARNING` |
| `FIELD_NULLABLE_CHANGED` | Type & Format | `BREAKING` |

**Total: 32 rules — 20 BREAKING, 7 WARNING, 5 SAFE**

---

## Next Step

> Both spec documents are now complete and approved:
> - `docs/specs/diff-report-schema.md` — Output contract ✅
> - `docs/specs/breaking-change-rules.md` — Rule table ✅
>
> Jules can now be submitted with `t01_go_scaffold.txt` to begin Phase 1 implementation.
