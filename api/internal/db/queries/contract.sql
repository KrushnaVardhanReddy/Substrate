-- name: UpsertContract :one
INSERT INTO contracts (repo_id, schema_type, spec_path, branch, latest_commit_sha, raw_content, synced_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
ON CONFLICT (repo_id, spec_path, branch) DO UPDATE
  SET schema_type       = EXCLUDED.schema_type,
      latest_commit_sha = EXCLUDED.latest_commit_sha,
      raw_content       = EXCLUDED.raw_content,
      synced_at         = NOW()
RETURNING id;

-- name: GetContractsByProviderFullName :many
SELECT c.id, c.repo_id, c.schema_type, c.spec_path, c.branch,
       c.latest_commit_sha, c.raw_content, c.synced_at,
       c.is_encrypted, COALESCE(c.kms_key_arn, '') AS kms_key_arn
FROM contracts c
JOIN repositories r ON c.repo_id = r.id
WHERE r.full_name = $1;

-- name: GetConsumersByProviderContract :many
SELECT
  d.consumer_repo_id,
  r.full_name         AS consumer_full_name,
  c.raw_content       AS contract_raw_content,
  d.required_notice_days
FROM dependencies d
JOIN repositories r ON d.consumer_repo_id = r.id
JOIN contracts    c ON c.repo_id = r.id
WHERE d.provider_contract_id = $1;
