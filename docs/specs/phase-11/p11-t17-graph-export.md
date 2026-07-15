# Spec: Graph Image Export (P11-T17)

## 1. Overview
To support architectural reviews, RFC writing, and compliance audits, users must be able to export their active dependency graph view as a high-resolution PNG image directly from the dashboard.

## 2. Requirements

### 2.1 Dependency
- The project MUST use the `html-to-image` package to convert the Svelte Flow DOM viewport into a data URL.

### 2.2 UI Implementation
- Add an "Export PNG" button with a download icon (Lucide `Download`) to the Control Panel (`<div class="control-panel">`) in `dashboard/src/routes/(app)/org/[org]/graph/+page.svelte`.
- The button MUST follow the premium dark mode aesthetic (sleek hover states, matching existing buttons).

### 2.3 Export Logic
- When clicked, the function MUST target the `.svelte-flow__viewport` node.
- It MUST optionally trigger Svelte Flow's `fitView` before capture to ensure all nodes are cleanly in the frame.
- The resulting image MUST have a background color applied (`backgroundColor: '#0f172a'`) to ensure the dark theme renders properly in the PNG.
- A simulated `<a>` tag download MUST be triggered, naming the file `substrate-graph-[timestamp].png`.

## 3. Success Criteria
- A user filters the graph, clicks the Export button, and instantly receives a high-resolution `.png` file of the current viewport.
