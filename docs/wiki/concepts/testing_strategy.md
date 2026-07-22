# Testing Strategy

## E2E Validation (No Mocks)
Substrate mandates a "No Mocks" policy for End-to-End (E2E) validation.
- **Location**: `scripts/e2e/`
- **Execution**: E2E suites require real, live infrastructure (e.g., live Postgres DB, running API binary, real GitHub hooks).
- **Graceful Skipping**: When running in constrained sandbox environments (like GitHub Actions with overlayfs issues), E2E tests are designed to compile but gracefully skip (`t.Skip`) rather than fail.
