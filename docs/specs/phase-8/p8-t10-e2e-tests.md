# Spec: P8-T10 - Phase 8 E2E Testing

## 1. Overview
To finalize Phase 8 (Enterprise Readiness & Scale), we must validate that all the newly introduced enterprise infrastructure components (RBAC, Job Queue, Telemetry, and Billing) function harmoniously together in a production-like environment. 

## 2. Requirements

### 2.1 E2E Test Suite (`make e2e-phase8`)
- Create `scripts/e2e/phase8_e2e_test.go`.
- Validate the `riverqueue/river` integration by triggering a webhook and verifying the job successfully processes asynchronously.
- Validate the Casbin/OpenFGA Authz middleware by attempting to access a protected dashboard route with an unauthorized JWT and ensuring a 403 Forbidden is returned, followed by a successful request with an Admin JWT.
- Validate the Cascading Rollback Gate (`substrate check-rollback`) correctly blocks an unsafe rollback.
- Validate the Billing Engine properly returns a neutral status code if the 90-day trial is artificially expired in the test database.

### 2.2 Makefile Target
- Add `e2e-phase8` to the Makefile to run these specific tests against a live local Postgres database, isolated from the Phase 7 suite.

## 3. Implementation Steps
1. Create the Phase 8 E2E Go test file.
2. Implement database fixtures to set up organizations with specific trial expiration dates and Authz configurations.
3. Use the Go `net/http/httptest` package or real HTTP requests against a local server instance to test the API and webhook endpoints.
4. Integrate the command into the CI/CD pipeline (`.github/workflows`).
