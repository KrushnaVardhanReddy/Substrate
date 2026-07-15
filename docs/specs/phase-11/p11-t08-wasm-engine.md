# Spec: P11-T08 - Substrate WASM Engine (In-Browser Diffing)

## 1. Overview
Compile the Go Diff Engine to WebAssembly (`GOOS=js GOARCH=wasm`) so users can test schema breakages instantly in the Svelte dashboard with zero backend latency.

## 2. Requirements
- Create `engine/cmd/wasm/main.go` that exports a `diff_schemas` function to JavaScript using the `syscall/js` package.
- The function should take two YAML strings (base, revision) and return a JSON DiffReport.
- Add a Makefile target `build-wasm` to compile this using `GOOS=js GOARCH=wasm go build -o ../dashboard/static/engine.wasm ./engine/cmd/wasm/main.go`.

## 3. Stitch & Jules Workflow
- **Jules:** Implement the Go WASM entrypoint and the Makefile target.
