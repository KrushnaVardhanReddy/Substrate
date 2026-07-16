# Spec: P12-T03 — AI Playground & Diff Viewer E2E

## 1. Overview
Add Playwright tests to validate the AI Playground flow: schema input → backend AI stream → result display → auto-fix application, ensuring malformed YAML does not crash Svelte state.

## 2. Owner
**Stitch** (Frontend / Playwright)

## 3. Files Modified
- `dashboard/tests/e2e/playground.spec.ts` ← expand with new test blocks

## 4. Requirements

### Test A — Happy Path: AI Analysis Stream
1. Mock `POST /api/v1/ai/analyze` (or equivalent) to return a successful SSE/JSON response:
   ```json
   { "findings": ["Field 'userId' was removed (BREAKING)"], "autofix": "openapi: 3.0.0\ninfo:\n  title: Fixed" }
   ```
2. Navigate to `/playground`.
3. Assert the page title or heading contains "AI Playground" or "API Studio".
4. Locate the schema editor textarea (or `[contenteditable]` div).
5. Clear it and type a clearly malformed YAML schema (e.g., `invalid: yaml: : broken`).
6. Click the "Analyze" button.
7. Assert the analysis result panel becomes visible.
8. Assert the findings panel contains text (e.g., `toContainText('BREAKING')` or `toContainText('userId')`).

### Test B — Auto-Fix Application
1. After Test A completes (or re-run setup), click "Apply Auto-Fix".
2. Assert the editor textarea content changes (is no longer the malformed input).
3. Assert no browser console errors appear matching `/$state` or `Cannot read properties`.

### Test C — Malformed YAML Resilience
1. Navigate to `/playground`.
2. Input `"{{ NULL_BYTE_\x00_GARBAGE" }}"` into the editor.
3. Click "Analyze".
4. Mock the API to return `400 Bad Request` with `{ "error": "Invalid schema" }`.
5. Assert an error state message is displayed in the UI.
6. Assert the page does NOT crash (assert `.main-content` or `.playground-container` is still visible).

## 5. Technical Constraints
- All API calls MUST be intercepted via `page.route()`. No real AI calls.
- The mock `autofix` YAML must be valid minimal OpenAPI to verify the editor accepts it.
- The malformed input test must not hang on a 30s timeout — mock endpoint must respond immediately.

## 6. Success Criteria
- `npx playwright test tests/e2e/playground.spec.ts` passes all 3 new tests.
- No modifications to any other E2E spec file.
