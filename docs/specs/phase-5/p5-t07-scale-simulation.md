# P5-T07: Dynamic Chaos & Scale Simulation

## Objective
Aggressively stress-test the Substrate backend (Registry API and PostgreSQL) and frontend (Svelte Dashboard visualization) using chaos engineering principles. 

This is not a static test. The system must ingest a massive, highly concurrent flood of repositories, accurately separate 9 distinct protocol clusters from malicious "poison pill" noise, mutate the graph dynamically over time, and output a detailed telemetry report—all without crashing or leaking memory.

## 1. Scale Simulation Requirements

### Generator Script (`scripts/e2e/scale_generator.go`)
Create a highly concurrent Go script that programmatically synthesizes parameterized repositories and fires them as GitHub Installation webhooks. It must accept a `--scale` flag (default 100) and a `--concurrency` flag (default 50).

#### A. Repository Generation
- **Noise & Poison Pills (60%):**
  - Disconnected repos containing random files.
  - Must include **Poison Pills**: a 50,000-line GraphQL file, a recursive infinite-loop OpenAPI file, and a corrupted binary masquerading as `package.json` to test panic recovery.
- **Protocol Clusters (40% across 10 Clusters):**
  - Divided into 10 distinct "Dependency Clusters".
  - 9 clusters strictly test isolated protocols: OpenAPI, GraphQL, Protobuf, AsyncAPI, Avro, SQL DDL, Terraform, AI/ML, SOAP.
  - 1 cluster is Hybrid Chaos (mixes OpenAPI, GraphQL, SQL).

#### B. Execution Phases
1. **The Flood:** Fire all webhooks concurrently using Go routines with random microsecond jitter to simulate CI/CD rush hour and test for Postgres deadlocks.
2. **The Mutation:** Wait 5 seconds, then randomly mutate the state: delete 5 repositories (orphaning edges), rename 5 endpoints, and introduce 3 breaking changes. Ensure the graph dynamically recalculates.
3. **The Assertion:** Query the Registry API (`GET /api/v1/graph/{org}`) passing the `Authorization: Bearer <local-dev-token>` header to assert that exactly the expected number of nodes and edges survived the mutations, with 0 false positives from the noise.

#### C. Observability & Reporting Matrix
Because we cannot manually monitor 100+ repos, the `scale_generator.go` script MUST act as an APM tool. At the end of the execution, it must output a terminal table displaying:
- Total Time & Average Latency per Protocol (Did GraphQL take longer to parse than OpenAPI?)
- HTTP Status Codes (200s vs 500s).
- Panic/Crash counts (did the Poison Pills successfully get caught and rejected?).

## 2. Dashboard Visualization Validation
- The Svelte Dashboard (`/dashboard/src/routes/+page.svelte`) must be manually or automatically validated to ensure it renders the `--scale` nodes without freezing the browser, automatically untangling the 10 protocol clusters from the floating noise.

## 3. Success Criteria
1. `make e2e-scale` executes the full Chaos Simulation (Flood, Mutation, Assertion).
2. The Go compiler does not crash when parsing the Poison Pills.
3. The Terminal Reporting Matrix outputs successfully.
