-- name: InsertMCPAuditLog :one
INSERT INTO mcp_audit_logs (
    agent_id, tool_name, request_payload, response_payload
) VALUES (
    $1, $2, $3, $4
) RETURNING id;

-- name: ListMCPAuditLogs :many
SELECT id, agent_id, tool_name, request_payload, response_payload, executed_at
FROM mcp_audit_logs
ORDER BY executed_at DESC;
