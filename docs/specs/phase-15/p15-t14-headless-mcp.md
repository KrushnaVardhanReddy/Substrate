# P15-T14: Headless Substrate (Full MCP Server Parity)

## Objective
As AI assistants (like Cursor, Claude Desktop, and autonomous agents) become the primary interface for software development, Substrate must evolve beyond a GUI/CLI into a fully "Headless" platform. The goal of this task is to ensure that 100% of Substrate's functionality—configuration, diffing, governance, and management—is accessible and controllable exclusively via the Model Context Protocol (MCP).

## Core Features
1. **100% REST API to MCP Mapping:**
   Every single endpoint in the Substrate API must be mapped to an MCP Tool or Resource.
   - **Resources:** Agents can read `substrate://schemas/{org}/{repo}`, `substrate://consumers/{org}/{repo}`, and `substrate://governance-rules`.
   - **Tools:** Agents can execute `bypass_breaking_change`, `configure_kms_byok`, `register_consumer_contract`, `generate_postmortem`, and `request_schema_review`.
2. **Interactive Prompts for Agents:**
   Substrate will expose MCP Prompts that agents can invoke to learn how to interact with the system (e.g., `prompt: substrate_onboarding` which returns the full context of how to write a `.substrate-consumer.yaml` file).
3. **Headless Operations Mode:**
   Substrate can be deployed with `--headless-mcp`, entirely disabling the Web UI and HTTP API, operating strictly over Stdio or SSE for pure agent-to-agent communication in heavily restricted VPCs.

## Deliverables
- `api/internal/mcp/tools.go`: Auto-generation mapping layer that converts standard HTTP route handlers into MCP Tool definitions.
- `api/internal/mcp/resources.go`: Dynamic URI resolution for Substrate's database models.
- Complete E2E testing of the MCP server ensuring an agent can successfully configure and operate a Substrate installation from scratch without human intervention.
