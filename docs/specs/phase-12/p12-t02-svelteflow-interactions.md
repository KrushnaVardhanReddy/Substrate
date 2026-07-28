# Spec: P12-T02 — Cytoscape Interaction E2E

## 1. Overview
Extend the graph Playwright test suite with deep canvas interaction tests: blast radius highlighting, heatmap color assertion, and edge tooltip visibility. These tests validate Svelte 5 reactive state under complex user interactions.

## 2. Owner
**Stitch** (Frontend / Playwright)

## 3. Files Modified
- `dashboard/tests/e2e/graph.spec.ts` ← **append new `test()` blocks only**. Do NOT edit any existing test.

## 4. Requirements

### Test A — Blast Radius CSS Assertion
1. Load the graph page with the standard `beforeEach` mock data.
2. Fill the search input with `/` to render all nodes.
3. Wait for `.service-node-card` to attach and be visible.
4. Click the first `.service-node-card.database` with `{ force: true }`.
5. Assert `.detail-panel` becomes visible.
6. Assert the clicked card has class `.service-node-card.origin`.
7. Assert at least one other `.service-node-card.faded` exists in the DOM.
8. Assert at least one `.service-node-card.affected` exists in the DOM.

### Test B — Edge Tooltip Hover
1. Load the graph page with the standard mock data.
2. Fill the search input with `/` to render all edges.
3. Wait for `.cytoscape__edge` to be present.
4. Hover over the first `.cytoscape__edge` element.
5. Assert an element matching `.edge-tooltip` appears in the DOM and `toBeVisible()`.

## 5. Technical Constraints
- Both tests append to the existing `test.describe('Dependency Graph', ...)` block.
- `beforeEach` mock data from the existing test suite must be reused as-is.
- Do not modify the 3 existing `test(...)` calls.
- Use `{ force: true }` on Cytoscape node clicks to bypass animation stability checks.

## 6. Success Criteria
- `npx playwright test tests/e2e/graph.spec.ts` passes all 5 tests (3 existing + 2 new).
- No regressions in any other test file.
