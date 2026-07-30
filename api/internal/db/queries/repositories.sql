-- name: UpsertRepo :one
-- Include metadata in upsert
INSERT INTO repositories (org_id, github_repo_id, name, full_name, metadata)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (github_repo_id) DO UPDATE
  SET name       = EXCLUDED.name,
      full_name  = EXCLUDED.full_name,
      org_id     = EXCLUDED.org_id,
      metadata   = EXCLUDED.metadata
RETURNING id;

-- name: ListReposByOrg :many
SELECT r.id, r.org_id, r.github_repo_id, r.name, r.full_name, r.metadata, r.created_at
FROM repositories r
JOIN organizations o ON r.org_id = o.id
WHERE o.github_org_name = $1
ORDER BY r.name;

-- name: GetAllRepositories :many
SELECT id, org_id, github_repo_id, name, full_name, metadata, created_at
FROM repositories;

-- name: GetDependencyGraph :many
SELECT
  c_repo.full_name  AS consumer_full_name,
  p_repo.full_name  AS provider_full_name,
  d.status,
  c_repo.metadata   AS consumer_metadata,
  p_repo.metadata   AS provider_metadata,
  COALESCE(rm.predictive_risk_score, 0) AS predictive_risk_score
FROM dependencies d
JOIN contracts    p_c  ON d.provider_contract_id = p_c.id
JOIN repositories p_repo ON p_c.repo_id = p_repo.id
JOIN repositories c_repo ON d.consumer_repo_id   = c_repo.id
LEFT JOIN repo_metrics rm ON rm.repo_id = p_repo.id
JOIN organizations o ON p_repo.org_id = o.id
WHERE o.github_org_name = $1;

-- name: UpsertRepoMetric :exec
INSERT INTO repo_metrics (repo_id, predictive_risk_score, calculated_at)
VALUES ($1, $2, NOW())
ON CONFLICT (repo_id) DO UPDATE
  SET predictive_risk_score = EXCLUDED.predictive_risk_score,
      calculated_at         = NOW();
