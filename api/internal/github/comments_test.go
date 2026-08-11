// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

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

func TestGenerateNegotiationComment(t *testing.T) {
	output := "Error: schema mismatch"
	mentions := []string{"@alice", "@bob"}
	comment := GenerateNegotiationComment(output, mentions)
	if !strings.Contains(comment, "@alice, @bob") {
		t.Errorf("expected comment to contain mentions, got %q", comment)
	}
	if !strings.Contains(comment, "React with 👍 to acknowledge and approve") {
		t.Errorf("expected comment to contain approval instructions, got %q", comment)
	}
	if !strings.Contains(comment, output) {
		t.Errorf("expected comment to contain output, got %q", comment)
	}
}
func TestParseNegotiationComment(t *testing.T) {
	comment := "Some text @alice, @bob\nMore text"
	mentions := ParseNegotiationComment(comment)
	if len(mentions) != 2 {
		t.Errorf("expected 2 mentions, got %d", len(mentions))
	}
}
