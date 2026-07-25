# Phase 1 E2E Spec: Core Diff Engine

## Objective
Validate the Phase 1 Core Diff Engine features, specifically testing the OpenAPI Diffing (oasdiff) and the CLI Override mechanisms. The tests will utilize the Full-Stack PGlite Harness (CC-T02).

## Database State
The PGlite database must be seeded with:
1. Organization: `mcp-org`
2. Repo: `mcp-org/core-api`
3. A baseline OpenAPI spec (v1) and a breaking change OpenAPI spec (v2).

## Scenarios to Test

### 1. Go API Tests (`scripts/e2e/phase1_api_test.go`)
- **POST /api/v1/diff/analyze**: Submit the v1 and v2 OpenAPI specs to the diff engine. Assert that the API correctly identifies the breaking changes using the oasdiff adapter and returns a structured diff report.
- **POST /api/v1/cli/override**: Submit a CLI override request for the breaking change and assert the API logs the override successfully.

### 2. Playwright UI Tests (`dashboard/playwright/diff_engine.spec.ts`)
- Navigate to `http://localhost:5173/org/mcp-org/diff-reports`.
- Wait for the Diff Reports list to load.
- Click on the newly generated diff report.
- Assert that the Side-by-Side Diff Viewer renders the removed endpoints with red highlighting (indicating a breaking change).
- Click the "Approve Override" button and assert the status changes to "Approved".
