package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Contract struct {
	ID              uuid.UUID `json:"id"`
	RepoID          uuid.UUID `json:"repo_id"`
	SchemaType      string    `json:"schema_type"`
	SpecPath        string    `json:"spec_path"`
	Branch          string    `json:"branch"`
	LatestCommitSHA string    `json:"latest_commit_sha"`
	RawContent      string    `json:"raw_content"`
	SyncedAt        time.Time `json:"synced_at"`
	IsEncrypted     bool      `json:"is_encrypted"`
	KMSKeyARN       string    `json:"kms_key_arn,omitempty"`
}

type Repository struct {
	ID           uuid.UUID       `json:"id"`
	OrgID        uuid.UUID       `json:"org_id"`
	GithubRepoID int64           `json:"github_repo_id"`
	Name         string          `json:"name"`
	FullName     string          `json:"full_name"`
	Metadata     json.RawMessage `json:"metadata"`
	CreatedAt    time.Time       `json:"created_at"`
}

type ConsumerDependency struct {
	ConsumerRepoID     uuid.UUID
	ConsumerFullName   string
	ContractRawContent string
	RequiredNoticeDays int
}

type DiffReportRecord struct {
	ReportData json.RawMessage
	CreatedAt  time.Time
}

type DependencyEdge struct {
	ConsumerFullName    string          `json:"consumer"`
	ProviderFullName    string          `json:"provider"`
	Status              string          `json:"status"`
	ConsumerMetadata    json.RawMessage `json:"consumer_metadata,omitempty"`
	ProviderMetadata    json.RawMessage `json:"provider_metadata,omitempty"`
	PredictiveRiskScore int             `json:"predictive_risk_score,omitempty"`
}

