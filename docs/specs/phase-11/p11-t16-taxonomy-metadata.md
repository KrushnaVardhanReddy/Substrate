# Spec: Taxonomy & Metadata Tagging (P11-T16)

## 1. Overview
As Substrate evolves from a strict CI/CD breaking change detector into a full **Internal Developer Portal (IDP)**, the dependency graph needs to provide more architectural context at a glance. 

This specification introduces a `metadata` block to the `substrate.yaml` file, allowing developers to tag their repositories with rich taxonomy (service type, owning team, connected databases). This metadata will be parsed, stored in the PostgreSQL database, and surfaced in the Cytoscape graph through visually distinct nodes and a detailed metadata panel.

## 2. Requirements

### 2.1 Configuration Schema (`substrate.yaml`)
- The `substrate.yaml` parser MUST be updated to accept a new `metadata` object.
- **Fields:**
  - `type` (string): The architectural type of the service. Expected values include `frontend`, `backend`, `database`, `mobile`, `gateway`, `cronjob`. If missing, default to `service`.
  - `team` (string): The organizational team that owns the repository (e.g., `platform-core`).
  - `databases` (array of strings): Datastores that this service connects to (e.g., `["postgres", "redis"]`).

### 2.2 Database Storage
- The `repositories` table MUST be extended with a `metadata` column of type `JSONB`, defaulting to `'{}'::jsonb`.
- A database migration file MUST be created to apply this change safely.
- The Go backend logic responsible for inserting/updating repositories (e.g., during the GitHub webhook sync in `push.go`) MUST serialize the parsed `metadata` object and persist it to the database.

### 2.3 API Modifications
- The Registry API endpoint responsible for serving the dependency graph (`/api/v1/graph/...`) MUST be updated to include `consumer_metadata` and `provider_metadata` JSON objects inside the returned dependency edge payload.

### 2.4 UI Visualization (Cytoscape)
- **Visual Node Icons:** The Cytoscape node style MUST render a distinct Lucide SVG icon embedded as a data URI (`background-image`) based on the `metadata.type` property:
  - `frontend` (`.frontend` class) → `AppWindow` Icon data URI with Purple tint border (`#8B5CF6`)
  - `database` (`.database` class) → `Database` Icon data URI with Cyan tint border (`#06B6D4`)
  - `mobile` (`.mobile` class) → `Smartphone` Icon data URI with Rose tint border (`#F43F5E`)
  - `backend` (`.backend` class) → `Server` Icon data URI with Emerald tint border (`#10B981`)
  - `service` / default (`.service` class) → Default border (`#a855f7`)
- **String Matching Fallback:** If `metadata.type` is missing from the API response, the UI MUST intelligently infer the node type by string-matching the repository label (e.g., if the label contains 'db' or 'postgres' it falls back to `database`, 'front' to `frontend`, etc.).
- **Aesthetics:** Nodes MUST use a 15% opacity background of the same stroke color to create a dynamic, premium "glow" aesthetic. The SVG background image should be centered (`background-position-x: 12px`, `background-position-y: center`) with a fixed `background-width`/`background-height` (e.g., `16px`).
- **Taxonomy Detail Panel:** When a node is selected, the right-hand Detail Panel (`<aside class="detail-panel">`) MUST dynamically render the taxonomy data if present:
  - Display the `team` name.
  - Display the `type` (capitalized).
  - Display the `databases` as a list of stylized pill badges.

## 3. Success Criteria
1. Submitting a `substrate.yaml` with a `metadata` block successfully updates the `repositories` table in PostgreSQL.
2. The Cytoscape canvas renders visually distinct icons for frontends, databases, and standard services.
3. Clicking a node populates the Detail Panel with the correct team, type, and database badges.
