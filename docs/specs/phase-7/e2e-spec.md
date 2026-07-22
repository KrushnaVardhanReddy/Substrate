# Phase 7 E2E Testing Specification (Real End-to-End)

## Overview
Unlike previous test phases which heavily relied on mocked downstream dependencies and simulated network traffic, the Phase 7 E2E test suite strictly enforces a **"No Mocks"** policy for Substrate's core services. 

This test suite spins up the real Substrate Go API, the real PostgreSQL database, the real Diff Engine CLI, and the real Proxy Sidecar to guarantee that Enterprise IT integrations behave exactly as they will in a production environment.

## Test Scenarios

### Scenario 1: Non-Blocking Audit Mode (P7-T06)
**Goal:** Verify that when Audit Mode is enabled, breaking changes do not cause the process to fail (exit 0) but still record the telemetry in the database correctly.
**Steps:**
1. Insert a baseline schema and a head schema containing a destructive breaking change (e.g., removing a required field).
2. Configure the diff engine or webhook payload with `mode: audit`.
3. Execute the diff engine / send the webhook.
4. **Assert:** The HTTP response or CLI exit code must be `0` (Success/Non-blocking).
5. **Assert:** The database `diff_reports` table must contain a new entry with `is_audit_mode = true`.

### Scenario 2: Custom Governance Rules via CEL (P7-T03)
**Goal:** Verify that a custom CEL rule injected via configuration successfully fails a schema that otherwise passes standard OpenAPI compliance.
**Steps:**
1. Define a CEL rule requiring all endpoints to have an `X-Correlation-ID` header.
2. Submit a schema *missing* this header.
3. **Assert:** The Diff Engine correctly identifies a `BREAKING` violation triggered by the custom CEL expression.
4. **Assert:** The error description matches the CEL rule's custom error message.

### Scenario 3: Runtime Drift Detection Sidecar (P7-T05)
**Goal:** Verify that the Substrate sidecar proxy accurately captures undocumented live traffic and reports anomalies back to the registry.
**Steps:**
1. Spin up a real HTTP dummy target server.
2. Spin up the real `substrate-proxy` pointing to the dummy target.
3. Inject the OpenAPI specification into the Substrate DB (or configure the proxy with it).
4. Send an HTTP request to the proxy with an undocumented payload field (e.g., `{"secret_admin": true}`).
5. **Assert:** The proxy detects the undocumented field and successfully POSTs an anomaly to the Substrate `/api/v1/telemetry/drift` endpoint.
6. **Assert:** Query the real database's `drift_anomalies` table and verify the payload was successfully recorded.

### Scenario 4: AI Autofix Cross-Repo PR Generation (P7-T04)
**Goal:** Verify the REST client correctly builds the payload and interacts with the GitHub App to open an automated draft PR.
**Steps:**
1. Run a local HTTP mock server strictly to simulate the GitHub API (since we cannot make real automated GitHub API calls in standard CI without a real token).
2. Trigger the `CrossRepoCheckHandler` with a breaking change impacting a known consumer.
3. **Assert:** The Substrate API successfully generates a unified diff patch.
4. **Assert:** The Substrate API fires a POST request to the GitHub API mock to open a Draft PR with the exact title `chore(substrate): Auto-fix breaking change from upstream [...]`.

## Execution Protocol
- **Infrastructure:** The test framework must use `testcontainers-go` or local Docker to spin up a real PostgreSQL instance.
- **Service Lifecycles:** The Go API and Sidecar Proxy must be started on local dynamic ports.
- **Database Teardown:** The database must be cleanly truncated or destroyed between test scenarios to prevent state pollution.
- **Execution Command:** The entire phase 7 E2E suite can be executed via `make e2e-phase7`. Ensure the background services are running (`make start-bg`) and a mock GitHub token is exported (`export GITHUB_TOKEN=mock`) before executing.
