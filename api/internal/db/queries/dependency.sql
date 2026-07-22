-- name: UpsertDependency :exec
INSERT INTO dependencies (consumer_repo_id, provider_contract_id, confidence_score, required_notice_days)
VALUES ($1, $2, $3, $4)
ON CONFLICT (consumer_repo_id, provider_contract_id) DO UPDATE
  SET confidence_score    = GREATEST(dependencies.confidence_score, EXCLUDED.confidence_score),
      required_notice_days = EXCLUDED.required_notice_days;

-- name: UpdateDependencyConfidence :exec
UPDATE dependencies
SET confidence_score = confidence_score + $3
FROM contracts c
JOIN repositories r ON c.repo_id = r.id
WHERE dependencies.provider_contract_id = c.id
  AND r.full_name = $1
  AND c.spec_path LIKE '%' || $2 || '%';

-- name: UpdateDependencyStatus :exec
UPDATE dependencies
SET status = $3
WHERE consumer_repo_id = $1
  AND provider_contract_id = $2;

-- name: CountDownstreamDependencies :one
SELECT COUNT(*)
FROM dependencies d
JOIN contracts c ON d.provider_contract_id = c.id
JOIN repositories r ON c.repo_id = r.id
WHERE r.full_name = $1;
