package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

import (
	"github.com/KrushnaVardhanReddy/substrate/api/internal/crypto"
)

// PGStore implements the Store interface
type PGStore struct {
	pool      *pgxpool.Pool
	kmsClient crypto.KMSClient
}

// NewPGStore creates a new PostgreSQL store implementation
func NewPGStore(pool *pgxpool.Pool) *PGStore {
	return &PGStore{pool: pool}
}

// WithKMSClient adds a KMS client to the store for Enterprise BYOK encryption
func (s *PGStore) WithKMSClient(client crypto.KMSClient) *PGStore {
	s.kmsClient = client
	return s
}

// UpsertOrg finds or creates an organization by github_installation_id.
func (s *PGStore) UpsertOrg(ctx context.Context, installationID int64, orgName string) (uuid.UUID, error) {
	return UpsertOrg(ctx, s.pool, installationID, orgName)
}

func UpsertOrg(ctx context.Context, pool *pgxpool.Pool, installationID int64, orgName string) (uuid.UUID, error) {
	var id uuid.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO organizations (github_installation_id, github_org_name, trial_ends_at)
		VALUES ($1, $2, NOW() + INTERVAL '90 days')
		ON CONFLICT (github_installation_id) DO UPDATE
		SET github_org_name = EXCLUDED.github_org_name
		RETURNING id
	`, installationID, orgName).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to upsert org: %w", err)
	}
	return id, nil
}

func (s *PGStore) UpdateStripeCustomerID(ctx context.Context, orgID uuid.UUID, stripeID string) error {
	return UpdateStripeCustomerID(ctx, s.pool, orgID, stripeID)
}

func UpdateStripeCustomerID(ctx context.Context, pool *pgxpool.Pool, orgID uuid.UUID, stripeID string) error {
	_, err := pool.Exec(ctx, `
		UPDATE organizations
		SET stripe_customer_id = $1
		WHERE id = $2
	`, stripeID, orgID)
	return err
}

func (s *PGStore) GetBillingStatus(ctx context.Context, orgName string) (time.Time, *string, error) {
	return GetBillingStatus(ctx, s.pool, orgName)
}

func GetBillingStatus(ctx context.Context, pool *pgxpool.Pool, orgName string) (time.Time, *string, error) {
	var trialEndsAt time.Time
	var stripeCustomerID *string
	err := pool.QueryRow(ctx, `
		SELECT trial_ends_at, stripe_customer_id
		FROM organizations
		WHERE github_org_name = $1
	`, orgName).Scan(&trialEndsAt, &stripeCustomerID)
	return trialEndsAt, stripeCustomerID, err
}

// UpsertRepo finds or creates a repository by github_repo_id.
func (s *PGStore) UpsertRepo(ctx context.Context, orgID uuid.UUID, githubRepoID int64, name, fullName string, metadata json.RawMessage) (uuid.UUID, error) {
	return UpsertRepo(ctx, s.pool, orgID, githubRepoID, name, fullName, metadata)
}

func UpsertRepo(ctx context.Context, pool *pgxpool.Pool, orgID uuid.UUID, githubRepoID int64, name, fullName string, metadata json.RawMessage) (uuid.UUID, error) {
	var id uuid.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO repositories (org_id, github_repo_id, name, full_name, metadata)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (github_repo_id) DO UPDATE
		SET name = EXCLUDED.name, full_name = EXCLUDED.full_name, org_id = EXCLUDED.org_id, metadata = EXCLUDED.metadata
		RETURNING id
	`, orgID, githubRepoID, name, fullName, metadata).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to upsert repo: %w", err)
	}
	return id, nil
}

