package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PGStore implements the Store interface
type PGStore struct {
	pool *pgxpool.Pool
}

// NewPGStore creates a new PostgreSQL store implementation
func NewPGStore(pool *pgxpool.Pool) *PGStore {
	return &PGStore{pool: pool}
}

// UpsertOrg finds or creates an organization by github_installation_id.
func (s *PGStore) UpsertOrg(ctx context.Context, installationID int64, orgName string) (uuid.UUID, error) {
	return UpsertOrg(ctx, s.pool, installationID, orgName)
}

func UpsertOrg(ctx context.Context, pool *pgxpool.Pool, installationID int64, orgName string) (uuid.UUID, error) {
	var id uuid.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO organizations (github_installation_id, github_org_name)
		VALUES ($1, $2)
		ON CONFLICT (github_installation_id) DO UPDATE
		SET github_org_name = EXCLUDED.github_org_name
		RETURNING id
	`, installationID, orgName).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to upsert org: %w", err)
	}
	return id, nil
}

// UpsertRepo finds or creates a repository by github_repo_id.
func (s *PGStore) UpsertRepo(ctx context.Context, orgID uuid.UUID, githubRepoID int64, name, fullName string) (uuid.UUID, error) {
	return UpsertRepo(ctx, s.pool, orgID, githubRepoID, name, fullName)
}

func UpsertRepo(ctx context.Context, pool *pgxpool.Pool, orgID uuid.UUID, githubRepoID int64, name, fullName string) (uuid.UUID, error) {
	var id uuid.UUID
	err := pool.QueryRow(ctx, `
		INSERT INTO repositories (org_id, github_repo_id, name, full_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (github_repo_id) DO UPDATE
		SET name = EXCLUDED.name, full_name = EXCLUDED.full_name, org_id = EXCLUDED.org_id
		RETURNING id
	`, orgID, githubRepoID, name, fullName).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to upsert repo: %w", err)
	}
	return id, nil
}

// UpsertContract stores (or updates) a schema snapshot for a repo/path/branch.
func (s *PGStore) UpsertContract(ctx context.Context, repoID uuid.UUID, schemaType, specPath, branch, commitSHA, rawContent string) (uuid.UUID, error) {
	return UpsertContract(ctx, s.pool, repoID, schemaType, specPath, branch, commitSHA, rawContent)
}

func UpsertContract(ctx context.Context, pool *pgxpool.Pool, repoID uuid.UUID, schemaType, specPath, branch, commitSHA, rawContent string) (uuid.UUID, error) {
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
func (s *PGStore) UpsertDependency(ctx context.Context, consumerRepoID, providerContractID uuid.UUID) error {
	return UpsertDependency(ctx, s.pool, consumerRepoID, providerContractID)
}

func UpsertDependency(ctx context.Context, pool *pgxpool.Pool, consumerRepoID, providerContractID uuid.UUID) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO dependencies (consumer_repo_id, provider_contract_id, last_checked_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (consumer_repo_id, provider_contract_id) DO UPDATE
		SET last_checked_at = EXCLUDED.last_checked_at
	`, consumerRepoID, providerContractID)
	if err != nil {
		return fmt.Errorf("failed to upsert dependency: %w", err)
	}
	return nil
}

// GetContractsByProviderFullName returns all contracts stored for a provider repo.
func (s *PGStore) GetContractsByProviderFullName(ctx context.Context, providerFullName string) ([]Contract, error) {
	return GetContractsByProviderFullName(ctx, s.pool, providerFullName)
}

func GetContractsByProviderFullName(ctx context.Context, pool *pgxpool.Pool, providerFullName string) ([]Contract, error) {
	rows, err := pool.Query(ctx, `
		SELECT c.id, c.repo_id, c.schema_type, c.spec_path, c.branch, c.latest_commit_sha, c.raw_content, c.synced_at
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
		if err := rows.Scan(&c.ID, &c.RepoID, &c.SchemaType, &c.SpecPath, &c.Branch, &commitSHA, &c.RawContent, &c.SyncedAt); err != nil {
			return nil, fmt.Errorf("failed to scan contract: %w", err)
		}
		if commitSHA != nil {
			c.LatestCommitSHA = *commitSHA
		}
		contracts = append(contracts, c)
	}
	return contracts, rows.Err()
}

// GetConsumersByProviderContract returns all consumer repos for a given provider contract ID.
func (s *PGStore) GetConsumersByProviderContract(ctx context.Context, providerContractID uuid.UUID) ([]ConsumerDependency, error) {
	return GetConsumersByProviderContract(ctx, s.pool, providerContractID)
}

func GetConsumersByProviderContract(ctx context.Context, pool *pgxpool.Pool, providerContractID uuid.UUID) ([]ConsumerDependency, error) {
	rows, err := pool.Query(ctx, `
		SELECT r.id, r.full_name, c.raw_content
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
		if err := rows.Scan(&c.ConsumerRepoID, &c.ConsumerFullName, &c.ContractRawContent); err != nil {
			return nil, fmt.Errorf("failed to scan consumer: %w", err)
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
		SELECT cr.full_name as consumer_full_name, pr.full_name as provider_full_name, d.status
		FROM dependencies d
		JOIN repositories cr ON d.consumer_repo_id = cr.id
		JOIN contracts pc ON d.provider_contract_id = pc.id
		JOIN repositories pr ON pc.repo_id = pr.id
		JOIN organizations o ON cr.org_id = o.id
		WHERE o.github_org_name = $1
	`, orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to query dependency graph: %w", err)
	}
	defer rows.Close()

	var edges []DependencyEdge
	for rows.Next() {
		var e DependencyEdge
		if err := rows.Scan(&e.ConsumerFullName, &e.ProviderFullName, &e.Status); err != nil {
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
