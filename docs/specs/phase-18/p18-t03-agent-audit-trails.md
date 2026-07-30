# Phase 18: Agent Execution Audit Trails (P18-T03)

## Overview
Substrate acts as the bridge between LLM agents (Claude/Cursor via MCP) and enterprise API governance. When an AI agent performs an action (e.g., retrieving a blast radius, rolling back a deployment, checking compatibility), we need a cryptographically sound, verifiable audit log of that execution to track AI hallucination rates and provide forensics.

This task introduces the "AI Forensics Dashboard" backend logging mechanism.

## Technical Requirements

### 1. Database Schema
Create a new SQL migration in `api/migrations` (e.g., `0058_add_mcp_audit_logs.up.sql`) to create the `mcp_audit_logs` table.
```sql
CREATE TABLE mcp_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id TEXT NOT NULL, -- e.g., 'cursor', 'claude-desktop', or user id
    tool_name TEXT NOT NULL, -- e.g., 'get_blast_radius'
    request_payload JSONB NOT NULL, -- The exact args the AI sent
    response_payload JSONB NOT NULL, -- The context Substrate injected back
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_mcp_audit_logs_tool ON mcp_audit_logs(tool_name);
```

### 2. SQLC Queries
Add a file `api/internal/db/queries/mcp_audit.sql`:
```sql
-- name: InsertMCPAuditLog :exec
INSERT INTO mcp_audit_logs (agent_id, tool_name, request_payload, response_payload)
VALUES ($1, $2, $3, $4);

-- name: ListMCPAuditLogs :many
SELECT * FROM mcp_audit_logs ORDER BY executed_at DESC LIMIT $1 OFFSET $2;
```
Run `make sqlc` from the `api` directory.

### 3. MCP Server Interceptor Logging
Update `api/internal/mcp/server.go`. Inside the `CallTool` handler (or a dedicated wrapper around tool executions), ensure that every time an MCP tool is invoked, it logs the execution asynchronously or synchronously into the `mcp_audit_logs` table.
- You can parse an `AgentID` from the incoming MCP request headers or connection metadata if available, otherwise default to `"unknown-agent"`.
- Log the exact JSON of the tool arguments as `request_payload`.
- Log the exact JSON response returned to the agent as `response_payload`.

### 4. REST API Endpoint
Expose `GET /api/v1/ai/audit` which returns the paginated audit logs, so the Svelte dashboard can eventually consume it.
Wire this up in `api/internal/server/router.go`.

## Rules
- No mocks! Save everything directly to PostgreSQL.
- Ensure the JSON marshaling of `request_payload` and `response_payload` does not crash if the payloads contain unmarshalable data (handle errors gracefully).
