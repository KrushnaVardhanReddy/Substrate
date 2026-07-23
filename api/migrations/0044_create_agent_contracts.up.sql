CREATE TABLE agent_consumers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_name TEXT NOT NULL,
    owner TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(owner, repo_name)
);

CREATE TABLE agent_tool_dependencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agent_consumers(id) ON DELETE CASCADE,
    tool_name TEXT NOT NULL,
    parameters_jsonb JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
