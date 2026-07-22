-- name: RecordBreakingChange :exec
INSERT INTO breaking_change_records (repo_id, org_name, repo_name, git_sha, breaking_changes)
VALUES ($1, $2, $3, $4, $5);

-- name: GetBreakingChangeHistory :many
SELECT id, repo_id, org_name, repo_name, git_sha, timestamp, breaking_changes
FROM breaking_change_records
WHERE org_name = $1 AND repo_name = $2
ORDER BY timestamp DESC
LIMIT $3;

-- name: GetBreakingChangesBetween :many
SELECT id, repo_id, org_name, repo_name, git_sha, timestamp, breaking_changes
FROM breaking_change_records
WHERE timestamp >= $1 AND timestamp <= $2
ORDER BY timestamp DESC;

-- name: CountRecentBreakingChanges :one
SELECT COUNT(*)
FROM breaking_change_records
WHERE repo_id = $1 AND timestamp >= $2;

-- name: GetCommitVelocity :one
SELECT COUNT(*)
FROM contracts
WHERE repo_id = $1 AND synced_at >= $2;

-- name: GetTimeSinceLastBreak :one
SELECT MAX(timestamp)
FROM breaking_change_records
WHERE repo_id = $1;
