package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
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

	q := sqlcgen.New(s.pool)
	row, err := q.GetROIMetrics(ctx, pgtype.Text{String: orgID, Valid: true})
	if err != nil {
		return metrics, err
	}

	metrics.TotalPreventedOutages = int(row.TotalPreventedOutages)
	metrics.TotalUndocumentedEndpoints = int(row.TotalUndocumentedEndpoints)
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
	q := sqlcgen.New(s.pool)
	return q.UpsertEndpointTraffic(ctx, sqlcgen.UpsertEndpointTrafficParams{
		RepoID:     pgtype.UUID{Bytes: repoID, Valid: true},
		Method:     method,
		Path:       path,
		LastSeenAt: pgtype.Timestamptz{Time: timestamp, Valid: true},
	})
}

func (s *PGStore) GetZeroTrafficEndpoints(ctx context.Context, orgName string, since time.Time) ([]EndpointTraffic, error) {
	q := sqlcgen.New(s.pool)
	rows, err := q.GetZeroTrafficEndpoints(ctx, sqlcgen.GetZeroTrafficEndpointsParams{
		GithubOrgName: orgName,
		LastSeenAt:    pgtype.Timestamptz{Time: since, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	var results []EndpointTraffic
	for _, r := range rows {
		var et EndpointTraffic
		if r.ID.Valid {
			et.ID = r.ID.Bytes
		}
		if r.RepoID.Valid {
			et.RepoID = r.RepoID.Bytes
		}
		et.Method = r.Method
		et.Path = r.Path
		if r.LastSeenAt.Valid {
			et.LastSeenAt = r.LastSeenAt.Time
		}
		if r.RequestCount.Valid {
			et.RequestCount = int64(r.RequestCount.Int32)
		}
		results = append(results, et)
	}
	return results, nil
}

func (s *PGStore) CreatePartner(ctx context.Context, arg sqlcgen.CreatePartnerParams) (sqlcgen.PartnerIntegration, error) {
	q := sqlcgen.New(s.pool)
	return q.CreatePartner(ctx, arg)
}

func (s *PGStore) ListPartners(ctx context.Context) ([]sqlcgen.PartnerIntegration, error) {
	q := sqlcgen.New(s.pool)
	return q.ListPartners(ctx)
}

func (s *PGStore) GetPartner(ctx context.Context, id uuid.UUID) (sqlcgen.PartnerIntegration, error) {
	pgUUID := pgtype.UUID{Bytes: id, Valid: true}
	q := sqlcgen.New(s.pool)
	return q.GetPartner(ctx, pgUUID)
}

func (s *PGStore) UpdatePartner(ctx context.Context, arg sqlcgen.UpdatePartnerParams) (sqlcgen.PartnerIntegration, error) {
	q := sqlcgen.New(s.pool)
	return q.UpdatePartner(ctx, arg)
}

func (s *PGStore) UpdatePartnerStatus(ctx context.Context, arg sqlcgen.UpdatePartnerStatusParams) (sqlcgen.PartnerIntegration, error) {
	q := sqlcgen.New(s.pool)
	return q.UpdatePartnerStatus(ctx, arg)
}

func (s *PGStore) DeletePartner(ctx context.Context, id uuid.UUID) error {
	pgUUID := pgtype.UUID{Bytes: id, Valid: true}
	q := sqlcgen.New(s.pool)
	_, err := q.DeletePartner(ctx, pgUUID)
	return err
}

// ─────────────────────────────────────────────────────────────────────────────
// API Key Methods
// ─────────────────────────────────────────────────────────────────────────────

func (s *PGStore) CreateAPIKey(ctx context.Context, orgID uuid.UUID, name, prefix, hash string) (*APIKey, error) {
	row, err := sqlcgen.New(s.pool).CreateAPIKey(ctx, sqlcgen.CreateAPIKeyParams{
		OrgID:  pgtype.UUID{Bytes: orgID, Valid: true},
		Name:   name,
		Prefix: prefix,
		Hash:   hash,
	})
	if err != nil {
		return nil, err
	}

	return &APIKey{
		ID:        row.ID.Bytes,
		OrgID:     row.OrgID.Bytes,
		Name:      row.Name,
		Prefix:    row.Prefix,
		Hash:      row.Hash,
		CreatedAt: row.CreatedAt.Time,
		LastUsedAt: func() *time.Time {
			if row.LastUsedAt.Valid {
				t := row.LastUsedAt.Time
				return &t
			}
			return nil
		}(),
	}, nil
}

func (s *PGStore) ListAPIKeys(ctx context.Context, orgID uuid.UUID) ([]*APIKey, error) {
	rows, err := sqlcgen.New(s.pool).ListAPIKeys(ctx, pgtype.UUID{Bytes: orgID, Valid: true})
	if err != nil {
		return nil, err
	}

	keys := make([]*APIKey, len(rows))
	for i, row := range rows {
		keys[i] = &APIKey{
			ID:        row.ID.Bytes,
			OrgID:     row.OrgID.Bytes,
			Name:      row.Name,
			Prefix:    row.Prefix,
			Hash:      row.Hash,
			CreatedAt: row.CreatedAt.Time,
			LastUsedAt: func() *time.Time {
				if row.LastUsedAt.Valid {
					t := row.LastUsedAt.Time
					return &t
				}
				return nil
			}(),
		}
	}
	return keys, nil
}

func (s *PGStore) DeleteAPIKey(ctx context.Context, id, orgID uuid.UUID) error {
	return sqlcgen.New(s.pool).DeleteAPIKey(ctx, sqlcgen.DeleteAPIKeyParams{
		ID:    pgtype.UUID{Bytes: id, Valid: true},
		OrgID: pgtype.UUID{Bytes: orgID, Valid: true},
	})
}

func (s *PGStore) InsertSchemaValidationGap(ctx context.Context, arg sqlcgen.InsertSchemaValidationGapParams) error {
	return sqlcgen.New(s.pool).InsertSchemaValidationGap(ctx, arg)
}

func (s *PGStore) GetSchemaValidationGaps(ctx context.Context) ([]sqlcgen.SchemaValidationGap, error) {
	return sqlcgen.New(s.pool).GetSchemaValidationGaps(ctx)
}

func (s *PGStore) GetSchema(ctx context.Context, org, repo string) (string, error) {
	contracts, err := s.GetContractsByProviderFullName(ctx, org+"/"+repo)
	if err != nil {
		return "", err
	}
	if len(contracts) == 0 {
		return "", ErrNotFound
	}
	return contracts[0].RawContent, nil
}

func (s *PGStore) GetPrunedSchemaCache(ctx context.Context, arg sqlcgen.GetPrunedSchemaCacheParams) (sqlcgen.SchemaPruneCache, error) {
	return sqlcgen.New(s.pool).GetPrunedSchemaCache(ctx, arg)
}

func (s *PGStore) UpsertPrunedSchemaCache(ctx context.Context, arg sqlcgen.UpsertPrunedSchemaCacheParams) error {
	return sqlcgen.New(s.pool).UpsertPrunedSchemaCache(ctx, arg)
}

func (s *PGStore) InsertEcosystemEvent(ctx context.Context, arg sqlcgen.InsertEcosystemEventParams) (sqlcgen.EcosystemEvent, error) {
	q := sqlcgen.New(s.pool)
	return q.InsertEcosystemEvent(ctx, arg)
}

func (s *PGStore) GetEcosystemEventsByOrg(ctx context.Context, arg sqlcgen.GetEcosystemEventsByOrgParams) ([]sqlcgen.EcosystemEvent, error) {
	q := sqlcgen.New(s.pool)
	return q.GetEcosystemEventsByOrg(ctx, arg)
}
