-- name: InsertSandboxAuditLog :exec
INSERT INTO sandbox_audit_log (org, repo, method, path, status_code, user_id)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetRepoBaseURL :one
SELECT base_url FROM repositories WHERE org_id = $1 AND name = $2 LIMIT 1;
