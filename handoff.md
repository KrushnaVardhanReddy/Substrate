# Substrate Handoff & Plan

## Current State
- **Backend API E2E Tests:** ✅ 100% Passing. We resolved the critical SQL schema mismatches (`organizations` / `repositories` tables) that were causing `TestPhase1f_AIMLAdapter` and `TestPhase1g_SalesforceAdapter` to fail.
- **Frontend Playwright Tests:** ❌ Currently experiencing 16 failures across 11 test suites. The failures are primarily timeout errors and incorrect UI locators (e.g., searching for text that has changed or UI elements that haven't fully rendered).

## Playwright UI Failures to Fix
1. `tests/e2e/preview-page.spec.ts` (4 failures - schema missing errors / fallback)
2. `tests/e2e/playground.spec.ts` (3 failures - AI streaming / autofix content)
3. `playwright/enterprise.spec.ts` (1 failure - missing "Active Webhook" text)
4. `tests/dashboard.spec.ts` (1 failure - empty state loading)
5. `tests/discovery.spec.ts` (1 failure - P5-T06 Phase 5 discovery scanners)
6. `tests/e2e/ai-autofix.spec.ts` (1 failure - AI patch application)
7. `tests/e2e/ai-copilot.spec.ts` (1 failure - Support widget streaming)
8. `tests/e2e/system-matrix-full.spec.ts` (1 failure - Microservices demo setup)
9. `tests/e2e/telemetry-roi.spec.ts` (1 failure - ROI dashboard metrics)
10. `tests/heatmap.spec.ts` (1 failure - Heatmap toggle)
11. `tests/stress-test.spec.ts` (1 failure - 1000 node graph crash)

## Plan to Fix
1. **Analyze Failed Logs & DOM state:** For each failing spec, review the specific line it fails on (e.g., `toBeVisible()` assertions timing out).
2. **Fix Locators:** Many tests use strict text matching (like `.getByText('Active Webhook')`). We need to update these to match the *actual* rendered text in SvelteKit or use more resilient `data-testid` attributes.
3. **Handle Asynchrony:** Some tests fail on "AI streaming" and "auto-fix" flows which take time. We need to increase timeout limits for AI-dependent tests or wait for specific network requests (`page.waitForResponse`) instead of just arbitrary UI delays.
4. **Graph / Heatmap Fixes:** Graph tests might be failing if the Canvas/Cytoscape instance isn't fully ready. Ensure the graph readiness state is verified before clicking nodes.
5. **Spec First Approach Note:** If the tests reveal that the actual backend API logic or input/output structures need to change, we must update the OpenAPI / AsyncAPI / GraphQL specs *first* as the single source of truth.

## How to Execute E2E Tests

### 1. Run the Full E2E Suite (Backend + Frontend)
This will spin up PGlite, Forgejo, seed data, run Go tests, and finally run the full Playwright UI test suite:
```bash
GITHUB_TOKEN=dummy make e2e
```
*(Note: Output is logged to `scripts/e2e/e2e_test.logs`)*

### 2. Run a Specific Playwright Test File (Faster for UI Debugging)
If the backend is already running (via `make e2e` or manual setup) and you just want to run one UI test file:
```bash
cd dashboard
npx playwright test tests/e2e/playground.spec.ts --project=chromium --headed
```
*(Use `--headed` to see the browser UI while the test runs)*
