CREATE TABLE IF NOT EXISTS consumer_manifests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_repo TEXT NOT NULL,
    consumer_repo TEXT NOT NULL,
    consumed_fields JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_consumer_manifests_provider_repo ON consumer_manifests(provider_repo);
CREATE INDEX idx_consumer_manifests_consumer_repo ON consumer_manifests(consumer_repo);
