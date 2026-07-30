# Cross-Cutting: Backfill Svelte Unit Tests (P-UNIT-01)

## Overview
The Substrate frontend has comprehensive Playwright E2E coverage, but certain utility pages and simple UI components lack fast Vitest unit tests. This task focuses on backfilling Vitest test coverage for the QA Dashboard and API Keys pages to ensure `npm run test` executes against these components.

## Technical Requirements

### 1. Identify Missing Tests
The `dashboard/src/routes/(app)/org/[org]/settings/apikeys` and `dashboard/src/routes/(app)/org/[org]/qa` folders contain Svelte components and `+page.svelte` files.

### 2. Write Vitest Suites
Using `@testing-library/svelte`, write standard Vitest test suites for:
1. `dashboard/src/routes/(app)/org/[org]/settings/apikeys/page.test.ts` (Testing API key generation button, masking of keys, and list rendering).
2. `dashboard/src/routes/(app)/org/[org]/qa/page.test.ts` (Testing the QA matrix rendering and mock data integration).

### 3. Ensure Reliability
- The tests MUST NOT hit the real backend. Use `@testing-library/svelte` to mount components with mock data props.
- Run `npm run test` to verify all existing and new tests pass cleanly.

## Rules
- Avoid barrel imports (e.g., `import { X } from 'lucide-svelte'`). Use direct imports (e.g., `import X from 'lucide-svelte/icons/x'`) to prevent Vitest memory leaks, as discovered in previous test fixing sessions.
