# P5-T07: 100-Repo Scale & Noise Simulation

## Objective
Stress-test the Substrate backend (Registry API and PostgreSQL) and frontend (Svelte Dashboard visualization) by generating a massive, enterprise-scale mock dependency graph. The system must process, discover, and visualize 100 repositories simultaneously, accurately distinguishing interdependent clusters from random noise.

## 1. Scale Simulation Requirements

### Generator Script (`scripts/e2e/scale_generator.go`)
Create a Go script that programmatically synthesizes a complex GitHub organization repository state.

- **Total Repositories:** 100
- **Noise Repositories (80):**
  - Standalone repos with no external consumers or providers.
  - Mix of mock `package.json`, generic `openapi.yaml`, and `terraform` outputs.
  - Represents the "cruft" of a large enterprise.
- **Cluster Repositories (20):**
  - Divided into 5 distinct "Dependency Clusters" (4 repos per cluster).
  - Each cluster must form a directed graph: `Gateway -> Auth Service -> User Service -> Database (Terraform)`.
  - These clusters use exact match OpenAPI urls or Terraform output injection to force the dependency discovery scanners (Env Scanner, Package Scanner, Terraform Scanner) to detect the edges.

### API Load Testing
- The generator must rapidly fire 100 `POST /api/v1/webhook/github` (installation events) to the local Mock Registry Server.
- Assert that the Registry handles the load without dropping connections, and accurately parses all 100 repositories.

### Dashboard Visualization Rendering
- The Svelte Dashboard (`/dashboard/src/routes/+page.svelte` or the graph component) must be validated (via Playwright or manual inspection checklist) to ensure it renders 100 nodes without browser lockup.
- The 5 distinct clusters must clearly separate from the 80 floating noise nodes using D3.js or the existing force-directed graph physics.

## 2. Success Criteria
1. `make e2e-scale` successfully runs the generator and populates the local Postgres DB with 100 repositories.
2. The Dependency Graph returns exactly 100 nodes and 15 edges (3 edges per cluster * 5 clusters).
3. The Svelte Dashboard loads and visually separates the clusters from the noise within 3 seconds of page load.
