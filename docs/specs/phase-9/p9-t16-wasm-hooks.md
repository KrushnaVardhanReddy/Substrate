# P9-T16: WASM Git Pre-Commit Hooks

## Objective
Prevent developers from accidentally committing breaking API changes by running Substrate's diff engine locally via a `git pre-commit` hook. The engine must compile to WebAssembly (WASM) for immediate execution (0.02s) without requiring a heavy Go installation or Docker container on the developer's laptop.

## Architecture
1. **WASM Compilation Target:**
   - The core `engine/` logic must be compiled to `substrate-cli.wasm`.
   - Ensure the WASM build strips out any CGO dependencies or database drivers (like SQLite or Postgres) that cannot be compiled to WASM.

2. **Node.js Wrapper (`@substrate/cli`):**
   - Create a lightweight npm package `packages/cli` that wraps the WASM binary.
   - It will use Node.js's native `fs` and `child_process` modules to read `base` and `head` files, pass them to the WASM binary, and parse the JSON response.
   - If a breaking change is detected, the Node.js script exits with code `1`, aborting the git commit.

3. **Husky Integration:**
   - Generate documentation and an init script so teams can simply run `npx substrate init --husky`.
   - This automatically adds the hook to `.husky/pre-commit`:
     ```bash
     npx @substrate/cli diff --base origin/main --head HEAD --fail-on-breaking
     ```

## Deliverables
- `engine/cmd/wasm/main.go` entrypoint for the WASM engine.
- `packages/cli/` npm module (Node wrapper for WASM).
- E2E tests simulating a `git commit` that gets rejected due to a breaking OpenAPI change.