// UpsertContract stores (or updates) a schema snapshot for a repo/path/branch.
func (s *PGStore) UpsertContract(ctx context.Context, repoID uuid.UUID, schemaType, specPath, branch, commitSHA, rawContent string) (uuid.UUID, error) {
	var id uuid.UUID

	isEncrypted := false
	var kmsKeyARN string
	var encryptedContent []byte
	var finalRawContent *string

	if s.kmsClient != nil {
		ciphertext, arn, err := s.kmsClient.Encrypt([]byte(rawContent))
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to encrypt contract content: %w", err)
		}
		isEncrypted = true
		kmsKeyARN = arn
		encryptedContent = ciphertext
		finalRawContent = nil
	} else {
		finalRawContent = &rawContent
	}

	err := s.pool.QueryRow(ctx, `
		INSERT INTO contracts (repo_id, schema_type, spec_path, branch, latest_commit_sha, raw_content, encrypted_content, is_encrypted, kms_key_arn, synced_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (repo_id, spec_path, branch) DO UPDATE
		SET schema_type = EXCLUDED.schema_type,
		    latest_commit_sha = EXCLUDED.latest_commit_sha,
		    raw_content = EXCLUDED.raw_content,
		    encrypted_content = EXCLUDED.encrypted_content,
		    is_encrypted = EXCLUDED.is_encrypted,
		    kms_key_arn = EXCLUDED.kms_key_arn,
		    synced_at = EXCLUDED.synced_at
		RETURNING id
	`, repoID, schemaType, specPath, branch, commitSHA, finalRawContent, encryptedContent, isEncrypted, kmsKeyARN).Scan(&id)

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to upsert contract: %w", err)
	}
	return id, nil
}

func UpsertContract(ctx context.Context, pool *pgxpool.Pool, repoID uuid.UUID, schemaType, specPath, branch, commitSHA, rawContent string) (uuid.UUID, error) {
	// This function is kept for backwards compatibility but is essentially what PGStore used to call.
	// Since encryption requires the kmsClient, package-level UpsertContract won't encrypt.
	var id uuid.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO contracts (repo_id, schema_type, spec_path, branch, latest_commit_sha, raw_content, synced_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (repo_id, spec_path, branch) DO UPDATE
		SET schema_type = EXCLUDED.schema_type,
		    latest_commit_sha = EXCLUDED.latest_commit_sha,
		    raw_content = EXCLUDED.raw_content,
		    synced_at = EXCLUDED.synced_at
		RETURNING id
	`, repoID, schemaType, specPath, branch, commitSHA, rawContent).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to upsert contract: %w", err)
	}
	return id, nil
}

// UpsertDependency links a consumer repo to a provider contract.
func (s *PGStore) UpsertDependency(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, confidenceScore int, requiredNoticeDays int) error {
	return UpsertDependency(ctx, s.pool, consumerRepoID, providerContractID, confidenceScore, requiredNoticeDays)
}

func UpsertDependency(ctx context.Context, pool *pgxpool.Pool, consumerRepoID, providerContractID uuid.UUID, confidenceScore int, requiredNoticeDays int) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO dependencies (consumer_repo_id, provider_contract_id, last_checked_at, confidence_score, required_notice_days)
		VALUES ($1, $2, NOW(), $3, $4)
		ON CONFLICT (consumer_repo_id, provider_contract_id) DO UPDATE
		SET last_checked_at = EXCLUDED.last_checked_at,
		    confidence_score = EXCLUDED.confidence_score,
		    required_notice_days = CASE
		        WHEN EXCLUDED.required_notice_days = 0 THEN dependencies.required_notice_days
		        ELSE EXCLUDED.required_notice_days
		    END
	`, consumerRepoID, providerContractID, confidenceScore, requiredNoticeDays)
	if err != nil {
		return fmt.Errorf("failed to upsert dependency: %w", err)
	}
	return nil
}

