-- name: UpsertAirbyteSource :one
INSERT INTO airbyte_sources (org, name, connector, config_enc)
VALUES ($1, $2, $3, $4)
ON CONFLICT (org, name) DO UPDATE
SET connector = EXCLUDED.connector,
    config_enc = EXCLUDED.config_enc,
    updated_at = NOW()
RETURNING *;

-- name: ListAirbyteSources :many
SELECT id, org, name, connector, config_enc, created_at, updated_at
FROM airbyte_sources
WHERE org = $1
ORDER BY created_at DESC;

-- name: GetAirbyteSource :one
SELECT id, org, name, connector, config_enc, created_at, updated_at
FROM airbyte_sources
WHERE org = $1 AND connector = $2;

-- name: GetAirbyteSourceByID :one
SELECT id, org, name, connector, config_enc, created_at, updated_at
FROM airbyte_sources
WHERE org = $1 AND id = $2;

-- name: DeleteAirbyteSource :exec
DELETE FROM airbyte_sources
WHERE org = $1 AND id = $2;

-- name: InsertAirbyteStagedRecord :one
INSERT INTO airbyte_staged_records (org, source_id, stream, data)
VALUES ($1, $2, $3, $4)
RETURNING *;
