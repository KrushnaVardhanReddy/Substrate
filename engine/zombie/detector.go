// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package zombie

import "time"

// IsZombie returns true if the lastSeen timestamp is older than thresholdDays.
// If lastSeen is the zero value, it assumes no traffic and returns true.
func IsZombie(lastSeen time.Time, thresholdDays int) bool {
	if lastSeen.IsZero() {
		return true
	}
	return time.Since(lastSeen).Hours() > float64(thresholdDays*24)
}