// GetContractsByProviderFullName returns all contracts stored for a provider repo.
func (s *PGStore) GetContractsByProviderFullName(ctx context.Context, providerFullName string) ([]Contract, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.id, c.repo_id, c.schema_type, c.spec_path, c.branch, c.latest_commit_sha,
		       c.raw_content, c.encrypted_content, c.is_encrypted, c.kms_key_arn, c.synced_at
		FROM contracts c
		JOIN repositories r ON c.repo_id = r.id
		WHERE r.full_name = $1
	`, providerFullName)
	if err != nil {
		return nil, fmt.Errorf("failed to query contracts: %w", err)
	}
	defer rows.Close()

	var contracts []Contract
	for rows.Next() {
		var c Contract
		var commitSHA *string
		var rawContent *string
		var encryptedContent []byte
		var isEncrypted bool
		var kmsKeyARN *string

		if err := rows.Scan(&c.ID, &c.RepoID, &c.SchemaType, &c.SpecPath, &c.Branch, &commitSHA,
			&rawContent, &encryptedContent, &isEncrypted, &kmsKeyARN, &c.SyncedAt); err != nil {
			return nil, fmt.Errorf("failed to scan contract: %w", err)
		}

		if commitSHA != nil {
			c.LatestCommitSHA = *commitSHA
		}

		c.IsEncrypted = isEncrypted
		if kmsKeyARN != nil {
			c.KMSKeyARN = *kmsKeyARN
		}

		if isEncrypted {
			if s.kmsClient == nil {
				return nil, fmt.Errorf("contract is encrypted but KMS client is not configured")
			}
			decrypted, err := s.kmsClient.Decrypt(encryptedContent, c.KMSKeyARN)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt contract content: %w", err)
			}
			c.RawContent = string(decrypted)
		} else if rawContent != nil {
			c.RawContent = *rawContent
		}

		contracts = append(contracts, c)
	}
	return contracts, rows.Err()
}

func GetContractsByProviderFullName(ctx context.Context, pool *pgxpool.Pool, providerFullName string) ([]Contract, error) {
	// This function is kept for backwards compatibility but does not decrypt.
	// Users of PGStore should call s.GetContractsByProviderFullName directly to ensure decryption works.
	rows, err := pool.Query(ctx, `
		SELECT c.id, c.repo_id, c.schema_type, c.spec_path, c.branch, c.latest_commit_sha, c.raw_content, c.is_encrypted, c.synced_at
		FROM contracts c
		JOIN repositories r ON c.repo_id = r.id
		WHERE r.full_name = $1
	`, providerFullName)
	if err != nil {
		return nil, fmt.Errorf("failed to query contracts: %w", err)
	}
	defer rows.Close()

	var contracts []Contract
	for rows.Next() {
		var c Contract
		var commitSHA *string
		var rawContent *string
		var isEncrypted bool
		if err := rows.Scan(&c.ID, &c.RepoID, &c.SchemaType, &c.SpecPath, &c.Branch, &commitSHA, &rawContent, &isEncrypted, &c.SyncedAt); err != nil {
			return nil, fmt.Errorf("failed to scan contract: %w", err)
		}
		if commitSHA != nil {
			c.LatestCommitSHA = *commitSHA
		}
		if isEncrypted {
			return nil, fmt.Errorf("contract %s is encrypted; must use PGStore method to decrypt", c.ID)
		}
		if rawContent != nil {
			c.RawContent = *rawContent
		}
		contracts = append(contracts, c)
	}
	return contracts, rows.Err()
}

// GetConsumersByProviderContract returns all consumer repos for a given provider contract ID.
func (s *PGStore) GetConsumersByProviderContract(ctx context.Context, providerContractID uuid.UUID) ([]ConsumerDependency, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT r.id, r.full_name, c.raw_content, c.encrypted_content, c.is_encrypted, c.kms_key_arn, COALESCE(d.required_notice_days, 0)
		FROM dependencies d
		JOIN repositories r ON d.consumer_repo_id = r.id
		JOIN contracts c ON d.provider_contract_id = c.id
		WHERE d.provider_contract_id = $1
	`, providerContractID)
	if err != nil {
		return nil, fmt.Errorf("failed to query consumers: %w", err)
	}
	defer rows.Close()

	var consumers []ConsumerDependency
	for rows.Next() {
		var c ConsumerDependency
		var rawContent *string
		var encryptedContent []byte
		var isEncrypted bool
		var kmsKeyARN *string

		if err := rows.Scan(&c.ConsumerRepoID, &c.ConsumerFullName, &rawContent, &encryptedContent, &isEncrypted, &kmsKeyARN, &c.RequiredNoticeDays); err != nil {
			return nil, fmt.Errorf("failed to scan consumer: %w", err)
		}

		if isEncrypted {
			if s.kmsClient == nil {
				return nil, fmt.Errorf("contract is encrypted but KMS client is not configured")
			}
			arn := ""
			if kmsKeyARN != nil {
				arn = *kmsKeyARN
			}
			decrypted, err := s.kmsClient.Decrypt(encryptedContent, arn)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt contract content: %w", err)
			}
			c.ContractRawContent = string(decrypted)
		} else if rawContent != nil {
			c.ContractRawContent = *rawContent
		}

		consumers = append(consumers, c)
	}
	return consumers, rows.Err()
}

