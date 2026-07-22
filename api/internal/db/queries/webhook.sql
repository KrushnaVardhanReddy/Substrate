-- name: RegisterWebhook :exec
INSERT INTO org_webhooks (org, url, secret)
VALUES ($1, $2, $3);

-- name: GetWebhooks :many
SELECT id::text, org, url, secret, created_at
FROM org_webhooks
WHERE org = $1;
