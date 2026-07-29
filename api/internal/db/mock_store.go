package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MockStore struct {
	pool                           *pgxpool.Pool
	UpsertRepoMetricFunc           func(ctx context.Context, repoID uuid.UUID, score int) error
	GetAllRepositoriesFunc         func(ctx context.Context) ([]Repository, error)
	CountRecentBreakingChangesFunc func(ctx context.Context, repoID uuid.UUID, since time.Time) (int, error)
	GetCommitVelocityFunc          func(ctx context.Context, repoID uuid.UUID, since time.Time) (int, error)
	GetTimeSinceLastBreakFunc      func(ctx context.Context, repoID uuid.UUID) (*time.Time, error)

	RecordDriftAnomalyFunc             func(ctx context.Context, anomaly DriftAnomaly) error
	GetDriftAnomaliesFunc              func(ctx context.Context, orgName, repoName string) ([]DriftAnomaly, error)
	GetOrgIDByNameFunc                 func(ctx context.Context, name string) (uuid.UUID, error)
	UpsertGovernanceRuleFunc           func(ctx context.Context, id uuid.UUID, text string) (uuid.UUID, error)
	GetGovernanceRulesByOrgFunc        func(ctx context.Context, id uuid.UUID) ([]GovernanceRule, error)
	DeleteGovernanceRuleFunc           func(ctx context.Context, rid uuid.UUID, oid uuid.UUID) error
	UpsertOrgFunc                      func(ctx context.Context, installationID int64, orgName string) (uuid.UUID, error)
	UpsertRepoFunc                     func(ctx context.Context, orgID uuid.UUID, githubRepoID int64, name, fullName string, metadata json.RawMessage) (uuid.UUID, error)
	UpsertContractFunc                 func(ctx context.Context, repoID uuid.UUID, schemaType, specPath, branch, commitSHA, rawContent string) (uuid.UUID, error)
	UpsertDependencyFunc               func(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, confidenceScore int, requiredNoticeDays int) error
	GetContractsByProviderFullNameFunc func(ctx context.Context, providerFullName string) ([]Contract, error)
	GetSchemaFunc                      func(ctx context.Context, org, repo string) (string, error)
	GetConsumersByProviderContractFunc func(ctx context.Context, providerContractID uuid.UUID) ([]ConsumerDependency, error)
	ListReposByOrgFunc                 func(ctx context.Context, orgName string) ([]Repository, error)
	GetDependencyGraphFunc             func(ctx context.Context, orgName string) ([]DependencyEdge, error)
	CountReposByOrgFunc                func(ctx context.Context, orgName string) (int, error)
	UpdateDependencyConfidenceFunc     func(ctx context.Context, consumerFullName, providerURL string, boostAmount float64) error
	CountDownstreamDependenciesFunc    func(ctx context.Context, providerFullName string) (int, error)
	RecordBreakingChangeFunc           func(ctx context.Context, repoID uuid.UUID, orgName, repoName, gitSHA string, breakingChanges json.RawMessage) error
	GetBreakingChangeHistoryFunc       func(ctx context.Context, orgName, repoName string, limit int) ([]BreakingChangeRecord, error)
	GetBreakingChangesBetweenFunc      func(ctx context.Context, since, until time.Time) ([]BreakingChangeRecord, error)
	RegisterWebhookFunc                func(ctx context.Context, config WebhookConfig) error
	GetWebhooksFunc                    func(ctx context.Context, org string) ([]WebhookConfig, error)
	GetROIMetricsFunc                  func(ctx context.Context, orgID string) (ROIMetrics, error)
	UpdateStripeCustomerIDFunc         func(ctx context.Context, orgID uuid.UUID, stripeID string) error
	GetBillingStatusFunc               func(ctx context.Context, orgName string) (time.Time, *string, error)
	SaveDiffReportFunc                 func(ctx context.Context, diffReport json.RawMessage, isAuditMode bool, orgName, repoName string) (uuid.UUID, error)
	GetDiffReportFunc                  func(ctx context.Context, id uuid.UUID) (json.RawMessage, error)
	GetDiffReportsByRepoFunc           func(ctx context.Context, orgName, repoName string, limit int) ([]DiffReportRecord, error)
	CreatePreviewSessionFunc           func(ctx context.Context, prNumber int, diffID uuid.UUID, expiresAt time.Time) (uuid.UUID, error)
	GetPreviewSessionFunc              func(ctx context.Context, token uuid.UUID) (json.RawMessage, time.Time, error)
	ExpirePreviewSessionsForPRFunc     func(ctx context.Context, prNumber int, orgName string, repoName string) error
	UpdateDependencyStatusFunc         func(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, status string) error
	GetConsumerManifestsFunc           func(ctx context.Context, providerRepo, consumerRepo string) (json.RawMessage, error)
	RegisterAgentFunc                  func(ctx context.Context, repoName, owner string, tools []AgentToolDependency) (uuid.UUID, error)
	GetAgentsByToolFunc                func(ctx context.Context, toolName string) ([]AgentConsumer, error)
	GetPublicSchemaFunc                func(ctx context.Context, namespace, name, version string) (*PublicSchema, error)
	PublishPublicSchemaFunc            func(ctx context.Context, namespace, name, version, schemaType, content string) error
	UpsertOrgKMSConfigFunc             func(ctx context.Context, orgName, provider, keyARN string) error
	GetOrgKMSConfigFunc                func(ctx context.Context, orgName string) (string, string, error)
	PublishPluginFunc                  func(ctx context.Context, name, description string, schemaContent json.RawMessage) (uuid.UUID, error)
	ListPluginsFunc                    func(ctx context.Context) ([]MarketplacePlugin, error)
	UpsertInsurancePolicyFunc          func(ctx context.Context, orgID uuid.UUID, policyLimitCents int64) (uuid.UUID, error)
	GetInsurancePolicyFunc             func(ctx context.Context, orgID uuid.UUID) (*InsurancePolicy, error)
	CreateInsuranceClaimFunc           func(ctx context.Context, claim InsuranceClaim) (uuid.UUID, error)
	GetInsuranceClaimsFunc             func(ctx context.Context, orgID uuid.UUID) ([]InsuranceClaim, error)
	GetPrunedSchemaCacheFunc           func(ctx context.Context, arg sqlcgen.GetPrunedSchemaCacheParams) (sqlcgen.SchemaPruneCache, error)
	UpsertPrunedSchemaCacheFunc        func(ctx context.Context, arg sqlcgen.UpsertPrunedSchemaCacheParams) error

	InsertEcosystemEventFunc    func(ctx context.Context, arg sqlcgen.InsertEcosystemEventParams) (sqlcgen.EcosystemEvent, error)
	GetEcosystemEventsByOrgFunc func(ctx context.Context, arg sqlcgen.GetEcosystemEventsByOrgParams) ([]sqlcgen.EcosystemEvent, error)
}