func GetConsumersByProviderContract(ctx context.Context, pool *pgxpool.Pool, providerContractID uuid.UUID) ([]ConsumerDependency, error) {
	// This function is kept for backwards compatibility but does not decrypt.
	// Users of PGStore should call s.GetConsumersByProviderContract directly to ensure decryption works.
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.full_name, c.raw_content, c.is_encrypted, COALESCE(d.required_notice_days, 0)
		FROM dependencies d
		JOIN repositories r ON d.consumer_repo_id = r.id
		JOIN contracts c ON d.provider_contract_id = c.id
		WHERE d.provider_contract_id = $1
	`, providerContractID)
	if err != nil {
		return nil, fmt.Errorf("failed to query consumers: %w", err)
	}
	defer rows.Close()

	var consumers []ConsumerDependency
	for rows.Next() {
		var c ConsumerDependency
		var rawContent *string
		var isEncrypted bool
		if err := rows.Scan(&c.ConsumerRepoID, &c.ConsumerFullName, &rawContent, &isEncrypted, &c.RequiredNoticeDays); err != nil {
			return nil, fmt.Errorf("failed to scan consumer: %w", err)
		}
		if isEncrypted {
			return nil, fmt.Errorf("contract %s is encrypted; must use PGStore method to decrypt", providerContractID)
		}
		if rawContent != nil {
			c.ContractRawContent = *rawContent
		}
		consumers = append(consumers, c)
	}
	return consumers, rows.Err()
}

// ListReposByOrg returns all repos for a given org name.
func (s *PGStore) ListReposByOrg(ctx context.Context, orgName string) ([]Repository, error) {
	return ListReposByOrg(ctx, s.pool, orgName)
}

func ListReposByOrg(ctx context.Context, pool *pgxpool.Pool, orgName string) ([]Repository, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.org_id, r.github_repo_id, r.name, r.full_name, r.created_at
		FROM repositories r
		JOIN organizations o ON r.org_id = o.id
		WHERE o.github_org_name = $1
	`, orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to query repos: %w", err)
	}
	defer rows.Close()

	var repos []Repository
	for rows.Next() {
		var r Repository
		var createdAt *time.Time
		if err := rows.Scan(&r.ID, &r.OrgID, &r.GithubRepoID, &r.Name, &r.FullName, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan repo: %w", err)
		}
		if createdAt != nil {
			r.CreatedAt = *createdAt
		}
		repos = append(repos, r)
	}
	return repos, rows.Err()
}

// GetDependencyGraph returns all dependency edges for an org (for graph visualization).
func (s *PGStore) GetDependencyGraph(ctx context.Context, orgName string) ([]DependencyEdge, error) {
	return GetDependencyGraph(ctx, s.pool, orgName)
}

func GetDependencyGraph(ctx context.Context, pool *pgxpool.Pool, orgName string) ([]DependencyEdge, error) {
	rows, err := pool.Query(ctx, `
		SELECT cr.full_name as consumer_full_name, pr.full_name as provider_full_name, d.status, cr.metadata as consumer_metadata, pr.metadata as provider_metadata, COALESCE(rm.predictive_risk_score, 0) as predictive_risk_score
		FROM dependencies d
		JOIN repositories cr ON d.consumer_repo_id = cr.id
		JOIN contracts pc ON d.provider_contract_id = pc.id
		JOIN repositories pr ON pc.repo_id = pr.id
		JOIN organizations o ON cr.org_id = o.id
		LEFT JOIN repo_metrics rm ON pr.id = rm.repo_id
		WHERE o.github_org_name = $1
	`, orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to query dependency graph: %w", err)
	}
	defer rows.Close()

	var edges []DependencyEdge
	for rows.Next() {
		var e DependencyEdge
		if err := rows.Scan(&e.ConsumerFullName, &e.ProviderFullName, &e.Status, &e.ConsumerMetadata, &e.ProviderMetadata); err != nil {
			return nil, fmt.Errorf("failed to scan dependency edge: %w", err)
		}
		edges = append(edges, e)
	}
	return edges, rows.Err()
}

// CountReposByOrg returns the number of repositories connected to an organization.
func (s *PGStore) CountReposByOrg(ctx context.Context, orgName string) (int, error) {
	return CountReposByOrg(ctx, s.pool, orgName)
}

