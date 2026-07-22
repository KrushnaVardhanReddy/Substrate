#!/bin/bash
set -e

# Change to the root directory
cd "$(dirname "$0")/.."

# Ensure the action directory exists
mkdir -p action

# Build the WASM binary
cd engine
GOOS=wasip1 GOARCH=wasm go build -o ../action/substrate.wasm ./cmd/substrate
echo "Build complete: action/substrate.wasm"
