# Phase 12 - Task 15: Advanced Feature Suites E2E (WASM, MCP, AI, Enterprise)

## 1. Goal
Complete the Phase 12 E2E testing phase by implementing Suites 8 through 14 defined in the system matrix. These suites validate the advanced capabilities of the Substrate platform beyond basic schema sync, ensuring our edge technologies (WASM Engine, MCP, AI Autofix, SSE) and Enterprise features are fully covered.

## 2. Requirements

Implement the following separate E2E test files in `dashboard/tests/e2e/`:

### 2.1 `wasm-engine.spec.ts` (Suite 8)
- Test navigation to `/playground`.
- Test real Protobuf, OpenAPI diffs entirely in-browser (intercept network to ensure no backend calls are made to `/diff`).
- Test large payload (60k lines) stability.
- Test malformed YAML input handling.

### 2.2 `impact-api.spec.ts` (Suite 9)
- Test `GET /api/v1/impact/{org}/{repo}`.
- Test the Can-Deploy and Can-Rollback gates.
- Validate the responses conform to the standard MCP schema.

### 2.3 `ai-autofix.spec.ts` (Suite 10)
- Test `POST /api/v1/ai/analyze` with a breaking change.
- Test `POST /api/v1/ai/autofix` to verify it returns a valid `safe_patch` for both Protobuf and OpenAPI schemas.

### 2.4 `enterprise-routes.spec.ts` (Suite 11)
- Validate all major Enterprise Dashboard routes render correctly (`/org/admin/graph`, `/org/admin/catalog`, `/org/admin/settings`, `/org/admin/telemetry`, `/org/admin/webhooks`).

### 2.5 `sse-realtime.spec.ts` (Suite 12)
- Validate Server-Sent Events (SSE). 
- Open two browser contexts, trigger a backend change, and verify the frontend receives the event without a page reload.

### 2.6 `blast-radius.spec.ts` (Suite 13)
- Validate Cross-Repo Blast Radius. Setup a multi-tier dependency chain (A -> B -> C) and break A. Ensure C displays a blast radius alert.

### 2.7 `telemetry-roi.spec.ts` (Suite 14)
- Validate `POST /api/v1/telemetry/track`.
- Validate the ROI metrics dashboard updates based on telemetry events.

## 3. Architecture Constraints
- Create **separate** files for each suite to avoid merge conflicts and keep tests modular.
- Use `system-matrix-full.spec.ts` as a reference for structuring Playwright assertions and interacting with the backend.

## 4. Deliverables
- [ ] `wasm-engine.spec.ts`
- [ ] `impact-api.spec.ts`
- [ ] `ai-autofix.spec.ts`
- [ ] `enterprise-routes.spec.ts`
- [ ] `sse-realtime.spec.ts`
- [ ] `blast-radius.spec.ts`
- [ ] `telemetry-roi.spec.ts`
