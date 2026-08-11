// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package schemaowners

import (
	"bufio"
	"strings"
)

// Parse parses a SCHEMAOWNERS file and returns a map of patterns to a list of owners
func Parse(content string) map[string][]string {
	result := make(map[string][]string)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) > 1 {
			pattern := parts[0]
			owners := parts[1:]
			result[pattern] = owners
		}
	}
	return result
}
