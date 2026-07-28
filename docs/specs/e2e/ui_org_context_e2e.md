# E2E Test Spec: UI Org Context & API Keys

## Objective
Add Playwright E2E coverage for the newly refactored UI Organization Context and API Keys scaffolding (CC-T04). This ensures that navigation works smoothly and the organizational context (`mcp-org`) is preserved when moving between features.

## Files to Create/Update
- `dashboard/tests/e2e/org-context.spec.ts` (New file for API keys and sidebar navigation checks)
- Update existing `playground.spec.ts` (if required) to assert context persistence.

## Test Scenarios

### 1. API Keys View (Org Scoped)
**File**: `org-context.spec.ts`
- **Action**: Navigate to `/org/mcp-org/apikeys`.
- **Assert**: Page header displays "API Keys".
- **Assert**: UI contains a "Generate New Key" button.
- **Assert**: Table or empty state renders correctly.
- **Assert**: Sidebar `.active` item is correctly set on "API Keys".

### 2. Context Persistence (Marketplace)
**File**: `org-context.spec.ts`
- **Action**: Navigate to `/org/mcp-org`.
- **Action**: Click "Marketplace" in the sidebar.
- **Assert**: URL is now `/org/mcp-org/marketplace`.
- **Assert**: The sidebar still displays the full suite of organization links (Repositories, Dependency Graph, API Keys, etc.), proving that the `mcp-org` context was not lost.

### 3. Context Persistence (AI Playground)
**File**: `org-context.spec.ts`
- **Action**: Navigate to `/org/mcp-org`.
- **Action**: Click "✨ AI Playground" in the sidebar.
- **Assert**: URL is now `/org/mcp-org/playground`.
- **Assert**: The sidebar still displays the full suite of organization links.

## Rules & Standards
- Tests must use standard Playwright assertions (`expect(page).toHaveURL`, `expect(locator).toBeVisible`).
- Timeout values should gracefully handle local dev server loads (e.g., standard 10s-30s navigation limits).
- Use `process.env.PUBLIC_ORG_NAME || 'mcp-org'` as the default organization context, consistent with other Playwright tests.
