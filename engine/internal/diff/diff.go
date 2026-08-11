// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package diff

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/treesitter"
)

// AnalyzeConsumerImpact walks the consumer directory to find exact files and line numbers
// where a removed API field is accessed using Tree-sitter AST parsing.
// It modifies the BreakingChanges array in-place and appends findings to ConsumerImpacts.
func AnalyzeConsumerImpact(consumerDir string, changes []report.Change) []report.Change {
	for i, change := range changes {
		if change.Severity != report.ChangeSeverityBreaking {
			continue
		}

		// Use the rule ID to determine if this is a removed field.
		if change.RuleID == "FIELD_REMOVED" {
			// Extract the target field name from the Path (e.g., "GET /users/{id} response.body.username")
			// We will just try to take the last component of the dot-separated path.
			parts := strings.Split(change.Path, ".")
			targetField := parts[len(parts)-1]
			if targetField == "" {
				continue
			}

			// Also handle cases where path isn't dot-separated or just uses a single word
			targetField = strings.TrimSpace(targetField)

			var impacts []string

			err := filepath.WalkDir(consumerDir, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return nil // ignore read errors
				}
				if !d.IsDir() {
					ext := strings.ToLower(filepath.Ext(path))
					if ext == ".js" || ext == ".ts" || ext == ".jsx" || ext == ".tsx" {
						lineNumbers, _ := treesitter.FindFieldUsages(path, targetField)
						for _, line := range lineNumbers {
							// Record relative path
							relPath, err := filepath.Rel(consumerDir, path)
							if err != nil {
								relPath = filepath.Base(path)
							}
							impacts = append(impacts, fmt.Sprintf("%s:%d", relPath, line))
						}
					}
				}
				return nil
			})

			if err == nil && len(impacts) > 0 {
				changes[i].ConsumerImpacts = impacts
			}
		}
	}
	return changes
}
