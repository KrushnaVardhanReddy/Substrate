// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package finops

// CalculateMonthlyCostDiff calculates the monthly egress cost difference.
// Math: ((proposed - base) * rps * 3600 * 730) / (1024^3) * costPerGB
func CalculateMonthlyCostDiff(baseBytes, proposedBytes int64, rps float64, costPerGB float64) float64 {
	diffBytes := float64(proposedBytes - baseBytes)

	// diff in bytes * RPS * seconds in an hour * hours in a month (~730)
	monthlyBytesDiff := diffBytes * rps * 3600 * 730

	// convert to GB
	monthlyGBDiff := monthlyBytesDiff / (1024 * 1024 * 1024)

	return monthlyGBDiff * costPerGB
}
