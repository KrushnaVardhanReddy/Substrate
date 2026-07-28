# Phase 5 E2E Spec: Discovery Scanners

## Objective
Validate the Phase 5 Discovery Scanners, including the Env Var & URL Registry Scanner, Package & Generator Scanners, and Terraform Confidence Scoring. The tests will utilize the Full-Stack PGlite Harness (CC-T02).

## Database State
The PGlite database must be seeded with:
1. Organization: `mcp-org`
2. Repo: `mcp-org/discovery-test-repo`
3. Pre-existing metadata representing an unscanned repo.

## Scenarios to Test

### 1. Go API Tests (`scripts/e2e/phase5_api_test.go`)
- **POST /api/v1/discovery/scan/{org}/{repo}**: Trigger a manual scan and assert the API returns a 202 Accepted.
- **GET /api/v1/discovery/status/{org}/{repo}**: Assert the discovery engine updates the status from `pending` to `completed`.
- **GET /api/v1/discovery/results/{org}/{repo}**: Assert the scan successfully identifies mocked environment variables, URLs, and Terraform dependencies.

### 2. Playwright UI Tests (`dashboard/playwright/discovery.spec.ts`)
- Navigate to `http://localhost:5173/org/mcp-org/discovery/mcp-org/discovery-test-repo`.
- Wait for the Discovery Dashboard to load.
- Click the "Run Full Scan" button.
- Assert a loading spinner appears, followed by a populated list of discovered dependencies (Env Vars, Packages, Terraform).
- Click on a discovered dependency and assert the Confidence Score side-panel opens.