func (m *MockStore) Pool() *pgxpool.Pool {
	// In memory pool for tests. We can just create a basic dummy or return nil.
	// Or we use pgxpool for real tests. Actually we use a local pool connected to postgres test database.
	// We'll add this to MockStore.
	return m.pool
}

func (m *MockStore) GetPublicSchema(ctx context.Context, namespace, name, version string) (*PublicSchema, error) {
	if m.GetPublicSchemaFunc != nil {
		return m.GetPublicSchemaFunc(ctx, namespace, name, version)
	}
	return nil, nil
}

func (m *MockStore) GetSchema(ctx context.Context, org, repo string) (string, error) {
	if m.GetSchemaFunc != nil {
		return m.GetSchemaFunc(ctx, org, repo)
	}
	return "", nil
}

func (m *MockStore) GetPrunedSchemaCache(ctx context.Context, arg sqlcgen.GetPrunedSchemaCacheParams) (sqlcgen.SchemaPruneCache, error) {
	if m.GetPrunedSchemaCacheFunc != nil {
		return m.GetPrunedSchemaCacheFunc(ctx, arg)
	}
	return sqlcgen.SchemaPruneCache{}, nil
}

func (m *MockStore) UpsertPrunedSchemaCache(ctx context.Context, arg sqlcgen.UpsertPrunedSchemaCacheParams) error {
	if m.UpsertPrunedSchemaCacheFunc != nil {
		return m.UpsertPrunedSchemaCacheFunc(ctx, arg)
	}
	return nil
}

