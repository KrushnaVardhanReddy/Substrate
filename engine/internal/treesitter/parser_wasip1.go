// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

//go:build wasip1
// +build wasip1

package treesitter

import (
	"fmt"
)

// FindFieldUsages parses the given file and returns a list of 1-indexed line numbers
// where the targetField is accessed (e.g., as a property access, object key, or destructuring).
// This is a stub for the wasip1 architecture which cannot compile go-tree-sitter.
func FindFieldUsages(filePath string, targetField string) ([]int, error) {
	return nil, fmt.Errorf("FindFieldUsages is not supported on wasip1 due to CGO dependencies")
}
