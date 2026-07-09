# Phase 3: Substrate MCP Server Specification

## 🎯 Overview

The Substrate Model Context Protocol (MCP) Server is what transforms Substrate from a CI/CD gatekeeper into a real-time AI coding assistant. 

By exposing the Substrate Registry API as an MCP server, developers using AI IDEs (Cursor, Windsurf) or AI chat interfaces (Claude Desktop) can natively query the cross-repository dependency graph *before* they even write code.

**Status:** ⏳ READY  
**Owner:** Antigravity (Spec) / Jules (Implementation)

---

## 1. Architecture & Setup

The MCP Server will be embedded directly into our existing Go REST API Server (`api/cmd/server/main.go`). 

We will expose a new endpoint: `POST /mcp`
This endpoint will implement the **JSON-RPC 2.0** protocol as defined by the MCP specification. 

Because we want Cursor/Claude to talk to the local API server while developing, the MCP server must support Standard Input/Output (stdio) integration via a small wrapper script, OR the AI client must support SSE (Server-Sent Events) HTTP transport. For maximum compatibility with IDEs, we will implement the **stdio** transport in a standalone CLI command: `substrate-mcp`.

---

## 2. Exposed MCP Tools

The server will expose the following tools to the LLM.

### Tool 1: `get_dependency_graph`
**Description:** Returns the full map of which repositories consume which providers. Useful for understanding the blast radius of a change.
**Input Schema:**
```json
{
  "type": "object",
  "properties": {
    "org": {
      "type": "string",
      "description": "The GitHub organization name (e.g. 'myorg')"
    }
  },
  "required": ["org"]
}
```
**Output:** A JSON array of `{ provider, consumer, status }`.

---

### Tool 2: `check_compatibility`
**Description:** Runs a simulated cross-repo compatibility check. The AI can pass in a drafted schema (e.g. OpenAPI YAML) before it's even committed, and Substrate will validate it against all registered downstream consumers.
**Input Schema:**
```json
{
  "type": "object",
  "properties": {
    "provider_repo": {
      "type": "string",
      "description": "The repository being modified (e.g. 'myorg/backend-api')"
    },
    "head_schema_content": {
      "type": "string",
      "description": "The raw string content of the proposed new schema"
    },
    "schema_type": {
      "type": "string",
      "description": "The format (e.g. 'openapi', 'graphql', 'sql', 'salesforce-object')"
    }
  },
  "required": ["provider_repo", "head_schema_content", "schema_type"]
}
```
**Output:** A detailed JSON report of breaking changes for every downstream consumer affected.

---

### Tool 3: `get_breaking_change_history`
**Description:** Retrieves a log of all historical breaking changes that have occurred on a specific repository. Useful for debugging production incidents.
**Input Schema:**
```json
{
  "type": "object",
  "properties": {
    "repo": {
      "type": "string",
      "description": "The repository to check (e.g. 'myorg/backend-api')"
    },
    "limit": {
      "type": "integer",
      "description": "Maximum number of records to return (default 10)"
    }
  },
  "required": ["repo"]
}
```
**Output:** An array of historical diff reports with timestamps and Git SHAs.

---

### Tool 4: `get_substrate_docs`
**Description:** Retrieves the official Substrate configuration guide and `substrate.yaml` schema. The AI uses this to learn the rules before helping the user configure their repository.
**Input Schema:**
```json
{
  "type": "object",
  "properties": {}
}
```
**Output:** A string containing the markdown documentation for `substrate.yaml`.

---

### Tool 5: `analyze_repository`
**Description:** Scans the user's local directory tree to automatically detect API contracts (e.g., finding `openapi.yaml`, `.proto` files, or `.graphql` files) to recommend a `substrate.yaml` configuration.
**Input Schema:**
```json
{
  "type": "object",
  "properties": {
    "directory": {
      "type": "string",
      "description": "The local directory path to scan (default: current directory)"
    }
  }
}
```
**Output:** A JSON list of detected contract files and their inferred types.

---

### Tool 6: `execute_cli_command`
**Description:** Allows the AI to safely execute the local `substrate` CLI binary (e.g., to run a local dry-run diff or initialize a repository).
**Input Schema:**
```json
{
  "type": "object",
  "properties": {
    "args": {
      "type": "array",
      "items": { "type": "string" },
      "description": "Arguments to pass to the substrate CLI (e.g., ['diff', '--base', 'a.yaml', '--head', 'b.yaml'])"
    }
  },
  "required": ["args"]
}
```
**Output:** The stdout and exit code of the CLI execution.

---

## 3. Implementation Plan (Jules)

1. Create `api/internal/mcp/server.go`.
2. Implement the standard MCP JSON-RPC handlers for `initialize`, `tools/list`, and `tools/call`.
3. Wire the `tools/call` requests directly to the existing Registry PostgreSQL database and the `diff` engine Go packages.
4. Create a new CLI entrypoint `engine/cmd/substrate-mcp/main.go` that reads JSON-RPC from `os.Stdin` and writes to `os.Stdout`, acting as a bridge to the `api/internal/mcp` logic.
