// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package risk

// Scorer input
type ScoreInput struct {
	BreakingChangesCount int
	BlastRadiusNodeCount int
	HasDBMigrations      bool
	E2ETestsPass         bool
}

// Scorer output
const (
	ScoreLow      = "LOW"
	ScoreMedium   = "MEDIUM"
	ScoreHigh     = "HIGH"
	ScoreCritical = "CRITICAL"
)

func CalculateRiskScore(input ScoreInput) string {
	if input.BreakingChangesCount > 0 || (input.HasDBMigrations && !input.E2ETestsPass) || input.BlastRadiusNodeCount > 10 {
		return ScoreCritical
	}
	if input.HasDBMigrations || input.BlastRadiusNodeCount > 5 {
		return ScoreHigh
	}
	if input.BlastRadiusNodeCount > 0 || !input.E2ETestsPass {
		return ScoreMedium
	}
	return ScoreLow
}
