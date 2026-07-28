package db

import (
	"context"
	"encoding/json"
	"time"
)

type AgentProfile struct {
	ID           int       `json:"id"`
	Org          string    `json:"org"`
	Name         string    `json:"name"`
	AllowedTools []string  `json:"allowed_tools"`
	HITLEnabled  bool      `json:"hitl_enabled"`
	CreatedAt    time.Time `json:"created_at"`
}

type HITLQueueItem struct {
	ID         int             `json:"id"`
	Org        string          `json:"org"`
	ProfileID  int             `json:"profile_id"`
	ToolName   string          `json:"tool_name"`
	Arguments  json.RawMessage `json:"arguments"`
	Status     string          `json:"status"`
	CreatedAt  time.Time       `json:"created_at"`
	ResolvedAt *time.Time      `json:"resolved_at"`
}

func (s *PGStore) CreateAgentProfile(ctx context.Context, profile AgentProfile) (AgentProfile, error) {
	var created AgentProfile
	err := s.pool.QueryRow(ctx, `
		INSERT INTO agent_profiles (org, name, allowed_tools, hitl_enabled)
		VALUES ($1, $2, $3, $4)
		RETURNING id, org, name, allowed_tools, hitl_enabled, created_at
	`, profile.Org, profile.Name, profile.AllowedTools, profile.HITLEnabled).Scan(
		&created.ID, &created.Org, &created.Name, &created.AllowedTools, &created.HITLEnabled, &created.CreatedAt,
	)
	if err != nil {
		return AgentProfile{}, err
	}
	return created, nil
}

func (s *PGStore) GetAgentProfile(ctx context.Context, id int) (AgentProfile, error) {
	var p AgentProfile
	err := s.pool.QueryRow(ctx, `
		SELECT id, org, name, allowed_tools, hitl_enabled, created_at
		FROM agent_profiles
		WHERE id = $1
	`, id).Scan(&p.ID, &p.Org, &p.Name, &p.AllowedTools, &p.HITLEnabled, &p.CreatedAt)
	if err != nil {
		return AgentProfile{}, err
	}
	return p, nil
}

func (s *PGStore) ListAgentProfiles(ctx context.Context, org string) ([]AgentProfile, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, org, name, allowed_tools, hitl_enabled, created_at
		FROM agent_profiles
		WHERE org = $1
		ORDER BY created_at DESC
	`, org)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []AgentProfile
	for rows.Next() {
		var p AgentProfile
		if err := rows.Scan(&p.ID, &p.Org, &p.Name, &p.AllowedTools, &p.HITLEnabled, &p.CreatedAt); err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}

func (s *PGStore) DeleteAgentProfile(ctx context.Context, id int) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM agent_profiles WHERE id = $1`, id)
	return err
}

func (s *PGStore) CreateHITLQueueItem(ctx context.Context, item HITLQueueItem) (HITLQueueItem, error) {
	var created HITLQueueItem
	err := s.pool.QueryRow(ctx, `
		INSERT INTO hitl_queue (org, profile_id, tool_name, arguments, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, org, profile_id, tool_name, arguments, status, created_at, resolved_at
	`, item.Org, item.ProfileID, item.ToolName, item.Arguments, item.Status).Scan(
		&created.ID, &created.Org, &created.ProfileID, &created.ToolName, &created.Arguments, &created.Status, &created.CreatedAt, &created.ResolvedAt,
	)
	if err != nil {
		return HITLQueueItem{}, err
	}
	return created, nil
}

func (s *PGStore) ListHITLQueue(ctx context.Context, org string) ([]HITLQueueItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, org, profile_id, tool_name, arguments, status, created_at, resolved_at
		FROM hitl_queue
		WHERE org = $1
		ORDER BY created_at DESC
	`, org)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []HITLQueueItem
	for rows.Next() {
		var i HITLQueueItem
		if err := rows.Scan(&i.ID, &i.Org, &i.ProfileID, &i.ToolName, &i.Arguments, &i.Status, &i.CreatedAt, &i.ResolvedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, nil
}

func (s *PGStore) ResolveHITLQueueItem(ctx context.Context, id int, status string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE hitl_queue
		SET status = $1, resolved_at = NOW()
		WHERE id = $2
	`, status, id)
	return err
}

func (s *PGStore) GetHITLQueueItem(ctx context.Context, id int) (HITLQueueItem, error) {
	var i HITLQueueItem
	err := s.pool.QueryRow(ctx, `
		SELECT id, org, profile_id, tool_name, arguments, status, created_at, resolved_at
		FROM hitl_queue
		WHERE id = $1
	`, id).Scan(&i.ID, &i.Org, &i.ProfileID, &i.ToolName, &i.Arguments, &i.Status, &i.CreatedAt, &i.ResolvedAt)
	if err != nil {
		return HITLQueueItem{}, err
	}
	return i, nil
}