func CountReposByOrg(ctx context.Context, pool *pgxpool.Pool, orgName string) (int, error) {
	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(r.id)
		FROM repositories r
		JOIN organizations o ON r.org_id = o.id
		WHERE o.github_org_name = $1
	`, orgName).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count repos: %w", err)
	}
	return count, nil
}

// RecordBreakingChange records a new breaking change entry.
func (s *PGStore) RecordBreakingChange(ctx context.Context, repoID uuid.UUID, orgName, repoName, gitSHA string, breakingChanges json.RawMessage) error {
	return RecordBreakingChange(ctx, s.pool, repoID, orgName, repoName, gitSHA, breakingChanges)
}

func RecordBreakingChange(ctx context.Context, pool *pgxpool.Pool, repoID uuid.UUID, orgName, repoName, gitSHA string, breakingChanges json.RawMessage) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO breaking_change_history (repo_id, org_name, repo_name, git_sha, breaking_changes)
		VALUES ($1, $2, $3, $4, $5)
	`, repoID, orgName, repoName, gitSHA, breakingChanges)
	if err != nil {
		return fmt.Errorf("failed to record breaking change: %w", err)
	}
	return nil
}

// GetBreakingChangeHistory retrieves recent breaking changes for a repository.
func (s *PGStore) GetBreakingChangeHistory(ctx context.Context, orgName, repoName string, limit int) ([]BreakingChangeRecord, error) {
	return GetBreakingChangeHistory(ctx, s.pool, orgName, repoName, limit)
}

func GetBreakingChangeHistory(ctx context.Context, pool *pgxpool.Pool, orgName, repoName string, limit int) ([]BreakingChangeRecord, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, repo_id, org_name, repo_name, git_sha, timestamp, breaking_changes
		FROM breaking_change_history
		WHERE org_name = $1 AND repo_name = $2
		ORDER BY timestamp DESC
		LIMIT $3
	`, orgName, repoName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get breaking change history: %w", err)
	}
	defer rows.Close()

	var records []BreakingChangeRecord
	for rows.Next() {
		var r BreakingChangeRecord
		if err := rows.Scan(&r.ID, &r.RepoID, &r.OrgName, &r.RepoName, &r.GitSHA, &r.Timestamp, &r.BreakingChanges); err != nil {
			return nil, fmt.Errorf("failed to scan breaking change record: %w", err)
		}
		records = append(records, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over breaking change history: %w", err)
	}

	if records == nil {
		records = []BreakingChangeRecord{}
	}

	return records, nil
}

// CountDownstreamDependencies returns the number of unique downstream consumers for a provider.
func (s *PGStore) CountDownstreamDependencies(ctx context.Context, providerFullName string) (int, error) {
	return CountDownstreamDependencies(ctx, s.pool, providerFullName)
}

func CountDownstreamDependencies(ctx context.Context, pool *pgxpool.Pool, providerFullName string) (int, error) {
	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT d.consumer_repo_id)
		FROM dependencies d
		JOIN contracts c ON d.provider_contract_id = c.id
		JOIN repositories r ON c.repo_id = r.id
		WHERE r.full_name = $1
	`, providerFullName).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count downstream dependencies: %w", err)
	}
	return count, nil
}

func (s *PGStore) UpdateDependencyConfidence(ctx context.Context, consumerFullName, providerURL string, boostAmount float64) error {
	return UpdateDependencyConfidence(ctx, s.pool, consumerFullName, providerURL, boostAmount)
}

func UpdateDependencyConfidence(ctx context.Context, pool *pgxpool.Pool, consumerFullName, providerURL string, boostAmount float64) error {
	_, err := pool.Exec(ctx, `
		UPDATE dependencies
		SET confidence_score = LEAST(100.0, confidence_score + $1)
		FROM repositories cr, contracts pc
		WHERE dependencies.consumer_repo_id = cr.id
		  AND dependencies.provider_contract_id = pc.id
		  AND cr.full_name = $2
		  AND pc.raw_content LIKE '%' || $3 || '%'
	`, boostAmount, consumerFullName, providerURL)
	if err != nil {
		return fmt.Errorf("failed to update confidence score: %w", err)
	}
	return nil
}

