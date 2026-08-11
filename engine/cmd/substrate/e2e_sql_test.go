// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "substrate")

	cmd := exec.Command("go", "build", "-o", binPath, ".")

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build binary: %v\nOutput: %s", err, string(out))
	}
	return binPath
}

func TestE2ESQL(t *testing.T) {
	binPath := buildBinary(t)

	// Since we are running from cmd/substrate, the path to testdata is:
	testdataDir := filepath.Join("..", "..", "internal", "sql", "testdata")

	tests := []struct {
		name         string
		base         string
		head         string
		format       string
		expectExit   int
		expectOutput []string
	}{
		{
			name:         "No change",
			base:         "base_no_change.sql",
			head:         "rev_no_change.sql",
			format:       "text",
			expectExit:   0,
			expectOutput: []string{"0 BREAKING"},
		},
		{
			name:         "Column removed",
			base:         "base_col_removed.sql",
			head:         "rev_col_removed.sql",
			format:       "text",
			expectExit:   2,
			expectOutput: []string{"COLUMN_REMOVED", "BREAKING"},
		},
		{
			name:         "Column added NOT NULL no DEFAULT",
			base:         "base_col_add_nn.sql",
			head:         "rev_col_add_nn.sql",
			format:       "text",
			expectExit:   2,
			expectOutput: []string{"COLUMN_ADDED_NOT_NULL_NO_DEFAULT"},
		},
		{
			name:         "Table removed JSON",
			base:         "base_table_removed.sql",
			head:         "rev_table_removed.sql",
			format:       "json",
			expectExit:   2,
			expectOutput: []string{`"breaking_changes":`, `"TABLE_REMOVED"`},
		},
		{
			name:         "Invalid SQL",
			base:         "base_invalid.sql",
			head:         "rev_no_change.sql",
			format:       "text",
			expectExit:   3,
			expectOutput: []string{"sql parse error"},
		},
		{
			name:         "Table renamed",
			base:         "base_table_renamed.sql",
			head:         "rev_table_renamed.sql",
			format:       "text",
			expectExit:   2,
			expectOutput: []string{"TABLE_RENAMED", "BREAKING"},
		},
		{
			name:         "View definition changed",
			base:         "base_view_def_changed.sql",
			head:         "rev_view_def_changed.sql",
			format:       "text",
			expectExit:   1,
			expectOutput: []string{"VIEW_DEFINITION_CHANGED"},
		},
		{
			name:         "Performance Risk FK",
			base:         "base_perf_fk.sql",
			head:         "rev_perf_fk.sql",
			format:       "text",
			expectExit:   2, // Contains BREAKING change FOREIGN_KEY_ADDED, wait actually exit 2
			expectOutput: []string{"🚀 PERFORMANCE RISKS", "FOREIGN_KEY_MISSING_INDEX"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bPath := filepath.Join(testdataDir, tc.base)
			hPath := filepath.Join(testdataDir, tc.head)

			// Execute binary
			cmd := exec.Command(binPath, "diff", bPath, hPath, "--format", tc.format)
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				} else {
					t.Fatalf("failed to run binary: %v\nStderr: %s", err, stderr.String())
				}
			}

			if exitCode != tc.expectExit {
				t.Errorf("expected exit code %d, got %d. Stderr: %s, Stdout: %s", tc.expectExit, exitCode, stderr.String(), stdout.String())
			}

			output := stdout.String() + stderr.String()
			for _, expected := range tc.expectOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("expected output to contain %q, but it didn't.\nOutput:\n%s", expected, output)
				}
			}
		})
	}
}
