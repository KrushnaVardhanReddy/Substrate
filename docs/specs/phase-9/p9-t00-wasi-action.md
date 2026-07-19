# P9-T00: GitHub Marketplace Launch (WASI)

## 1. Objective
Package the Substrate CLI into a WASM binary (`GOOS=wasip1 GOARCH=wasm`) and wrap it in a GitHub Action. This provides a sub-second, Docker-less CI/CD diffing experience for our users, allowing us to publish to the GitHub Marketplace as a native Action.

## 2. Architecture & Rationale
Currently, users pull a Docker container to run the Substrate CLI in their CI/CD pipelines. This adds 15-30 seconds of overhead just to spin up the container.
By compiling Go to `wasip1/wasm`, we can bundle the binary directly into a JavaScript-based GitHub Action. Node.js (which GitHub Actions natively supports) will execute the WASM binary using the `@bytecodealliance/wasi` bindings in milliseconds.

## 3. Implementation Steps
1. **The Build Pipeline (`scripts/build_wasi.sh`):**
   - Create a build script that runs `GOOS=wasip1 GOARCH=wasm go build -o action/substrate.wasm ./engine/cmd/substrate`.
2. **The GitHub Action (`action/action.yml`):**
   - Define a standard `runs: using: "node20", main: "index.js"` metadata file.
   - Accept inputs for `base_schema`, `head_schema`, and `config`.
3. **The Node Wrapper (`action/index.js`):**
   - Import the built-in Node `fs` and `wasi` modules.
   - Instantiate the WebAssembly module with `substrate.wasm`.
   - Pass the GitHub Action input parameters as arguments to the WASM binary execution.

## 4. UX / Go-to-Market
Once merged, we will tag a v1 release and publish `action.yml` to the GitHub Actions Marketplace, advertising "0.02s API Breakage Detection."
