# Spec: P11-T01 - Cascading Blast Radius

## 1. Overview
When an enterprise architect clicks a microservice node, the graph should dynamically calculate and highlight the cascading "blast radius" (downstream consumers that would break if this service fails).

## 2. Requirements
- Modify the existing Cytoscape Custom Node to accept an `isActive` and `isFaded` state.
- When a node is clicked, run a DFS/BFS traversal on the Cytoscape edges to find all downstream dependents (1st, 2nd, Nth degree).
- Highlight the origin node with a glowing red border (`var(--danger)`).
- Highlight downstream nodes with a warning border (`var(--safe)`).
- Fade out (opacity: 0.2) all other unrelated nodes and edges.
- Add a "Clear Selection" button floating at the top right of the graph to reset the view.

## 3. Stitch & Jules Workflow
- **Jules:** Implement the traversal logic in Svelte 5 and update the Cytoscape store. No Stitch mockup needed since this is purely a visual state change of existing nodes.
