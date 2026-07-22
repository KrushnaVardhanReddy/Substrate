# P13-T04: Enterprise E2E Validation

## 1. Objective
Create a comprehensive Go-based End-to-End (E2E) test suite in the `scripts/e2e` directory to programmatically validate the newly merged Phase 10 and Phase 13 enterprise features. This ensures that features like Protobuf/gRPC diffing, API Gateway Auto-Sync mock, FinOps egress calculations, and Tree-sitter AST impact analysis do not regress.

## 2. Context & Existing Patterns
We heavily rely on Go-based E2E tests (e.g., `phase7_e2e_test.go`, `phase12_sse_test.go`) to test the integration between the Substrate CLI (`engine/`) and the central registry (`api/`). 
This task will strictly adhere to the established boilerplate in those files (spinning up a mock HTTP server, running `os/exec` to trigger the CLI, and asserting the JSON output or API side effects).

## 3. Implementation Details
Create `scripts/e2e/phase10_13_e2e_test.go` encompassing the following core scenarios:

### Scenario 1: Protobuf Schema Registry
- Provide inline `.proto` definitions (Base and Head).
- Drop an RPC method or Enum value.
- Execute `substrate diff`.
- Assert that the exit code is `2` (Breaking) and the JSON output identifies the exact removed Protobuf element.

### Scenario 2: FinOps Cost Prediction
- Mock the Datadog traffic response using the existing `helpers.MockServer`.
- Execute a schema diff that increases the payload size.
- Assert that the `DiffReport` calculates a positive USD cost increase.

### Scenario 3: API Gateway Push Mock
- Since we changed the gateway architecture to a CI/CD "Push" model, mock the Kong Admin API endpoint.
- Assert that running `substrate gateway sync` successfully POSTs the schema to the mocked Kong URL.
