CREATE TABLE breaking_change_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_id UUID REFERENCES repositories(id) ON DELETE CASCADE,
    org_name TEXT NOT NULL,
    repo_name TEXT NOT NULL,
    git_sha TEXT NOT NULL,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    breaking_changes JSONB NOT NULL
);
