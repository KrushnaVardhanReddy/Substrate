# P12-T13: Zero-Config Developer Portal (Catalog UI)

## Objective
Transform the Substrate Dashboard from a simple graph visualization tool into a full-fledged Developer Portal (a "Backstage Killer"). Because Substrate already parses and stores the dependency graph and schemas, we will expose this data as a searchable Service Catalog and interactive API Documentation hub, requiring zero manual configuration from the user.

## Architecture & Requirements

### 1. The Catalog List View (`/org/[orgId]/catalog`)
Create a new Svelte route that fetches all nodes from the Go API (`GET /api/v1/nodes`) and displays them as a data table or a grid of cards.
**Fields to display:**
*   **Service Name:** The repository name.
*   **Owner:** Team owner (extracted from `substrate.yaml` or mocked for now).
*   **Protocol:** e.g., OpenAPI, GraphQL, Protobuf.
*   **Status:** Safe (Green) vs. Breaking (Red).
*   **Upstream/Downstream Count:** Number of edges connected to this node.

### 2. The Global Navigation Updates
Update the left sidebar or top navigation (likely `dashboard/src/routes/org/[orgId]/+layout.svelte`) to include two primary tabs:
*   **Graph:** The existing interactive dependency map (`/org/[orgId]/graph`).
*   **Catalog:** The new service directory (`/org/[orgId]/catalog`).

### 3. The API Details View (`/org/[orgId]/catalog/[repoName]`)
When a user clicks on a service in the catalog, route them to a detail page.
*   **Header:** Show service metadata and a button to "View on GitHub".
*   **API Documentation Rendering:**
    *   Substrate stores the raw schema strings in the database.
    *   For OpenAPI specs, use a lightweight, framework-agnostic documentation renderer like **Stoplight Elements** (`@stoplight/elements`) via its Web Component API.
    *   *Implementation Note:* Elements can be loaded via a CDN `<script>` tag in the `+page.svelte` and initialized with `<elements-api apiDescriptionDocument={rawYaml}></elements-api>`.

### 4. Integration with Go API
The frontend must fetch the latest raw schema to render the documentation. 
Ensure the Go API endpoint for fetching a specific node includes the raw schema payload, or create a mock function in the Svelte `load` function temporarily if the API doesn't return the full schema payload yet.

## Success Criteria
*   Users can navigate between the Graph and the Catalog.
*   The Catalog displays all services discovered by Substrate.
*   Clicking a service opens a detailed view with auto-rendered API documentation.
*   Zero manual `catalog-info.yaml` files are required.
