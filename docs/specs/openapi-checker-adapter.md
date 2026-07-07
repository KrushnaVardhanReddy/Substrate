# Substrate — oasdiff Checker Adapter Specification (P1-T06)

> **Status:** APPROVED ✅
> **Spec-First Gate:** Jules MUST NOT implement P1-T06 until this document is approved.
> **Scope:** Replaces the shallow structural diff in `engine/internal/diff/openapi.go` with the oasdiff semantic `checker` package, enabling all 37 Substrate breaking-change rules to fire correctly.
> **Depends on:** `docs/specs/breaking-change-rules.md` (rule table — the authoritative rule-to-severity mapping)

---

## Problem Statement

The current `CompareOpenAPI` implementation in `engine/internal/diff/openapi.go` calls `diff.Get()` from the oasdiff `diff` package and manually inspects the raw structural diff tree. This produces two classes of inaccuracy:

**False Positives (over-reporting):**
- Any schema modification (e.g., adding an optional field `Customer.phone`) fires `FIELD_REMOVED` as BREAKING because the code treats any `SchemasDiff.Modified` entry as breaking — regardless of whether a field was removed, added, or just annotated.

**False Negatives (under-reporting):**
- `REQUIRED_FIELD_ADDED`, `ENUM_VALUE_REMOVED`, `METHOD_REMOVED`, `PARAMETER_ADDED_REQUIRED`, `RESPONSE_SCHEMA_FIELD_REMOVED`, and 20+ other rules from `breaking-change-rules.md` are completely undetected.

**Root cause:** The oasdiff library has two separate packages:
- `github.com/oasdiff/oasdiff/diff` — raw structural diff tree (what we currently use)
- `github.com/oasdiff/oasdiff/checker` — semantic breaking-change analyzer (what we should use)

The `checker` package applies all of oasdiff's 160+ rules and returns a typed `[]checker.BackwardCompatibilityError` list — exactly what maps to Substrate's `DiffReport`.

---

## Solution: Replace `diff.Get()` with `checker.CheckBackwardCompatibility()`

### New Function Signature

The function signature in `engine/internal/diff/openapi.go` remains unchanged:

```go
func CompareOpenAPI(basePath, revisionPath string, flattenAllOf bool) (*report.DiffReport, error)
```

The internal implementation is replaced entirely.

---

## Implementation Specification

### Step 1 — Load Specs (unchanged)

Use the existing `openapi3.NewLoader()` approach. No change here.

```go
loader := openapi3.NewLoader()
loader.IsExternalRefsAllowed = true

base, err := loader.LoadFromFile(basePath)
// ...
revision, err := loader.LoadFromFile(revisionPath)
```

### Step 2 — Pre-diff Spec Validation

Before calling the checker, validate both specs using `openapi3.Validate()`. This implements `SPEC_INVALID_REF`, `SPEC_INVALID_TYPE`, and `SPEC_MISSING_REQUIRED_FIELD` from `breaking-change-rules.md`.

```go
if err := base.Validate(context.Background()); err != nil {
    return nil, fmt.Errorf("base spec invalid: %w", err)  // exit code 3
}
if err := revision.Validate(context.Background()); err != nil {
    return nil, fmt.Errorf("revision spec invalid: %w", err)  // exit code 3
}
```

> **Exit code for validation failures:** The CLI must exit with code `3` (not `1` or `2`) when either spec fails validation, so upstream tooling can distinguish parse errors from breaking changes. Document this in the CLI layer (`cmd/substrate/main.go`).

### Step 3 — Compute Structural Diff (required by checker)

The `checker` package requires the raw `*diff.Diff` object as input. Call `diff.Get()` as before, but only to feed the checker — not to manually inspect results.

```go
diffConfig := diff.NewConfig()
if flattenAllOf {
    diffConfig = diffConfig.WithFlattenAllOf()
}

diffObj, err := diff.Get(diffConfig, base, revision)
if err != nil {
    return nil, fmt.Errorf("failed to compute diff: %w", err)
}
```

### Step 4 — Run Semantic Checker

```go
import "github.com/oasdiff/oasdiff/checker"

checkerConfig := checker.NewConfig()  // uses oasdiff defaults
errors, err := checker.CheckBackwardCompatibility(checkerConfig, diffObj, nil)
if err != nil {
    return nil, fmt.Errorf("checker error: %w", err)
}
```

`errors` is of type `checker.BackwardCompatibilityErrors` — a slice of `checker.BackwardCompatibilityError`.

Each `BackwardCompatibilityError` has:
- `.GetId()` — the oasdiff rule ID (e.g., `"response-success-status-removed"`)
- `.GetLevel()` — `checker.ERR` (breaking) or `checker.WARN` (warning)
- `.GetOperation()` — the HTTP method
- `.GetPath()` — the URL path
- `.GetText(lang)` — human-readable description

