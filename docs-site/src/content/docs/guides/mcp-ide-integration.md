---
title: MCP Server & AI IDE Integration
description: Connect Substrate to Cursor, Windsurf, or Claude Desktop for real-time schema intelligence.
---

The Substrate MCP (Model Context Protocol) Server transforms your AI IDE into a schema-aware coding assistant. Instead of asking the AI to guess about your API contracts, it can directly query Substrate for live, organizational data.

## What the AI Can Do With MCP

Once connected, your AI assistant gains access to 7 specialized tools:

| Tool | Runs | What it does |
|---|---|---|
| `analyze_repository` | Locally | Scans your directory and detects schema files (`openapi.yaml`, `.proto`, `.graphql`, etc.) to auto-generate your `substrate.yaml`. |
| `check_compatibility` | Locally | Validates a *drafted* schema change against the diff engine before you even commit. |
| `get_substrate_docs` | Locally | Returns the full `substrate.yaml` configuration guide so the AI can answer config questions accurately. |
| `execute_cli_command` | Locally | Runs `substrate diff` or `substrate init` on your behalf. |
| `get_dependency_graph` | Registry API | Returns the full cross-repo dependency map for your organization. |
| `get_schema_file` | Registry API | Fetches the exact contract schema of any downstream consumer team. |
| `get_breaking_change_history` | Registry API | Queries the history of past breaking changes on a repository. |

## Connecting to Cursor

Add the following to your Cursor MCP configuration (`~/.cursor/mcp.json`):

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

## Connecting to Claude Desktop

Add the following to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

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

## Local Development

For local development, omit `REGISTRY_API_URL` and the server will default to `http://localhost:8090`, which is the default port for the Substrate Registry API.

```json
{
  "mcpServers": {
    "substrate": {
      "command": "substrate-mcp"
    }
  }
}
```

> **Note:** The `substrate-mcp` binary must be installed in your `$PATH`. See the [Getting Started](/tutorials/getting-started/) guide for installation instructions.
