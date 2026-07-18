# P10-T14: Embedded JS Governance Rules

## 1. Context
Substrate currently identifies breaking changes automatically. However, enterprise platform teams need to enforce custom API design policies (e.g., "All endpoints must have an X-Correlation-ID header" or "No pagination limits over 100"). Rather than forcing them to use a separate linter or compile WASM plugins, we will embed a pure-Go JavaScript engine (`github.com/dop251/goja`) directly into the Substrate CLI.

## 2. Requirements
- Add the `github.com/dop251/goja` dependency to `engine/go.mod`.
- Modify `engine/internal/config/config.go` to support a new `javascript` field in the `CustomRule` struct.
- In `engine/internal/checker/checker.go` or a new `engine/internal/rules/js_engine.go`, initialize a Goja runtime.
- For every custom rule defined in `substrate.yaml` with a `javascript` block, execute the script. The script should expose a `validate(schema)` function.
- We must pass the parsed JSON schema map to the JS environment.
- If the JS function returns a string, it is treated as a validation failure (the string is the error message). If it returns `true`, it passes.
- **MCP Server Integration:** Add a new MCP tool `test_js_rule` to `engine/cmd/substrate-mcp/main.go`. This tool should accept a `javascript` string and a `schema` (JSON string or object) to allow AI agents to instantly test custom rules while generating `substrate.yaml` configs.

## 3. Example `substrate.yaml` UX
```yaml
custom_rules:
  - id: "require-correlation-id"
    description: "Every endpoint must trace requests"
    javascript: |
      function validate(schema) {
        for (const path in schema.paths) {
          for (const method in schema.paths[path]) {
            const params = schema.paths[path][method].parameters || [];
            const hasHeader = params.some(p => p.in === 'header' && p.name === 'X-Correlation-ID');
            if (!hasHeader) {
              return "Missing X-Correlation-ID header on " + method.toUpperCase() + " " + path;
            }
          }
        }
        return true;
      }
```

## 4. Implementation Details
- Ensure the JS engine is heavily sandboxed (no file system access, no network access). Goja is inherently sandboxed as it has no native event loop or I/O bindings by default.
- Integrate the returned errors into the standard `report.DiffReport` as `Warnings` or `BreakingChanges` based on the rule severity.
