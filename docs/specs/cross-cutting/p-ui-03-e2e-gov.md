# P-UI-03: E2E Governance Rules UI

## Goal
Add Playwright E2E coverage for the Governance Rules page (`/org/{org}/governance`).

## Scenarios
1. **List Rules**: Displays existing CEL rules fetched from the backend.
2. **Create Rule**: Filling out the new rule form and clicking "Save" calls the POST API and optimistic-updates the list.
3. **Delete Rule**: Clicking the trash icon calls the DELETE API and removes the rule from the UI.

## Constraints
- File: `dashboard/tests/e2e/governance.spec.ts`
- Must use Playwright UI testing against mocked endpoints.
