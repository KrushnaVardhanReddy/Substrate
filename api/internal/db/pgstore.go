package db

import (
	"context"
)

func (s *PGStore) RegisterWebhook(ctx context.Context, config WebhookConfig) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO org_webhooks (org, url, secret)
		VALUES ($1, $2, $3)
	`, config.Org, config.URL, config.Secret)
	return err
}

func (s *PGStore) GetWebhooks(ctx context.Context, org string) ([]WebhookConfig, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, org, url, secret, created_at
		FROM org_webhooks
		WHERE org = $1
	`, org)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []WebhookConfig
	for rows.Next() {
		var w WebhookConfig
		if err := rows.Scan(&w.ID, &w.Org, &w.URL, &w.Secret, &w.CreatedAt); err != nil {
			return nil, err
		}
		webhooks = append(webhooks, w)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return webhooks, nil
}
