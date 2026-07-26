# Spec: P12-T01 — Zero-to-One Onboarding E2E

## 1. Overview
Automate the complete new-user journey from the landing page through GitHub token entry to a fully rendered Dependency Graph using Playwright.

## 2. Owner
**Stitch** (Frontend / Playwright)

## 3. Files Modified
- `dashboard/tests/e2e/onboarding.spec.ts` ← expand with new test blocks

## 4. Requirements

### Happy Path Journey
1. Navigate to `/` (root URL).
2. Expect the onboarding wizard component to render (assert heading contains "Connect GitHub").
3. Fill the GitHub token input field with a mock token string.
4. Click the "Connect" button.
5. Assert the "Scanning Repositories" progress indicator / spinner appears.
6. Mock `GET /api/v1/repos/{org}` to return a JSON array of 3 repositories.
7. Assert the page navigates to `/org/{org}/graph` (assert `page.url()` contains `/graph`).
8. Assert `.cytoscape` canvas is visible.
9. Assert `.filter-panel` sidebar is visible.

### Error Path
10. Navigate to `/` with an invalid token.
11. Mock the API route to return a `401 Unauthorized`.
12. Assert an error message appears in the UI (e.g., text contains "Invalid token" or similar).

## 5. Technical Constraints
- Must use `page.route()` to intercept API calls — no real GitHub requests.
- The mock data should include at least 1 repo with `full_name: "testorg/core-service"`.
- Tests must not depend on real network or real GitHub auth.

## 6. Success Criteria
- `npx playwright test tests/e2e/onboarding.spec.ts` exits with code 0.
- Both happy path and error path tests pass.
- No changes to any other existing E2E test file.
