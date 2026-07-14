# Spec: P11-T10 - Svelte Flow Migration

## 1. Overview
Migrate the primary dependency graph visualization from `cytoscape.js` to `@xyflow/svelte` (Svelte Flow) for better reactivity, custom HTML nodes, and built-in minimap functionality.

## 2. Requirements
- Replace `cytoscape` dependencies with `@xyflow/svelte`.
- Implement a custom Node component (`ServiceNode.svelte`) that displays the service name, health status, and repository badge.
- Implement an animated Edge component to visualize data flow direction.
- Integrate the Svelte Flow `MiniMap` and `Controls` components.
- Ensure the layout remains performant for 100+ nodes (using Dagre or a similar layout engine for auto-positioning).

## 3. Stitch & Jules Workflow
- **Stitch:** Generate a static mockup of a "Service Node" card.
- **Jules:** Implement the `ServiceNode.svelte` component based on the mockup and integrate `@xyflow/svelte` into the main Graph view.
