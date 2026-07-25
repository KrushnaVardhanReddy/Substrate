INSERT INTO repositories (id, org, name, full_name, github_repo_id, metadata, status, created_at, updated_at) VALUES 
('mcp-org/backend', 'mcp-org', 'backend', 'mcp-org/backend', 101, '{"type": "backend", "team": "core"}', 'SAFE', NOW(), NOW()),
('mcp-org/frontend', 'mcp-org', 'frontend', 'mcp-org/frontend', 102, '{"type": "frontend", "team": "product"}', 'SAFE', NOW(), NOW()),
('mcp-org/discovery-test-repo', 'mcp-org', 'discovery-test-repo', 'mcp-org/discovery-test-repo', 103, '{"type": "backend", "team": "discovery"}', 'SAFE', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO schemas (id, repo, org, format, version, raw_content, parsed_json, commit_sha, file_path, status, created_at) VALUES 
('backend-schema-1', 'mcp-org/backend', 'mcp-org', 'openapi', '1.0.0', 'openapi: 3.0.0', '{}', 'sha-backend', 'openapi.yaml', 'active', NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO edges (id, org, provider, consumer, protocol, status, meta, created_at, updated_at) VALUES 
('edge-1', 'mcp-org', 'mcp-org/backend', 'mcp-org/frontend', 'openapi', 'SAFE', '{}', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
