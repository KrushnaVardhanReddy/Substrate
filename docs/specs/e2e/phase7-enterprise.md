# Phase 7 E2E Spec: Enterprise Rollout & Drift

## Objective
Validate the Phase 7 Enterprise features, specifically testing the Enterprise Webhook/Event Egress, Custom Rules Engine (CEL), and Runtime Drift Detection Sidecar. The tests will utilize the Full-Stack PGlite Harness (CC-T02).

## Database State
The PGlite database must be seeded with:
1. Organization: `mcp-org` (Enterprise Tier enabled)
2. Repo: `mcp-org/enterprise-repo`
3. A custom CEL rule and an active webhook endpoint.

## Scenarios to Test

### 1. Go API Tests (`scripts/e2e/phase7_api_test.go`)
- **POST /api/v1/enterprise/webhook**: Trigger a manual webhook ping and assert the API returns a 200 OK after dispatching.
- **POST /api/v1/enterprise/rules/validate**: Send a payload against a custom CEL rule and assert it successfully evaluates (pass/fail).
- **GET /api/v1/enterprise/drift/{org}/{repo}**: Assert the API returns the mock sidecar drift report.

### 2. Playwright UI Tests (`dashboard/playwright/enterprise.spec.ts`)
- Navigate to `http://localhost:5173/org/mcp-org/enterprise`.
- Wait for the Enterprise Dashboard to load.
- Click the "Webhooks" tab and assert the active webhook is displayed.
- Click the "Custom Rules" tab. Edit the CEL rule text area and click "Save Rule".
- Navigate to the "Drift Detection" tab and assert the drift report is visible with a "Resolve" button.
