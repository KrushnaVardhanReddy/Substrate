// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package risk

import "testing"

func TestCalculateRiskScore(t *testing.T) {
	tests := []struct {
		name     string
		input    ScoreInput
		expected string
	}{
		{
			name:     "Low Risk - No Breaking Changes, No Migrations, Passes E2E, No Blast Radius",
			input:    ScoreInput{BreakingChangesCount: 0, BlastRadiusNodeCount: 0, HasDBMigrations: false, E2ETestsPass: true},
			expected: ScoreLow,
		},
		{
			name:     "Medium Risk - Fails E2E",
			input:    ScoreInput{BreakingChangesCount: 0, BlastRadiusNodeCount: 0, HasDBMigrations: false, E2ETestsPass: false},
			expected: ScoreMedium,
		},
		{
			name:     "Medium Risk - Blast radius > 0",
			input:    ScoreInput{BreakingChangesCount: 0, BlastRadiusNodeCount: 3, HasDBMigrations: false, E2ETestsPass: true},
			expected: ScoreMedium,
		},
		{
			name:     "High Risk - Has Migrations",
			input:    ScoreInput{BreakingChangesCount: 0, BlastRadiusNodeCount: 0, HasDBMigrations: true, E2ETestsPass: true},
			expected: ScoreHigh,
		},
		{
			name:     "High Risk - Blast radius > 5",
			input:    ScoreInput{BreakingChangesCount: 0, BlastRadiusNodeCount: 6, HasDBMigrations: false, E2ETestsPass: true},
			expected: ScoreHigh,
		},
		{
			name:     "Critical Risk - Breaking changes",
			input:    ScoreInput{BreakingChangesCount: 1, BlastRadiusNodeCount: 0, HasDBMigrations: false, E2ETestsPass: true},
			expected: ScoreCritical,
		},
		{
			name:     "Critical Risk - Blast radius > 10",
			input:    ScoreInput{BreakingChangesCount: 0, BlastRadiusNodeCount: 11, HasDBMigrations: false, E2ETestsPass: true},
			expected: ScoreCritical,
		},
		{
			name:     "Critical Risk - Migrations and Failed E2E",
			input:    ScoreInput{BreakingChangesCount: 0, BlastRadiusNodeCount: 0, HasDBMigrations: true, E2ETestsPass: false},
			expected: ScoreCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := CalculateRiskScore(tt.input)
			if actual != tt.expected {
				t.Errorf("Expected %s, but got %s", tt.expected, actual)
			}
		})
	}
}