func (m *MockStore) PublishPublicSchema(ctx context.Context, namespace, name, version, schemaType, content string) error {
	if m.PublishPublicSchemaFunc != nil {
		return m.PublishPublicSchemaFunc(ctx, namespace, name, version, schemaType, content)
	}
	return nil
}

func (m *MockStore) UpsertOrg(ctx context.Context, installationID int64, orgName string) (uuid.UUID, error) {
	if m.UpsertOrgFunc != nil {
		return m.UpsertOrgFunc(ctx, installationID, orgName)
	}
	return uuid.New(), nil
}

func (m *MockStore) UpsertRepo(ctx context.Context, orgID uuid.UUID, githubRepoID int64, name, fullName string, metadata json.RawMessage) (uuid.UUID, error) {
	if m.UpsertRepoFunc != nil {
		return m.UpsertRepoFunc(ctx, orgID, githubRepoID, name, fullName, metadata)
	}
	return uuid.New(), nil
}

func (m *MockStore) UpsertContract(ctx context.Context, repoID uuid.UUID, schemaType, specPath, branch, commitSHA, rawContent string) (uuid.UUID, error) {
	if m.UpsertContractFunc != nil {
		return m.UpsertContractFunc(ctx, repoID, schemaType, specPath, branch, commitSHA, rawContent)
	}
	return uuid.New(), nil
}

func (m *MockStore) UpsertDependency(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, confidenceScore int, requiredNoticeDays int) error {
	if m.UpsertDependencyFunc != nil {
		return m.UpsertDependencyFunc(ctx, consumerRepoID, providerContractID, confidenceScore, requiredNoticeDays)
	}
	return nil
}

func (m *MockStore) GetContractsByProviderFullName(ctx context.Context, providerFullName string) ([]Contract, error) {
	if m.GetContractsByProviderFullNameFunc != nil {
		return m.GetContractsByProviderFullNameFunc(ctx, providerFullName)
	}
	return nil, nil
}

func (m *MockStore) UpsertOrgKMSConfig(ctx context.Context, orgName, provider, keyARN string) error {
	if m.UpsertOrgKMSConfigFunc != nil {
		return m.UpsertOrgKMSConfigFunc(ctx, orgName, provider, keyARN)
	}
	return nil
}

func (m *MockStore) GetOrgKMSConfig(ctx context.Context, orgName string) (string, string, error) {
	if m.GetOrgKMSConfigFunc != nil {
		return m.GetOrgKMSConfigFunc(ctx, orgName)
	}
	return "", "", nil
}

func (m *MockStore) PublishPlugin(ctx context.Context, name, description string, schemaContent json.RawMessage) (uuid.UUID, error) {
	if m.PublishPluginFunc != nil {
		return m.PublishPluginFunc(ctx, name, description, schemaContent)
	}
	return uuid.Nil, nil
}

func (m *MockStore) ListPlugins(ctx context.Context) ([]MarketplacePlugin, error) {
	if m.ListPluginsFunc != nil {
		return m.ListPluginsFunc(ctx)
	}
	return nil, nil
}

func (m *MockStore) SaveDiffReport(ctx context.Context, diffReport json.RawMessage, isAuditMode bool, orgName, repoName string) (uuid.UUID, error) {
	if m.SaveDiffReportFunc != nil {
		return m.SaveDiffReportFunc(ctx, diffReport, isAuditMode, orgName, repoName)
	}
	return uuid.New(), nil
}

func (m *MockStore) CreatePreviewSession(ctx context.Context, prNumber int, diffID uuid.UUID, expiresAt time.Time) (uuid.UUID, error) {
	if m.CreatePreviewSessionFunc != nil {
		return m.CreatePreviewSessionFunc(ctx, prNumber, diffID, expiresAt)
	}
	return uuid.New(), nil
}

func (m *MockStore) GetPreviewSession(ctx context.Context, token uuid.UUID) (json.RawMessage, time.Time, error) {
	if m.GetPreviewSessionFunc != nil {
		return m.GetPreviewSessionFunc(ctx, token)
	}
	return json.RawMessage("{}"), time.Now().Add(7 * 24 * time.Hour), nil
}

