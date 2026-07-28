# Phase 18 - Task 01: Granular MCP Tool Governance

## 1. Goal
Allow Forward Deployed AI Engineers to define "Agent Profiles" in Substrate that restrict which MCP tools an AI agent is allowed to invoke, and enforce Human-in-the-Loop (HITL) approval gates for destructive operations.

## 2. Requirements
- Add `agent_profiles` table storing profile definitions (name, allowed tools list, HITL-required methods).
- Expose CRUD endpoints:
  - `POST /api/v1/mcp/profiles` — create a profile.
  - `GET /api/v1/mcp/profiles/{org}` — list all profiles for an org.
  - `DELETE /api/v1/mcp/profiles/{id}` — delete a profile.
- When an MCP session is initiated, accept an optional `profile_id` query param. The MCP server filters its `tools/list` response to only include tools allowed by the profile.
- If an agent invokes a `POST`/`DELETE` tool and the profile has `hitl: true`, the MCP server must return a `pending_approval` response instead of executing immediately, and store the pending call in a `hitl_queue` table.
- Expose `GET /api/v1/mcp/hitl-queue/{org}` so a human dashboard can review and approve/reject pending calls.

## 3. Database Schema
```sql
CREATE TABLE agent_profiles (
  id            SERIAL PRIMARY KEY,
  org           TEXT NOT NULL,
  name          TEXT NOT NULL,
  allowed_tools TEXT[] NOT NULL,
  hitl_enabled  BOOLEAN NOT NULL DEFAULT false,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE hitl_queue (
  id            SERIAL PRIMARY KEY,
  org           TEXT NOT NULL,
  profile_id    INT REFERENCES agent_profiles(id),
  tool_name     TEXT NOT NULL,
  arguments     JSONB NOT NULL,
  status        TEXT NOT NULL DEFAULT 'pending', -- pending | approved | rejected
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  resolved_at   TIMESTAMPTZ
);
```
