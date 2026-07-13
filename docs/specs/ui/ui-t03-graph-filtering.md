# UI-T03: Graph Filtering & Navigation

## Objective
As the dependency graph scales to enterprise proportions (100+ nodes), a full unconstrained graph becomes visually noisy and difficult to navigate. We must implement dynamic filtering capabilities within the Cytoscape graph canvas to allow users to quickly identify critical failure points (e.g., breaking changes) and filter out irrelevant protocols.

## Requirements

### 1. The Filtering UI Controls
- **Location:** Update `dashboard/src/routes/org/[org]/graph/+page.svelte` to add a new "Filter Panel" inside the `.graph-controls` or as a sticky sidebar overlay.
- **Controls Needed:**
  - **Status Toggle:** A checkbox or toggle button to "Show Only BREAKING Changes". When active, this hides any consumer or provider node that is strictly `SAFE`.
  - **Protocol Filter:** A select dropdown to filter by protocol (e.g., `OpenAPI`, `GraphQL`, `Avro`).
  - **Search Input:** A text box allowing the user to type a repository name.
  - **Include Neighbors Toggle:** A checkbox (e.g. "Highlight connected neighbors") placed near the filters. Default: Off.

### 2. Search & Filtering UX Flow (Two-Step Exploration)
- **Step 1: Strict Isolation.** By default, when a user applies a Protocol Filter or a Search Query, the graph must **strictly highlight only the exact matching nodes**. All other nodes and edges MUST be dimmed (`opacity: 0.2`). This prevents visual clutter and allows the user to immediately locate the matching nodes in the haystack.
- **Step 2: Click to Explore (Blast Radius).** When a user clicks/taps on any node, the graph must enter a "Focus Mode". It should fully highlight the clicked node AND its first-degree connected edges/neighbors (upstream providers and downstream consumers). All other graph elements remain dimmed.
- **Neighbor Override.** If the "Highlight connected neighbors" toggle is checked, Step 1 is overridden: the initial search/filter will automatically highlight the matching nodes *plus* all of their connected neighbors and edges immediately.

### 3. Cytoscape Implementation Details
- When a filter is applied, do **not** destroy the Cytoscape instance.
- Instead, use Cytoscape collections to apply CSS classes or manipulate visibility:
  - Example: `cy.nodes().removeClass('hidden')` and `cy.nodes('[status != "BREAKING"]').addClass('hidden')`.
  - Ensure the stylesheet contains a `.hidden` class (`{ display: 'none' }`) and a `.dimmed` class (`{ opacity: 0.2 }`).
- **Critical API Constraint:** When retrieving adjacent nodes from a node collection, you MUST use `matchedNodes.neighborhood('node')` or `matchedNodes.neighborhood()`. Do not use `.connectedNodes()` directly on a node collection, as Cytoscape 3.x returns an empty collection when that method is called on nodes rather than edges.
- Re-run the `dagre` layout automatically after hiding/showing nodes so the layout reflows and compacts properly: `cy.layout({ name: 'dagre', ... }).run()`.

### 3. State Preservation during Polling
- Since `E2E-T01` implemented a 5-second polling loop, the filtering logic must persist.
- If a user has "Show Only BREAKING" toggled, any new nodes fetched during the poll must immediately respect that filter before the layout is run.

## Technical Constraints
- Do not add any new heavy third-party UI libraries (stick to vanilla HTML/CSS and Svelte primitives where possible).
- All Cytoscape styles must remain in the `cy` initialization block (adding new classes is fine).
