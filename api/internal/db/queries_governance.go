// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package db

import (
	"context"

	"github.com/google/uuid"
)

func (s *PGStore) GetOrgIDByName(ctx context.Context, orgName string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, "SELECT id FROM organizations WHERE github_org_name = $1", orgName).Scan(&id)
	return id, err
}

func (s *PGStore) UpsertGovernanceRule(ctx context.Context, orgID uuid.UUID, ruleText string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, `
		INSERT INTO governance_rules (org_id, rule_text)
		VALUES ($1, $2)
		RETURNING id
	`, orgID, ruleText).Scan(&id)
	return id, err
}

func (s *PGStore) GetGovernanceRulesByOrg(ctx context.Context, orgID uuid.UUID) ([]GovernanceRule, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, org_id, rule_text, created_at
		FROM governance_rules
		WHERE org_id = $1
		ORDER BY created_at DESC
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []GovernanceRule
	for rows.Next() {
		var r GovernanceRule
		if err := rows.Scan(&r.ID, &r.OrgID, &r.RuleText, &r.CreatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, nil
}

func (s *PGStore) DeleteGovernanceRule(ctx context.Context, ruleID uuid.UUID, orgID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM governance_rules WHERE id = $1 AND org_id = $2", ruleID, orgID)
	return err
}
