# P-SYNC-01: Contract Sync Pipeline E2E Validation

## 1. Overview
`POST /api/v1/sync` is the entry point for all consumer schema snapshots — if it regresses, no contract ever gets stored and the entire registry silently breaks. `GET /api/v1/schema` and `GET /api/v1/spec` are how the CLI and dashboard read those stored schemas back. None of these have direct E2E assertions.

## 2. APIs Under Test

| Route | Auth | Description |
|-------|------|-------------|
| `POST /api/v1/sync` | Service Token | Push a schema snapshot for a consumer repo |
| `GET /api/v1/schema/{owner}/{repo}` | Service Token or JWT | Get stored schema for a repo |
| `GET /api/v1/spec/{org}/{repo}` | Service Token or JWT | Get full OpenAPI spec JSON for a repo |

## 3. Test Scenarios

### Scenario 1: Sync Stores Schema and Schema Retrieval Works
1. `setupPSync_Database` — clean `contracts`, `repositories`, `organizations`.
2. Seed org + repo `sync-org/backend` via pgxpool.
3. `POST /api/v1/sync` with payload:
   ```json
   {
     "org": "sync-org",
     "provider_repo": "sync-org/backend",
     "consumer_repo": "sync-org/frontend",
     "commit_sha": "sha-sync-001",
     "schema_type": "openapi",
     "raw_content": "openapi: 3.0.0\ninfo:\n  title: Sync Test\n  version: 1.0.0\npaths: {}"
   }
   ```
   → assert **200** or **201**.
4. `GET /api/v1/schema/sync-org/backend` → assert **200**.
5. Assert response body contains `"openapi"` and `"sync-org/backend"` or non-empty `raw_content`.

### Scenario 2: Spec Endpoint Returns Parsed JSON
1. Reuse seeded contract from Scenario 1.
2. `GET /api/v1/spec/sync-org/backend` → assert **200**.
3. Unmarshal as `map[string]interface{}` — assert it has `"openapi"` key.

### Scenario 3: Sync Idempotent — Re-syncing Same SHA Updates Timestamp
1. `POST /api/v1/sync` with same `commit_sha` as Scenario 1 but different content → assert **200** or **201** (no duplicate key error).

### Scenario 4: Schema 404 for Unknown Repo
1. `GET /api/v1/schema/sync-org/nonexistent` → assert **404**.

## 4. Constraints
- File: `scripts/e2e/phase_sync_e2e_test.go`
- Prefix: `pSync`
- No mocks.

## 5. Implementation Status
⏳ **NOT YET IMPLEMENTED** — created 2026-07-23
