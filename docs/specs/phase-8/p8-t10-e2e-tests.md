# Phase 8 E2E Testing Specification (Real End-to-End)

## Overview
Phase 8 introduces complex backend infrastructure required for Enterprise Scale: a Postgres-backed Job Queue, strict RBAC authorization, and critical deployment safety gates. Similar to Phase 7, this E2E suite enforces a **"No Mocks"** policy for core services.

This test suite runs against the real Substrate Go API and the real PostgreSQL database running locally via `make start-bg`. It intentionally avoids `testcontainers-go` (matching Phase 7) to prevent local Docker/Podman daemon socket issues across diverse developer environments.

## Test Scenarios

### Scenario 1: Durable Job Queue & Webhook Egress (P8-T01)
**Goal:** Verify that synchronous API handlers correctly enqueue jobs into River and that the asynchronous workers process them successfully, specifically the Webhook Egress delivery.
**Steps:**
1. Spin up a local mock target server to receive the outbound JSON webhook.
2. Submit a breaking change payload to the `POST /api/v1/diff` endpoint.
3. **Assert:** The API immediately returns a `202 Accepted` instead of blocking.
4. **Assert:** Query the PostgreSQL `river_job` table and confirm the `EgressWebhookJob` is in an `available` or `completed` state.
5. **Assert:** The local mock target server receives the valid webhook payload with the correct HMAC signature.

### Scenario 2: Enterprise Authz & RBAC (P8-T02)
**Goal:** Verify the Casbin/OpenFGA middleware correctly enforces Role-Based Access Control on sensitive dashboard and administrative API endpoints.
**Steps:**
1. Generate two JWTs: one with `Viewer` role, one with `Admin` role.
2. Attempt to access `DELETE /api/v1/org/acme/repo/billing-api` using the `Viewer` JWT.
3. **Assert:** The API returns `403 Forbidden`.
4. Attempt to access the exact same endpoint using the `Admin` JWT.
5. **Assert:** The API returns `200 OK` (or successfully executes the deletion).

### Scenario 3: Cascading Rollback Gate (P8-T03)
**Goal:** Verify the `substrate check-rollback` CLI command accurately detects when a downstream consumer has already adopted a newer schema, preventing an unsafe rollback.
**Steps:**
1. Register `billing-api` at `v2`.
2. Register a consumer `invoice-service` that depends on `billing-api@v2`.
3. Execute `substrate check-rollback --service billing-api --target-version v1`.
4. **Assert:** The CLI exits with a non-zero code (Failure).
5. **Assert:** The CLI outputs a clear error stating that `invoice-service` is actively consuming `v2` and rolling back will break production.

### Scenario 4: Billing Engine & Trial Enforcement (P8-T07)
**Goal:** Verify that Substrate gracefully enforces audit mode or pauses execution when an enterprise trial expires.
**Steps:**
1. Manually update the database fixture for `acme-corp` to set `trial_ends_at` to a date in the past (e.g., 2020-01-01).
2. Submit a destructive breaking change schema for an API owned by `acme-corp`.
3. **Assert:** The Diff Engine bypasses blocking rules and returns `202 Accepted`.
4. **Assert:** The payload response is a JSON object with `{"status": "paused_due_to_billing"}`, gracefully pausing execution.

## Execution Protocol
- **Infrastructure:** The test framework relies on the background services being spun up natively via `make start-bg` (or `make postgres` and `make api`).
- **Service Lifecycles:** The tests connect to `localhost:5432` for the database and `localhost:8090` for the API, verifying `/health` before proceeding.
- **Database Teardown:** The database must be cleanly truncated or destroyed between test scenarios to prevent state pollution.
- **Execution Command:** The entire phase 8 E2E suite can be executed via `make e2e-phase8`.
