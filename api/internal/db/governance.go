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
