// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package zombie

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsZombie(t *testing.T) {
	tests := []struct {
		name          string
		lastSeen      time.Time
		thresholdDays int
		expected      bool
	}{
		{
			name:          "Zero time is zombie",
			lastSeen:      time.Time{},
			thresholdDays: 30,
			expected:      true,
		},
		{
			name:          "Recent traffic is not zombie",
			lastSeen:      time.Now().Add(-15 * 24 * time.Hour), // 15 days ago
			thresholdDays: 30,
			expected:      false,
		},
		{
			name:          "Old traffic is zombie",
			lastSeen:      time.Now().Add(-35 * 24 * time.Hour), // 35 days ago
			thresholdDays: 30,
			expected:      true,
		},
		{
			name:          "Exact threshold (minus epsilon) is not zombie",
			lastSeen:      time.Now().Add(-29 * 24 * time.Hour).Add(-23 * time.Hour),
			thresholdDays: 30,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsZombie(tt.lastSeen, tt.thresholdDays))
		})
	}
}
