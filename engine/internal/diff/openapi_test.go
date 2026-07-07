package diff

import (
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
)

func TestCompareOpenAPI(t *testing.T) {
	tests := []struct {
		name          string
		baseFile      string
		revisionFile  string
		expectedSev   report.Severity
		breakingCount int
	}{
		{
			name:          "No changes",
			baseFile:      "testdata/base.yaml",
			revisionFile:  "testdata/base.yaml",
			expectedSev:   report.SeverityNoChanges,
			breakingCount: 0,
		},
		{
			name:          "Breaking changes",
			baseFile:      "testdata/base.yaml",
			revisionFile:  "testdata/rev_breaking.yaml",
			expectedSev:   report.SeverityBreaking,
			breakingCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep, err := CompareOpenAPI(tt.baseFile, tt.revisionFile, false)
			if err != nil {
				t.Fatalf("CompareOpenAPI failed: %v", err)
			}
			if rep.Summary.OverallSeverity != tt.expectedSev {
				t.Errorf("Expected severity %s, got %s", tt.expectedSev, rep.Summary.OverallSeverity)
			}
			if rep.Summary.BreakingCount != tt.breakingCount {
				t.Errorf("Expected %d breaking changes, got %d", tt.breakingCount, rep.Summary.BreakingCount)
			}
		})
	}
}
