package ports

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
)

// ─────────────────────────────────────────────────────────────────────────────
// MCP Profiles & HITL port
// ─────────────────────────────────────────────────────────────────────────────

// MCPProfileStore is the port for MCP Agent Profiles and HITL Queue DB operations.
type MCPProfileStore interface {
	CreateAgentProfile(ctx context.Context, profile db.AgentProfile) (db.AgentProfile, error)
	GetAgentProfile(ctx context.Context, id int) (db.AgentProfile, error)
	ListAgentProfiles(ctx context.Context, org string) ([]db.AgentProfile, error)
	DeleteAgentProfile(ctx context.Context, id int) error
	CreateHITLQueueItem(ctx context.Context, item db.HITLQueueItem) (db.HITLQueueItem, error)
	ListHITLQueue(ctx context.Context, org string) ([]db.HITLQueueItem, error)
	GetHITLQueueItem(ctx context.Context, id int) (db.HITLQueueItem, error)
	ResolveHITLQueueItem(ctx context.Context, id int, status string) error
}

// ─────────────────────────────────────────────────────────────────────────────
// Org port
// ─────────────────────────────────────────────────────────────────────────────

// OrgStore is the port for organisation-level DB operations.
type OrgStore interface {
	GetOrgIDByName(ctx context.Context, orgName string) (uuid.UUID, error)
	UpsertOrg(ctx context.Context, installationID int64, orgName string) (uuid.UUID, error)
	CountReposByOrg(ctx context.Context, orgName string) (int, error)
	UpsertOrgIntegration(ctx context.Context, orgID uuid.UUID, provider, apiKeyEncrypted string) (db.OrgIntegration, error)
	GetOrgIntegration(ctx context.Context, orgID uuid.UUID, provider string) (db.OrgIntegration, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Repo port
// ─────────────────────────────────────────────────────────────────────────────

// RepoStore is the port for repository-level DB operations.
type RepoStore interface {
	UpsertRepo(ctx context.Context, orgID uuid.UUID, githubRepoID int64, name, fullName string, metadata json.RawMessage) (uuid.UUID, error)
	ListReposByOrg(ctx context.Context, orgName string) ([]db.Repository, error)
	GetAllRepositories(ctx context.Context) ([]db.Repository, error)
	UpsertRepoMetric(ctx context.Context, repoID uuid.UUID, score int) error
	GetTimeSinceLastBreak(ctx context.Context, repoID uuid.UUID) (*time.Time, error)
	GetDependencyGraph(ctx context.Context, orgName string) ([]db.DependencyEdge, error)
	UpsertRepoGuide(ctx context.Context, org, repo, filePath, title, content string) error
	ListRepoGuides(ctx context.Context, org, repo string) ([]db.RepoGuide, error)
	GetRepoGuide(ctx context.Context, org, repo, slug string) (*db.RepoGuide, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Contract port
// ─────────────────────────────────────────────────────────────────────────────

// ContractStore is the port for schema contract operations.
type ContractStore interface {
	UpsertContract(ctx context.Context, repoID uuid.UUID, schemaType, specPath, branch, commitSHA, rawContent string) (uuid.UUID, error)
	GetContractsByProviderFullName(ctx context.Context, providerFullName string) ([]db.Contract, error)
	GetSchema(ctx context.Context, org, repo string) (string, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Dependency port
// ─────────────────────────────────────────────────────────────────────────────

// DependencyStore is the port for the cross-repo dependency graph.
type DependencyStore interface {
	UpsertDependency(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, confidenceScore int, requiredNoticeDays int) error
	GetDependencyGraph(ctx context.Context, orgName string) ([]db.DependencyEdge, error)
	GetConsumersByProviderContract(ctx context.Context, providerContractID uuid.UUID) ([]db.ConsumerDependency, error)
	UpdateDependencyConfidence(ctx context.Context, consumerFullName, providerURL string, boostAmount float64) error
	UpdateDependencyStatus(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, status string) error
	CountDownstreamDependencies(ctx context.Context, providerFullName string) (int, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Diff port
// ─────────────────────────────────────────────────────────────────────────────

// DiffStore is the port for persisting and retrieving diff reports.
type DiffStore interface {
	SaveDiffReport(ctx context.Context, diffReport json.RawMessage, isAuditMode bool, orgName, repoName string) (uuid.UUID, error)
	GetDiffReport(ctx context.Context, id uuid.UUID) (json.RawMessage, error)
	GetDiffReportsByRepo(ctx context.Context, orgName, repoName string, limit int) ([]db.DiffReportRecord, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Change (breaking-change history) port
// ─────────────────────────────────────────────────────────────────────────────

// ChangeStore is the port for breaking-change event history.
type ChangeStore interface {
	RecordBreakingChange(ctx context.Context, repoID uuid.UUID, orgName, repoName, gitSHA string, breakingChanges json.RawMessage) error
	GetBreakingChangeHistory(ctx context.Context, orgName, repoName string, limit int) ([]db.BreakingChangeRecord, error)
	GetBreakingChangesBetween(ctx context.Context, since, until time.Time) ([]db.BreakingChangeRecord, error)
	CountRecentBreakingChanges(ctx context.Context, repoID uuid.UUID, since time.Time) (int, error)
	GetCommitVelocity(ctx context.Context, repoID uuid.UUID, since time.Time) (int, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Preview port
// ─────────────────────────────────────────────────────────────────────────────

// PreviewStore is the port for ephemeral PR preview sessions.
type PreviewStore interface {
	CreatePreviewSession(ctx context.Context, prNumber int, diffID uuid.UUID, expiresAt time.Time) (uuid.UUID, error)
	GetPreviewSession(ctx context.Context, token uuid.UUID) (json.RawMessage, time.Time, error)
	ExpirePreviewSessionsForPR(ctx context.Context, prNumber int, orgName string, repoName string) error
}

// ─────────────────────────────────────────────────────────────────────────────
// Telemetry port
// ─────────────────────────────────────────────────────────────────────────────

// TelemetryStore is the port for runtime telemetry and drift anomalies.
type TelemetryStore interface {
	RecordDriftAnomaly(ctx context.Context, anomaly db.DriftAnomaly) error
	GetDriftAnomalies(ctx context.Context, orgName, repoName string) ([]db.DriftAnomaly, error)
	UpsertEndpointTraffic(ctx context.Context, repoID uuid.UUID, method, path string, timestamp time.Time) error
	GetZeroTrafficEndpoints(ctx context.Context, orgName string, since time.Time) ([]db.EndpointTraffic, error)
	GetROIMetrics(ctx context.Context, orgID string) (db.ROIMetrics, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Webhook port
// ─────────────────────────────────────────────────────────────────────────────

// WebhookStore is the port for egress webhook configuration.
type WebhookStore interface {
	RegisterWebhook(ctx context.Context, config db.WebhookConfig) error
	GetWebhooks(ctx context.Context, org string) ([]db.WebhookConfig, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Billing port
// ─────────────────────────────────────────────────────────────────────────────

// BillingStore is the port for Stripe/billing operations.
type BillingStore interface {
	UpdateStripeCustomerID(ctx context.Context, orgID uuid.UUID, stripeID string) error
	GetBillingStatus(ctx context.Context, orgName string) (time.Time, *string, error)
	GetConsumerManifests(ctx context.Context, providerRepo, consumerRepo string) (json.RawMessage, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Agent (MCP) port
// ─────────────────────────────────────────────────────────────────────────────

// AgentStore is the port for AI agent contract registry operations.
type AgentStore interface {
	RegisterAgent(ctx context.Context, repoName, owner string, tools []db.AgentToolDependency) (uuid.UUID, error)
	GetAgentsByTool(ctx context.Context, toolName string) ([]db.AgentConsumer, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Public Registry port
// ─────────────────────────────────────────────────────────────────────────────

// RegistryStore is the port for the public schema registry.
type RegistryStore interface {
	GetPublicSchema(ctx context.Context, namespace, name, version string) (*db.PublicSchema, error)
	PublishPublicSchema(ctx context.Context, namespace, name, version, schemaType, content string) error
}

// ─────────────────────────────────────────────────────────────────────────────
// Governance port
// ─────────────────────────────────────────────────────────────────────────────

// GovernanceStore is the port for natural-language governance rule operations.
type GovernanceStore interface {
	GetOrgIDByName(ctx context.Context, orgName string) (uuid.UUID, error)
	UpsertGovernanceRule(ctx context.Context, orgID uuid.UUID, ruleText string) (uuid.UUID, error)
	GetGovernanceRulesByOrg(ctx context.Context, orgID uuid.UUID) ([]db.GovernanceRule, error)
	DeleteGovernanceRule(ctx context.Context, ruleID uuid.UUID, orgID uuid.UUID) error
}

// ─────────────────────────────────────────────────────────────────────────────
// Insurance port
// ─────────────────────────────────────────────────────────────────────────────

// InsuranceStore is the port for schema insurance policy operations.
type InsuranceStore interface {
	GetOrgIDByName(ctx context.Context, orgName string) (uuid.UUID, error)
	UpsertInsurancePolicy(ctx context.Context, orgID uuid.UUID, policyLimitCents int64) (uuid.UUID, error)
	GetInsurancePolicy(ctx context.Context, orgID uuid.UUID) (*db.InsurancePolicy, error)
	CreateInsuranceClaim(ctx context.Context, claim db.InsuranceClaim) (uuid.UUID, error)
	GetInsuranceClaims(ctx context.Context, orgID uuid.UUID) ([]db.InsuranceClaim, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Marketplace port
// ─────────────────────────────────────────────────────────────────────────────

// MarketplaceStore is the port for the community plugin marketplace.
type MarketplaceStore interface {
	PublishPlugin(ctx context.Context, name, description string, schemaContent json.RawMessage) (uuid.UUID, error)
	ListPlugins(ctx context.Context) ([]db.MarketplacePlugin, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// KMS / Enterprise BYOK port
// ─────────────────────────────────────────────────────────────────────────────

// KMSStore is the port for enterprise BYOK key management.
type KMSStore interface {
	UpsertOrgKMSConfig(ctx context.Context, orgName, provider, keyARN string) error
	GetOrgKMSConfig(ctx context.Context, orgName string) (string, string, error)
	GetCRMSecrets(ctx context.Context, orgName string) (stripeKey, sfURL, sfToken, sfClientID, sfClientSecret, sfUsername, sfPassword string, err error)
}

// ─────────────────────────────────────────────────────────────────────────────
// Partner port
// ─────────────────────────────────────────────────────────────────────────────

type PartnerStore interface {
	CreatePartner(ctx context.Context, arg sqlcgen.CreatePartnerParams) (sqlcgen.PartnerIntegration, error)
	ListPartners(ctx context.Context) ([]sqlcgen.PartnerIntegration, error)
	GetPartner(ctx context.Context, id uuid.UUID) (sqlcgen.PartnerIntegration, error)
	UpdatePartner(ctx context.Context, arg sqlcgen.UpdatePartnerParams) (sqlcgen.PartnerIntegration, error)
	UpdatePartnerStatus(ctx context.Context, arg sqlcgen.UpdatePartnerStatusParams) (sqlcgen.PartnerIntegration, error)
	DeletePartner(ctx context.Context, id uuid.UUID) error
}

// ─────────────────────────────────────────────────────────────────────────────
// Pool accessor — needed for sub-packages that build their own queries
// (e.g. handler.PartnersHandler, handler.PublicProfileHandler)
// ─────────────────────────────────────────────────────────────────────────────

type PoolProvider interface {
	InsertSchemaValidationGap(ctx context.Context, arg sqlcgen.InsertSchemaValidationGapParams) error
	GetSchemaValidationGaps(ctx context.Context) ([]sqlcgen.SchemaValidationGap, error)
	GetPrunedSchemaCache(ctx context.Context, arg sqlcgen.GetPrunedSchemaCacheParams) (sqlcgen.SchemaPruneCache, error)
	UpsertPrunedSchemaCache(ctx context.Context, arg sqlcgen.UpsertPrunedSchemaCacheParams) error
	SaveGatewayConfig(ctx context.Context, org, repo, gatewayType, crdYaml, prUrl string) error
	Pool() *pgxpool.Pool
}

// ─────────────────────────────────────────────────────────────────────────────
// Composite helpers — multi-domain handlers use these instead of the full Store
// ─────────────────────────────────────────────────────────────────────────────

// CanDeployStore is the composite port for the can-deploy registry check,
// which needs both change history and the dependency graph.
type CanDeployStore interface {
	ChangeStore
	RepoStore
}

// LimitsStore is the port used by the tier-limits middleware to enforce
// per-org resource quotas.
type LimitsStore interface {
	OrgStore
	RepoStore
	ContractStore
	DependencyStore
}

// CrossRepoStore is the composite port needed by the cross-repo check service.
type CrossRepoStore interface {
	ContractStore
	BillingStore
	DependencyStore
	KMSStore
}

// EgressStore is the narrow port needed by the egress (webhook dispatch) layer.
type EgressStore interface {
	WebhookStore
}

// BadgesStore is the composite port for the contract score badge handler.
type BadgesStore interface {
	RepoStore
	ChangeStore
}

// RollbackStore is the composite port for the can-rollback handler,
// which needs contracts and the cross-repo check.
type RollbackStore interface {
	ContractStore
	CrossRepoStore
}

// DiscoveryStore is the composite port used by the telemetry/OTel handler
// and the runtime signal discovery package (UpdateDependencyConfidence).
type DiscoveryStore interface {
	TelemetryStore
	DependencyStore
}

// SaveDiffStore is the composite port for the diff save flow, which spans
// diff persistence, governance rule look-up, org resolution, cross-repo check,
// and egress webhook dispatch.
type SaveDiffStore interface {
	DiffStore
	GovernanceStore
	DependencyStore
	ContractStore
	CrossRepoStore // needed for services.PerformCrossRepoCheck
	EgressStore    // needed for egress.DispatchEvent
}

// ─────────────────────────────────────────────────────────────────────────────
// Composite — convenience alias used by the router and wiring layer.
// All concrete sub-interfaces above are embedded here so that db.PGStore,
// which implements every individual method, automatically satisfies this type.
// ─────────────────────────────────────────────────────────────────────────────

// Store is the composite port that embeds every domain-specific port.
// Use the narrow sub-interfaces in handler constructors; use Store only in
// wiring code (router, main) where the full surface is needed.
type Store interface {
	OrgStore
	RepoStore
	ContractStore
	DependencyStore
	DiffStore
	ChangeStore
	MCPProfileStore
	PreviewStore
	TelemetryStore
	WebhookStore
	BillingStore
	APIKeyStore
	AgentStore
	RegistryStore
	GovernanceStore
	InsuranceStore
	MarketplaceStore
	KMSStore
	PartnerStore
	PoolProvider

	InsertEcosystemEvent(ctx context.Context, arg sqlcgen.InsertEcosystemEventParams) (sqlcgen.EcosystemEvent, error)
	GetEcosystemEventsByOrg(ctx context.Context, arg sqlcgen.GetEcosystemEventsByOrgParams) ([]sqlcgen.EcosystemEvent, error)
}

// ─────────────────────────────────────────────────────────────────────────────
// API Key port
// ─────────────────────────────────────────────────────────────────────────────

type APIKeyStore interface {
	CreateAPIKey(ctx context.Context, orgID uuid.UUID, name, prefix, hash string) (*db.APIKey, error)
	ListAPIKeys(ctx context.Context, orgID uuid.UUID) ([]*db.APIKey, error)
	DeleteAPIKey(ctx context.Context, id, orgID uuid.UUID) error
}
