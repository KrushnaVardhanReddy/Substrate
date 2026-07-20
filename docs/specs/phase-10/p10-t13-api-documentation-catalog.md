# P10-T13: Integrated API Documentation Catalog

## Overview
Evolve the registry into an internal Developer Portal by embedding interactive API reference viewers (Stoplight Elements or ReDoc) directly into the dashboard.

## Requirements
1. **Viewer Integration**: Embed `@stoplight/elements` (or ReDoc) React/Svelte component in the catalog repo detail view.
2. **Spec Serving**: Create `GET /api/v1/spec/:org/:repo` to serve the latest merged OpenAPI spec as JSON for the viewer.
3. **Search**: Add full-text search across all endpoint paths and descriptions.
4. **Navigation**: Add a top-level "Catalog" link in the dashboard sidebar.
