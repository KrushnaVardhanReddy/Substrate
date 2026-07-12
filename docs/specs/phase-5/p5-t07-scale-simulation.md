# P5-T07: 100-Repo Scale & Universal Protocol Simulation

## Objective
Stress-test the Substrate backend (Registry API and PostgreSQL) and frontend (Svelte Dashboard visualization) by generating a massive, enterprise-scale mock dependency graph. 

Crucially, this simulation will validate the universal nature of the Substrate Discovery Engine by generating dependencies across **all 9 supported protocol adapters**. The system must process, discover, and visualize 100 repositories simultaneously, accurately distinguishing the complex 9-protocol interdependent clusters from random noise.

## 1. Scale Simulation Requirements

### Generator Script (`scripts/e2e/scale_generator.go`)
Create a Go script that programmatically synthesizes 100 repositories and fires them as GitHub Installation webhooks to the local Registry API.

- **Total Repositories:** 100
- **Noise Repositories (60):**
  - Standalone repos with no external consumers or providers.
  - Contain random, malformed, or disconnected `openapi.yaml`, `schema.graphql`, and `package.json` files.
  - Proves the "Signal vs. Noise" discovery algorithms do not generate false-positive edges.
- **Protocol Clusters (40 Repositories across 10 Clusters):**
  - Divided into 10 distinct "Dependency Clusters" (roughly 4 repos per cluster).
  - Each cluster must test a completely different underlying technology and protocol adapter:
    1. **REST / OpenAPI:** API Gateway -> NextJS Frontend -> Node Backend
    2. **GraphQL Federation:** Apollo Gateway -> Users Subgraph -> Posts Subgraph
    3. **gRPC / Protobuf:** Mobile App -> gRPC Gateway -> C++ Backend
    4. **AsyncAPI (Events):** Order Service -> Kafka Topic -> Fulfillment Service
    5. **Apache Avro:** Data Pipeline -> Schema Registry -> Analytics Engine
    6. **SQL DDL:** Migration Repo -> Postgres DB -> BI Tool
    7. **Terraform (IaC):** VPC Module -> EKS Cluster -> RDS Database
    8. **AI/ML Models:** Training Pipeline -> Model Registry -> Inference API
    9. **Enterprise SOAP:** Legacy Salesforce -> Enterprise Service Bus -> Payment Gateway
    10. **Hybrid Chaos:** A single cluster that mixes OpenAPI, GraphQL, and SQL in one flow.

### API Load Testing
- The generator must rapidly fire 100 `POST /api/v1/webhook/github` (installation events) to the local Mock Registry Server.
- Assert that the Registry handles the load without dropping connections, and accurately parses all 100 repositories using the 9 different protocol scanners.

### Dashboard Visualization Rendering
- The Svelte Dashboard (`/dashboard/src/routes/+page.svelte` or the graph component) must be validated (via Playwright or manual inspection checklist) to ensure it renders 100 nodes without browser lockup.
- The 10 distinct protocol clusters must clearly separate from the 60 floating noise nodes using D3.js or the existing force-directed graph physics.

## 2. Success Criteria
1. `make e2e-scale` successfully runs the generator and populates the local Postgres DB with 100 repositories.
2. The Dependency Graph returns exactly 100 nodes and exactly the correct number of edges for the 10 defined clusters (0 false positives from the 60 noise repos).
3. The Svelte Dashboard loads and visually separates the clusters from the noise within 3 seconds of page load.
