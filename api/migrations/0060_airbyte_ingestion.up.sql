CREATE TABLE airbyte_sources (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org         TEXT NOT NULL,
    name        TEXT NOT NULL,                  -- human label, e.g. "HubSpot CRM"
    connector   TEXT NOT NULL,                  -- e.g. "source-hubspot"
    config_enc  BYTEA NOT NULL,                 -- AES-256-GCM encrypted JSON config
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org, name)
);

CREATE TABLE airbyte_staged_records (
    id          BIGSERIAL PRIMARY KEY,
    org         TEXT NOT NULL,
    source_id   UUID NOT NULL REFERENCES airbyte_sources(id) ON DELETE CASCADE,
    stream      TEXT NOT NULL,                  -- Airbyte stream name, e.g. "contacts"
    data        JSONB NOT NULL,
    ingested_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_airbyte_staged_records_org_stream
    ON airbyte_staged_records (org, stream, ingested_at DESC);
