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

func (s *PGStore) GetROIMetrics(ctx context.Context, orgID string) (ROIMetrics, error) {
	var metrics ROIMetrics

	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM webhook_events
		WHERE (is_audit_mode = true OR status = 'blocked')
		  AND org_name = $1
		  AND timestamp > NOW() - INTERVAL '30 days'
	`, orgID).Scan(&metrics.TotalPreventedOutages)
	if err != nil {
		return metrics, err
	}

	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM drift_anomalies
		WHERE org_name = $1
		  AND timestamp > NOW() - INTERVAL '30 days'
	`, orgID).Scan(&metrics.TotalUndocumentedEndpoints)
	if err != nil {
		return metrics, err
	}

	metrics.HoursSaved = metrics.TotalPreventedOutages * 4
	metrics.EstimatedDollarValueSaved = metrics.HoursSaved * 100

	return metrics, nil
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
