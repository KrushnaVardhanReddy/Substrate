# P7-T03: Custom Rules Engine (CEL)

## Objective
Different organizations have different API design standards. While Substrate provides out-of-the-box breaking change detection, enterprises need the ability to enforce custom governance rules (e.g., "All endpoints must have an `X-Correlation-ID` header" or "No path parameters can contain uppercase letters"). 
Substrate will implement a Custom Rules Engine using Google's Common Expression Language (CEL).

## Requirements

### 1. CEL Integration
- Target: `engine/internal/diff/evaluator.go`
- Import `github.com/google/cel-go/cel`.
- Expose the parsed schema Abstract Syntax Tree (AST) to the CEL environment.

### 2. Custom Rule Definition
- In `substrate.yaml`, allow users to define custom governance rules:
```yaml
custom_rules:
  - id: REQUIRE_CORRELATION_ID
    description: "All APIs must accept X-Correlation-ID"
    severity: BREAKING
    match: "endpoint.headers.exists(h, h.name == 'X-Correlation-ID')"
```

### 3. Engine Execution
- During the diff process, after the standard breaking change rules run, the engine will compile and evaluate the custom CEL expressions against the `head` schema.
- If a custom rule evaluates to `false`, a new `report.Change` is generated using the custom `id` and `severity`.
