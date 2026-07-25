# Phase 13 E2E Spec: God Mode & FinOps

## Objective
Validate the Phase 13 God Mode & FinOps features. The tests will utilize the Full-Stack PGlite Harness (CC-T02).

## Database State
The PGlite database must be seeded with:
1. Organization: `mcp-org`
2. Repo: `mcp-org/core-repo`
3. Phase-specific mock data.

## Scenarios to Test

### 1. Go API Tests (`scripts/e2e/phase13_api_test.go`)
- Assert standard API behaviors and validation logic.

### 2. Playwright UI Tests (`dashboard/playwright/phase13.spec.ts`)
- Navigate to `http://localhost:5173/org/mcp-org/`.
- Assert the UI renders the correct state.
