# Spec: P11-T12 - Rich Side-by-Side Diff Viewer & Sign Out

## 1. Overview
Build an interactive, syntax-highlighted side-by-side YAML diff viewer in the dashboard to review exact schema changes, mirroring the GitHub PR experience. Additionally, implement the missing "Sign Out" functionality in the global TopNav.

## 2. Requirements
- Create a new component `DiffViewer.svelte` that takes a `before` and `after` YAML string.
- Render the strings side-by-side with line numbers and syntax highlighting.
- Highlight added lines in green (`var(--safe)` or similar) and removed lines in red (`var(--danger)`).
- Update `dashboard/src/routes/(app)/diff/[id]/+page.svelte` to utilize this new `DiffViewer` component, mocking some data if necessary.
- **TopNav Sign Out:** Modify `dashboard/src/lib/components/TopNav.svelte` to add a "Sign Out" button or icon next to the user avatar.
- When the Sign Out button is clicked, it must:
  1. Remove `github_token` from `localStorage`.
  2. Call `goto('/onboarding')` to redirect the user back to the onboarding flow.

## 3. Stitch & Jules Workflow
- **Stitch:** Generate mockups for the Diff Viewer layout.
- **Jules:** Implement the `DiffViewer.svelte` component using vanilla CSS matching our Premium Aesthetics design system, and implement the Sign Out logic in the TopNav.
