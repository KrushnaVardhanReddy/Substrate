package workers

import (
	"context"
	"log"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/scoring"
	"github.com/riverqueue/river"
)

type RecalculateRiskScoresArgs struct{}

func (RecalculateRiskScoresArgs) Kind() string { return "RecalculateRiskScores" }

type RecalculateRiskScoresWorker struct {
	river.WorkerDefaults[RecalculateRiskScoresArgs]
	store db.Store
}

func (w *RecalculateRiskScoresWorker) Work(ctx context.Context, job *river.Job[RecalculateRiskScoresArgs]) error {
	log.Println("Starting RecalculateRiskScores background job...")

	repos, err := w.store.GetAllRepositories(ctx)
	if err != nil {
		log.Printf("Failed to get all repositories: %v", err)
		return err
	}

	for _, repo := range repos {
		score, err := scoring.CalculateRiskScore(ctx, w.store, repo.ID, repo.FullName)
		if err != nil {
			log.Printf("Failed to calculate risk score for repo %s: %v", repo.FullName, err)
			continue
		}

		err = w.store.UpsertRepoMetric(ctx, repo.ID, score)
		if err != nil {
			log.Printf("Failed to save risk score for repo %s: %v", repo.FullName, err)
			continue
		}
	}

	log.Println("Successfully finished RecalculateRiskScores job.")
	return nil
}
