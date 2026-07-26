# Phase 1f and 1g E2E Spec: AI/ML and Enterprise Salesforce Adapters

## Objective
Validate the Phase 1f (AI/ML) and Phase 1g (Enterprise Salesforce/SOAP) schema adapters in the Core Diff Engine. The tests will utilize the Full-Stack PGlite Harness.

## Database State
The PGlite database must be seeded with:
1. Organization: `mcp-org`
2. Repos: `mcp-org/ml-models` and `mcp-org/salesforce-crm`

## Scenarios to Test

### 1. Go API Tests (`scripts/e2e/phase1f_1g_e2e_test.go`)
- Create integration tests for AI/ML and Salesforce SOAP schema diffing.
- **Red Path (Breaking Change):** Remove a required output from an AI/ML schema. Remove a custom field from a Salesforce object. Assert that the Substrate Go API returns a `BREAKING` status.
- **Green Path (Safe Change):** Add an optional input to an AI/ML schema. Add a new custom field to a Salesforce object. Assert that the Substrate Go API returns a `SAFE` status.
- Use `pglite` (port `54320`) and the Go `httpexpect` framework for validation.

### 2. Playwright UI Tests
- *Note:* The Playwright UI validations for AI/ML and Salesforce have already been successfully added to `dashboard/tests/e2e/system-matrix-full.spec.ts`. Jules DOES NOT need to implement Playwright tests for this task. Focus purely on the Go Backend API test suite.