### Step 5 — Map oasdiff Errors → Substrate `DiffReport`

This is the core mapping logic. Each `BackwardCompatibilityError` maps to one `report.Change`.

#### Severity Mapping

| oasdiff Level | Substrate Severity |
|---|---|
| `checker.ERR` | `report.ChangeSeverityBreaking` |
| `checker.WARN` | `report.ChangeSeverityWarning` |
| `checker.INFO` | `report.ChangeSeveritySafe` |

#### Rule ID Mapping

oasdiff uses its own internal rule ID strings. Substrate exposes the rule IDs defined in `breaking-change-rules.md`. The adapter must translate between them.

Implement a `var oasdiffRuleMap = map[string]string{}` lookup table. Key = oasdiff rule ID, Value = Substrate rule ID. If an oasdiff rule ID has no mapping, use the oasdiff ID as a fallback (prefixed with `OASDIFF_` to make it identifiable).

**Mapping table (starter set — expand as oasdiff coverage is confirmed):**

| oasdiff Rule ID | Substrate Rule ID |
|---|---|
| `response-property-removed` | `FIELD_REMOVED` |
| `request-property-removed` | `FIELD_REMOVED` |
| `response-required-property-removed` | `FIELD_REMOVED` |
| `request-required-property-removed` | `FIELD_REMOVED` |
| `response-required-property-added` | `REQUIRED_FIELD_ADDED` |
| `request-required-property-added` | `REQUIRED_FIELD_ADDED` |
| `response-property-became-required` | `REQUIRED_FIELD_ADDED` |
| `request-property-became-required` | `REQUIRED_FIELD_ADDED` |
| `response-property-became-optional` | `REQUIRED_FIELD_MADE_OPTIONAL` |
| `request-property-became-optional` | `REQUIRED_FIELD_MADE_OPTIONAL` |
| `response-optional-property-added` | `FIELD_ADDED_OPTIONAL` |
| `request-optional-property-added` | `FIELD_ADDED_OPTIONAL` |
| `response-enum-value-removed` | `ENUM_VALUE_REMOVED` |
| `request-enum-value-removed` | `ENUM_VALUE_REMOVED` |
| `response-enum-value-added` | `ENUM_VALUE_ADDED` |
| `request-enum-value-added` | `ENUM_VALUE_ADDED` |
| `endpoint-removed` | `ENDPOINT_REMOVED` |
| `endpoint-added` | `ENDPOINT_ADDED` |
| `api-path-removed-without-deprecation` | `ENDPOINT_REMOVED` |
| `endpoint-deprecated` | `ENDPOINT_DEPRECATED` |
| `http-method-removed` | `METHOD_REMOVED` |
| `request-body-removed` | `REQUEST_BODY_REMOVED` |
| `request-body-became-required` | `REQUEST_BODY_MADE_REQUIRED` |
| `request-body-media-type-removed` | `REQUEST_BODY_CONTENT_TYPE_REMOVED` |
| `request-body-media-type-added` | `REQUEST_BODY_CONTENT_TYPE_ADDED` |
| `response-status-removed` | `RESPONSE_CODE_REMOVED` |
| `response-status-added` | `RESPONSE_CODE_ADDED` |
| `response-body-removed` | `RESPONSE_SCHEMA_FIELD_REMOVED` |
| `api-parameter-removed` | `PARAMETER_REMOVED` |
| `required-api-parameter-added` | `PARAMETER_ADDED_REQUIRED` |
| `api-parameter-added-with-default-value` | `PARAMETER_ADDED_OPTIONAL` |
| `api-parameter-became-required` | `PARAMETER_MADE_REQUIRED` |
| `request-parameter-type-changed` | `PARAMETER_TYPE_CHANGED` |
| `response-property-type-changed` | `FIELD_TYPE_CHANGED` |
| `request-property-type-changed` | `FIELD_TYPE_CHANGED` |
| `response-property-format-changed` | `FIELD_FORMAT_CHANGED` |
| `request-property-format-changed` | `FIELD_FORMAT_CHANGED` |
| `api-security-removed` | `SECURITY_SCHEME_REMOVED` |
| `api-operation-security-added` | `SECURITY_REQUIREMENT_ADDED` |
| `api-operation-security-removed` | `SECURITY_REQUIREMENT_REMOVED` |
| `response-property-became-nullable` | `FIELD_NULLABLE_CHANGED` |
| `request-property-became-nullable` | `FIELD_NULLABLE_CHANGED` |

> **Note:** This mapping table will need to be verified against the actual oasdiff rule IDs at implementation time. Use `oasdiff diff --fail-on ERR --format text` against test fixtures to discover the exact strings oasdiff emits. Update the table during implementation if the strings differ. Do **not** hardcode oasdiff internal constants — use string comparison so the mapping remains decoupled.

