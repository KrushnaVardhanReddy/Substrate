-- name: UpsertOrgIntegration :one
INSERT INTO org_integrations (org_id, provider, api_key_encrypted, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (org_id, provider) DO UPDATE
  SET api_key_encrypted = EXCLUDED.api_key_encrypted,
      updated_at = NOW()
RETURNING id, org_id, provider, api_key_encrypted, created_at, updated_at;

-- name: GetOrgIntegration :one
SELECT id, org_id, provider, api_key_encrypted, created_at, updated_at
FROM org_integrations
WHERE org_id = $1 AND provider = $2;
