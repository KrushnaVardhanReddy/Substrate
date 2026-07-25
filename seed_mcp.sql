INSERT INTO organizations (id, github_installation_id, github_org_name) VALUES
('11111111-1111-1111-1111-111111111111', 123456, 'mcp-org')
ON CONFLICT (id) DO NOTHING;

INSERT INTO repositories (id, org_id, github_repo_id, name, full_name) VALUES 
('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 101, 'backend', 'mcp-org/backend'),
('33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 102, 'frontend', 'mcp-org/frontend'),
('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 103, 'discovery-test-repo', 'mcp-org/discovery-test-repo')
ON CONFLICT (id) DO NOTHING;

INSERT INTO contracts (id, repo_id, schema_type, spec_path, branch, latest_commit_sha, raw_content) VALUES 
('55555555-5555-5555-5555-555555555555', '22222222-2222-2222-2222-222222222222', 'openapi', 'openapi.yaml', 'main', 'sha-backend', 'openapi: 3.0.0')
ON CONFLICT (id) DO NOTHING;

INSERT INTO dependencies (id, consumer_repo_id, provider_contract_id) VALUES 
('66666666-6666-6666-6666-666666666666', '33333333-3333-3333-3333-333333333333', '55555555-5555-5555-5555-555555555555')
ON CONFLICT (id) DO NOTHING;