func (m *MockStore) ExpirePreviewSessionsForPR(ctx context.Context, prNumber int, orgName string, repoName string) error {
	if m.ExpirePreviewSessionsForPRFunc != nil {
		return m.ExpirePreviewSessionsForPRFunc(ctx, prNumber, orgName, repoName)
	}
	return nil
}

func (m *MockStore) GetDiffReportsByRepo(ctx context.Context, orgName, repoName string, limit int) ([]DiffReportRecord, error) {
	if m.GetDiffReportsByRepoFunc != nil {
		return m.GetDiffReportsByRepoFunc(ctx, orgName, repoName, limit)
	}
	return nil, nil
}

func (m *MockStore) GetDiffReport(ctx context.Context, id uuid.UUID) (json.RawMessage, error) {
	if m.GetDiffReportFunc != nil {
		return m.GetDiffReportFunc(ctx, id)
	}
	return []byte("{}"), nil
}

func (m *MockStore) UpdateDependencyStatus(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, status string) error {
	if m.UpdateDependencyStatusFunc != nil {
		return m.UpdateDependencyStatusFunc(ctx, consumerRepoID, providerContractID, status)
	}
	return nil
}

func (m *MockStore) GetROIMetrics(ctx context.Context, orgID string) (ROIMetrics, error) {
	if m.GetROIMetricsFunc != nil {
		return m.GetROIMetricsFunc(ctx, orgID)
	}
	return ROIMetrics{}, nil
}

func (m *MockStore) UpdateStripeCustomerID(ctx context.Context, orgID uuid.UUID, stripeID string) error {
	if m.UpdateStripeCustomerIDFunc != nil {
		return m.UpdateStripeCustomerIDFunc(ctx, orgID, stripeID)
	}
	return nil
}

func (m *MockStore) GetBillingStatus(ctx context.Context, orgName string) (time.Time, *string, error) {
	if m.GetBillingStatusFunc != nil {
		return m.GetBillingStatusFunc(ctx, orgName)
	}
	importTime := time.Now().Add(24 * time.Hour)
	return importTime, nil, nil
}
func (m *MockStore) CountReposByOrg(ctx context.Context, orgName string) (int, error) {
	if m.CountReposByOrgFunc != nil {
		return m.CountReposByOrgFunc(ctx, orgName)
	}
	return 0, nil
}

func (m *MockStore) RecordBreakingChange(ctx context.Context, repoID uuid.UUID, orgName, repoName, gitSHA string, breakingChanges json.RawMessage) error {
	if m.RecordBreakingChangeFunc != nil {
		return m.RecordBreakingChangeFunc(ctx, repoID, orgName, repoName, gitSHA, breakingChanges)
	}
	return nil
}

func (m *MockStore) GetBreakingChangesBetween(ctx context.Context, since, until time.Time) ([]BreakingChangeRecord, error) {
	if m.GetBreakingChangesBetweenFunc != nil {
		return m.GetBreakingChangesBetweenFunc(ctx, since, until)
	}
	return nil, nil
}

func (m *MockStore) GetBreakingChangeHistory(ctx context.Context, orgName, repoName string, limit int) ([]BreakingChangeRecord, error) {
	if m.GetBreakingChangeHistoryFunc != nil {
		return m.GetBreakingChangeHistoryFunc(ctx, orgName, repoName, limit)
	}
	return []BreakingChangeRecord{}, nil
}

func (m *MockStore) CountDownstreamDependencies(ctx context.Context, providerFullName string) (int, error) {
	if m.CountDownstreamDependenciesFunc != nil {
		return m.CountDownstreamDependenciesFunc(ctx, providerFullName)
	}
	return 0, nil
}

func (m *MockStore) GetConsumersByProviderContract(ctx context.Context, providerContractID uuid.UUID) ([]ConsumerDependency, error) {
	if m.GetConsumersByProviderContractFunc != nil {
		return m.GetConsumersByProviderContractFunc(ctx, providerContractID)
	}
	return nil, nil
}

