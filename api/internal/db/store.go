package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Contract struct {
	ID              uuid.UUID
	RepoID          uuid.UUID
	SchemaType      string
	SpecPath        string
	Branch          string
	LatestCommitSHA string
	RawContent      string
	SyncedAt        time.Time
}

type Repository struct {
	ID           uuid.UUID
	OrgID        uuid.UUID
	GithubRepoID int64
	Name         string
	FullName     string
	Metadata     json.RawMessage
	CreatedAt    time.Time
}

type ConsumerDependency struct {
	ConsumerRepoID     uuid.UUID
	ConsumerFullName   string
	ContractRawContent string
}

type DependencyEdge struct {
	ConsumerFullName string          `json:"consumer"`
	ProviderFullName string          `json:"provider"`
	Status           string          `json:"status"`
	ConsumerMetadata json.RawMessage `json:"consumer_metadata,omitempty"`
	ProviderMetadata json.RawMessage `json:"provider_metadata,omitempty"`
}

type BreakingChangeRecord struct {
	ID              uuid.UUID       `json:"id"`
	RepoID          uuid.UUID       `json:"repo_id"`
	OrgName         string          `json:"org_name"`
	RepoName        string          `json:"repo_name"`
	GitSHA          string          `json:"git_sha"`
	Timestamp       time.Time       `json:"timestamp"`
	BreakingChanges json.RawMessage `json:"breaking_changes"`
}

type DriftAnomaly struct {
	ID           uuid.UUID `json:"id"`
	OrgName      string    `json:"org_name"`
	RepoName     string    `json:"repo_name"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	ErrorMessage string    `json:"error_message"`
	Timestamp    time.Time `json:"timestamp"`
}

type WebhookConfig struct {
	ID        string
	Org       string
	URL       string
	Secret    string
	CreatedAt time.Time
}

type ROIMetrics struct {
	TotalPreventedOutages      int `json:"total_prevented_outages"`
	TotalUndocumentedEndpoints int `json:"total_undocumented_endpoints"`
	HoursSaved                 int `json:"hours_saved"`
	EstimatedDollarValueSaved  int `json:"estimated_dollar_value_saved"`
}

type Store interface {
	UpsertOrg(ctx context.Context, installationID int64, orgName string) (uuid.UUID, error)
	UpsertRepo(ctx context.Context, orgID uuid.UUID, githubRepoID int64, name, fullName string, metadata json.RawMessage) (uuid.UUID, error)
	UpsertContract(ctx context.Context, repoID uuid.UUID, schemaType, specPath, branch, commitSHA, rawContent string) (uuid.UUID, error)
	UpsertDependency(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, confidenceScore int) error
	GetContractsByProviderFullName(ctx context.Context, providerFullName string) ([]Contract, error)
	GetConsumersByProviderContract(ctx context.Context, providerContractID uuid.UUID) ([]ConsumerDependency, error)
	ListReposByOrg(ctx context.Context, orgName string) ([]Repository, error)
	GetDependencyGraph(ctx context.Context, orgName string) ([]DependencyEdge, error)
	CountReposByOrg(ctx context.Context, orgName string) (int, error)
	CountDownstreamDependencies(ctx context.Context, providerFullName string) (int, error)
	RecordBreakingChange(ctx context.Context, repoID uuid.UUID, orgName, repoName, gitSHA string, breakingChanges json.RawMessage) error
	GetBreakingChangeHistory(ctx context.Context, orgName, repoName string, limit int) ([]BreakingChangeRecord, error)
	UpdateDependencyConfidence(ctx context.Context, consumerFullName, providerURL string, boostAmount float64) error
	SaveDiffReport(ctx context.Context, diffReport json.RawMessage, isAuditMode bool) (uuid.UUID, error)
	GetDiffReport(ctx context.Context, id uuid.UUID) (json.RawMessage, error)
	RecordDriftAnomaly(ctx context.Context, anomaly DriftAnomaly) error
	GetDriftAnomalies(ctx context.Context, orgName, repoName string) ([]DriftAnomaly, error)
	RegisterWebhook(ctx context.Context, config WebhookConfig) error
	GetWebhooks(ctx context.Context, org string) ([]WebhookConfig, error)
	GetROIMetrics(ctx context.Context, orgID string) (ROIMetrics, error)
	UpdateStripeCustomerID(ctx context.Context, orgID uuid.UUID, stripeID string) error
	GetBillingStatus(ctx context.Context, orgName string) (time.Time, *string, error)
}
