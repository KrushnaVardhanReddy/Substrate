# P-INS-01: E2E Schema Insurance API

## Goal
Validate the Schema Insurance APIs (`GET /api/v1/org/{org}/insurance/policy`, `GET /api/v1/org/{org}/insurance/claims`).

## Scenarios
1. **Get Policy**: Fetches the active SLA insurance policy details.
2. **List Claims**: Fetches the history of submitted claims.
3. **Not Found**: Returns 404 for an org without an active policy.

## Constraints
- File: `scripts/e2e/phase_insurance_e2e_test.go`
- Must use live PostgreSQL database connections, no mocks.
