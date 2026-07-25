# PGlite E2E Infrastructure Spec

## Objective
Replace the external PostgreSQL database dependency in the E2E test suite (`scripts/e2e/*.go`) with a completely self-contained, in-memory PGlite instance.

## Requirements

1. **Vendored Dependency:** 
   - The `@electric-sql/pglite` WASM package must be downloaded and vendored into `scripts/e2e/vendor/pglite`. No network calls during test execution.

2. **Node.js Wrapper (`pglite_server.js`):**
   - A lightweight script that spins up a local TCP server acting as a PostgreSQL wire protocol listener using the vendored PGlite WASM.
   - Must automatically execute the existing `seed.sql` to initialize the schema and base data immediately upon startup.
   - Must bind to a dynamic port or a configurable local port (e.g., `5432`).

3. **Test Runner Integration (`run.sh`):**
   - Must start the `pglite_server.js` as a background process.
   - Must poll/wait for the TCP port to become available.
   - Must export the `DATABASE_URL` environment variable pointing to the PGlite port.
   - Must execute the Go E2E test suite (`go test ./...`).
   - Must trap exit signals and definitively kill the Node.js background process upon test completion (pass or fail).

4. **Go Test Suite Modifications:**
   - Existing Go E2E tests must remain in Go.
   - Tests must read the `DATABASE_URL` environment variable injected by the runner.
   - Tests must assume the database is fresh and seeded at the start of the entire suite run.
