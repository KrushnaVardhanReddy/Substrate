# P-UI-01: Missing Playwright E2E Coverage — High Priority Pages

## 1. Overview
Three high-value UI pages have zero Playwright E2E coverage. A regression in any of them would not be caught before reaching production:
- `/diff/[id]` — the interactive diff viewer (core product page)
- `/org/{org}/repo/[repo]/impact` — the blast radius impact detail page
- `/org/{org}/governance` — the NL→CEL rules management page

## 2. Pages Under Test

### Page A: `/diff/[id]` — Interactive Diff Viewer
**What it shows:** Side-by-side schema diff with breaking change highlights, AI autofix button, severity scores.
**Risk:** If the diff retrieval or rendering breaks, engineers clicking the Substrate PR comment link see a blank/error page.

**Test scenarios:**
1. Navigate to `/diff/{known_id}` (seed a diff report via API first).
2. Assert the page title contains "Diff" or "Breaking Changes".
3. Assert at least one breaking change row is rendered (`.breaking-change` or similar).
4. Assert "AI Autofix" button is visible.
5. Assert severity badge is rendered with a numeric count.

### Page B: `/org/{org}/repo/[repo]/impact` — Impact Page
**What it shows:** Which consumer repos would break if this provider repo's API changes.
**Risk:** Core product value prop. A JS error here silently hides blast radius.

**Test scenarios:**
1. Navigate to `/org/testorg/repo/backend/impact`.
2. Assert page loads with 200 (no redirect to 404).
3. Assert at least a "No downstream consumers" or consumer list renders (not a JS crash).
4. Assert "Check Deploy" button or safe/blocked badge is visible.

### Page C: `/org/{org}/governance` — Governance Rules Page
**What it shows:** List of active CEL rules, "Add Rule" button, NL rule input.
**Risk:** Rule CRUD is tested at API level but the UI that enables this workflow has zero coverage.

**Test scenarios:**
1. Navigate to `/org/testorg/governance`.
2. Assert page loads without error.
3. Assert "Governance Rules" heading or equivalent is visible.
4. Assert the rule input textarea is present.
5. Click "Add Rule", type a plain-English rule, assert a CEL rule appears in the list (or a success toast).

## 3. Constraints
- New files: `dashboard/tests/e2e/diff-viewer.spec.ts`, `dashboard/tests/e2e/impact-page.spec.ts`, `dashboard/tests/e2e/governance.spec.ts`
- Use the existing `playwright.config.ts` setup (`baseURL: http://localhost:5173`).
- **No `page.route()` mocking.** All tests run against the live Go API (`:8090`) with seeded database state. This ensures tests catch real regressions, not just UI rendering issues.
- Follow the pattern in `dashboard/tests/e2e/graph.spec.ts` and `enterprise-routes.spec.ts`.

## 4. Implementation Status
⏳ **NOT YET IMPLEMENTED** — created 2026-07-23
