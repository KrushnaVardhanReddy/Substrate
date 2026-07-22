-- name: UpsertInsurancePolicy :one
INSERT INTO insurance_policies (org_id, policy_limit_cents)
VALUES ($1, $2)
ON CONFLICT (org_id) DO UPDATE
  SET policy_limit_cents = $2, updated_at = NOW()
RETURNING id;

-- name: GetInsurancePolicy :one
SELECT id, org_id, policy_limit_cents, created_at, updated_at
FROM insurance_policies
WHERE org_id = $1;

-- name: CreateInsuranceClaim :one
INSERT INTO insurance_claims (id, org_id, policy_id, github_pr_url, incident_date, status, amount_cents)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;

-- name: GetInsuranceClaims :many
SELECT id, org_id, policy_id, github_pr_url, incident_date, status, amount_cents, created_at, updated_at
FROM insurance_claims
WHERE org_id = $1
ORDER BY created_at DESC;
