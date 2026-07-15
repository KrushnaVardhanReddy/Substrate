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

## 3. Stitch & Jules Workflow
- **Jules:** Add the mock `volatilityScore` to the node data, implement the Toggle UI, and dynamically update the Custom Node styling based on the toggle state.
