# E2E-T01: 1-Hour Chaos Endurance Mode & Live UI Polling

## Objective
Elevate the Substrate stress test from a simple 20ms burst into a full 1-hour sustained chaos engineering endurance test. This ensures the Go API, Postgres database, and UI can handle prolonged, high-concurrency "Poison Pill" attacks and mutations without memory leaks, deadlocks, or GC degradation. Additionally, the UI must be updated to poll for these changes in real-time.

## Requirements

### 1. The 1-Hour Siege (`scale_generator.go`)
- **CLI Flag:** Add a `--duration` flag to `scripts/e2e/scale_generator.go` (defaulting to `1h`). 
- **Continuous Loop:** Instead of iterating exactly `config.Scale` times and exiting, the `runFlood` and `runMutation` functions must operate continuously inside a loop for the given `--duration`.
- **Concurrent Workers:** Maintain a worker pool of exactly `config.Concurrency` (e.g., 50). These workers should randomly select an ID between `0` and `config.Scale` and fire a webhook (either a valid schema, a poison pill, or a mutation/delete).
- **Graceful Shutdown:** Listen for `os.Interrupt` (Ctrl+C) or the expiration of the duration timer to gracefully stop the workers and print the final "Assertion & Reporting Matrix".

### 2. Live Graph Polling (`+page.svelte` & `+page.ts`)
- **Polling Loop:** In `dashboard/src/routes/org/[org]/graph/+page.svelte`, implement a 5-second polling loop using Svelte 5's `$effect()` or `setInterval`.
- **API Fetching:** The component must fetch `GET /api/v1/graph/{org}` (with the `Authorization: Bearer <local-dev-token>` header) every 5 seconds.
- **Cytoscape Updates:** When new data is fetched, do NOT destroy the entire cytoscape instance. Instead:
  - Add any new nodes/edges that don't exist.
  - Remove any nodes/edges that are no longer in the backend data.
  - Re-run the `dagre` layout (`cy.layout({ name: 'dagre' }).run()`) so the graph physically animates and restructures as nodes appear and disappear during the 1-hour chaos simulation.

### 3. Observability
- The terminal output of `scale_generator.go` must occasionally (e.g., every 1 minute) print a periodic Latency & Panic check so the user knows the API is surviving the sustained load.
