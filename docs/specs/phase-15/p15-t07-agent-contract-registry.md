# P15-T07: Substrate for AI Agents (Agent Contract Registry)

## Objective
As the industry moves towards Agentic workflows, autonomous AI agents rely entirely on API contracts (like Model Context Protocol) to function. If an API drops a required tool parameter, the agent hallucinates or loops indefinitely. Substrate will become the "API Governance Platform for the Agentic Era" by tracking which AI agents consume which API tools, and blocking breaking changes that would break agent workflows.

## Core Features
1. **Agent Consumer Identification:**
   Agents register themselves with Substrate, similar to frontend consumer manifests (P10-T10), but explicitly defining the **Tools** and **Prompts** they rely on.
2. **Agent-Aware Diffing:**
   When the MCP Runtime Diff engine (P10-T07) detects a breaking change to a tool (e.g., `execute_sql` removes the `database` parameter), Substrate queries the registry.
3. **Automated Agent Prompt Updates:**
   Substrate uses an LLM to generate an automated PR in the AI Agent's repository, updating its system prompt or tool-calling JSON schema to accommodate the upstream breaking change.

## Deliverables
- `api/internal/agents/`: Registry module for storing Agent tool dependencies.
- `api/internal/handlers/agents.go`: API endpoints for agents to register their required tools.
- `engine/pkg/ai/agent_fix.go`: Logic to auto-generate a fix PR for the broken AI agent.