func (s *PGStore) SaveDiffReport(ctx context.Context, diffReport json.RawMessage, isAuditMode bool, orgName, repoName string) (uuid.UUID, error) {
	return SaveDiffReport(ctx, s.pool, diffReport, isAuditMode, orgName, repoName)
}

func SaveDiffReport(ctx context.Context, pool *pgxpool.Pool, diffReport json.RawMessage, isAuditMode bool, orgName, repoName string) (uuid.UUID, error) {
	var id uuid.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO diff_reports (report_data, is_audit_mode, org_name, repo_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, diffReport, isAuditMode, orgName, repoName).Scan(&id)
	return id, err
}

func (s *PGStore) GetDiffReportsByRepo(ctx context.Context, orgName, repoName string, limit int) ([]DiffReportRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT report_data, created_at
		FROM diff_reports
		WHERE org_name = $1 AND repo_name = $2
		ORDER BY created_at DESC
		LIMIT $3
	`, orgName, repoName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get diff reports: %w", err)
	}
	defer rows.Close()

	var reports []DiffReportRecord
	for rows.Next() {
		var record DiffReportRecord
		if err := rows.Scan(&record.ReportData, &record.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan report: %w", err)
		}
		reports = append(reports, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return reports, nil
}

func (s *PGStore) GetDiffReport(ctx context.Context, id uuid.UUID) (json.RawMessage, error) {
	return GetDiffReport(ctx, s.pool, id)
}

func GetDiffReport(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID) (json.RawMessage, error) {
	var reportData json.RawMessage
	err := pool.QueryRow(ctx, `
		SELECT report_data
		FROM diff_reports
		WHERE id = $1
	`, id).Scan(&reportData)
	return reportData, err
}

func (s *PGStore) RecordDriftAnomaly(ctx context.Context, anomaly DriftAnomaly) error {
	query := `
		INSERT INTO drift_anomalies (org_name, repo_name, method, path, error_message)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.pool.Exec(ctx, query, anomaly.OrgName, anomaly.RepoName, anomaly.Method, anomaly.Path, anomaly.ErrorMessage)
	return err
}