func (m *MockStore) ListReposByOrg(ctx context.Context, orgName string) ([]Repository, error) {
	if m.ListReposByOrgFunc != nil {
		return m.ListReposByOrgFunc(ctx, orgName)
	}
	return nil, nil
}

func (m *MockStore) GetDependencyGraph(ctx context.Context, orgName string) ([]DependencyEdge, error) {
	if m.GetDependencyGraphFunc != nil {
		return m.GetDependencyGraphFunc(ctx, orgName)
	}
	return nil, nil
}

func (m *MockStore) UpdateDependencyConfidence(ctx context.Context, consumerFullName, providerURL string, boostAmount float64) error {
	if m.UpdateDependencyConfidenceFunc != nil {
		return m.UpdateDependencyConfidenceFunc(ctx, consumerFullName, providerURL, boostAmount)
	}
	return nil
}

func (m *MockStore) RecordDriftAnomaly(ctx context.Context, anomaly DriftAnomaly) error {
	if m.RecordDriftAnomalyFunc != nil {
		return m.RecordDriftAnomalyFunc(ctx, anomaly)
	}
	return nil

}

func (m *MockStore) GetDriftAnomalies(ctx context.Context, orgName, repoName string) ([]DriftAnomaly, error) {
	if m.GetDriftAnomaliesFunc != nil {
		return m.GetDriftAnomaliesFunc(ctx, orgName, repoName)
	}
	return nil, nil
}

func (m *MockStore) RegisterWebhook(ctx context.Context, config WebhookConfig) error {
	if m.RegisterWebhookFunc != nil {
		return m.RegisterWebhookFunc(ctx, config)
	}
	return nil
}

func (m *MockStore) GetWebhooks(ctx context.Context, org string) ([]WebhookConfig, error) {
	if m.GetWebhooksFunc != nil {
		return m.GetWebhooksFunc(ctx, org)
	}
	return nil, nil
}

func (m *MockStore) UpsertRepoMetric(ctx context.Context, repoID uuid.UUID, score int) error {
	if m.UpsertRepoMetricFunc != nil {
		return m.UpsertRepoMetricFunc(ctx, repoID, score)
	}
	return nil
}

func (m *MockStore) GetAllRepositories(ctx context.Context) ([]Repository, error) {
	if m.GetAllRepositoriesFunc != nil {
		return m.GetAllRepositoriesFunc(ctx)
	}
	return nil, nil
}

func (m *MockStore) CountRecentBreakingChanges(ctx context.Context, repoID uuid.UUID, since time.Time) (int, error) {
	if m.CountRecentBreakingChangesFunc != nil {
		return m.CountRecentBreakingChangesFunc(ctx, repoID, since)
	}
	return 0, nil
}

func (m *MockStore) GetCommitVelocity(ctx context.Context, repoID uuid.UUID, since time.Time) (int, error) {
	if m.GetCommitVelocityFunc != nil {
		return m.GetCommitVelocityFunc(ctx, repoID, since)
	}
	return 0, nil
}

func (m *MockStore) GetTimeSinceLastBreak(ctx context.Context, repoID uuid.UUID) (*time.Time, error) {
	if m.GetTimeSinceLastBreakFunc != nil {
		return m.GetTimeSinceLastBreakFunc(ctx, repoID)
	}
	return nil, nil
}

func (m *MockStore) GetConsumerManifests(ctx context.Context, providerRepo, consumerRepo string) (json.RawMessage, error) {
	if m.GetConsumerManifestsFunc != nil {
		return m.GetConsumerManifestsFunc(ctx, providerRepo, consumerRepo)
	}
	return nil, nil
}

func (m *MockStore) RegisterAgent(ctx context.Context, repoName, owner string, tools []AgentToolDependency) (uuid.UUID, error) {
	if m.RegisterAgentFunc != nil {
		return m.RegisterAgentFunc(ctx, repoName, owner, tools)
	}
	return uuid.New(), nil
}

func (m *MockStore) GetAgentsByTool(ctx context.Context, toolName string) ([]AgentConsumer, error) {
	if m.GetAgentsByToolFunc != nil {
		return m.GetAgentsByToolFunc(ctx, toolName)
	}
	return nil, nil
}

