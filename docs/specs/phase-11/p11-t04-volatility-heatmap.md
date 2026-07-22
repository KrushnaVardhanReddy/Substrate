# Spec: P11-T04 - Historical Volatility Heatmap

## 1. Overview
A toggle switch that shifts the graph into "Heatmap Mode". Node colors transition from safe green to warning red based on their frequency of breaking changes.

## 2. Requirements
- Add a "Heatmap Mode" toggle switch at the top left of the graph.
- When toggled, update all nodes to reflect a `volatilityScore` (0 to 100).
- Score 0-30: Green (`var(--accent)`)
- Score 31-70: Yellow (`var(--safe)`)
- Score 71-100: Red (`var(--danger)`)
- The nodes should smoothly transition their background colors.

## 3. Implementation Notes & Constraints
- **Svelte 5 Reactivity:** When dynamically computing inline styles for the heatmap based on `$props`, ensure `$derived` runes that return computed values use `$derived.by(() => { ... })` instead of `$derived(() => { ... })` if you intend to evaluate the block. Otherwise, it will return the arrow function itself, breaking inline CSS rendering.
- **E2E Testing:** Playwright tests checking for the visibility of node cards in Heatmap Mode MUST simulate a search input first, as the "Search-First" layout model means the graph is empty on initial load.

## 4. Stitch & Jules Workflow
- **Jules:** Added the mock `volatilityScore` to the node data, implemented the Toggle UI, and dynamically updated the Custom Node styling based on the toggle state.
- **Stitch:** Fixed Svelte 5 `$derived.by` reactivity syntax for inline styles and added E2E search bypass (✅ Done).
