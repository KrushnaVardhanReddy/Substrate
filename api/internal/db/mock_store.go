package db

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type MockStore struct {
	UpsertRepoMetricFunc           func(ctx context.Context, repoID uuid.UUID, score int) error
	GetAllRepositoriesFunc         func(ctx context.Context) ([]Repository, error)
	CountRecentBreakingChangesFunc func(ctx context.Context, repoID uuid.UUID, since time.Time) (int, error)
	GetCommitVelocityFunc          func(ctx context.Context, repoID uuid.UUID, since time.Time) (int, error)
	GetTimeSinceLastBreakFunc      func(ctx context.Context, repoID uuid.UUID) (*time.Time, error)

	RecordDriftAnomalyFunc             func(ctx context.Context, anomaly DriftAnomaly) error
	GetDriftAnomaliesFunc              func(ctx context.Context, orgName, repoName string) ([]DriftAnomaly, error)
	UpsertOrgFunc                      func(ctx context.Context, installationID int64, orgName string) (uuid.UUID, error)
	UpsertRepoFunc                     func(ctx context.Context, orgID uuid.UUID, githubRepoID int64, name, fullName string, metadata json.RawMessage) (uuid.UUID, error)
	UpsertContractFunc                 func(ctx context.Context, repoID uuid.UUID, schemaType, specPath, branch, commitSHA, rawContent string) (uuid.UUID, error)
	UpsertDependencyFunc               func(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, confidenceScore int, requiredNoticeDays int) error
	GetContractsByProviderFullNameFunc func(ctx context.Context, providerFullName string) ([]Contract, error)
	GetConsumersByProviderContractFunc func(ctx context.Context, providerContractID uuid.UUID) ([]ConsumerDependency, error)
	ListReposByOrgFunc                 func(ctx context.Context, orgName string) ([]Repository, error)
	GetDependencyGraphFunc             func(ctx context.Context, orgName string) ([]DependencyEdge, error)
	CountReposByOrgFunc                func(ctx context.Context, orgName string) (int, error)
	UpdateDependencyConfidenceFunc     func(ctx context.Context, consumerFullName, providerURL string, boostAmount float64) error
	CountDownstreamDependenciesFunc    func(ctx context.Context, providerFullName string) (int, error)
	RecordBreakingChangeFunc           func(ctx context.Context, repoID uuid.UUID, orgName, repoName, gitSHA string, breakingChanges json.RawMessage) error
	GetBreakingChangeHistoryFunc       func(ctx context.Context, orgName, repoName string, limit int) ([]BreakingChangeRecord, error)
	RegisterWebhookFunc                func(ctx context.Context, config WebhookConfig) error
	GetWebhooksFunc                    func(ctx context.Context, org string) ([]WebhookConfig, error)
	GetROIMetricsFunc                  func(ctx context.Context, orgID string) (ROIMetrics, error)
	UpdateStripeCustomerIDFunc         func(ctx context.Context, orgID uuid.UUID, stripeID string) error
	GetBillingStatusFunc               func(ctx context.Context, orgName string) (time.Time, *string, error)
	SaveDiffReportFunc                 func(ctx context.Context, diffReport json.RawMessage, isAuditMode bool, orgName, repoName string) (uuid.UUID, error)
	GetDiffReportFunc                  func(ctx context.Context, id uuid.UUID) (json.RawMessage, error)
	GetDiffReportsByRepoFunc           func(ctx context.Context, orgName, repoName string, limit int) ([]DiffReportRecord, error)
	UpdateDependencyStatusFunc         func(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, status string) error
	GetConsumerManifestsFunc             func(ctx context.Context, providerRepo, consumerRepo string) (json.RawMessage, error)
	RegisterAgentFunc                  func(ctx context.Context, repoName, owner string, tools []AgentToolDependency) (uuid.UUID, error)
	GetAgentsByToolFunc                func(ctx context.Context, toolName string) ([]AgentConsumer, error)
	GetPublicSchemaFunc                func(ctx context.Context, namespace, name, version string) (*PublicSchema, error)
	PublishPublicSchemaFunc            func(ctx context.Context, namespace, name, version, schemaType, content string) error
	UpsertOrgKMSConfigFunc               func(ctx context.Context, orgName, provider, keyARN string) error
	GetOrgKMSConfigFunc                  func(ctx context.Context, orgName string) (string, string, error)
}

func (m *MockStore) GetPublicSchema(ctx context.Context, namespace, name, version string) (*PublicSchema, error) {
	if m.GetPublicSchemaFunc != nil {
		return m.GetPublicSchemaFunc(ctx, namespace, name, version)
	}
	return nil, nil
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

func (m *MockStore) SaveDiffReport(ctx context.Context, diffReport json.RawMessage, isAuditMode bool, orgName, repoName string) (uuid.UUID, error) {
	if m.SaveDiffReportFunc != nil {
		return m.SaveDiffReportFunc(ctx, diffReport, isAuditMode, orgName, repoName)
	}
	return uuid.New(), nil
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
