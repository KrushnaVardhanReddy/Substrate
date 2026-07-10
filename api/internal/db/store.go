package db

import (
	"context"
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
	CreatedAt    time.Time
}

type ConsumerDependency struct {
	ConsumerRepoID     uuid.UUID
	ConsumerFullName   string
	ContractRawContent string
}

type DependencyEdge struct {
	ConsumerFullName string `json:"consumer"`
	ProviderFullName string `json:"provider"`
	Status           string `json:"status"`
}

type Store interface {
	UpsertOrg(ctx context.Context, installationID int64, orgName string) (uuid.UUID, error)
	UpsertRepo(ctx context.Context, orgID uuid.UUID, githubRepoID int64, name, fullName string) (uuid.UUID, error)
	UpsertContract(ctx context.Context, repoID uuid.UUID, schemaType, specPath, branch, commitSHA, rawContent string) (uuid.UUID, error)
	UpsertDependency(ctx context.Context, consumerRepoID, providerContractID uuid.UUID) error
	GetContractsByProviderFullName(ctx context.Context, providerFullName string) ([]Contract, error)
	GetConsumersByProviderContract(ctx context.Context, providerContractID uuid.UUID) ([]ConsumerDependency, error)
	ListReposByOrg(ctx context.Context, orgName string) ([]Repository, error)
	GetDependencyGraph(ctx context.Context, orgName string) ([]DependencyEdge, error)
	CountReposByOrg(ctx context.Context, orgName string) (int, error)
	CountDownstreamDependencies(ctx context.Context, providerFullName string) (int, error)
}
