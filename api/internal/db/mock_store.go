package db

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type MockStore struct {
	UpsertOrgFunc                      func(ctx context.Context, installationID int64, orgName string) (uuid.UUID, error)
	UpsertRepoFunc                     func(ctx context.Context, orgID uuid.UUID, githubRepoID int64, name, fullName string) (uuid.UUID, error)
	UpsertContractFunc                 func(ctx context.Context, repoID uuid.UUID, schemaType, specPath, branch, commitSHA, rawContent string) (uuid.UUID, error)
	UpsertDependencyFunc               func(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, confidenceScore int) error
	GetContractsByProviderFullNameFunc func(ctx context.Context, providerFullName string) ([]Contract, error)
	GetConsumersByProviderContractFunc func(ctx context.Context, providerContractID uuid.UUID) ([]ConsumerDependency, error)
	ListReposByOrgFunc                 func(ctx context.Context, orgName string) ([]Repository, error)
	GetDependencyGraphFunc             func(ctx context.Context, orgName string) ([]DependencyEdge, error)
	CountReposByOrgFunc                func(ctx context.Context, orgName string) (int, error)
	UpdateDependencyConfidenceFunc     func(ctx context.Context, consumerFullName, providerURL string, boostAmount float64) error
	CountDownstreamDependenciesFunc    func(ctx context.Context, providerFullName string) (int, error)
	RecordBreakingChangeFunc           func(ctx context.Context, repoID uuid.UUID, orgName, repoName, gitSHA string, breakingChanges json.RawMessage) error
	GetBreakingChangeHistoryFunc       func(ctx context.Context, orgName, repoName string, limit int) ([]BreakingChangeRecord, error)
}

func (m *MockStore) UpsertOrg(ctx context.Context, installationID int64, orgName string) (uuid.UUID, error) {
	if m.UpsertOrgFunc != nil {
		return m.UpsertOrgFunc(ctx, installationID, orgName)
	}
	return uuid.New(), nil
}

func (m *MockStore) UpsertRepo(ctx context.Context, orgID uuid.UUID, githubRepoID int64, name, fullName string) (uuid.UUID, error) {
	if m.UpsertRepoFunc != nil {
		return m.UpsertRepoFunc(ctx, orgID, githubRepoID, name, fullName)
	}
	return uuid.New(), nil
}

func (m *MockStore) UpsertContract(ctx context.Context, repoID uuid.UUID, schemaType, specPath, branch, commitSHA, rawContent string) (uuid.UUID, error) {
	if m.UpsertContractFunc != nil {
		return m.UpsertContractFunc(ctx, repoID, schemaType, specPath, branch, commitSHA, rawContent)
	}
	return uuid.New(), nil
}

func (m *MockStore) UpsertDependency(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, confidenceScore int) error {
	if m.UpsertDependencyFunc != nil {
		return m.UpsertDependencyFunc(ctx, consumerRepoID, providerContractID, confidenceScore)
	}
	return nil
}

func (m *MockStore) GetContractsByProviderFullName(ctx context.Context, providerFullName string) ([]Contract, error) {
	if m.GetContractsByProviderFullNameFunc != nil {
		return m.GetContractsByProviderFullNameFunc(ctx, providerFullName)
	}
	return nil, nil
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

func (m *MockStore) SaveDiffReport(ctx context.Context, diffReport json.RawMessage) (uuid.UUID, error) {
	id := uuid.New()
	return id, nil
}

func (m *MockStore) GetDiffReport(ctx context.Context, id uuid.UUID) (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}
