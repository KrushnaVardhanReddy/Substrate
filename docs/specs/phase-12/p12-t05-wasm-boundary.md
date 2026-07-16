# Spec: P12-T05 — WASM Engine Boundary Tests

## 1. Overview
Validate the Go-compiled WASM diff engine (`engine/`) handles all edge cases and malformed inputs safely, returning structured JSON errors instead of Go panics.

## 2. Owner
**Jules** (Go Backend)

## 3. Files Modified
- `scripts/e2e/wasm_boundary_test.go` ← **new file** in the same Go package as `phase12_sse_test.go`

## 4. Requirements

### Core Rule
Every test MUST assert: **no `panic`** and the return value is a valid JSON string (either a result or a `{ "error": "..." }` payload).

### Test Cases

| Test Name | Input | Expected |
|---|---|---|
| `TestDiff_EmptyStrings` | `""`, `""` | Returns JSON with empty diff or no-change signal |
| `TestDiff_ValidYAML` | Two valid OpenAPI YAML strings | Returns valid JSON diff result |
| `TestDiff_LargePayload` | 10MB random bytes each | Returns JSON error, no panic |
| `TestDiff_NullBytes` | `"\x00\x00\x00"` | Returns JSON error, no panic |
| `TestDiff_UnicodeGarbage` | `"日本語テスト\u0000☃️"` | Returns JSON error, no panic |
| `TestDiff_MismatchedSchemaTypes` | YAML vs JSON | Handled gracefully (error or best-effort diff) |
| `TestDiff_TruncatedYAML` | Valid YAML truncated mid-field | Returns JSON parse error |

### Fuzz Test
```go
func FuzzDiffSchemas(f *testing.F) {
    f.Add("", "")
    f.Fuzz(func(t *testing.T, a, b string) {
        defer func() {
            if r := recover(); r != nil {
                t.Fatalf("panic: %v", r)
            }
        }()
        result := engine.DiffSchemas(a, b)
        if !json.Valid([]byte(result)) {
            t.Errorf("result is not valid JSON: %s", result)
        }
    })
}
```

## 5. Technical Constraints
- Tests call the native Go `engine.DiffSchemas(a, b string) string` function directly (NOT via WASM binary — that would require a browser).
- If `engine.DiffSchemas` is unexported or named differently, Jules must adjust the import path.
- The 10MB payload test should use `strings.Repeat("x", 10_000_000)`.
- All tests must complete in < 5 seconds per case (excluding fuzz time).

## 6. Success Criteria
- `go test ./scripts/e2e/ -run TestDiff` passes all 7 table tests.
- `go test ./scripts/e2e/ -fuzz=FuzzDiffSchemas -fuzztime=30s` finds no panics.
- `go test -race` passes with no data races.
