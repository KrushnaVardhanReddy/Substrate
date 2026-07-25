# Full-Stack PGlite E2E Infrastructure Spec

## Objective
Replace the external PostgreSQL database dependency in the E2E test suite with a self-contained, in-memory PGlite instance, and expand the test harness to perform Full-Stack testing. The new harness will spin up the database, backend API, MCP server, and SvelteKit frontend to execute API, MCP, and Playwright UI tests in a single offline pass.

## Requirements

1. **Vendored Database:** 
   - The `@electric-sql/pglite` WASM package must be vendored into `scripts/e2e/vendor/pglite`.
   - A Node.js wrapper (`pglite_server.js`) must spin up a local TCP server acting as a PostgreSQL wire protocol listener.
   - The wrapper must automatically execute `seed.sql` and schema files upon startup.

2. **The 4-Tier Tear-Up Runner (`run_full_e2e.sh`):**
   - Must be a resilient bash script that handles background process management.
   - **Tier 1:** Start `pglite_server.js` and wait for the port to be ready.
   - **Tier 2:** Export `DATABASE_URL` and start the Go backend API (`go run ./api/cmd/server`). Wait for `:8090`.
   - **Tier 3:** Start the MCP Server.
   - **Tier 4:** Start the SvelteKit frontend (`npm run dev` inside `/dashboard`). Wait for `:5173`.
   
3. **Test Execution Sequence:**
   - Run Go API Tests: `go test -v ./scripts/e2e/...`
   - Run MCP Tool Tests.
   - Run Playwright UI Tests: `cd dashboard && npx playwright test`

4. **Robust Tear-Down:**
   - Must trap `EXIT`, `SIGINT`, and `SIGTERM`.
   - Must definitively kill the Vite process, MCP process, Go API process, and PGlite Node process regardless of test pass/fail state to prevent zombie ports.

5. **Playwright Initialization:**
   - The dashboard must be initialized with Playwright (`@playwright/test`).
   - A basic skeleton UI test must be provided to prove the browser can hit `:5173` and see the seeded data.
