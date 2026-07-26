-- name: InsertQAShadowTraffic :exec
INSERT INTO qa_shadow_traffic (
    repo_id, method, path, request_payload, response_payload, status_code
) VALUES (
    $1, $2, $3, $4, $5, $6
);

-- name: CreateQAReplayJob :one
INSERT INTO qa_replay_jobs (
    repo_id, timestamp_target, status
) VALUES (
    $1, $2, 'pending'
) RETURNING id;

-- name: UpdateQAReplayJobStatus :exec
UPDATE qa_replay_jobs
SET status = $1, coverage_score = $2, completed_at = NOW()
WHERE id = $3;

-- name: GetQAShadowTrafficByRepo :many
SELECT * FROM qa_shadow_traffic
WHERE repo_id = $1
ORDER BY captured_at DESC
LIMIT 1000;

-- name: GetLatestQAReplayJob :one
SELECT * FROM qa_replay_jobs
WHERE repo_id = $1
ORDER BY created_at DESC
LIMIT 1;
