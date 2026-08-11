// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package scoring

import (
	"context"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/google/uuid"
)

func TestCalculateRiskScore(t *testing.T) {
	ctx := context.Background()
	repoID := uuid.New()
	providerFullName := "org/repo"

	now := time.Now()
	twoDaysAgo := now.AddDate(0, 0, -2)

	tests := []struct {
		name             string
		breaksCount      int
		velocityCount    int
		consumerCount    int
		lastBreakTime    *time.Time
		expectedScoreMin int
		expectedScoreMax int
	}{
		{
			name:             "High Risk",
			breaksCount:      5,           // 40 pts
			velocityCount:    20,          // 20 pts
			consumerCount:    100,         // ~30 pts (log10(101)*15 = 30)
			lastBreakTime:    &twoDaysAgo, // ~10 pts
			expectedScoreMin: 95,
			expectedScoreMax: 100,
		},
		{
			name:             "Low Risk - No Consumers, No Breaks",
			breaksCount:      0,
			velocityCount:    5, // 5 pts
			consumerCount:    0,
			lastBreakTime:    nil,
			expectedScoreMin: 5,
			expectedScoreMax: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &db.MockStore{
				CountRecentBreakingChangesFunc: func(ctx context.Context, repoID uuid.UUID, since time.Time) (int, error) {
					return tt.breaksCount, nil
				},
				GetCommitVelocityFunc: func(ctx context.Context, repoID uuid.UUID, since time.Time) (int, error) {
					return tt.velocityCount, nil
				},
				CountDownstreamDependenciesFunc: func(ctx context.Context, providerFullName string) (int, error) {
					return tt.consumerCount, nil
				},
				GetTimeSinceLastBreakFunc: func(ctx context.Context, repoID uuid.UUID) (*time.Time, error) {
					return tt.lastBreakTime, nil
				},
			}

			score, err := CalculateRiskScore(ctx, mockStore, repoID, providerFullName)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if score < tt.expectedScoreMin || score > tt.expectedScoreMax {
				t.Errorf("expected score between %d and %d, got %d", tt.expectedScoreMin, tt.expectedScoreMax, score)
			}
		})
	}
}
