// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package db

import (
	"time"

	"github.com/google/uuid"
)

type GovernanceRule struct {
	ID        uuid.UUID `json:"id"`
	OrgID     uuid.UUID `json:"org_id"`
	RuleText  string    `json:"rule_text"`
	CreatedAt time.Time `json:"created_at"`
}
