# Spec: P11-T03 - Interactive Edge Tooltips

## 1. Overview
Hovering over the dependency lines (edges) between services should reveal a tooltip detailing the exact API contract (e.g., HTTP POST, gRPC Method) that connects them.

## 2. Requirements
- Create a Custom Edge in Cytoscape (`InteractiveEdge.svelte`).
- The edge should be thicker and change color when hovered.
- On hover, display a floating HTML tooltip exactly at the mouse cursor position or edge center, showing the mock contract details (e.g., `Protocol: gRPC, Method: GetUser`).
- Use Svelte 5 `$state` to track hovered edge data.

## 3. Stitch & Jules Workflow
- **Jules:** Implement the Custom Edge and Tooltip component using standard Vanilla CSS.
