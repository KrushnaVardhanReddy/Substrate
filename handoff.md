# Substrate Handoff & Plan

## Current State
- **System E2E Status:** 🟩 100% Passing (94 Playwright tests, Go API suites, MCP tools).
- **Recent Work:** E2E Validation for UI Org Context (`CC-T05`) completed. `CC-T06` (API Keys Backend) and `P10-T01` (QA Fuzzing) PRs merged successfully.
- **Next Up:** Move on to Phase 11 (Advanced Graph Visualization) or any remaining unassigned tasks.

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
npx playwright test tests/e2e/org-context.spec.ts --project=chromium --headed
```
*(Use `--headed` to see the browser UI while the test runs)*
