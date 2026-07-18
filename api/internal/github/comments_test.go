package github

import (
	"strings"
	"testing"
)

func TestGeneratePRComment(t *testing.T) {
	output := "Error: assertion failed"
	comment := GeneratePRComment(output)
	if !strings.Contains(comment, "## Proof of Breakage") {
		t.Errorf("expected comment to contain '## Proof of Breakage', got %q", comment)
	}
	if !strings.Contains(comment, output) {
		t.Errorf("expected comment to contain %q, got %q", output, comment)
	}
}