func (m *MockStore) UpsertInsurancePolicy(ctx context.Context, orgID uuid.UUID, policyLimitCents int64) (uuid.UUID, error) {
	if m.UpsertInsurancePolicyFunc != nil {
		return m.UpsertInsurancePolicyFunc(ctx, orgID, policyLimitCents)
	}
	return uuid.Nil, nil
}

func (m *MockStore) GetInsurancePolicy(ctx context.Context, orgID uuid.UUID) (*InsurancePolicy, error) {
	if m.GetInsurancePolicyFunc != nil {
		return m.GetInsurancePolicyFunc(ctx, orgID)
	}
	return nil, fmt.Errorf("not found")
}

func (m *MockStore) CreateInsuranceClaim(ctx context.Context, claim InsuranceClaim) (uuid.UUID, error) {
	if m.CreateInsuranceClaimFunc != nil {
		return m.CreateInsuranceClaimFunc(ctx, claim)
	}
	return claim.ID, nil
}

func (m *MockStore) GetInsuranceClaims(ctx context.Context, orgID uuid.UUID) ([]InsuranceClaim, error) {
	if m.GetInsuranceClaimsFunc != nil {
		return m.GetInsuranceClaimsFunc(ctx, orgID)
	}
	return nil, nil
}

func (m *MockStore) UpsertEndpointTraffic(ctx context.Context, repoID uuid.UUID, method, path string, timestamp time.Time) error {
	return nil
}

func (m *MockStore) GetZeroTrafficEndpoints(ctx context.Context, orgName string, since time.Time) ([]EndpointTraffic, error) {
	return nil, nil
}

func (m *MockStore) GetOrgIDByName(ctx context.Context, orgName string) (uuid.UUID, error) {
	if m.GetOrgIDByNameFunc != nil {
		return m.GetOrgIDByNameFunc(ctx, orgName)
	}
	return uuid.Nil, nil
}

func (m *MockStore) UpsertGovernanceRule(ctx context.Context, orgID uuid.UUID, ruleText string) (uuid.UUID, error) {
	if m.UpsertGovernanceRuleFunc != nil {
		return m.UpsertGovernanceRuleFunc(ctx, orgID, ruleText)
	}
	return uuid.Nil, nil
}

func (m *MockStore) GetGovernanceRulesByOrg(ctx context.Context, orgID uuid.UUID) ([]GovernanceRule, error) {
	if m.GetGovernanceRulesByOrgFunc != nil {
		return m.GetGovernanceRulesByOrgFunc(ctx, orgID)
	}
	return nil, nil
}

func (m *MockStore) DeleteGovernanceRule(ctx context.Context, ruleID uuid.UUID, orgID uuid.UUID) error {
	if m.DeleteGovernanceRuleFunc != nil {
		return m.DeleteGovernanceRuleFunc(ctx, ruleID, orgID)
	}
	return nil
}

func (m *MockStore) GetCRMSecrets(ctx context.Context, orgName string) (stripeKey, sfURL, sfToken, sfClientID, sfClientSecret, sfUsername, sfPassword string, err error) {
	return "", "", "", "", "", "", "", nil
}

func (m *MockStore) CreatePartner(ctx context.Context, arg sqlcgen.CreatePartnerParams) (sqlcgen.PartnerIntegration, error) {
	return sqlcgen.PartnerIntegration{}, nil
}

func (m *MockStore) ListPartners(ctx context.Context) ([]sqlcgen.PartnerIntegration, error) {
	return []sqlcgen.PartnerIntegration{}, nil
}

func (m *MockStore) GetPartner(ctx context.Context, id uuid.UUID) (sqlcgen.PartnerIntegration, error) {
	return sqlcgen.PartnerIntegration{}, nil
}

func (m *MockStore) UpdatePartner(ctx context.Context, arg sqlcgen.UpdatePartnerParams) (sqlcgen.PartnerIntegration, error) {
	return sqlcgen.PartnerIntegration{}, nil
}

