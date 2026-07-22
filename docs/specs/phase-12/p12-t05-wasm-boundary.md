# Spec: P12-T05 — WASM Engine Boundary Tests

## 1. Overview
Validate the Go-compiled WASM diff engine (`engine/`) handles all edge cases and malformed inputs safely, returning structured JSON errors instead of Go panics.

## 2. Owner
**Jules** (Go Backend)

## 3. Files Modified
- `scripts/e2e/wasm_boundary_test.go` ← **new file** in the same package as `phase12_sse_test.go` (`package main`)

> **Decision Note (Spec > Prompt):** The test file lives in `scripts/e2e/`, NOT `engine/`. Reason:
> The `engine/` source files are compiled with `//go:build js && wasm` build tags, meaning the Go toolchain **cannot import them in a native test binary**. Placing the test inside `engine/` would require removing the build tag from production code — a risky refactor. Instead, the `scripts/e2e/` package already imports the main API server and has access to the full module; the boundary tests call the engine function via a direct package import path, or test a thin unwrapped version extracted into a build-tag-free file. This is the safer, correct approach.

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
- File lives in `scripts/e2e/wasm_boundary_test.go`, `package main` (same as `phase12_sse_test.go`).
- Because `engine/` source is compiled with `//go:build js && wasm`, Jules MUST first inspect the engine source and either:
  - **Option A (preferred):** Import the engine's Go module directly if any source file in `engine/` exists WITHOUT the `js && wasm` build tag (e.g., a helper or utility file).
  - **Option B:** Create a thin `engine/diff_core.go` file (no build tag) that exposes the raw diff function for native Go import, and keep the WASM JS glue in the existing tagged file. This is a minimal, safe addition.
- In the fuzz test snippet above, replace `engine.DiffSchemas(a, b)` with the correct import path after inspecting the module. The import will be something like `"github.com/KrushnaVardhanReddy/Substrate/engine"` — check `scripts/e2e/go.mod` for the exact module name.
- The 10MB payload test should use `strings.Repeat("x", 10_000_000)`.
- All tests must complete in < 5 seconds per case (excluding fuzz time).
- Do NOT use `testing.Short()` skips — these boundary tests must always run.

## 6. Success Criteria
- `go test ./scripts/e2e/ -run TestDiff` passes all 7 table tests.
- `go test ./scripts/e2e/ -fuzz=FuzzDiffSchemas -fuzztime=30s` finds no panics.
- `go test -race ./scripts/e2e/` passes with no data races.

