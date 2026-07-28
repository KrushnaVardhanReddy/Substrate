INSERT INTO contracts (id, repo_id, schema_type, spec_path, branch, latest_commit_sha, raw_content) VALUES 
('c0000000-0000-0000-0000-000000000001', '5469940f-3399-4fdb-8684-403fc181b102', 'openapi', 'openapi.yaml', 'main', 'sha123', 'openapi: 3.0.0');

INSERT INTO dependencies (consumer_repo_id, provider_contract_id, status) VALUES 
('a882fd25-c6d0-4713-bb3d-113228d12b7e', 'c0000000-0000-0000-0000-000000000001', 'active');