func (s *PGStore) GetDriftAnomalies(ctx context.Context, orgName, repoName string) ([]DriftAnomaly, error) {
	query := `
		SELECT id, org_name, repo_name, method, path, error_message, timestamp
		FROM drift_anomalies
		WHERE org_name = $1 AND repo_name = $2
		ORDER BY timestamp DESC
	`
	rows, err := s.pool.Query(ctx, query, orgName, repoName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var anomalies []DriftAnomaly
	for rows.Next() {
		var a DriftAnomaly
		if err := rows.Scan(&a.ID, &a.OrgName, &a.RepoName, &a.Method, &a.Path, &a.ErrorMessage, &a.Timestamp); err != nil {
			return nil, err
		}
		anomalies = append(anomalies, a)
	}
	return anomalies, nil
}

func (s *PGStore) UpdateDependencyStatus(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, status string) error {
	return UpdateDependencyStatus(ctx, s.pool, consumerRepoID, providerContractID, status)
}

func UpdateDependencyStatus(ctx context.Context, pool *pgxpool.Pool, consumerRepoID, providerContractID uuid.UUID, status string) error {
	_, err := pool.Exec(ctx, `
		UPDATE dependencies
		SET status = $3, last_checked_at = NOW()
		WHERE consumer_repo_id = $1 AND provider_contract_id = $2
	`, consumerRepoID, providerContractID, status)
	if err != nil {
		return fmt.Errorf("failed to update dependency status: %w", err)
	}
	return nil
}

func (s *PGStore) GetConsumerManifests(ctx context.Context, providerRepo, consumerRepo string) (json.RawMessage, error) {
	return GetConsumerManifests(ctx, s.pool, providerRepo, consumerRepo)
}

func (s *PGStore) UpsertOrgKMSConfig(ctx context.Context, orgName, provider, keyARN string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO org_kms_config (org_name, provider, key_arn)
		VALUES ($1, $2, $3)
		ON CONFLICT (org_name) DO UPDATE
		SET provider = EXCLUDED.provider, key_arn = EXCLUDED.key_arn`,
		orgName, provider, keyARN,
	)
	return err
}

func (s *PGStore) GetOrgKMSConfig(ctx context.Context, orgName string) (string, string, error) {
	var provider, keyARN string
	err := s.pool.QueryRow(ctx, `
		SELECT provider, key_arn
		FROM org_kms_config
		WHERE org_name = $1`, orgName).Scan(&provider, &keyARN)
	return provider, keyARN, err
}

func GetConsumerManifests(ctx context.Context, pool *pgxpool.Pool, providerRepo, consumerRepo string) (json.RawMessage, error) {
	var consumedFields json.RawMessage
	err := pool.QueryRow(ctx, `
		SELECT consumed_fields
		FROM consumer_manifests
		WHERE provider_repo = $1 AND consumer_repo = $2
		ORDER BY created_at DESC
		LIMIT 1
	`, providerRepo, consumerRepo).Scan(&consumedFields)

	if err != nil {
		return nil, err
	}
	return consumedFields, nil
}

func (s *PGStore) RegisterAgent(ctx context.Context, repoName, owner string, tools []AgentToolDependency) (uuid.UUID, error) {
	return RegisterAgent(ctx, s.pool, repoName, owner, tools)
}

func RegisterAgent(ctx context.Context, pool *pgxpool.Pool, repoName, owner string, tools []AgentToolDependency) (uuid.UUID, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var agentID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO agent_consumers (repo_name, owner)
		VALUES ($1, $2)
		ON CONFLICT (owner, repo_name) DO UPDATE SET repo_name = EXCLUDED.repo_name
		RETURNING id
	`, repoName, owner).Scan(&agentID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to insert agent consumer: %w", err)
	}

	_, err = tx.Exec(ctx, "DELETE FROM agent_tool_dependencies WHERE agent_id = $1", agentID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to delete old tools: %w", err)
	}

	for _, tool := range tools {
		_, err = tx.Exec(ctx, `
			INSERT INTO agent_tool_dependencies (agent_id, tool_name, parameters_jsonb)
			VALUES ($1, $2, $3)
		`, agentID, tool.ToolName, tool.ParametersJSON)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to insert agent tool dependency: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return agentID, nil
}

func (s *PGStore) GetAgentsByTool(ctx context.Context, toolName string) ([]AgentConsumer, error) {
	return GetAgentsByTool(ctx, s.pool, toolName)
}

func GetAgentsByTool(ctx context.Context, pool *pgxpool.Pool, toolName string) ([]AgentConsumer, error) {
	rows, err := pool.Query(ctx, `
		SELECT ac.id, ac.repo_name, ac.owner, ac.created_at
		FROM agent_consumers ac
		JOIN agent_tool_dependencies atd ON ac.id = atd.agent_id
		WHERE atd.tool_name = $1
	`, toolName)
	if err != nil {
		return nil, fmt.Errorf("failed to query agents by tool: %w", err)
	}
	defer rows.Close()

	var agents []AgentConsumer
	for rows.Next() {
		var a AgentConsumer
		if err := rows.Scan(&a.ID, &a.RepoName, &a.Owner, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan agent consumer: %w", err)
		}
		agents = append(agents, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return agents, nil
}

// GetBreakingChangesBetween retrieves breaking changes that occurred between since and until.
func (s *PGStore) GetBreakingChangesBetween(ctx context.Context, since, until time.Time) ([]BreakingChangeRecord, error) {
	return GetBreakingChangesBetween(ctx, s.pool, since, until)
}

func GetBreakingChangesBetween(ctx context.Context, pool *pgxpool.Pool, since, until time.Time) ([]BreakingChangeRecord, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, repo_id, org_name, repo_name, git_sha, timestamp, breaking_changes
		FROM breaking_change_history
		WHERE timestamp >= $1 AND timestamp <= $2
		ORDER BY timestamp DESC
	`, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to get breaking changes between: %w", err)
	}
	defer rows.Close()

	var records []BreakingChangeRecord
	for rows.Next() {
		var r BreakingChangeRecord
		if err := rows.Scan(&r.ID, &r.RepoID, &r.OrgName, &r.RepoName, &r.GitSHA, &r.Timestamp, &r.BreakingChanges); err != nil {
			return nil, fmt.Errorf("failed to scan breaking change record: %w", err)
		}
		records = append(records, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over breaking change history: %w", err)
	}

	if records == nil {
		records = []BreakingChangeRecord{}
	}

	return records, nil
}
