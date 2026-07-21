CREATE TABLE IF NOT EXISTS org_kms_config (
    org_name TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    key_arn TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
