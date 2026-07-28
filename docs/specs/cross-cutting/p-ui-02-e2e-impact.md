# P-UI-02: E2E Impact Analysis UI

## Goal
Add Playwright E2E coverage for the Impact Analysis page (`/org/{org}/repo/[repo]/impact`).

## Scenarios
1. **Successful Impact Load**: When visiting the impact page for a provider, it displays a list of downstream consumer repositories.
2. **Empty Impact**: When visiting the impact page for a leaf node (no consumers), it displays the "No downstream impact" empty state.
3. **Network Error**: If the backend `/api/v1/impact` endpoint returns a 500, the UI gracefully displays an error boundary rather than crashing.

## Constraints
- File: `dashboard/tests/e2e/impact_analysis.spec.ts`
- Must use Playwright test fixtures with **real API responses** from the live Go backend (`:8090`). No `page.route()` mocking. Relies on seeded database state from `seed_via_api.sh`.
