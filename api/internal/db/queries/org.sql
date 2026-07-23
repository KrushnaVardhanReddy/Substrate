-- name: UpsertOrg :one
-- Finds or creates an organization by github installation ID.
INSERT INTO organizations (github_installation_id, github_org_name, trial_ends_at)
VALUES ($1, $2, NOW() + INTERVAL '90 days')
ON CONFLICT (github_installation_id) DO UPDATE
  SET github_org_name = EXCLUDED.github_org_name
RETURNING id;


-- name: CountReposByOrg :one
SELECT COUNT(*) FROM repositories r
JOIN organizations o ON r.org_id = o.id
WHERE o.github_org_name = $1;

-- name: UpdateStripeCustomerID :exec
UPDATE organizations
SET stripe_customer_id = $1
WHERE id = $2;

-- name: GetBillingStatus :one
SELECT trial_ends_at, stripe_customer_id
FROM organizations
WHERE github_org_name = $1;

-- name: UpsertOrgKMSConfig :exec
INSERT INTO org_kms_config (org_name, provider, key_arn)
VALUES ($1, $2, $3)
ON CONFLICT (org_name) DO UPDATE
  SET provider = EXCLUDED.provider,
      key_arn  = EXCLUDED.key_arn;

-- name: GetOrgKMSConfig :one
SELECT provider, key_arn
FROM org_kms_config
WHERE org_name = $1;
