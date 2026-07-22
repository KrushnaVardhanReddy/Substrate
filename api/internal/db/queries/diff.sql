-- name: SaveDiffReport :one
INSERT INTO diff_reports (report_data, is_audit_mode, org_name, repo_name)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: GetDiffReport :one
SELECT report_data FROM diff_reports WHERE id = $1;

-- name: GetDiffReportsByRepo :many
SELECT report_data, created_at
FROM diff_reports
WHERE org_name = $1 AND repo_name = $2
ORDER BY created_at DESC
LIMIT $3;

-- name: CreatePreviewSession :one
INSERT INTO preview_sessions (pr_number, diff_id, expires_at)
VALUES ($1, $2, $3)
RETURNING id;

-- name: GetPreviewSession :one
SELECT d.report_data, ps.expires_at
FROM preview_sessions ps
JOIN diff_reports d ON ps.diff_id = d.id
WHERE ps.id = $1;

-- name: ExpirePreviewSessionsForPR :exec
UPDATE preview_sessions
SET expires_at = NOW() - INTERVAL '1 second'
WHERE pr_number = $1
  AND org_name  = $2
  AND repo_name = $3
  AND expires_at > NOW();
