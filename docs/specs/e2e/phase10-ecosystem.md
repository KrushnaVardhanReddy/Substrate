# Phase 10 E2E Spec: Ecosystem & Zombie Pruning

## Objective
Validate the Phase 10 Ecosystem & Zombie Pruning features. The tests will utilize the Full-Stack PGlite Harness (CC-T02).

## Database State
The PGlite database must be seeded with:
1. Organization: `mcp-org`
2. Repo: `mcp-org/core-repo`
3. Phase-specific mock data.

## Scenarios to Test

### 1. Go API Tests (`scripts/e2e/phase10_api_test.go`)
- Assert standard API behaviors and validation logic.

### 2. Playwright UI Tests (`dashboard/playwright/phase10.spec.ts`)
- Navigate to `http://localhost:5173/org/mcp-org/`.
- Assert the UI renders the correct state.
