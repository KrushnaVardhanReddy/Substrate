// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package gates

import (
	"testing"
)

func TestEvaluateGate(t *testing.T) {
	tests := []struct {
		name          string
		gate          string
		breakingCount int
		warningCount  int
		crossRepoSafe bool
		expected      bool
	}{
		// Strict gate
		{"strict_all_zero_safe", "strict", 0, 0, true, true},
		{"strict_warnings_unsafe", "strict", 0, 1, true, false},
		{"strict_breaking_unsafe", "strict", 1, 0, true, false},
		{"strict_all_zero_cross_unsafe", "strict", 0, 0, false, false},

		// Standard gate
		{"standard_all_zero_safe", "standard", 0, 0, true, true},
		{"standard_warnings_safe", "standard", 0, 1, true, true},
		{"standard_breaking_unsafe", "standard", 1, 0, true, false},
		{"standard_all_zero_cross_unsafe", "standard", 0, 0, false, false},

		// Default gate (same as standard)
		{"default_all_zero_safe", "", 0, 0, true, true},
		{"default_warnings_safe", "", 0, 1, true, true},
		{"default_breaking_unsafe", "", 1, 0, true, false},
		{"default_all_zero_cross_unsafe", "", 0, 0, false, false},

		// Permissive gate
		{"permissive_all_zero_safe", "permissive", 0, 0, true, true},
		{"permissive_warnings_safe", "permissive", 0, 1, true, true},
		{"permissive_breaking_safe", "permissive", 1, 0, true, true},
		{"permissive_all_zero_cross_unsafe", "permissive", 0, 0, false, false},
		{"permissive_breaking_cross_unsafe", "permissive", 1, 0, false, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := EvaluateGate(tc.gate, tc.breakingCount, tc.warningCount, tc.crossRepoSafe)
			if result != tc.expected {
				t.Errorf("EvaluateGate(%q, %d, %d, %v) = %v; want %v",
					tc.gate, tc.breakingCount, tc.warningCount, tc.crossRepoSafe, result, tc.expected)
			}
		})
	}
}
