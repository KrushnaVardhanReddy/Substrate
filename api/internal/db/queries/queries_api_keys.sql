-- name: CreateAPIKey :one
INSERT INTO api_keys (
    org_id,
    name,
    prefix,
    hash
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: ListAPIKeys :many
SELECT * FROM api_keys
WHERE org_id = $1
ORDER BY created_at DESC;

-- name: DeleteAPIKey :exec
DELETE FROM api_keys
WHERE id = $1 AND org_id = $2;
