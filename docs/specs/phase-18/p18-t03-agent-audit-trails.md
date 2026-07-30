# Phase 18 - Task 03: Agent Execution Audit Trails

## 1. Goal
Build an AI Forensics logging system that records every MCP tool request, the injected prompt context, and the AI's final payload. This provides a verifiable audit log for tracking LLM hallucinations and tracing AI actions against enterprise APIs.

## 2. Requirements
- Ensure the DB migration perfectly handles JSONB insertion. Unmarshalable payloads must be handled gracefully.
- Do NOT break the existing MCP functionality. Wrap or defer the logging so it does not block or error the actual tool execution.
- No frontend UI is required in this task; only the Go backend logging and REST API endpoint.

## 3. Implementation details
- Create a migration file `0058_add_mcp_audit_logs.up.sql` to add the `mcp_audit_logs` table (agent_id, tool_name, request_payload, response_payload, executed_at).
- Write `InsertMCPAuditLog` and `ListMCPAuditLogs` inside a new `api/internal/db/queries/mcp_audit.sql` file.
- Run `make sqlc` from the `api` folder to generate the types.
- In `api/internal/mcp/server.go`, intercept or wrap tool invocations (like `get_blast_radius`). Serialize the input and output into JSON strings/bytes to store as `request_payload` and `response_payload`.
- Use a non-blocking goroutine or a safe defer pattern to perform the `InsertMCPAuditLog` so MCP latency is not impacted.

## 4. Deliverables
1. FILE: api/migrations/0058_add_mcp_audit_logs.up.sql (CREATE) (and down.sql)
2. FILE: api/internal/db/queries/mcp_audit.sql (CREATE)
3. MODIFY: api/internal/mcp/server.go (Add tool execution logging)
4. FILE: api/internal/handlers/mcp_audit.go (CREATE REST handler)
5. MODIFY: api/internal/server/router.go (Wire up `GET /api/v1/ai/audit`)
6. Run `go test ./...` in the `api` folder.
