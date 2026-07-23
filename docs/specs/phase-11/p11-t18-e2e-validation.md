# P11-T18: Phase 11 Backend E2E Validation

## 1. Overview
Phase 11 introduced three backend APIs that power the frontend graph/SSE features. While the frontend is tested via Playwright, the backend API contracts have **no E2E test coverage**. This spec defines the Go-based E2E test suite to close that gap.

## 2. APIs Under Test

| API | Route | Auth | Description |
|-----|-------|------|-------------|
| Graph / Blast Radius | `GET /api/v1/graph/{org}` | Service Token | Returns dependency graph as JSON `DependencyEdge[]` |
| Impact Analysis | `GET /api/v1/impact/{org}/{repo}` | Service Token | Returns list of repos impacted if this repo breaks |
| SSE Stream | `GET /api/v1/events` | Service Token | Returns `text/event-stream` with heartbeat every 5s |
| Diff Report | `GET /api/v1/diff/{id}` | Service Token | Returns stored diff report JSON |
| Preview Session | `POST /api/v1/preview` | Service Token | Creates a preview session token for a PR diff |

## 3. Test Scenarios

### Scenario 1: Dependency Graph Returns Correct Edges (P11-T01 Backend)
1. Seed DB: org → provider repo → contract → consumer repo → dependency edge.
2. `GET /api/v1/graph/{org}` → assert HTTP 200.
3. Assert response body contains a JSON array with at least one `DependencyEdge`.
4. Assert `consumer` and `provider` full names match seeded data.

### Scenario 2: Impact Analysis Enumerates Downstream Consumers (P11-T01 Backend)
1. Reuse DB state from Scenario 1.
2. `GET /api/v1/impact/{org}/{repo}` where repo is the provider → assert HTTP 200.
3. Assert response contains the consumer repo name.

### Scenario 3: SSE Stream Connects and Delivers Heartbeat (P11-T09)
1. Open HTTP GET to `/api/v1/events` with `Authorization: Bearer local-dev-token`.
2. Assert response status 200 and `Content-Type: text/event-stream`.
3. Read the stream for up to 6 seconds; assert at least one `data:` line is received (heartbeat).
4. Cancel context and assert connection closes cleanly without goroutine leak.

### Scenario 4: Diff Report Persistence and Retrieval (V1 Regression)
1. `POST /api/v1/diff` with a valid diff report payload → assert HTTP 201 and returns `id`.
2. `GET /api/v1/diff/{id}` using the returned ID → assert HTTP 200 and `report_data` is non-empty JSON.

## 4. Constraints
- No mocks for DB or API — use real local Postgres (`postgres://postgres:postgres@localhost:5432/substrate`).
- Use the same `waitForP11Services` + `setupP11Database` boilerplate pattern as `phase7_e2e_test.go`.
- Test file: `scripts/e2e/phase11_e2e_test.go`.
- All scenarios run inside a single `TestPhase11SystemE2E(t)` parent test with `t.Run` sub-tests.

## 5. Implementation Status
⏳ **NOT YET IMPLEMENTED** — created 2026-07-23.
