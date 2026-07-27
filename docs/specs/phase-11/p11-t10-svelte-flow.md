# Spec: P11-T10 - Cytoscape Migration

## 1. Overview
Migrate the primary dependency graph visualization from `cytoscape.js` to `cytoscape` (Cytoscape) for better reactivity, custom HTML nodes, and built-in minimap functionality.

## 2. Requirements
- Replace `cytoscape.js` with modern Cytoscape integrations using `$effect` for reactivity.
- Implement custom styling for nodes that displays the service name, health status, and repository badge. Nodes MUST utilize dynamic Lucide icons (e.g., AppWindow, Database, Smartphone) embedded as SVG data URIs via `background-image` in Cytoscape styles. Use a tinted background (`background-color`) matching the SVG stroke color (Cyan for Database, Purple for Frontend, Rose for Mobile) to create a premium, dynamic aesthetic.
- Implement an animated Edge component to visualize data flow direction. Fall back to standard `straight` lines when edge count exceeds 150 to preserve GPU performance. **Crucially, graph edges MUST be mapped as `source: Provider` and `target: Consumer`.** This guarantees Dagre's default Bottom-to-Top (`BT`) layout naturally anchors downstream consumers (Frontends) at the top and visually cascades data flow upwards from upstream providers (Databases) at the bottom. A Rotate feature MUST be provided in the UI to cycle through orientations (`BT`, `LR`, `TB`, `RL`).
- Integrate the Cytoscape `MiniMap` and `Controls` components.
- **Search-First Exploration Model:** To handle massive enterprise graphs (200+ nodes), the initial graph state MUST be empty with a prompt for the user to search. 
- **Asymmetrical Recursive Filtering:** When generating the sub-graph based on a search or filter, the logic MUST pull in only 1 layer of upstream providers (direct dependencies) to reduce visual noise, but MUST recursively pull in ALL layers of downstream consumers. This guarantees that the entire cascading "Blast Radius" of a breaking change is available on the canvas for highlighting.
- **Sub-Graph Layout Optimization:** The heavy Dagre layout calculation MUST only run on the filtered subset of nodes (e.g., the searched node + its neighbors). This ensures layout computes in <1ms and prevents massive layout scattering (the "hairball" problem). **Critical Constraint:** When constructing the subset array for Dagre, the filtering logic MUST always include the `parentId` group nodes for any surviving children, otherwise Dagre will throw an error attempting to position a compound child with an undefined parent.
- **Interactive State Separation:** Style updates (e.g., node selection, blast radius highlighting) MUST be separated from the layout algorithm using Svelte 5 `$derived` runes. Clicking a node only applies CSS opacity updates, maintaining 60fps responsiveness.
- **Reactivity Considerations (Svelte 5):** When implementing debounce functionality for the search input using `$effect`, reactive dependencies (e.g., `searchQuery`) MUST be read synchronously before any async callback (like `setTimeout`). Failure to do so prevents Svelte 5 from tracking the dependency, causing the graph to remain empty.
- **Detail Panel Integrations:** The side panel MUST provide quick developer actions for the selected node. The "View Logs" button MUST link directly to the node's GitHub Actions URL (`https://github.com/[repo]/actions`) to review CI failures. The "Open in IDE" button MUST utilize deep linking (e.g., `vscode://`) to instantly clone and open the repository locally.

## 3. Stitch & Jules Workflow
- **Stitch:** Generate a static mockup of a "Service Node" card.
- **Jules:** Implement the `ServiceNode.svelte` component based on the mockup and integrate `cytoscape` into the main Graph view.
