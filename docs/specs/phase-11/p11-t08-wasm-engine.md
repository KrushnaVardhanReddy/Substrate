# Spec: P11-T08 - Substrate WASM Engine (In-Browser Diffing)

## 1. Overview
Compile the Go Diff Engine to WebAssembly (`GOOS=js GOARCH=wasm`) so users can test schema breakages instantly in the Svelte dashboard with zero backend latency.

## 2. Requirements
- Create `engine/cmd/wasm/main.go` that exports a `diff_schemas` function to JavaScript using the `syscall/js` package.
- The function should take two YAML strings (base, revision) and return a JSON DiffReport.
- Add a Makefile target `build-wasm` to compile this using `GOOS=js GOARCH=wasm go build -o ../dashboard/static/engine.wasm ./engine/cmd/wasm/main.go`.

## 3. Implementation Steps

### Phase A: Backend Compilation (Completed via PR 113)
- ✅ Create `engine/cmd/wasm/main.go` that exports a `diff_schemas` function to JavaScript using the `syscall/js` package.
- ✅ The function should take two YAML strings (base, revision) and return a JSON DiffReport.
- ✅ Add a Makefile target `build-wasm` to compile this into `dashboard/static/engine.wasm`.

### Phase B: Frontend WASM Loader Integration (Pending)
- Fetch the official `wasm_exec.js` script (matching the Go version used to compile the WASM binary) and place it in `dashboard/static/`.
- Include `wasm_exec.js` in the `app.html` header.
- Create a reusable Svelte/JS WASM Loader utility (`dashboard/src/lib/utils/wasmLoader.ts`) that initializes the WebAssembly instance via `WebAssembly.instantiateStreaming` or `WebAssembly.instantiate`.
- Update `DiffViewer.svelte` (or a dedicated WASM testing component) to call `window.diff_schemas()` and render the JSON DiffReport dynamically, proving zero-latency in-browser diffing.

## 4. Stitch & Jules Workflow
- **Jules:** Implemented the Go WASM entrypoint and the Makefile target (✅ Done).
- **Stitch:** Implement the Frontend WASM Loader, copy `wasm_exec.js`, and wire it up to the `DiffViewer` (⏳ Next).
