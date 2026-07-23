# P10-T20: Phase 10 E2E Validation (No Mocks)

## Overview
End-to-end testing of OTel webhook ingestion, zombie API detection, governance rules CRUD, and auto-SDK generation. Must use a real local Postgres instance and zero mock APIs for database/API interactions.

## Requirements
1. **Test Environments**: Use a real local PostgreSQL instance (connection: `postgres://postgres:postgres@localhost:5432/substrate`). The API server must be running at `http://localhost:8090`.
2. **Coverage**: Tests must cover OTel metric ingestion, zombie endpoint detection, governance rule creation, and SDK generation triggers.
3. **No Mocks for DB/API**: All database and HTTP calls use real connections. GitHub/Forgejo PR comment assertions are skipped gracefully when a local Forgejo instance is not available (not a test failure).
4. **CI Integration**: The test suite runs as part of `go test ./...` in `scripts/e2e/`.

## Test Scenarios (`scripts/e2e/phase10_e2e_test.go`)

| # | Scenario | Endpoint | Expected |
|---|----------|----------|----------|
| 1 | OTel webhook ingestion | `POST /api/v1/otel/webhook` | `202 Accepted` |
| 2 | Zombie detection query | `GET /api/v1/org/{org}/zombies` | `200 OK` |
| 3 | Governance rules CRUD | `POST /api/v1/org/{org}/rules` | `201 Created` (PR comment assertion skipped if Forgejo absent) |
| 4 | Auto-SDK generation trigger | `POST /api/v1/webhook` | `202 Accepted` |

## Authentication Model
- Scenarios 1 and 4 use the **service token** (`REGISTRY_API_TOKEN`) via `Authorization: Bearer <token>`.
- Scenarios 2 and 3 also use the **service token** (bypasses JWT requirement, allowing full access).
- JWT/PASETO tokens are used in production for user-scoped org access; the service token grants full admin access for internal service calls.

## Setup
Test setup (`setupP10Database`) performs the following before each run:
1. Deletes existing `endpoint_traffic`, `governance_rules`, and `organizations` rows to ensure a clean state.
2. Inserts a fresh `organizations` row with `github_installation_id=1010`, `github_org_name='phase10-org'`.
3. Inserts a `repositories` row (`id=00000000-0000-0000-0000-000000000001`) linked to the org, required for foreign key integrity on `endpoint_traffic`.
