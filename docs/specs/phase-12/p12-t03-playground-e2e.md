# Spec: P12-T03 — AI Playground & Diff Viewer E2E

## 1. Overview
Playwright tests to validate the AI Playground flow end-to-end against the live backend:
schema input → real `POST /api/v1/ai/analyze` → SSE stream → result display → auto-fix application,
ensuring malformed YAML does not crash Svelte state.

**Philosophy: NO page.route() mocks, NO browser-level interception. All requests hit the live Go API.**

When `SUBSTRATE_AI_BASE_URL` is not configured, the Go API returns a deterministic fallback SSE response.
Tests must be written to pass against that real server behavior.

## 2. Owner
**Stitch** (Frontend / Playwright)

## 3. Files Modified
- `dashboard/tests/e2e/playground.spec.ts`

## 4. Requirements

### Test A — Happy Path: AI Analysis Stream
1. Navigate to `/playground`.
2. Assert the page title contains "AI Schema Validator Playground".
3. Locate the schema editor textarea.
4. Fill it with a valid schema sample.
5. Click "Analyze with Substrate AI".
6. Assert the `.analysis-panel` becomes visible (timeout: 10s).
7. Assert it contains "BREAKING" (from server's real or fallback SSE response).

### Test B — Auto-Fix Application
1. Navigate to `/playground`.
2. Fill editor and click Analyze.
3. Wait for `.auto-fix-section` to appear.
4. Click "Apply Fix".
5. Assert the proposed schema editor's value is updated (contains `deprecated: true`).
6. Assert no console errors matching `$state` or `Cannot read properties`.

### Test C — Malformed YAML Resilience
1. Navigate to `/playground`.
2. Fill editor with malformed content.
3. Click Analyze.
4. Assert the page does NOT crash (`.playground-container` or `main` still visible).
5. No `page.route()` interception — let the real server parse and respond.

## 5. Technical Constraints
- **No `page.route()` interception.** All requests go to the live Go API at `localhost:8090`.
- The server returns a deterministic SSE stream even when no LLM is configured (fallback behavior is part of the real API contract).
- Tests must handle streaming — wait for `.analysis-panel` with sufficient timeout.

## 6. Success Criteria
- `npx playwright test tests/e2e/playground.spec.ts` passes all tests.
- Zero `page.route()` calls in the test file.
