# Substrate MCP Server: Testing and Usage Guide

The Substrate Model Context Protocol (MCP) Server transforms your IDE (Cursor, Windsurf) or AI chat interfaces (Claude Desktop) into an autonomous schema intelligence agent. This guide covers how to set up, test, and use the MCP server locally.

## Prerequisites

Before using the MCP server, ensure you have built the binaries and placed them in your system's `$PATH`.

```bash
# Build the core CLI
cd engine
go build -o substrate ./cmd/substrate/

# Build the MCP server
go build -o substrate-mcp ./cmd/substrate-mcp/main.go

# Make sure they are in your PATH (example for local testing)
export PATH=$PATH:$(pwd)
```

## How the MCP Server Works

The MCP server operates statelessly over standard input and output using JSON-RPC 2.0. It exposes tools that an AI can use to reason about your repository's API schemas and breaking changes.

### 1. Repository Scanning (`analyze_repository`)
The AI can automatically detect schemas (OpenAPI, GraphQL, SQL, etc.) in your workspace.

**Example test:** Create a sample OpenAPI file in a new directory:
```yaml
# test-mcp-dir/openapi.yaml
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /hello:
    get:
      responses:
        '200':
          description: OK
```
When the `analyze_repository` tool is called on this directory, it will detect the file and correctly identify its format as `openapi`, allowing the AI to auto-generate your `substrate.yaml` configuration.

### 2. Draft Schema Diffing (`check_compatibility`)
Before you even commit a file, the AI can simulate a schema change to see if it causes any breaking changes across your organization.

By passing the raw text of your drafted schema to the MCP server, Substrate will:
1. Create a temporary mock of the schema's base state.
2. Run the `substrate diff` engine locally.
3. Return a JSON DiffReport to the AI.

If the change is safe (like adding a new endpoint), the AI knows it can proceed. If it's a breaking change, the AI will suggest a safer alternative (like adding a `@deprecated` flag).

### 3. Exploring Substrate Documentation (`get_substrate_docs`)
The AI can read the internal Substrate configuration rules. If you ask it "How do I configure a consumer in Substrate?", it will call this tool to retrieve the complete `substrate.yaml` Markdown guide and answer your question accurately.

### 4. Running CLI Commands (`execute_cli_command`)
The AI can invoke the standard `substrate` CLI to perform dry runs or initialize projects. *(Note: This requires the `substrate` binary to be installed in your `$PATH`)*.

### 5. Accessing the Live Registry
For Cross-Repo intelligence, the MCP server connects to your organization's Substrate Registry API.
- **`get_schema_file`**: Fetches the exact schema of a downstream consumer repo so the AI can write integration code against it.
- **`get_breaking_change_history`**: Queries past breaking changes to understand historical deployment patterns.

By default, the server expects the Registry API to be available at `http://localhost:8090`. You can override this by setting the `REGISTRY_API_URL` environment variable.

## Connecting to Cursor

To use this with an AI IDE like Cursor, add the following to your Cursor MCP configuration:

```json
{
  "mcpServers": {
    "substrate": {
      "command": "substrate-mcp",
      "env": {
        "REGISTRY_API_URL": "https://api.substrate.yourcompany.com"
      }
    }
  }
}
```
Once connected, the AI will have full access to your organization's schema intelligence!
