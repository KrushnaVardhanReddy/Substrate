package db

import (
	"context"
	"github.com/google/uuid"
	"time"
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

func (s *PGStore) UpsertInsurancePolicy(ctx context.Context, orgID uuid.UUID, policyLimitCents int64) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, `
		INSERT INTO insurance_policies (org_id, policy_limit_cents)
		VALUES ($1, $2)
		ON CONFLICT (org_id) DO UPDATE SET policy_limit_cents = $2, updated_at = NOW()
		RETURNING id
	`, orgID, policyLimitCents).Scan(&id)
	return id, err
}

func (s *PGStore) GetInsurancePolicy(ctx context.Context, orgID uuid.UUID) (*InsurancePolicy, error) {
	var p InsurancePolicy
	err := s.pool.QueryRow(ctx, `
		SELECT id, org_id, policy_limit_cents, created_at, updated_at
		FROM insurance_policies
		WHERE org_id = $1
	`, orgID).Scan(&p.ID, &p.OrgID, &p.PolicyLimitCents, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *PGStore) CreateInsuranceClaim(ctx context.Context, claim InsuranceClaim) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, `
		INSERT INTO insurance_claims (id, org_id, policy_id, github_pr_url, incident_date, status, amount_cents)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, claim.ID, claim.OrgID, claim.PolicyID, claim.GithubPRUrl, claim.IncidentDate, claim.Status, claim.AmountCents).Scan(&id)
	return id, err
}

func (s *PGStore) GetInsuranceClaims(ctx context.Context, orgID uuid.UUID) ([]InsuranceClaim, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, org_id, policy_id, github_pr_url, incident_date, status, amount_cents, created_at, updated_at
		FROM insurance_claims
		WHERE org_id = $1
		ORDER BY created_at DESC
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var claims []InsuranceClaim
	for rows.Next() {
		var c InsuranceClaim
		if err := rows.Scan(&c.ID, &c.OrgID, &c.PolicyID, &c.GithubPRUrl, &c.IncidentDate, &c.Status, &c.AmountCents, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		claims = append(claims, c)
	}
	return claims, nil
}

func (s *PGStore) UpsertEndpointTraffic(ctx context.Context, repoID uuid.UUID, method, path string, timestamp time.Time) error {
	query := `
		INSERT INTO endpoint_traffic (repo_id, method, path, last_seen_at, request_count)
		VALUES ($1, $2, $3, $4, 1)
		ON CONFLICT (repo_id, method, path)
		DO UPDATE SET
			last_seen_at = GREATEST(endpoint_traffic.last_seen_at, EXCLUDED.last_seen_at),
			request_count = endpoint_traffic.request_count + 1
	`
	_, err := s.pool.Exec(ctx, query, repoID, method, path, timestamp)
	return err
}

func (s *PGStore) GetZeroTrafficEndpoints(ctx context.Context, orgName string, since time.Time) ([]EndpointTraffic, error) {
	query := `
		SELECT et.id, et.repo_id, et.method, et.path, et.last_seen_at, et.request_count
		FROM endpoint_traffic et
		JOIN repositories r ON et.repo_id = r.id
		JOIN organizations o ON r.org_id = o.id
		WHERE o.name = $1 AND (et.last_seen_at IS NULL OR et.last_seen_at < $2)
	`
	rows, err := s.pool.Query(ctx, query, orgName, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []EndpointTraffic
	for rows.Next() {
		var et EndpointTraffic
		if err := rows.Scan(&et.ID, &et.RepoID, &et.Method, &et.Path, &et.LastSeenAt, &et.RequestCount); err != nil {
			return nil, err
		}
		results = append(results, et)
	}
	return results, nil
}
