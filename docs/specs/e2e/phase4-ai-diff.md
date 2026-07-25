# Phase 4 E2E Spec: AI Diff Engine & Analyzers

## Objective
Validate the Phase 4 AI Diff Engine, specifically testing the Streaming SSE AI Analyze Handler, Safe Schema Patch Generator, and the VSCode Extension endpoints. The tests will utilize the Full-Stack PGlite Harness (CC-T02).

## Database State
The PGlite database must be seeded with:
1. Organization: `mcp-org`
2. Repo: `mcp-org/backend`
3. A baseline OpenAPI schema and a modified OpenAPI schema representing a diff.
4. Active PR integration tokens.

## Scenarios to Test

### 1. Go API Tests (`scripts/e2e/phase4_api_test.go`)
- **GET /api/v1/diff/stream/{org}/{repo}**: Send a mock diff payload and assert that Server-Sent Events (SSE) stream back the AI patch generator response correctly.
- **POST /api/v1/diff/upgrade**: Simulate a PR comment upgrade. Assert the API constructs the correct response payload meant for the GitHub PR bot.

### 2. Playwright UI Tests (`dashboard/playwright/ai_diff.spec.ts`)
- Navigate to `http://localhost:5173/org/mcp-org/diff/{diff_id}`.
- Wait for the diff comparison viewer to load.
- Click the "Generate AI Patch" button.
- Assert that the SSE stream connects and streams the code patch into the UI diff viewer dynamically.
- Assert the "Apply Patch" button becomes enabled once streaming completes.
