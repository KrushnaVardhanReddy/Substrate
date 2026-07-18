package diff_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/diff"
)

func protoTestdataPath(subdir string) string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return filepath.Join(wd, "testdata", "proto", subdir)
}

func TestCompareProto(t *testing.T) {
	tests := []struct {
		name         string
		baseDir      string
		headDir      string
		wantBreaking int
		wantWarnings int
		wantRuleID   string
	}{
		{
			name:         "No changes",
			baseDir:      "base_no_change",
			headDir:      "head_no_change",
			wantBreaking: 0,
			wantWarnings: 0,
			wantRuleID:   "",
		},
		{
			name:         "Field removed",
			baseDir:      "base_field_removed",
			headDir:      "head_field_removed",
			wantBreaking: 1,
			wantWarnings: 0,
			wantRuleID:   "PROTO_FIELD_REMOVED",
		},
		{
			name:         "Field type changed",
			baseDir:      "base_field_type_changed",
			headDir:      "head_field_type_changed",
			wantBreaking: 1,
			wantWarnings: 0,
			wantRuleID:   "PROTO_FIELD_TYPE_CHANGED",
		},
		{
			name:         "Field number changed",
			baseDir:      "base_field_number_changed",
			headDir:      "head_field_number_changed",
			wantBreaking: 1,
			wantWarnings: 0,
			wantRuleID:   "PROTO_FIELD_NUMBER_CHANGED",
		},
		{
			name:         "Service removed",
			baseDir:      "base_service_removed",
			headDir:      "head_service_removed",
			wantBreaking: 1,
			wantWarnings: 0,
			wantRuleID:   "PROTO_SERVICE_REMOVED",
		},
		{
			name:         "RPC method removed",
			baseDir:      "base_rpc_removed",
			headDir:      "head_rpc_removed",
			wantBreaking: 1,
			wantWarnings: 0,
			wantRuleID:   "PROTO_RPC_REMOVED",
		},
		{
			name:         "Enum value removed",
			baseDir:      "base_enum_value_removed",
			headDir:      "head_enum_value_removed",
			wantBreaking: 1,
			wantWarnings: 0,
			wantRuleID:   "PROTO_ENUM_VALUE_REMOVED",
		},
		{
			name:         "New field added",
			baseDir:      "base_safe_field_added",
			headDir:      "head_safe_field_added",
			wantBreaking: 0,
			wantWarnings: 0,
			wantRuleID:   "",
		},
		{
			name:         "New RPC added",
			baseDir:      "base_safe_rpc_added",
			headDir:      "head_safe_rpc_added",
			wantBreaking: 0,
			wantWarnings: 0,
			wantRuleID:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basePath := protoTestdataPath(tt.baseDir)
			headPath := protoTestdataPath(tt.headDir)

			rep, err := diff.CompareProto(basePath, headPath)
			if err != nil {
				t.Fatalf("CompareProto() error = %v", err)
			}

			if rep.Summary.BreakingCount != tt.wantBreaking {
				t.Errorf("CompareProto() BreakingCount = %d, want %d", rep.Summary.BreakingCount, tt.wantBreaking)
			}
			if rep.Summary.WarningCount != tt.wantWarnings {
				t.Errorf("CompareProto() WarningCount = %d, want %d", rep.Summary.WarningCount, tt.wantWarnings)
			}

			if tt.wantRuleID != "" {
				found := false
				for _, bc := range rep.BreakingChanges {
					if bc.RuleID == tt.wantRuleID {
						found = true
						break
					}
				}
				for _, wc := range rep.Warnings {
					if wc.RuleID == tt.wantRuleID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("CompareProto() expected rule %s not found in report", tt.wantRuleID)
				}
			}
		})
	}
}

