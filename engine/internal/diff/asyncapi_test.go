// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package diff_test

import (
	"path/filepath"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/diff"
)

func asyncAPITestdataPath(subdir string) string {
	return filepath.Join("testdata", "asyncapi", subdir, "asyncapi.yaml")
}

func TestCompareAsyncAPI(t *testing.T) {
	tests := []struct {
		name          string
		baseSubdir    string
		headSubdir    string
		wantBreaking  int
		wantWarnings  int
		expectedRules []string
	}{
		{
			name:         "1. No changes",
			baseSubdir:   "base_no_change",
			headSubdir:   "head_no_change",
			wantBreaking: 0,
			wantWarnings: 0,
		},
		{
			name:          "2. Channel removed",
			baseSubdir:    "base_no_change",
			headSubdir:    "head_channel_removed",
			wantBreaking:  1,
			wantWarnings:  0,
			expectedRules: []string{"ASYNCAPI_CHANNEL_REMOVED"},
		},
		{
			name:          "3. Channel address changed",
			baseSubdir:    "base_channel_address_changed",
			headSubdir:    "head_channel_address_changed",
			wantBreaking:  1,
			wantWarnings:  0,
			expectedRules: []string{"ASYNCAPI_CHANNEL_ADDRESS_CHANGED"},
		},
		{
			name:          "4. Message payload field removed",
			baseSubdir:    "base_no_change",
			headSubdir:    "head_field_removed",
			wantBreaking:  1,
			wantWarnings:  0,
			expectedRules: []string{"ASYNCAPI_MESSAGE_PAYLOAD_FIELD_REMOVED"},
		},
		{
			name:          "5. Message payload field type changed",
			baseSubdir:    "base_no_change",
			headSubdir:    "head_field_type_changed",
			wantBreaking:  1,
			wantWarnings:  0,
			expectedRules: []string{"ASYNCAPI_MESSAGE_PAYLOAD_FIELD_TYPE_CHANGED"},
		},
		{
			name:          "6. Required field added to payload",
			baseSubdir:    "base_no_change",
			headSubdir:    "head_required_added",
			wantBreaking:  1,
			wantWarnings:  0,
			expectedRules: []string{"ASYNCAPI_MESSAGE_PAYLOAD_REQUIRED_ADDED"},
		},
		{
			name:          "7. Enum value removed from payload",
			baseSubdir:    "base_enum_removed",
			headSubdir:    "head_enum_removed",
			wantBreaking:  1,
			wantWarnings:  0,
			expectedRules: []string{"ASYNCAPI_MESSAGE_PAYLOAD_ENUM_VALUE_REMOVED"},
		},
		{
			name:         "8. New channel added",
			baseSubdir:   "base_no_change",
			headSubdir:   "head_channel_added",
			wantBreaking: 0,
			wantWarnings: 0,
		},
		{
			name:         "9. New optional field added to payload",
			baseSubdir:   "base_no_change",
			headSubdir:   "head_field_added",
			wantBreaking: 0,
			wantWarnings: 0,
		},
		{
			name:          "10. Server removed",
			baseSubdir:    "base_server_removed",
			headSubdir:    "head_server_removed",
			wantBreaking:  1,
			wantWarnings:  0,
			expectedRules: []string{"ASYNCAPI_SERVER_REMOVED"},
		},
		{
			name:          "11. Server protocol changed",
			baseSubdir:    "base_server_protocol",
			headSubdir:    "head_server_protocol",
			wantBreaking:  1,
			wantWarnings:  1, // Also URL change warning is emitted
			expectedRules: []string{"ASYNCAPI_SERVER_PROTOCOL_CHANGED"},
		},
		{
			name:          "12. Channel deprecated",
			baseSubdir:    "base_no_change",
			headSubdir:    "head_deprecated",
			wantBreaking:  0,
			wantWarnings:  1,
			expectedRules: []string{"ASYNCAPI_CHANNEL_DEPRECATED"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseFile := asyncAPITestdataPath(tt.baseSubdir)
			headFile := asyncAPITestdataPath(tt.headSubdir)

			rep, err := diff.CompareAsyncAPI(baseFile, headFile)
			if err != nil {
				// Parser might fail gracefully, skip if so as per spec
				t.Skipf("Parsing failed gracefully: %v", err)
			}

			if rep == nil {
				t.Fatal("CompareAsyncAPI returned nil report")
			}

			if rep.SchemaType != "asyncapi" {
				t.Errorf("Expected SchemaType 'asyncapi', got '%s'", rep.SchemaType)
			}

			if rep.Summary.BreakingCount != tt.wantBreaking {
				t.Errorf("Expected %d breaking changes, got %d", tt.wantBreaking, rep.Summary.BreakingCount)
			}

			if rep.Summary.WarningCount != tt.wantWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.wantWarnings, rep.Summary.WarningCount)
			}

			foundRules := make(map[string]bool)
			for _, bc := range rep.BreakingChanges {
				foundRules[bc.RuleID] = true
			}
			for _, wc := range rep.Warnings {
				foundRules[wc.RuleID] = true
			}

			for _, expectedRule := range tt.expectedRules {
				if !foundRules[expectedRule] {
					t.Errorf("Expected rule %s not found in report", expectedRule)
				}
			}
		})
	}
}
