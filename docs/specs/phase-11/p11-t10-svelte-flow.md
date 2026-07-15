# Spec: P11-T10 - Svelte Flow Migration

## 1. Overview
Migrate the primary dependency graph visualization from `cytoscape.js` to `@xyflow/svelte` (Svelte Flow) for better reactivity, custom HTML nodes, and built-in minimap functionality.

## 2. Requirements
- Replace `cytoscape` dependencies with `@xyflow/svelte`.
- Implement a custom Node component (`ServiceNode.svelte`) that displays the service name, health status, and repository badge.
- Implement an animated Edge component to visualize data flow direction. Fall back to standard `straight` lines when edge count exceeds 150 to preserve GPU performance.
- Integrate the Svelte Flow `MiniMap` and `Controls` components.
- **Search-First Exploration Model:** To handle massive enterprise graphs (200+ nodes), the initial graph state MUST be empty with a prompt for the user to search. 
- **Sub-Graph Layout Optimization:** The heavy Dagre layout calculation MUST only run on the filtered subset of nodes (e.g., the searched node + its neighbors). This ensures layout computes in <1ms and prevents massive layout scattering (the "hairball" problem).
- **Interactive State Separation:** Style updates (e.g., node selection, blast radius highlighting) MUST be separated from the layout algorithm using Svelte 5 `$derived` runes. Clicking a node only applies CSS opacity updates, maintaining 60fps responsiveness.
- **Reactivity Considerations (Svelte 5):** When implementing debounce functionality for the search input using `$effect`, reactive dependencies (e.g., `searchQuery`) MUST be read synchronously before any async callback (like `setTimeout`). Failure to do so prevents Svelte 5 from tracking the dependency, causing the graph to remain empty.

## 3. Stitch & Jules Workflow
- **Stitch:** Generate a static mockup of a "Service Node" card.
- **Jules:** Implement the `ServiceNode.svelte` component based on the mockup and integrate `@xyflow/svelte` into the main Graph view.
