# P10-T07: MCP (Model Context Protocol) Runtime Diffing

## Objective
As AI Agents become first-class citizens in enterprise architecture, the tools and resources they rely on (exposed via MCP) become critical infrastructure. Unlike OpenAPI, MCP schemas are dynamically discovered at runtime. Substrate must introduce a specialized "Runtime Diff Engine" capable of spinning up an MCP server during CI/CD, querying its capabilities, and detecting breaking changes before they disrupt autonomous agent workflows.

## Breaking Change Definitions for MCP
Because AI agents rely on precise JSON schemas for tool execution, even minor changes can break an autonomous workflow. Substrate will monitor for the following breaking changes:

### 1. Tool Breaking Changes
- **Tool Removal:** Removing an existing tool (e.g., `execute_sql`).
- **New Required Parameter:** Adding a required argument to an existing tool (e.g., adding `region` when the agent only knows to send `query`).
- **Parameter Type Change:** Changing a parameter from `string` to `integer`.
- **Parameter Removal:** Removing a parameter that agents might still be attempting to pass (causing strict schema validation failures).

### 2. Resource Breaking Changes
- **Resource URI Removal:** Deleting a previously available resource path (e.g., `file:///logs/system.log`).
- **MIME Type Change:** Changing the format of a resource (e.g., switching from `application/json` to `text/csv`).

### 3. Prompt Breaking Changes
- **Prompt Removal:** Deleting a predefined prompt.
- **Required Argument Addition:** Adding a new required argument to a prompt template.

## Implementation Plan
1. **Dynamic SDK Runner:** Substrate will integrate a lightweight MCP client (via `@modelcontextprotocol/sdk`).
2. **Runtime Introspection:** During the CI/CD GitHub Action, Substrate will execute the server (e.g., `npx @smithery/cli run`) in a sandbox.
3. **State Capture:** Substrate calls `tools/list`, `resources/list`, and `prompts/list`.
4. **Diff Engine:** The captured JSON state is compared against the previously captured state in the Substrate database.
5. **PR Blocking:** If a breaking change is detected, the PR is blocked, preventing the AI ecosystem from breaking.
