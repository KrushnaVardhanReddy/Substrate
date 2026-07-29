# Substrate Handoff & Plan

## Current State
- **System E2E Status:** 🟩 100% Passing (94 Playwright tests, Go API suites, MCP tools).
- **Recent Work:** Merged the third batch of Jules sessions (API Maturity Scorecards, PagerDuty Injection, ArgoCD Rollback) and resolved DB store append conflicts. Dispatched the final Phase 17 batch.
- **Active Jules Sessions:**
  1. `feat/p17-t02-finops-translation` - Legacy API FinOps Translation
  2. `feat/p17-t99-e2e-validation` - Phase 17 E2E Validation
- **Next Up:** Monitor the final Phase 17 Jules sessions. Review the generated PRs and merge them to formally close out Phase 17!

## 🚀 2-Day Execution & Merge Strategy

To complete Phases 16, 17, and 18 in 48 hours without blocking ourselves on merge conflicts, we follow this strict merge protocol for the parallel Jules sessions:

### 1. Database Migration Sequencing (The Bottleneck)
Parallel Jules agents may create conflicting SQL migration timestamps (e.g., two agents creating `20260728_...sql` files). 
- **Action:** When merging, accept the first PR as-is. For the second and third PRs, *manually rename* the SQL migration files to sequence them sequentially before merging, and ensure they don't step on the same tables.

### 2. Batch Dispatch & Review
- **Day 1:** Run and merge the current batch (`Sandbox`, `GitOps Sync`, `Schema Pruning`). Then dispatch the next batch (`Scorecards`, `Markdown Ingestion`, `MCP Governance`). 
- **Merge Order:** Merge backend API handlers first, then merge Frontend UI components, then merge E2E tests (`T99`).

### 3. E2E Safety Net
Run `make e2e` at the end of each day. We focus our human debugging time entirely on fixing any Playwright UI flakiness in the container environment, letting Jules handle all boilerplate CRUD.

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
