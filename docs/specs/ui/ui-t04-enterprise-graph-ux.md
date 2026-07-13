# UI-T04: Enterprise Graph UX Overhaul

## Objective
As the Dependency Graph scales to support 100+ repositories (verified by the Phase 6 E2E load tests), a simple left-to-right DAG layout with highlighting becomes insufficient. To meet enterprise usability standards, the graph requires semantic visual cues, noise reduction features, and dedicated "Blast Radius" drill-down views to ensure developers can quickly assess impact.

## Requirements

### 1. Semantic Node Coloring
- **Concept:** Differentiate the types of architectural components visually at a glance.
- **Heuristic:** The UI must infer the component type based on substrings in the repository name (since protocol/type is not strictly enforced by the backend payload).
- **Color Mapping:**
  - **Frontend/UI** (Matches `ui`, `web`, `frontend`, `dashboard`, `app`): Color = **Purple** (`#a855f7`)
  - **Data/Infra** (Matches `db`, `data`, `postgres`, `redis`, `kafka`): Color = **Green** (`#22c55e`)
  - **Backend/API** (Matches `service`, `api`, `core`, `backend`, or default): Color = **Blue** (`#3b82f6`)

### 2. Hide Orphaned Nodes
- **Concept:** Repositories without any dependencies (neither upstream nor downstream) create massive visual clutter in large organizations.
- **Implementation:** Add a new checkbox in the Filter Panel: `[x] Hide Orphaned Nodes`.
- **Default State:** **Checked (True)** by default.
- **Behavior:** When enabled, any node with a degree of 0 (no connected edges) MUST be removed from the view (`.hidden`).

### 3. Blast Radius Modal (Focus Mode)
- **Concept:** When a user investigates a specific repository, the background context of 90 other dimmed repositories is distracting.
- **Implementation:** When a user clicks a node, instead of just dimming the background, it should completely isolate the "Blast Radius".
- **Focus Mode Capabilities:**
  - **Isolated Sub-graph:** Adds `.hidden` to all unrelated nodes. Because the layout engine ignores `.hidden` nodes, the remaining targeted node and its 1st-degree neighbors will naturally be pulled together into a tight, compact cluster.
  - **Cinematic Camera Pan:** The layout MUST be executed with `animate: true` and `fit: true` so the camera smoothly pans and zooms perfectly onto the newly compacted cluster, avoiding any jarring layout jumps out of the user's viewport.
  - **Impact Table:** A clean table listing the Downstream Consumers (who will break if this service changes) and Upstream Providers (who this service depends on) in the side panel.
  - **Action Buttons:** Standard actions (View Logs, Open in IDE) preserved from the current side-panel.

## Implementation Details (Cytoscape & Svelte)
- Update `dashboard/src/routes/org/[org]/graph/+page.svelte`.
- Expand the `style` block in Cytoscape to include classes for `.node-frontend`, `.node-backend`, and `.node-data`.
- Dynamically assign these classes during the `elements.push()` iteration loop.
- Use Svelte 5 snippets or standard modal patterns for the Blast Radius overlay.
