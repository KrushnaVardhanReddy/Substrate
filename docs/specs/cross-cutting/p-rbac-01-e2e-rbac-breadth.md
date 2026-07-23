# P-RBAC-01: RBAC Breadth E2E Validation

## 1. Overview
Currently only `POST /api/v1/org/{org}/webhooks` is tested with role-based tokens (Phase 8 Scenario 2). All other `authzMW`-protected routes are assumed to behave correctly via shared middleware but are never explicitly asserted. A refactor to the middleware would silently break auth on all untested routes.

## 2. Routes Under Test

All routes guarded by `authzMW` (requires admin role for writes):

| Route | Method | Expected: Read-Only | Expected: Admin |
|-------|--------|---------------------|-----------------|
| `/api/v1/org/{org}/rules` | POST | 403 | 201 |
| `/api/v1/org/{org}/rules/{ruleID}` | DELETE | 403 | 200/204 |
| `/api/v1/telemetry/roi/{org}` | GET | 200 *(read — all roles)* | 200 |
| `/api/v1/org/{org}/insurance/claims` | POST | 403 | 201 |
| `/api/v1/org/{org}/zombies` | GET | 200 *(read — all roles)* | 200 |
| `/api/v1/org/{org}/zombies/pr` | POST | 403 | 200/201 |
| `/api/v1/org/{org}/partners` | POST | 403 | 201 |

## 3. Test Scenarios

### Scenario 1: Read-Only Token Blocked on All Write Endpoints
1. Mint PASETO token: `orgs: {"rbac-org": "Read-Only"}` using `createJWT()` from `phase8_e2e_test.go`.
2. For each write endpoint above, send request with Read-Only token.
3. Assert **403 Forbidden** for every write route.

### Scenario 2: Admin Token Allowed on All Write Endpoints
1. Mint PASETO token: `orgs: {"rbac-org": "admin"}`.
2. Seed minimal DB data (org, repo) for FK integrity.
3. For each write endpoint, send request with Admin token and valid payload.
4. Assert **201** or **200** (not 403, not 401).

### Scenario 3: Service Token Allowed on Service-Token Routes
1. Confirm `POST /api/v1/sync`, `POST /api/v1/diff`, `POST /api/v1/otel/webhook` accept `Authorization: Bearer local-dev-token` (service token).
2. Confirm they return **401/403** when called with a user JWT instead of service token.

### Scenario 4: No Token Returns 401
1. Send request to `POST /api/v1/org/rbac-org/rules` with no `Authorization` header.
2. Assert **401 Unauthorized**.

## 4. Constraints
- Extend file: `scripts/e2e/phase8_e2e_test.go` — add new `t.Run("Scenario 5: RBAC Breadth", ...)` block.
- Reuse `createJWT()` already defined in that file.
- Seed DB using `setupP8Database()` pattern.
- No new dependencies.

## 5. Implementation Status
⏳ **NOT YET IMPLEMENTED** — created 2026-07-23
