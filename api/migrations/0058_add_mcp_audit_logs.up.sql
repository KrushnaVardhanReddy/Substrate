CREATE TABLE mcp_audit_logs (
    id SERIAL PRIMARY KEY,
    agent_id INT,
    tool_name TEXT NOT NULL,
    request_payload JSONB NOT NULL,
    response_payload JSONB,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