func (m *MockStore) UpdatePartnerStatus(ctx context.Context, arg sqlcgen.UpdatePartnerStatusParams) (sqlcgen.PartnerIntegration, error) {
	return sqlcgen.PartnerIntegration{}, nil
}

func (m *MockStore) DeletePartner(ctx context.Context, id uuid.UUID) error {
	// Simulate Not Found
	if id.String() == "99999999-9999-9999-9999-999999999999" {
		return fmt.Errorf("no rows in result set")
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// API Key Methods
// ─────────────────────────────────────────────────────────────────────────────

func (m *MockStore) CreateAPIKey(ctx context.Context, orgID uuid.UUID, name, prefix, hash string) (*APIKey, error) {
	panic("not implemented")
}

func (m *MockStore) ListAPIKeys(ctx context.Context, orgID uuid.UUID) ([]*APIKey, error) {
	panic("not implemented")
}

func (m *MockStore) DeleteAPIKey(ctx context.Context, id, orgID uuid.UUID) error {
	panic("not implemented")
}

func (m *MockStore) InsertSchemaValidationGap(ctx context.Context, arg sqlcgen.InsertSchemaValidationGapParams) error {
	return nil
}

func (m *MockStore) GetSchemaValidationGaps(ctx context.Context) ([]sqlcgen.SchemaValidationGap, error) {
	return []sqlcgen.SchemaValidationGap{}, nil
}

func (m *MockStore) SaveGatewayConfig(ctx context.Context, org, repo, gatewayType, crdYaml, prUrl string) error {
	return nil
}

func (m *MockStore) UpsertRepoGuide(ctx context.Context, org, repo, filePath, title, content string) error {
	return nil
}

func (m *MockStore) ListRepoGuides(ctx context.Context, org, repo string) ([]RepoGuide, error) {
	return nil, nil
}

func (m *MockStore) GetRepoGuide(ctx context.Context, org, repo, slug string) (*RepoGuide, error) {
	return nil, nil
}

func (m *MockStore) CreateAgentProfile(ctx context.Context, profile AgentProfile) (AgentProfile, error) {
	return profile, nil
}

func (m *MockStore) GetAgentProfile(ctx context.Context, id int) (AgentProfile, error) {
	return AgentProfile{}, nil
}

func (m *MockStore) ListAgentProfiles(ctx context.Context, org string) ([]AgentProfile, error) {
	return []AgentProfile{}, nil
}

func (m *MockStore) DeleteAgentProfile(ctx context.Context, id int) error {
	return nil
}

func (m *MockStore) CreateHITLQueueItem(ctx context.Context, item HITLQueueItem) (HITLQueueItem, error) {
	return item, nil
}

func (m *MockStore) ListHITLQueue(ctx context.Context, org string) ([]HITLQueueItem, error) {
	return []HITLQueueItem{}, nil
}

func (m *MockStore) GetHITLQueueItem(ctx context.Context, id int) (HITLQueueItem, error) {
	return HITLQueueItem{}, nil
}

func (m *MockStore) ResolveHITLQueueItem(ctx context.Context, id int, status string) error {
	return nil
}

func (m *MockStore) InsertEcosystemEvent(ctx context.Context, arg sqlcgen.InsertEcosystemEventParams) (sqlcgen.EcosystemEvent, error) {
	if m.InsertEcosystemEventFunc != nil {
		return m.InsertEcosystemEventFunc(ctx, arg)
	}
	return sqlcgen.EcosystemEvent{}, nil
}

func (m *MockStore) GetEcosystemEventsByOrg(ctx context.Context, arg sqlcgen.GetEcosystemEventsByOrgParams) ([]sqlcgen.EcosystemEvent, error) {
	if m.GetEcosystemEventsByOrgFunc != nil {
		return m.GetEcosystemEventsByOrgFunc(ctx, arg)
	}
	return nil, nil
}
func (m *MockStore) UpsertOrgIntegration(ctx context.Context, orgID uuid.UUID, provider, apiKeyEncrypted string) (OrgIntegration, error) {
	return OrgIntegration{}, nil
}

func (m *MockStore) GetOrgIntegration(ctx context.Context, orgID uuid.UUID, provider string) (OrgIntegration, error) {
	return OrgIntegration{}, nil
}
