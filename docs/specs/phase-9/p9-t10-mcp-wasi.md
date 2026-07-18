# P9-T10: MCP Server WASI Distribution

## Overview
To provide a secure, sandboxed execution environment for AI coding assistants (Cursor, Claude Desktop), we must distribute the Substrate MCP server as a WebAssembly System Interface (WASI) binary. 

## Requirements
1. **Build Matrix**: Extend the `Makefile` and GitHub Actions pipeline to compile the MCP server using `GOOS=wasip1 GOARCH=wasm`.
2. **Wasmtime Execution**: Document and ensure compatibility with Node.js WASI implementation and `wasmtime` for executing the binary locally.
3. **Stdio Protocol**: Validate that the JSON-RPC Model Context Protocol functions flawlessly over `stdio` within the WASI sandbox.
4. **Publishing**: Automate the publishing of the `.wasm` binary to the GitHub releases page.

## Acceptance Criteria
- `substrate-mcp.wasm` successfully handles `tools/list` and `tools/call` requests over standard IO in a WASI runtime.
- GitHub Actions automatically builds the `.wasm` file.
