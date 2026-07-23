# P3-T12: Phase 3 Contract Registry E2E Validation

## 1. Overview
Phase 3 introduced the cross-repo contract registry. While `TestV1SystemE2E` covers the `check-deploy` flow, it never explicitly validates the **dependency graph API** or the **cross-repo compatibility check endpoint**. This spec defines targeted E2E tests for those gaps.

## 2. APIs Under Test

| API | Route | Auth | Description |
|-----|-------|------|-------------|
| Dependency Graph | `GET /api/v1/graph/{org}` | Service Token | Returns `DependencyEdge[]` JSON |
| Can Deploy | `GET /api/v1/registry/can-deploy` | Service Token | Returns 200 (safe) or 409 (blocked) |
| Can Rollback | `GET /api/v1/registry/can-rollback` | Service Token | Returns 200 (safe) or 409 (blocked) |
| Cross-Repo Check | `POST /api/v1/cross-repo-check` | Service Token | Returns consumer impact list |
| List Repos | `GET /api/v1/repos/{org}` | Service Token | Lists repos registered for an org |

## 3. Test Scenarios

### Scenario 1: Dependency Graph Returns Correct JSON
1. Seed: org `p3-org` → provider repo `p3-org/backend` → contract → consumer repo `p3-org/frontend` → dependency.
2. `GET /api/v1/graph/p3-org` → assert HTTP 200.
3. Assert `[]DependencyEdge` contains exactly one edge: `consumer: "p3-org/frontend"`, `provider: "p3-org/backend"`.
4. Assert `status: "active"`.

### Scenario 2: Can-Deploy Gate Blocks Breaking Change
1. Reuse DB state from Scenario 1.
2. Insert `breaking_change_history` record for `p3-org/backend` at commit `sha-new`.
3. `GET /api/v1/registry/can-deploy?repo=p3-org/backend&commit=sha-new` → assert HTTP **409 Conflict**.
4. Assert response body contains the consumer name `p3-org/frontend`.

### Scenario 3: Can-Deploy Gate Passes Safe Commit
1. `GET /api/v1/registry/can-deploy?repo=p3-org/backend&commit=sha-safe` (no breaking history for this SHA) → assert HTTP **200 OK**.

### Scenario 4: Repos List
1. `GET /api/v1/repos/p3-org` → assert HTTP 200 and JSON array contains `p3-org/backend` and `p3-org/frontend`.

## 4. Constraints
- No mocks for DB or API.
- Prefix all DB setup vars/functions with `p3` to avoid conflicts with other test files.
- File: `scripts/e2e/phase3_e2e_test.go`.
- All scenarios run inside `TestPhase3ContractRegistry(t)` with `t.Run` sub-tests.

## 5. Implementation Status
⏳ **NOT YET IMPLEMENTED** — created 2026-07-23.
