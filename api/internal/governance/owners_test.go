package governance

import (
	"reflect"
	"sort"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/services"
)

func TestParseSchemaOwners(t *testing.T) {
	content := []byte(`
- path: "/v1/payments/*"
  reviewers:
    - "@myorg/api-platform"
- rules:
    - SECURITY_HEADER_MODIFIED
  reviewers:
    - "@myorg/security-team"
`)
	rules, err := ParseSchemaOwners(content)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}

	if rules[0].Path != "/v1/payments/*" {
		t.Errorf("expected path /v1/payments/*, got %s", rules[0].Path)
	}
	if len(rules[0].Reviewers) != 1 || rules[0].Reviewers[0] != "@myorg/api-platform" {
		t.Errorf("expected @myorg/api-platform, got %v", rules[0].Reviewers)
	}

	if len(rules[1].Rules) != 1 || rules[1].Rules[0] != "SECURITY_HEADER_MODIFIED" {
		t.Errorf("expected rule SECURITY_HEADER_MODIFIED, got %v", rules[1].Rules)
	}
	if len(rules[1].Reviewers) != 1 || rules[1].Reviewers[0] != "@myorg/security-team" {
		t.Errorf("expected @myorg/security-team, got %v", rules[1].Reviewers)
	}
}

func TestGetReviewersForChanges(t *testing.T) {
	rules := []SchemaOwnerRule{
		{
			Path:      "/v1/payments/*",
			Reviewers: []string{"@myorg/api-platform", "@user1"},
		},
		{
			Rules:     []string{"SECURITY_HEADER_MODIFIED"},
			Reviewers: []string{"@myorg/security-team"},
		},
	}

	rep := &services.DiffReport{
		Breaking: []interface{}{
			map[string]interface{}{"path": "/v1/payments/process", "rule_id": "FIELD_REMOVED"},
		},
		Warning: []interface{}{
			map[string]interface{}{"path": "/v1/auth", "rule_id": "SECURITY_HEADER_MODIFIED"},
		},
		Info: []interface{}{
			map[string]interface{}{"path": "/v1/users", "rule_id": "FIELD_ADDED"},
		},
	}

	reviewers := GetReviewersForChanges(rules, rep)
	sort.Strings(reviewers)

	expected := []string{"@myorg/api-platform", "@myorg/security-team", "@user1"}
	sort.Strings(expected)

	if !reflect.DeepEqual(reviewers, expected) {
		t.Errorf("expected %v, got %v", expected, reviewers)
	}
}
