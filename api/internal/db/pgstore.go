package db

import (
	"context"
	"github.com/google/uuid"
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

func (s *PGStore) GetPublicSchema(ctx context.Context, namespace, name, version string) (*PublicSchema, error) {
	var schema PublicSchema
	err := s.pool.QueryRow(ctx, `
		SELECT s.id, s.namespace_id, n.name as namespace_name, s.name, s.version, s.schema_type, s.schema_content, s.created_at
		FROM public_schemas s
		JOIN public_namespaces n ON s.namespace_id = n.id
		WHERE n.name = $1 AND s.name = $2 AND s.version = $3
	`, namespace, name, version).Scan(
		&schema.ID, &schema.NamespaceID, &schema.NamespaceName, &schema.Name,
		&schema.Version, &schema.SchemaType, &schema.SchemaContent, &schema.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &schema, nil
}

func (s *PGStore) PublishPublicSchema(ctx context.Context, namespace, name, version, schemaType, content string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var namespaceID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO public_namespaces (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, namespace).Scan(&namespaceID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO public_schemas (namespace_id, name, version, schema_type, schema_content)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (namespace_id, name, version) DO UPDATE SET
			schema_type = EXCLUDED.schema_type,
			schema_content = EXCLUDED.schema_content
	`, namespaceID, name, version, schemaType, content)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
