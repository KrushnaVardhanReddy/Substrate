# Substrate MCP Server Deployment and IDE Integration

## 1. Overview

The Substrate MCP (Model Context Protocol) server allows LLMs to query the Substrate contract registry to understand cross-repo dependencies and schemas. This provides AI coding assistants with deep, context-aware intelligence about how different services and data contracts interact across your entire organization.

The MCP server supports two operating modes:
- **`stdio`**: For local IDE integration (communicates via standard input/output).
- **`sse`**: For hosted Server-Sent Events (communicates over HTTP).

Available tools provided by the Substrate MCP server include:
- `get_schema_file`: Retrieves a specific schema file from a repository.
- `analyze_repository`: Analyzes a repository's dependencies and consumers.
- `get_substrate_docs`: Fetches documentation about Substrate integration and usage.

## 2. Authentication

All MCP tools must communicate with the upstream Registry API server to fetch the required schema and dependency data.

To do this, the MCP server requires the `REGISTRY_API_TOKEN` environment variable to be set. This token authenticates the MCP server's requests against the Registry API, ensuring secure access to your organization's schema contracts.

## 3. IDE Integration Guides

### A. Cursor IDE

To configure the Substrate MCP server in Cursor, create or update the `.cursor/mcp.json` file in your project or workspace root with the following configuration:

**File:** `.cursor/mcp.json`

```json
{
  "mcpServers": {
    "substrate": {
      "command": "substrate-mcp",
      "args": [
        "--registry-url",
        "http://localhost:8090",
        "--token",
        "${REGISTRY_API_TOKEN}"
      ]
    }
  }
}
```
*(Assumes `substrate-mcp` is installed via `go install` or is otherwise available in your PATH).*

### B. Claude Desktop

To configure the Substrate MCP server in Claude Desktop on macOS, update your Claude Desktop configuration file with the following setup:

**File:** `~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "substrate": {
      "command": "substrate-mcp",
      "args": [
        "--registry-url",
        "http://localhost:8090",
        "--token",
        "${REGISTRY_API_TOKEN}"
      ]
    }
  }
}
```

## 4. Example Prompts

Once the MCP server is successfully connected to your IDE, you can ask your AI assistant questions like:

- "What are the downstream consumers of the checkout-service schema?"
- "Fetch the schema for org/repo and tell me if dropping the 'email' field breaks any consumers."
