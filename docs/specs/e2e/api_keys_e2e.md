# E2E Test Spec: API Keys Backend

## Objective
Add full end-to-end coverage for the new API Keys backend integration (CC-T06) ensuring that both the Go API and the Playwright UI correctly generate, store, and revoke organization-scoped API Keys.

## Files to Create/Update
- `scripts/e2e/api_keys_e2e_test.go` (New file for backend Go testing)
- `dashboard/tests/e2e/api-keys.spec.ts` (New file for Playwright UI testing)

## Test Scenarios

### 1. Backend Go Tests (`api_keys_e2e_test.go`)
- **Action**: Call `POST /api/v1/org/{org}/apikeys` with a valid token.
- **Assert**: Response is `201 Created` and returns a raw API key.
- **Action**: Call `GET /api/v1/org/{org}/apikeys`.
- **Assert**: Response is `200 OK` and contains the newly created key (prefix only).
- **Action**: Call `DELETE /api/v1/org/{org}/apikeys/{id}` using the generated ID.
- **Assert**: Response is `204 No Content`.
- **Action**: Call `GET /api/v1/org/{org}/apikeys` again.
- **Assert**: The deleted key is no longer in the list.

### 2. Frontend Playwright Tests (`api-keys.spec.ts`)
- **Action**: Navigate to `/org/mcp-org/apikeys`.
- **Action**: Click "Generate New Key".
- **Assert**: A modal/dialog appears showing the raw API key exactly once.
- **Action**: Close the modal.
- **Assert**: The new key (by prefix) appears in the data table.
- **Action**: Click the "Revoke" button on the newly created row.
- **Assert**: The row is removed from the table.

## Rules & Standards
- Ensure tests use standard Playwright assertions (`expect(page).toHaveURL`, `expect(locator).toBeVisible`).
- Follow the structure of `phase3_e2e_test.go` for the backend tests.
- Use `process.env.PUBLIC_ORG_NAME || 'mcp-org'` as the default organization context for frontend testing.
