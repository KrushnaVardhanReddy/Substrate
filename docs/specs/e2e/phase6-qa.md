# Phase 6 E2E Spec: QA & Shadow API

## Objective
Validate the Phase 6 QA features, specifically testing the Auto-Updating Postman Collections, Shadow API Test Coverage, and the Mock Server Time Machine. The tests will utilize the Full-Stack PGlite Harness (CC-T02).

## Database State
The PGlite database must be seeded with:
1. Organization: `mcp-org`
2. Repo: `mcp-org/shadow-api-repo`
3. Pre-existing shadow API traffic logs.

## Scenarios to Test

### 1. Go API Tests (`scripts/e2e/phase6_api_test.go`)
- **GET /api/v1/qa/postman/{org}/{repo}**: Assert the API generates a valid Postman collection JSON payload for the target repository.
- **POST /api/v1/qa/shadow/replay**: Trigger a Shadow API replay job and assert the engine returns 200 OK.
- **GET /api/v1/qa/coverage/{org}/{repo}**: Assert the API computes a shadow coverage score based on the mocked traffic logs.

### 2. Playwright UI Tests (`dashboard/playwright/qa_shadow.spec.ts`)
- Navigate to `http://localhost:5173/org/mcp-org/qa/mcp-org/shadow-api-repo`.
- Wait for the QA Dashboard to load.
- Click "Export Postman Collection" and assert the download/modal triggers.
- Navigate to the "Time Machine" tab.
- Select a timestamp and click "Replay Traffic".
- Assert the UI reflects a "Replaying..." status and eventually shows the coverage results.
