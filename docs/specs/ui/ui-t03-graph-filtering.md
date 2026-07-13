# UI-T03: Graph Filtering & Navigation

## Objective
As the dependency graph scales to enterprise proportions (100+ nodes), a full unconstrained graph becomes visually noisy and difficult to navigate. We must implement dynamic filtering capabilities within the Cytoscape graph canvas to allow users to quickly identify critical failure points (e.g., breaking changes) and filter out irrelevant protocols.

## Requirements

### 1. The Filtering UI Controls
- **Location:** Update `dashboard/src/routes/org/[org]/graph/+page.svelte` to add a new "Filter Panel" inside the `.graph-controls` or as a sticky sidebar overlay.
- **Controls Needed:**
  - **Status Toggle:** A checkbox or toggle button to "Show Only BREAKING Changes". When active, this hides any consumer or provider node that is strictly `SAFE`.
  - **Protocol Filter:** A select dropdown to filter by protocol (e.g., `OpenAPI`, `GraphQL`, `Avro`). Note: When filtering for a specific protocol, the UI **must** retain not only the matched provider nodes but also all of their connected consumer nodes so that the edges remain visible.
  - **Search Input:** A text box allowing the user to type a repository name. Nodes matching the substring should be highlighted, and non-matching nodes should be dimmed (`opacity: 0.2`).

### 2. Cytoscape Filtering Logic
- When a filter is applied, do **not** destroy the Cytoscape instance.
- Instead, use Cytoscape collections to apply CSS classes or manipulate visibility:
  - Example: `cy.nodes().removeClass('hidden')` and `cy.nodes('[status != "BREAKING"]').addClass('hidden')`.
  - Ensure the stylesheet contains a `.hidden` class (`{ display: 'none' }`) and a `.dimmed` class (`{ opacity: 0.2 }`).
- Re-run the `dagre` layout automatically after hiding/showing nodes so the layout reflows and compacts properly: `cy.layout({ name: 'dagre', ... }).run()`.

### 3. State Preservation during Polling
- Since `E2E-T01` implemented a 5-second polling loop, the filtering logic must persist.
- If a user has "Show Only BREAKING" toggled, any new nodes fetched during the poll must immediately respect that filter before the layout is run.

## Technical Constraints
- Do not add any new heavy third-party UI libraries (stick to vanilla HTML/CSS and Svelte primitives where possible).
- All Cytoscape styles must remain in the `cy` initialization block (adding new classes is fine).
