# Phase 3 E2E Spec: Dependency Graph & Can-Deploy

## Objective
Validate the Phase 3 Contract Registry APIs (`/api/v1/graph/{org}` and `/api/v1/registry/can-deploy`) and their corresponding UI components using the Full-Stack PGlite E2E Harness.

## Database State
The PGlite database must be seeded with:
1. Organization: `mcp-org`
2. Provider Repo: `mcp-org/backend`
3. Consumer Repo: `mcp-org/frontend`
4. Contract: `openapi.yaml` for backend
5. Dependency: `mcp-org/frontend` depends on `mcp-org/backend` (status: active).

## Scenarios to Test

### 1. Go API Tests (`scripts/e2e/phase3_api_test.go`)
- **GET /api/v1/graph/mcp-org**: Must return 200 OK. Assert the JSON response contains an edge where consumer is `mcp-org/frontend` and provider is `mcp-org/backend`.
- **GET /api/v1/registry/can-deploy**: Seed a breaking change in the database. Assert this endpoint returns 409 Conflict and lists `mcp-org/frontend` in the block reason.

### 2. Playwright UI Tests (`dashboard/playwright/graph.spec.ts`)
- Navigate to `http://localhost:5173/org/mcp-org/graph`.
- Wait for the cytoscape canvas (`cyContainer`) to load.
- Assert that two distinct nodes exist in the DOM (representing the backend and frontend).
- Click on the `mcp-org/backend` node and assert that the side-panel opens displaying the "openapi.yaml" schema details.
