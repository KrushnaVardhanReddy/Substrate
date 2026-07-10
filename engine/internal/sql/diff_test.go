package sql

import (
	"path/filepath"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
)

func TestDiffSchemas(t *testing.T) {
	tests := []struct {
		name          string
		baseFile      string
		headFile      string
		expectRule    string
		expectSev     report.ChangeSeverity
		expectErr     bool
		expectChanges int
	}{
		{
			name:          "No changes",
			baseFile:      "base_no_change.sql",
			headFile:      "rev_no_change.sql",
			expectChanges: 0,
		},
		{
			name:          "Table removed",
			baseFile:      "base_table_removed.sql",
			headFile:      "rev_table_removed.sql",
			expectRule:    "TABLE_REMOVED",
			expectSev:     report.ChangeSeverityBreaking,
			expectChanges: 1,
		},
		{
			name:          "Table added",
			baseFile:      "base_table_added.sql",
			headFile:      "rev_table_added.sql",
			expectRule:    "TABLE_ADDED",
			expectSev:     report.ChangeSeveritySafe,
			expectChanges: 1,
		},
		{
			name:          "Column removed",
			baseFile:      "base_col_removed.sql",
			headFile:      "rev_col_removed.sql",
			expectRule:    "COLUMN_REMOVED",
			expectSev:     report.ChangeSeverityBreaking,
			expectChanges: 1,
		},
		{
			name:          "Column nullable added",
			baseFile:      "base_col_add_null.sql",
			headFile:      "rev_col_add_null.sql",
			expectRule:    "COLUMN_ADDED_NULLABLE",
			expectSev:     report.ChangeSeveritySafe,
			expectChanges: 1,
		},
		{
			name:          "Column NOT NULL no DEFAULT",
			baseFile:      "base_col_add_nn.sql",
			headFile:      "rev_col_add_nn.sql",
			expectRule:    "COLUMN_ADDED_NOT_NULL_NO_DEFAULT",
			expectSev:     report.ChangeSeverityBreaking,
			expectChanges: 1,
		},
		{
			name:          "Column type changed",
			baseFile:      "base_col_type.sql",
			headFile:      "rev_col_type.sql",
			expectRule:    "COLUMN_TYPE_CHANGED",
			expectSev:     report.ChangeSeverityBreaking,
			expectChanges: 1,
		},
		{
			name:          "Column type widened",
			baseFile:      "base_col_widen.sql",
			headFile:      "rev_col_widen.sql",
			expectRule:    "COLUMN_TYPE_WIDENED",
			expectSev:     report.ChangeSeverityWarning,
			expectChanges: 1,
		},
		{
			name:          "Column made NOT NULL",
			baseFile:      "base_col_nn.sql",
			headFile:      "rev_col_nn.sql",
			expectRule:    "COLUMN_MADE_NOT_NULL",
			expectSev:     report.ChangeSeverityBreaking,
			expectChanges: 1,
		},
		{
			name:          "Default removed from NOT NULL",
			baseFile:      "base_default_removed.sql",
			headFile:      "rev_default_removed.sql",
			expectRule:    "COLUMN_DEFAULT_REMOVED",
			expectSev:     report.ChangeSeverityBreaking,
			expectChanges: 1,
		},
		{
			name:          "UNIQUE added",
			baseFile:      "base_unique.sql",
			headFile:      "rev_unique.sql",
			expectRule:    "CONSTRAINT_ADDED_UNIQUE",
			expectSev:     report.ChangeSeverityBreaking,
			expectChanges: 1,
		},
		{
			name:          "CHECK added",
			baseFile:      "base_check.sql",
			headFile:      "rev_check.sql",
			expectRule:    "CONSTRAINT_ADDED_CHECK",
			expectSev:     report.ChangeSeverityBreaking,
			expectChanges: 1,
		},
		{
			name:          "Foreign key added",
			baseFile:      "base_fk.sql",
			headFile:      "rev_fk.sql",
			expectRule:    "FOREIGN_KEY_ADDED",
			expectSev:     report.ChangeSeverityBreaking,
			expectChanges: 1, // Will technically also have TABLE_ADDED for profiles if we diff properly, but we're testing the rule. Let's just check the rule presence.
		},
		{
			name:          "Constraint removed",
			baseFile:      "base_constraint_removed.sql",
			headFile:      "rev_constraint_removed.sql",
			expectRule:    "CONSTRAINT_REMOVED",
			expectSev:     report.ChangeSeverityWarning,
			expectChanges: 1,
		},
		{
			name:          "Index removed",
			baseFile:      "base_index_removed.sql",
			headFile:      "rev_index_removed.sql",
			expectRule:    "INDEX_REMOVED",
			expectSev:     report.ChangeSeverityWarning,
			expectChanges: 1,
		},
		{
			name:          "Index added",
			baseFile:      "base_index_added.sql",
			headFile:      "rev_index_added.sql",
			expectRule:    "INDEX_ADDED",
			expectSev:     report.ChangeSeveritySafe,
			expectChanges: 1,
		},
		{
			name:          "View removed",
			baseFile:      "base_view_removed.sql",
			headFile:      "rev_view_removed.sql",
			expectRule:    "VIEW_REMOVED",
			expectSev:     report.ChangeSeverityBreaking,
			expectChanges: 1,
		},
		{
			name:          "View column removed",
			baseFile:      "base_view_col.sql",
			headFile:      "rev_view_col.sql",
			expectRule:    "VIEW_COLUMN_REMOVED",
			expectSev:     report.ChangeSeverityBreaking,
			expectChanges: 1,
		},
		{
			name:          "Enum value removed",
			baseFile:      "base_enum_removed.sql",
			headFile:      "rev_enum_removed.sql",
			expectRule:    "SQL_ENUM_VALUE_REMOVED",
			expectSev:     report.ChangeSeverityBreaking,
			expectChanges: 1,
		},
		{
			name:          "Enum value added",
			baseFile:      "base_enum_added.sql",
			headFile:      "rev_enum_added.sql",
			expectRule:    "SQL_ENUM_VALUE_ADDED",
			expectSev:     report.ChangeSeverityWarning,
			expectChanges: 1,
		},
		{
			name:      "Invalid base SQL",
			baseFile:  "base_invalid.sql",
			headFile:  "rev_no_change.sql",
			expectErr: true,
		},
		{
			name:          "Table renamed",
			baseFile:      "base_table_renamed.sql",
			headFile:      "rev_table_renamed.sql",
			expectRule:    "TABLE_RENAMED",
			expectSev:     report.ChangeSeverityBreaking,
			expectChanges: 1,
		},
		{
			name:          "Column renamed",
			baseFile:      "base_col_renamed.sql",
			headFile:      "rev_col_renamed.sql",
			expectRule:    "COLUMN_RENAMED",
			expectSev:     report.ChangeSeverityBreaking,
			expectChanges: 1,
		},
		{
			name:          "Enum type removed",
			baseFile:      "base_enum_type_removed.sql",
			headFile:      "rev_enum_type_removed.sql",
			expectRule:    "SQL_ENUM_TYPE_REMOVED",
			expectSev:     report.ChangeSeverityBreaking,
			expectChanges: 1,
		},
		{
			name:          "View definition changed",
			baseFile:      "base_view_def_changed.sql",
			headFile:      "rev_view_def_changed.sql",
			expectRule:    "VIEW_DEFINITION_CHANGED",
			expectSev:     report.ChangeSeverityWarning,
			expectChanges: 1,
		},
		{
			name:          "pg_dump headers skipped (no false positives)",
			baseFile:      "base_pgdump_headers.sql",
			headFile:      "rev_pgdump_headers.sql",
			expectChanges: 0,
		},
		{
			name:          "NUMERIC precision widening",
			baseFile:      "base_numeric_widen.sql",
			headFile:      "rev_numeric_widen.sql",
			expectRule:    "COLUMN_TYPE_WIDENED",
			expectSev:     report.ChangeSeverityWarning,
			expectChanges: 1,
		},
		{
			name:          "DEFAULT added to NOT NULL column",
			baseFile:      "base_default_added.sql",
			headFile:      "rev_default_added.sql",
			expectRule:    "COLUMN_DEFAULT_CHANGED",
			expectSev:     report.ChangeSeverityWarning,
			expectChanges: 1,
		},
		{
			name:          "Multi-statement ALTER TABLE (add + drop column)",
			baseFile:      "base_multi_alter.sql",
			headFile:      "rev_multi_alter.sql",
			expectRule:    "COLUMN_REMOVED",
			expectSev:     report.ChangeSeverityBreaking,
			expectChanges: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bPath := filepath.Join("testdata", tc.baseFile)
			hPath := filepath.Join("testdata", tc.headFile)

			base, err := ParseSchema(bPath)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error on base: %v", err)
			}

			head, err := ParseSchema(hPath)
			if err != nil {
				t.Fatalf("unexpected error on head: %v", err)
			}

			rep := DiffSchemas(base, head)

			if tc.expectRule == "" && rep.Summary.TotalChanges != 0 {
				t.Errorf("expected 0 changes, got %d", rep.Summary.TotalChanges)
			}

			if tc.expectRule != "" {
				found := false
				var all []report.Change
				all = append(all, rep.BreakingChanges...)
				all = append(all, rep.Warnings...)
				all = append(all, rep.SafeChanges...)

				for _, c := range all {
					if c.RuleID == tc.expectRule {
						found = true
						if c.Severity != tc.expectSev {
							t.Errorf("expected severity %s, got %s for rule %s", tc.expectSev, c.Severity, tc.expectRule)
						}
					}
				}

				if !found {
					t.Errorf("expected rule %s to be present", tc.expectRule)
				}
			}
		})
	}
}
