package workers

import (
	"context"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/google/uuid"
	"github.com/riverqueue/river"
)

func TestRecalculateRiskScoresWorker(t *testing.T) {
	ctx := context.Background()

	repoID1 := uuid.New()
	repoID2 := uuid.New()

	mockStore := &db.MockStore{
		GetAllRepositoriesFunc: func(ctx context.Context) ([]db.Repository, error) {
			return []db.Repository{
				{ID: repoID1, FullName: "org/repo1"},
				{ID: repoID2, FullName: "org/repo2"},
			}, nil
		},
		CountRecentBreakingChangesFunc: func(ctx context.Context, repoID uuid.UUID, since time.Time) (int, error) {
			if repoID == repoID1 {
				return 5, nil
			}
			return 0, nil
		},
		GetCommitVelocityFunc: func(ctx context.Context, repoID uuid.UUID, since time.Time) (int, error) {
			return 10, nil
		},
		CountDownstreamDependenciesFunc: func(ctx context.Context, providerFullName string) (int, error) {
			return 0, nil
		},
		GetTimeSinceLastBreakFunc: func(ctx context.Context, repoID uuid.UUID) (*time.Time, error) {
			return nil, nil
		},
		UpsertRepoMetricFunc: func(ctx context.Context, repoID uuid.UUID, score int) error {
			return nil
		},
	}

	worker := &RecalculateRiskScoresWorker{
		store: mockStore,
	}

	job := &river.Job[RecalculateRiskScoresArgs]{
		Args: RecalculateRiskScoresArgs{},
	}

	err := worker.Work(ctx, job)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
