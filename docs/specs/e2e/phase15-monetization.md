# Phase 15 E2E Spec: Monetization & Insurance

## Objective
Validate the Phase 15 Monetization & Insurance features. The tests will utilize the Full-Stack PGlite Harness (CC-T02).

## Database State
The PGlite database must be seeded with:
1. Organization: `mcp-org`
2. Repo: `mcp-org/core-repo`
3. Phase-specific mock data.

## Scenarios to Test

### 1. Go API Tests (`scripts/e2e/phase15_api_test.go`)
- Assert standard API behaviors and validation logic.

### 2. Playwright UI Tests (`dashboard/playwright/phase15.spec.ts`)
- Navigate to `http://localhost:5173/org/mcp-org/`.
- Assert the UI renders the correct state.
