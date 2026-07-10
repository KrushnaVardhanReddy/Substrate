CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE organizations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  github_installation_id BIGINT UNIQUE NOT NULL,
  github_org_name TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE repositories (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
  github_repo_id BIGINT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  full_name TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE contracts (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  repo_id UUID REFERENCES repositories(id) ON DELETE CASCADE,
  schema_type TEXT NOT NULL,
  spec_path TEXT NOT NULL,
  branch TEXT NOT NULL DEFAULT 'main',
  latest_commit_sha TEXT,
  raw_content TEXT NOT NULL,
  synced_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(repo_id, spec_path, branch)
);

CREATE TABLE dependencies (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  consumer_repo_id UUID REFERENCES repositories(id) ON DELETE CASCADE,
  provider_contract_id UUID REFERENCES contracts(id) ON DELETE CASCADE,
  status TEXT NOT NULL DEFAULT 'active',
  last_checked_at TIMESTAMPTZ,
  UNIQUE(consumer_repo_id, provider_contract_id)
);
