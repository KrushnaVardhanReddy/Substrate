package scoring

import (
	"context"
	"math"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/google/uuid"
)

// CalculateRiskScore computes a 0-100 predictive risk score for a repository.
//
// Heuristics:
// 1. Historical Break Frequency (Weight: 40%): Breaking changes attempted in last 90 days.
// 2. Commit Velocity (Weight: 20%): Commit volume on schema files. (Currently a proxy/simplified).
// 3. Downstream Blast Radius (Weight: 30%): Number of consumers, scaled logarithmically.
// 4. Time Since Last Break (Weight: 10%): Decay factor based on time since last break.
func CalculateRiskScore(ctx context.Context, store db.Store, repoID uuid.UUID, providerFullName string) (int, error) {
	now := time.Now()
	ninetyDaysAgo := now.AddDate(0, 0, -90)

	// 1. Historical Break Frequency (40%)
	breaksCount, err := store.CountRecentBreakingChanges(ctx, repoID, ninetyDaysAgo)
	if err != nil {
		return 0, err
	}

	// Max out at 5 breaks for full 40 points
	breakScore := float64(breaksCount) * 8.0
	if breakScore > 40.0 {
		breakScore = 40.0
	}

	// 2. Commit Velocity (20%)
	// Note: We might not have granular schema commit tracking yet, using general commits or mock for now.
	velocityCount, err := store.GetCommitVelocity(ctx, repoID, ninetyDaysAgo)
	if err != nil {
		return 0, err
	}

	// Max out at 20 commits for full 20 points
	velocityScore := float64(velocityCount) * 1.0
	if velocityScore > 20.0 {
		velocityScore = 20.0
	}

	// 3. Downstream Blast Radius (30%)
	consumerCount, err := store.CountDownstreamDependencies(ctx, providerFullName)
	if err != nil {
		return 0, err
	}

	blastScore := 0.0
	if consumerCount > 0 {
		// Logarithmic scale: 1 consumer -> low, 50 consumers -> close to max
		// using math.Log10(count + 1) * factor
		blastScore = math.Log10(float64(consumerCount)+1.0) * 15.0
		if blastScore > 30.0 {
			blastScore = 30.0
		}
	}

	// 4. Time Since Last Break (10%)
	lastBreakTime, err := store.GetTimeSinceLastBreak(ctx, repoID)
	if err != nil {
		return 0, err
	}

	timeScore := 0.0
	if lastBreakTime != nil {
		daysSince := now.Sub(*lastBreakTime).Hours() / 24.0
		// If broken recently (0 days), penalty is 10.
		// If broken 365 days ago, penalty is 0.
		if daysSince < 365.0 {
			timeScore = 10.0 * (1.0 - (daysSince / 365.0))
		}
	} else {
		// Never broken? No penalty.
		timeScore = 0.0
	}

	totalScore := int(math.Round(breakScore + velocityScore + blastScore + timeScore))
	if totalScore > 100 {
		totalScore = 100
	} else if totalScore < 0 {
		totalScore = 0
	}

	return totalScore, nil
}
