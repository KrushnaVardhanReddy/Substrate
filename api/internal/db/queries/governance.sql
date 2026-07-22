-- name: GetOrgIDByName :one
SELECT id FROM organizations WHERE github_org_name = $1;

-- name: UpsertGovernanceRule :one
INSERT INTO governance_rules (org_id, rule_text)
VALUES ($1, $2)
RETURNING id;

-- name: GetGovernanceRulesByOrg :many
SELECT id, org_id, rule_text, created_at
FROM governance_rules
WHERE org_id = $1
ORDER BY created_at DESC;

-- name: DeleteGovernanceRule :exec
DELETE FROM governance_rules
WHERE id = $1 AND org_id = $2;
