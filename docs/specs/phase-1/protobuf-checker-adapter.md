# Substrate — Protobuf & gRPC Checker Adapter Specification (Phase 1d)

> **Status:** APPROVED ✅ — Implemented & merged (PR #17)
> **Spec-First Gate:** Jules MUST NOT implement Phase 1d until this document is approved.
> **Scope:** Add Protobuf & gRPC breaking-change detection to the Substrate diff engine using `bufbuild/buf` as the underlying checker. Follows the identical adapter pattern established by the oasdiff checker adapter (P1-T06).
> **Depends on:** `docs/specs/phase-1/diff-report-schema.md` (the DiffReport contract)

---

## Problem Statement

Protobuf + gRPC is the dominant schema format for internal microservice APIs at scale. A single breaking Protobuf change (removing a field, changing a field number, renaming an enum value) can silently corrupt serialized data or crash clients on the wire with no runtime error until deserialization fails.

Substrate currently only supports OpenAPI and SQL. Adding Protobuf support lets Substrate protect every gRPC service that uses `.proto` files.

---

## Library Decision: `bufbuild/buf` ✅

**Library:** `buf` CLI (`github.com/bufbuild/buf`) — Apache 2.0 license.

**Rationale:**
- `buf breaking` is the `oasdiff` of Protobuf — it enforces 100+ wire-compatibility rules covering field numbers, type changes, service/method removals, and message renames.
- The Substrate adapter pattern is: shell out to `buf breaking`, capture stdout/stderr as JSON, parse into `DiffReport`. This mirrors how the CI ecosystem uses `buf`.
- Do NOT import buf as a Go library dependency — it is designed to be invoked as a binary (`buf breaking --against ...`). The binary is self-contained and available via `go install github.com/bufbuild/buf/cmd/buf@latest`.
- **Single-binary promise:** The Substrate engine binary does NOT need to bundle buf. Instead, the adapter checks for `buf` on PATH and fails gracefully with a clear error message if not found.

---

## buf Breaking Change Rules → Substrate Severity Mapping

`buf breaking` outputs violations as structured JSON. Each violation has a `type` field that maps to a Substrate rule.

### Severity Classification

| Severity | buf Category | Description |
|---|---|---|
| `BREAKING` | `FILE`, `WIRE`, `WIRE_JSON` | Changes that break wire compatibility |
| `WARNING` | `PACKAGE`, `SOURCE_BREAK` | Package-level renames that may break source but not wire |
| `SAFE` | — | No buf violations at this category |

### Rule Mapping Table

| buf Violation Type | Substrate Rule ID | Severity |
|---|---|---|
| `FIELD_SAME_TYPE` | `PROTO_FIELD_TYPE_CHANGED` | BREAKING |
| `FIELD_SAME_NUMBER` | `PROTO_FIELD_NUMBER_CHANGED` | BREAKING |
| `FIELD_SAME_NAME` | `PROTO_FIELD_RENAMED` | BREAKING |
| `FIELD_NO_DELETE` | `PROTO_FIELD_REMOVED` | BREAKING |
| `FIELD_SAME_LABEL` | `PROTO_FIELD_LABEL_CHANGED` | BREAKING |
| `FIELD_SAME_ONEOF` | `PROTO_FIELD_ONEOF_CHANGED` | BREAKING |
| `ENUM_NO_DELETE` | `PROTO_ENUM_REMOVED` | BREAKING |
| `ENUM_VALUE_NO_DELETE` | `PROTO_ENUM_VALUE_REMOVED` | BREAKING |
| `ENUM_VALUE_SAME_NUMBER` | `PROTO_ENUM_VALUE_NUMBER_CHANGED` | BREAKING |
| `ENUM_VALUE_SAME_NAME` | `PROTO_ENUM_VALUE_RENAMED` | BREAKING |
| `MESSAGE_NO_DELETE` | `PROTO_MESSAGE_REMOVED` | BREAKING |
| `RPC_NO_DELETE` | `PROTO_RPC_REMOVED` | BREAKING |
| `RPC_SAME_REQUEST_TYPE` | `PROTO_RPC_REQUEST_TYPE_CHANGED` | BREAKING |
| `RPC_SAME_RESPONSE_TYPE` | `PROTO_RPC_RESPONSE_TYPE_CHANGED` | BREAKING |
| `RPC_SAME_CLIENT_STREAMING` | `PROTO_RPC_STREAMING_CHANGED` | BREAKING |
| `RPC_SAME_SERVER_STREAMING` | `PROTO_RPC_STREAMING_CHANGED` | BREAKING |
| `SERVICE_NO_DELETE` | `PROTO_SERVICE_REMOVED` | BREAKING |
| `FILE_SAME_PACKAGE` | `PROTO_PACKAGE_CHANGED` | WARNING |
| `FILE_NO_DELETE` | `PROTO_FILE_REMOVED` | BREAKING |
| `PACKAGE_ENUM_NO_DELETE` | `PROTO_ENUM_REMOVED` | BREAKING |
| `PACKAGE_MESSAGE_NO_DELETE` | `PROTO_MESSAGE_REMOVED` | BREAKING |
| `PACKAGE_SERVICE_NO_DELETE` | `PROTO_SERVICE_REMOVED` | BREAKING |
| *(any unmapped type)* | `PROTO_UNKNOWN_BREAK` | BREAKING |

> **Implementation note:** Unmapped violation types should be treated as BREAKING (safe default) and should use the raw buf type string as the rule ID with a `PROTO_` prefix. Log a warning when an unmapped type is encountered.

---

## `buf breaking` Command Contract

**Input:** Two directories (or zip archives) containing `.proto` files.

**Command format:**
```bash
buf breaking <head_dir> --against <base_dir> --error-format json
```

**JSON output format** (one line per violation, newline-delimited):
```json
{"path":"user/v1/user.proto","start_line":12,"start_column":3,"end_line":12,"end_column":15,"type":"FIELD_NO_DELETE","message":"Field \"email\" with name \"email\" on message \"User\" was deleted."}
```

**Exit codes:**
- `0` — no breaking changes detected
- `1` — breaking changes detected (violations in stdout)
- `1` — also used for `buf` errors (check stderr to distinguish)

---

## Implementation Specification

### New File: `engine/internal/diff/proto.go`

Function signature (matches existing diff pattern):

```go
// CompareProto compares two directories of .proto files and returns a DiffReport.
// baseDir and headDir are paths to directories containing .proto files.
// Returns an error if buf is not on PATH or if proto parsing fails.
func CompareProto(baseDir, headDir string) (*report.DiffReport, error)
```

**Implementation steps:**

#### Step 1 — Verify `buf` is available on PATH

```go
_, err := exec.LookPath("buf")
if err != nil {
    return nil, fmt.Errorf("buf is not installed or not on PATH — install with: go install github.com/bufbuild/buf/cmd/buf@latest")
}
```

#### Step 2 — Create temp directories for base and head `.proto` content

The function receives `baseDir` and `headDir` as paths to directories already on disk.
Write a `buf.yaml` workspace file into each if one doesn't already exist:

```yaml
# buf.yaml (minimal — written by Substrate if not present)
version: v2
```

Only write the `buf.yaml` if the file does not already exist in that directory.
This respects user-provided `buf.yaml` configurations.

#### Step 3 — Run `buf breaking`

```go
cmd := exec.Command("buf", "breaking", headDir, "--against", baseDir, "--error-format", "json")
var stdout, stderr bytes.Buffer
cmd.Stdout = &stdout
cmd.Stderr = &stderr
err := cmd.Run()
// Exit code 0 = no violations, exit code 1 = violations OR buf error
// Distinguish by checking if stdout has JSON or stderr has error text
```

If stderr is non-empty and stdout is empty → buf itself errored → return error.
If stdout has JSON lines → parse violations (even if exit code is 1).
If both empty and exit code 0 → no violations.

#### Step 4 — Parse buf JSON output into violations

```go
type bufViolation struct {
    Path        string `json:"path"`
    StartLine   int    `json:"start_line"`
    StartColumn int    `json:"start_column"`
    Type        string `json:"type"`
    Message     string `json:"message"`
}
```

Parse each newline-delimited JSON line from stdout. Skip blank lines.

#### Step 5 — Map violations → `report.Change`

Apply the rule mapping table above. For each violation:

```go
change := report.Change{
    ID:          fmt.Sprintf("proto_%s_%s_%d", strings.ToLower(violation.Type), slugify(violation.Path), violation.StartLine),
    RuleID:      mapBufType(violation.Type),   // from mapping table
    Severity:    mapBufSeverity(violation.Type), // BREAKING or WARNING
    Path:        fmt.Sprintf("%s:%d", violation.Path, violation.StartLine),
    Description: violation.Message,
    Recommendation: recommendationFor(violation.Type),
}
```

`recommendationFor()` returns a human-readable suggestion for each rule ID:
- `PROTO_FIELD_REMOVED` → "Do not remove fields. Mark as deprecated instead using `[deprecated=true]`."
- `PROTO_FIELD_TYPE_CHANGED` → "Field type changes break wire compatibility. Add a new field instead."
- `PROTO_FIELD_NUMBER_CHANGED` → "Field numbers are wire-protocol identifiers and must never change."
- `PROTO_RPC_REMOVED` → "Removing an RPC method breaks all clients. Deprecate it first."
- `PROTO_ENUM_VALUE_REMOVED` → "Removing enum values breaks clients that send or receive that value."
- `PROTO_SERVICE_REMOVED` → "Removing a service breaks all clients. Use a deprecation notice first."
- *(all others)* → "This change breaks wire or source compatibility. Review buf documentation for migration guidance."

#### Step 6 — Build `DiffReport`

```go
rep := &report.DiffReport{
    SubstrateVersion: "0.1.0",
    SchemaType:       "protobuf",
    ComparedAt:       time.Now().UTC().Format(time.RFC3339),
    BreakingChanges:  []report.Change{},
    Warnings:         []report.Change{},
    SafeChanges:      []report.Change{},
}
// distribute changes into correct bucket by severity
rep.Summary = computeSummary(rep)
```

---

## CLI Integration

### `substrate` CLI — New `schema_type: protobuf` support

The existing `substrate` CLI already dispatches based on `schema_type` in `substrate.yaml`.
Add a new case in the dispatch function in `engine/cmd/substrate/main.go`:

```go
case "protobuf", "proto":
    report, err = diff.CompareProto(baseSchema, headSchema)
```

Where `base_schema` and `head_schema` from `substrate.yaml` are interpreted as **directory paths** (not single files) for Protobuf.

### New `substrate.yaml` config format for Protobuf:

```yaml
service: payment-service
schema_type: protobuf
base_schema: proto/          # directory of .proto files on base branch
head_schema: proto/          # directory of .proto files on head branch
on_breaking_change: block
```

### GitHub Action — `schema_type: protobuf` support

The action already passes `base_schema` and `head_schema` through. No action.yml change needed.
The user sets `schema_type: protobuf` in their `substrate.yaml` and provides proto directory paths.

Users must also ensure `buf` is installed in the CI runner:
```yaml
- name: Install buf
  run: go install github.com/bufbuild/buf/cmd/buf@latest
```

Document this in the action README.

---

## Test Requirements

### Unit Tests: `engine/internal/diff/proto_test.go`

All tests are table-driven. Use testdata fixture directories.

| Test Case | Input | Expected Output |
|---|---|---|
| No changes | identical proto dirs | `DiffReport` with 0 changes |
| Field removed | `User.email` field deleted from proto | 1 BREAKING: `PROTO_FIELD_REMOVED` |
| Field type changed | `User.id` changes from `string` to `int32` | 1 BREAKING: `PROTO_FIELD_TYPE_CHANGED` |
| Field number changed | field 1 becomes field 2 | 1 BREAKING: `PROTO_FIELD_NUMBER_CHANGED` |
| Service removed | `UserService` deleted from proto | 1 BREAKING: `PROTO_SERVICE_REMOVED` |
| RPC method removed | `GetUser` RPC removed from service | 1 BREAKING: `PROTO_RPC_REMOVED` |
| Enum value removed | `STATUS_ACTIVE` removed from enum | 1 BREAKING: `PROTO_ENUM_VALUE_REMOVED` |
| New field added | `User.phone` added with new field number | 0 BREAKING, 0 WARNINGS (safe) |
| New RPC added | `DeleteUser` RPC added to service | 0 BREAKING (safe) |
| buf not on PATH | mock exec.LookPath to return error | returns error containing "buf is not installed" |

### Testdata Fixtures: `engine/internal/diff/testdata/proto/`

Create these fixture directories:

```
engine/internal/diff/testdata/proto/
  base_no_change/
    user.proto     (base proto with User message, UserService)
  head_no_change/
    user.proto     (identical to base)
  base_field_removed/
    user.proto     (User message with email field)
  head_field_removed/
    user.proto     (User message WITHOUT email field)
  base_field_type_changed/
    user.proto     (User.id: string)
  head_field_type_changed/
    user.proto     (User.id: int32)
  base_service_removed/
    user.proto     (has UserService)
  head_service_removed/
    user.proto     (UserService deleted)
  base_rpc_removed/
    user.proto     (UserService with GetUser + ListUsers RPCs)
  head_rpc_removed/
    user.proto     (UserService with only ListUsers — GetUser deleted)
  base_enum_value_removed/
    user.proto     (Status enum with ACTIVE + INACTIVE)
  head_enum_value_removed/
    user.proto     (Status enum with only INACTIVE — ACTIVE deleted)
  base_safe_field_added/
    user.proto     (User with id, email)
  head_safe_field_added/
    user.proto     (User with id, email, phone — new field number)
```

All test proto files MUST be valid Protobuf syntax (buf can validate them).
Use `syntax = "proto3";` and `package substrate.test.v1;` in all test fixtures.

---

## Files to Create / Modify

| File | Action |
|---|---|
| `docs/specs/phase-1/protobuf-checker-adapter.md` | ✅ This file (spec) |
| `engine/internal/diff/proto.go` | CREATE — new adapter |
| `engine/internal/diff/proto_test.go` | CREATE — table-driven tests |
| `engine/internal/diff/testdata/proto/*/` | CREATE — fixture directories |
| `engine/cmd/substrate/main.go` | MODIFY — add `protobuf` case to schema_type dispatch |

---

## Out of Scope for Phase 1d

- `buf lint` integration — linting is separate from breaking change detection
- Support for `buf.work.yaml` workspaces with multiple modules — single-module only
- `.proto` file discovery across nested subdirectories (only top-level proto dir)
- JSON transcoding rules (gRPC-Web) — future task
- gRPC reflection — not needed for static analysis

---

## Jules Prompt Location

`prompts/phase-1d-protobuf/t01_buf_adapter.txt`

> This spec is the source of truth. Jules must implement exactly what is described here.
> All 10 test cases in the Test Requirements table are mandatory.
> The `buf` binary must be invoked via `exec.Command` — do NOT import buf as a Go library.
