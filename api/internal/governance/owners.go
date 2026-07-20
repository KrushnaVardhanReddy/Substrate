package governance

import (
	"path/filepath"
	"regexp"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/services"
	"gopkg.in/yaml.v3"
)

type SchemaOwnerRule struct {
	Path      string   `yaml:"path,omitempty"`
	Rules     []string `yaml:"rules,omitempty"`
	Reviewers []string `yaml:"reviewers"`
}

func ParseSchemaOwners(content []byte) ([]SchemaOwnerRule, error) {
	var rules []SchemaOwnerRule
	if err := yaml.Unmarshal(content, &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func GetReviewersForChanges(rules []SchemaOwnerRule, rep *services.DiffReport) []string {
	if rep == nil {
		return nil
	}

	reviewerMap := make(map[string]bool)

	// We need a helper to check rules on interface{} items
	checkItem := func(item interface{}) {
		m, ok := item.(map[string]interface{})
		if !ok {
			return
		}

		path, _ := m["path"].(string)
		ruleID, _ := m["rule_id"].(string)

		for _, rule := range rules {
			// Check if rule matches by rule ID
			ruleMatched := false
			for _, r := range rule.Rules {
				if r == ruleID {
					ruleMatched = true
					break
				}
			}

			// Check if rule matches by path glob
			pathMatched := false
			if rule.Path != "" && path != "" {
				matched, err := filepath.Match(rule.Path, path)
				if err == nil && matched {
					pathMatched = true
				}
				if path == rule.Path {
					pathMatched = true
				}
				if !pathMatched {
					regexPattern := "^" + regexp.QuoteMeta(rule.Path) + "$"
					regexPattern = regexp.MustCompile(`\\\*`).ReplaceAllString(regexPattern, ".*")
					if matched, _ := regexp.MatchString(regexPattern, path); matched {
						pathMatched = true
					}
				}
			}

			if ruleMatched || pathMatched {
				for _, reviewer := range rule.Reviewers {
					reviewerMap[reviewer] = true
				}
			}
		}
	}

	for _, item := range rep.Breaking {
		checkItem(item)
	}
	for _, item := range rep.Warning {
		checkItem(item)
	}
	for _, item := range rep.Info {
		checkItem(item)
	}

	var reviewers []string
	for reviewer := range reviewerMap {
		reviewers = append(reviewers, reviewer)
	}

	return reviewers
}