type RepoMetric struct {
	ID                  uuid.UUID `json:"id"`
	RepoID              uuid.UUID `json:"repo_id"`
	PredictiveRiskScore int       `json:"predictive_risk_score"`
	CalculatedAt        time.Time `json:"calculated_at"`
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

type EndpointTraffic struct {
	ID           uuid.UUID `json:"id"`
	RepoID       uuid.UUID `json:"repo_id"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	LastSeenAt   time.Time `json:"last_seen_at"`
	RequestCount int64     `json:"request_count"`
}

type ROIMetrics struct {
	TotalPreventedOutages      int `json:"total_prevented_outages"`
	TotalUndocumentedEndpoints int `json:"total_undocumented_endpoints"`
	HoursSaved                 int `json:"hours_saved"`
	EstimatedDollarValueSaved  int `json:"estimated_dollar_value_saved"`
}

type AgentConsumer struct {
	ID        uuid.UUID `json:"id"`
	RepoName  string    `json:"repo_name"`
	Owner     string    `json:"owner"`
	CreatedAt time.Time `json:"created_at"`
}

type AgentToolDependency struct {
	ID             uuid.UUID       `json:"id"`
	AgentID        uuid.UUID       `json:"agent_id"`
	ToolName       string          `json:"tool_name"`
	ParametersJSON json.RawMessage `json:"parameters_json"`
	CreatedAt      time.Time       `json:"created_at"`
}

type Store interface {
	GetOrgIDByName(ctx context.Context, orgName string) (uuid.UUID, error)
	UpsertGovernanceRule(ctx context.Context, orgID uuid.UUID, ruleText string) (uuid.UUID, error)
	GetGovernanceRulesByOrg(ctx context.Context, orgID uuid.UUID) ([]GovernanceRule, error)
	DeleteGovernanceRule(ctx context.Context, ruleID uuid.UUID, orgID uuid.UUID) error
	UpsertOrg(ctx context.Context, installationID int64, orgName string) (uuid.UUID, error)
	UpsertRepo(ctx context.Context, orgID uuid.UUID, githubRepoID int64, name, fullName string, metadata json.RawMessage) (uuid.UUID, error)
	UpsertContract(ctx context.Context, repoID uuid.UUID, schemaType, specPath, branch, commitSHA, rawContent string) (uuid.UUID, error)
	UpsertDependency(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, confidenceScore int, requiredNoticeDays int) error
	GetContractsByProviderFullName(ctx context.Context, providerFullName string) ([]Contract, error)
	GetConsumersByProviderContract(ctx context.Context, providerContractID uuid.UUID) ([]ConsumerDependency, error)
	ListReposByOrg(ctx context.Context, orgName string) ([]Repository, error)
	GetDependencyGraph(ctx context.Context, orgName string) ([]DependencyEdge, error)
	UpsertRepoMetric(ctx context.Context, repoID uuid.UUID, score int) error
	GetAllRepositories(ctx context.Context) ([]Repository, error)
	CountRecentBreakingChanges(ctx context.Context, repoID uuid.UUID, since time.Time) (int, error)
	GetCommitVelocity(ctx context.Context, repoID uuid.UUID, since time.Time) (int, error)
	GetTimeSinceLastBreak(ctx context.Context, repoID uuid.UUID) (*time.Time, error)
	CountReposByOrg(ctx context.Context, orgName string) (int, error)
	CountDownstreamDependencies(ctx context.Context, providerFullName string) (int, error)
	RecordBreakingChange(ctx context.Context, repoID uuid.UUID, orgName, repoName, gitSHA string, breakingChanges json.RawMessage) error
	GetBreakingChangeHistory(ctx context.Context, orgName, repoName string, limit int) ([]BreakingChangeRecord, error)
	GetBreakingChangesBetween(ctx context.Context, since, until time.Time) ([]BreakingChangeRecord, error)
	UpdateDependencyConfidence(ctx context.Context, consumerFullName, providerURL string, boostAmount float64) error
	UpdateDependencyStatus(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, status string) error
	SaveDiffReport(ctx context.Context, diffReport json.RawMessage, isAuditMode bool, orgName, repoName string) (uuid.UUID, error)
	GetDiffReport(ctx context.Context, id uuid.UUID) (json.RawMessage, error)
	GetDiffReportsByRepo(ctx context.Context, orgName, repoName string, limit int) ([]DiffReportRecord, error)
	CreatePreviewSession(ctx context.Context, prNumber int, diffID uuid.UUID, expiresAt time.Time) (uuid.UUID, error)
	GetPreviewSession(ctx context.Context, token uuid.UUID) (json.RawMessage, time.Time, error)
	ExpirePreviewSessionsForPR(ctx context.Context, prNumber int, orgName string, repoName string) error
	RecordDriftAnomaly(ctx context.Context, anomaly DriftAnomaly) error
	GetDriftAnomalies(ctx context.Context, orgName, repoName string) ([]DriftAnomaly, error)
	RegisterWebhook(ctx context.Context, config WebhookConfig) error
	GetWebhooks(ctx context.Context, org string) ([]WebhookConfig, error)
	GetROIMetrics(ctx context.Context, orgID string) (ROIMetrics, error)
	UpdateStripeCustomerID(ctx context.Context, orgID uuid.UUID, stripeID string) error
	GetBillingStatus(ctx context.Context, orgName string) (time.Time, *string, error)
	GetConsumerManifests(ctx context.Context, providerRepo, consumerRepo string) (json.RawMessage, error)
	RegisterAgent(ctx context.Context, repoName, owner string, tools []AgentToolDependency) (uuid.UUID, error)
	GetAgentsByTool(ctx context.Context, toolName string) ([]AgentConsumer, error)
	UpsertEndpointTraffic(ctx context.Context, repoID uuid.UUID, method, path string, timestamp time.Time) error
	GetZeroTrafficEndpoints(ctx context.Context, orgName string, since time.Time) ([]EndpointTraffic, error)
	GetPublicSchema(ctx context.Context, namespace, name, version string) (*PublicSchema, error)
	PublishPublicSchema(ctx context.Context, namespace, name, version, schemaType, content string) error
	UpsertOrgKMSConfig(ctx context.Context, orgName, provider, keyARN string) error
	GetOrgKMSConfig(ctx context.Context, orgName string) (string, string, error) // Returns provider, keyARN
	PublishPlugin(ctx context.Context, name, description string, schemaContent json.RawMessage) (uuid.UUID, error)
	ListPlugins(ctx context.Context) ([]MarketplacePlugin, error)
	UpsertInsurancePolicy(ctx context.Context, orgID uuid.UUID, policyLimitCents int64) (uuid.UUID, error)
	GetInsurancePolicy(ctx context.Context, orgID uuid.UUID) (*InsurancePolicy, error)
	CreateInsuranceClaim(ctx context.Context, claim InsuranceClaim) (uuid.UUID, error)
	GetInsuranceClaims(ctx context.Context, orgID uuid.UUID) ([]InsuranceClaim, error)
	Pool() *pgxpool.Pool
}

type PublicNamespace struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type PublicSchema struct {
	ID            uuid.UUID `json:"id"`
	NamespaceID   uuid.UUID `json:"namespace_id"`
	NamespaceName string    `json:"namespace_name,omitempty"`
	Name          string    `json:"name"`
	Version       string    `json:"version"`
	SchemaType    string    `json:"schema_type"`
	SchemaContent string    `json:"schema_content"`
	CreatedAt     time.Time `json:"created_at"`
}

type MarketplacePlugin struct {
	ID            uuid.UUID       `json:"id"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	SchemaContent json.RawMessage `json:"schema_content"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type InsurancePolicy struct {
	ID               uuid.UUID `json:"id"`
	OrgID            uuid.UUID `json:"org_id"`
	PolicyLimitCents int64     `json:"policy_limit_cents"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type InsuranceClaim struct {
	ID           uuid.UUID `json:"id"`
	OrgID        uuid.UUID `json:"org_id"`
	PolicyID     uuid.UUID `json:"policy_id"`
	GithubPRUrl  string    `json:"github_pr_url"`
	IncidentDate time.Time `json:"incident_date"`
	Status       string    `json:"status"`
	AmountCents  int64     `json:"amount_cents"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
