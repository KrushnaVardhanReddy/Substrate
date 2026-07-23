-- name: CreatePartner :one
INSERT INTO partner_integrations (vendor_name, webhook_url, webhook_secret)
VALUES ($1, $2, $3)
RETURNING id, vendor_name, status, webhook_url, webhook_secret, created_at, updated_at;

-- name: ListPartners :many
SELECT id, vendor_name, status, webhook_url, webhook_secret, created_at, updated_at
FROM partner_integrations
ORDER BY created_at DESC;

-- name: GetPartner :one
SELECT id, vendor_name, status, webhook_url, webhook_secret, created_at, updated_at
FROM partner_integrations
WHERE id = $1;

-- name: UpdatePartner :one
UPDATE partner_integrations
SET
    vendor_name = COALESCE(NULLIF(sqlc.arg(vendor_name)::text, ''), vendor_name),
    webhook_url = COALESCE(NULLIF(sqlc.arg(webhook_url)::text, ''), webhook_url),
    webhook_secret = COALESCE(NULLIF(sqlc.arg(webhook_secret)::text, ''), webhook_secret),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING id, vendor_name, status, webhook_url, webhook_secret, created_at, updated_at;

-- name: UpdatePartnerStatus :one
UPDATE partner_integrations
SET status = $1, updated_at = NOW()
WHERE id = $2
RETURNING id, vendor_name, status, webhook_url, webhook_secret, created_at, updated_at;

-- name: DeletePartner :one
DELETE FROM partner_integrations
WHERE id = $1
RETURNING id;
