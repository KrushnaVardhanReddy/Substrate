# Substrate Handoff & Plan

## Current State
- **System E2E Status:** 🟩 100% Passing (Phase 17 UI/API tests fully integrated).
- **Recent Work:** 
  - ✅ Completed Phase 17 DevOps & Management Intelligence (FinOps, Auto-Rollback, GitOps Sync, Scorecards).
  - ✅ All automated validation for Phase 17 passes reliably (including live Forgejo seeding).
- **Next Up:** 
  - **Phase 16** (The Absolute SSOT API Documentation)
  - **Phase 18** (AI Agent Governance)

## How to Execute E2E Tests

### 1. Run the Full E2E Suite (Backend + Frontend)
This will spin up PGlite, Forgejo, seed data, run Go tests, and finally run the full Playwright UI test suite:
```bash
export GITHUB_TOKEN=dummy; make e2e
```
*(Note: Output is logged to `scripts/e2e/e2e_test.logs`)*

### 2. Run a Specific Playwright Test File (Faster for UI Debugging)
If the backend is already running (via `make e2e` or manual setup):
```bash
cd dashboard
npx playwright test tests/e2e/org-context.spec.ts --project=chromium --headed
```