#### Path Construction

The `report.Change.Path` field uses dot-notation (e.g., `paths./customers/{id}.get.requestBody`). Build it from the `BackwardCompatibilityError` fields:

```go
path := fmt.Sprintf("paths.%s.%s", err.GetPath(), strings.ToLower(err.GetOperation()))
```

Add schema component path if available from the error's source context.

#### Change ID Construction

```go
id := fmt.Sprintf("chg_%s_%s_%s",
    strings.ToLower(substrateRuleID),
    strings.ReplaceAll(err.GetPath(), "/", "_"),
    strings.ToLower(err.GetOperation()),
)
```

### Step 6 — Build `DiffReport`

```go
rep := &report.DiffReport{
    SubstrateVersion: "0.1.0",
    SchemaType:       report.SchemaTypeOpenAPI,
    ComparedAt:       time.Now().UTC().Format(time.RFC3339),
    BreakingChanges:  []report.Change{},
    Warnings:         []report.Change{},
    SafeChanges:      []report.Change{},
}

for _, e := range errors {
    change := mapCheckerError(e)
    switch change.Severity {
    case report.ChangeSeverityBreaking:
        rep.BreakingChanges = append(rep.BreakingChanges, change)
    case report.ChangeSeverityWarning:
        rep.Warnings = append(rep.Warnings, change)
    default:
        rep.SafeChanges = append(rep.SafeChanges, change)
    }
}

rep.Summary = computeSummary(rep)
```

---

## Test Requirements

The implementation MUST include tests in `engine/internal/diff/openapi_test.go` covering:

| Test Case | Input | Expected Output |
|---|---|---|
| No changes | identical specs | `DiffReport{Summary: {TotalChanges: 0, OverallSeverity: "NO_CHANGES"}}` |
| Optional field added | `Customer` gains `phone?: string` | `SafeChanges` contains `FIELD_ADDED_OPTIONAL` |
| Required field added | `CreateOrder.required` gains `quantity` | `BreakingChanges` contains `REQUIRED_FIELD_ADDED` |
| Field removed | `Customer.email` removed | `BreakingChanges` contains `FIELD_REMOVED` |
| Endpoint removed | `/customers/{id}` path deleted | `BreakingChanges` contains `ENDPOINT_REMOVED` |
| Method removed | `GET /customers` deleted, `POST` kept | `BreakingChanges` contains `METHOD_REMOVED` |
| Enum value removed | `OrderStatus.CANCELLED` removed | `BreakingChanges` contains `ENUM_VALUE_REMOVED` |
| Enum value added | `OrderStatus.REFUNDED` added | `Warnings` contains `ENUM_VALUE_ADDED` |
| Parameter made required | `?limit` becomes `required: true` | `BreakingChanges` contains `PARAMETER_MADE_REQUIRED` |
| Type changed | `Customer.id: string` → `integer` | `BreakingChanges` contains `FIELD_TYPE_CHANGED` |
| Invalid base spec | malformed YAML | error returned, no `DiffReport` |

Use the existing `engine/internal/diff/testdata/` directory. Add new fixture pairs (e.g., `base_optional_add.yaml` / `rev_optional_add.yaml`) following the existing naming convention.

---

## Files Changed

| File | Change |
|---|---|
| `engine/internal/diff/openapi.go` | Full replacement of implementation (signature unchanged) |
| `engine/internal/diff/openapi_test.go` | Add 10+ new test cases per table above |
| `engine/internal/diff/testdata/` | Add new fixture YAML pairs for each new test case |
| `engine/go.mod` / `engine/go.sum` | No changes expected — `oasdiff` already in deps |

---

## Out of Scope for P1-T06

The following are **explicitly excluded** from this task to keep scope tight:

- `FIELD_RENAMED` heuristic detection (Levenshtein similarity) — future task
- `STABILITY_LEVEL_CHANGED` detection — requires `x-stability` extension handling — future task
- Overrides integration (already implemented in P1-T03 — checker output feeds into existing override matching)
- SQL, GraphQL, Protobuf rules — separate sub-phases (1b–1d)

---

## Jules Prompt Location

`prompts/phase-1-diff-engine/t06_checker_adapter.txt`

> This spec is the source of truth. Jules must implement **exactly** what is described here.
> The mapping table in Step 5 must be verified at implementation time against real oasdiff output.
> All 10 test cases in the Test Requirements table are mandatory — Jules must not skip any.

---

## Next Step

> Once this spec is approved and P1-T06 is complete:
> - **P1-T05** (`NOTICES` file) — write Apache 2.0 attribution file
> - **P1b-T01** (`breaking-change-rules-sql.md`) — begin SQL Migration spec
> - **P1-MVP-4** (\"Optic Alternative\" blog post) — marketing launch
