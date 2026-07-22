# P9-T01: Shift-Left IDE Plugins (VSCode Extension / MCP)

## Objective
Develop a VSCode Extension and integrate with the existing MCP (Model Context Protocol) server to bring Substrate's blast radius and dependency graph intelligence directly into the developer's IDE, enabling a "shift-left" workflow.

## Context
Developers currently must commit their changes or use the web dashboard to see the downstream impact of their schema changes. By shifting this intelligence left, developers can get real-time feedback within VSCode as they edit `substrate.yaml` or OpenAPI specs, catching breaking changes before they are ever committed.

## Requirements

### 1. VSCode Extension
- **Syntax Highlighting & Validation:** Provide autocomplete, linting, and hover documentation for `substrate.yaml` using JSON Schema.
- **Inline Blast Radius:** When a developer modifies a schema (e.g., an OpenAPI file), query the local or remote Substrate engine to calculate the blast radius and display warnings inline (using VSCode Diagnostics) if downstream consumers will be broken.
- **Dependency Graph View:** Embed a webview panel within VSCode that renders a localized version of the SvelteFlow dependency graph for the current repository.

### 2. MCP Integration
- Extend the `substrate-mcp` server to expose tools for:
  - Validating local `substrate.yaml` files.
  - Querying the local API graph.
  - Analyzing the local impact of staged git changes.
- Ensure the MCP server can be consumed by AI assistants like Claude Desktop or Cursor for contextual code generation.

## Acceptance Criteria
1. VSCode extension can be installed and provides syntax validation for `substrate.yaml`.
2. Modifying a tracked schema file highlights breaking changes directly in the editor using diagnostics.
3. The MCP server provides the necessary tools and returns formatted impact analysis data.
