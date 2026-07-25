# Phase 2 E2E Spec: GitHub App Integration

## Objective
Validate the Phase 2 GitHub App Integration features, specifically testing the Auto-Discovery Webhook, PR Commenting, and Commit Status Checks. The tests will utilize the Full-Stack PGlite Harness (CC-T02).

## Database State
The PGlite database must be seeded with:
1. Organization: `mcp-org`
2. Repo: `mcp-org/github-app-repo`
3. A mock GitHub Installation ID and mapped repository.

## Scenarios to Test

### 1. Go API Tests (`scripts/e2e/phase2_api_test.go`)
- **POST /api/v1/github/webhook**: Simulate an incoming `push` webhook payload from GitHub. Assert that the API correctly triggers the auto-discovery engine and responds with 200 OK.
- **POST /api/v1/github/status**: Simulate a PR status update request and assert the API formats the check-run payload correctly for the GitHub API.

### 2. Playwright UI Tests (`dashboard/playwright/github_app.spec.ts`)
- Navigate to `http://localhost:5173/org/mcp-org/settings/github`.
- Wait for the GitHub Settings panel to load.
- Assert the mock GitHub Installation ID is visible and marked as "Connected".
- Click the "Sync Repositories" button.
- Assert a success toast appears confirming that the repositories have been synced with GitHub.
