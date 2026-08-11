// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package discovery

import (
	"bufio"
	"regexp"
	"strings"
)

var (
	// High-signal patterns from the spec
	// *_API_URL, *_SERVICE_URL, *_ENDPOINT, *_BASE_URL, *_HOST (when value is a URL), *_API_BASE, *_GATEWAY_URL
	highSignalRegexStr = `(?i)_API_URL$|_SERVICE_URL$|_ENDPOINT$|_BASE_URL$|_HOST$|_API_BASE$|_GATEWAY_URL$`
	highSignalRegex    = regexp.MustCompile(highSignalRegexStr)

	// Ensure we exclude secrets
	secretRegex = regexp.MustCompile(`(?i)SECRET|KEY|TOKEN|PASSWORD`)

	// URL extractor for k8s yaml/docker-compose/env format
	// For simplicity, we just want to match http:// or https:// patterns or typical k8s service names
	urlRegex = regexp.MustCompile(`https?://[^\s"'<>]+|[a-zA-Z0-9-]+\.[a-zA-Z0-9-]+\.svc\.cluster\.local`)
)

type DiscoveredDependency struct {
	VarName         string
	VarValue        string
	ConfidenceScore int
}

type EnvScanner struct {
	regex *regexp.Regexp
}

func NewEnvScanner(patterns []string) *EnvScanner {
	if len(patterns) == 0 {
		return &EnvScanner{regex: highSignalRegex}
	}

	patternStr := strings.Join(patterns, "|")

	mergedRegex, err := regexp.Compile(highSignalRegexStr + "|" + patternStr)
	if err != nil {
		return &EnvScanner{regex: highSignalRegex}
	}

	return &EnvScanner{regex: mergedRegex}
}

func (s *EnvScanner) parseEnvLine(line string) *DiscoveredDependency {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return nil
	}
	name := strings.TrimSpace(parts[0])
	value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

	if secretRegex.MatchString(name) {
		return nil
	}

	if s.regex.MatchString(name) {
		score := 15 // Base score for name pattern match

		// If it has a parsable URL, we might want to check the registry.
		// For Tier 1, finding a value could mean +15 if it resolves.
		// To match the spec strictly: "15 pts for name pattern match".
		// We'll leave value resolution to the URL->Repo registry lookup later.

		return &DiscoveredDependency{
			VarName:         name,
			VarValue:        value,
			ConfidenceScore: score,
		}
	}
	return nil
}

func (s *EnvScanner) ScanEnvFile(content string) []DiscoveredDependency {
	var deps []DiscoveredDependency
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}

		if dep := s.parseEnvLine(line); dep != nil {
			deps = append(deps, *dep)
		}
	}
	return deps
}

// Global functions for backward compatibility where patterns are nil
func ScanEnvFile(content string) []DiscoveredDependency {
	return NewEnvScanner(nil).ScanEnvFile(content)
}

func (s *EnvScanner) ScanDockerCompose(content string) []DiscoveredDependency {
	// A rudimentary scanner. A real one might use go-yaml, but regex is acceptable for string extraction
	// In docker compose, env variables are under "environment:" list or dict.
	var deps []DiscoveredDependency
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Strip comments
		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		// list format: - VAR=value
		if strings.HasPrefix(line, "- ") {
			if dep := s.parseEnvLine(line[2:]); dep != nil {
				deps = append(deps, *dep)
			}
		} else if strings.Contains(line, ": ") {
			// dict format: VAR: value
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[0])
				value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

				if secretRegex.MatchString(name) {
					continue
				}

				if s.regex.MatchString(name) {
					deps = append(deps, DiscoveredDependency{
						VarName:         name,
						VarValue:        value,
						ConfidenceScore: 15, // Using base 15 here
					})
				}
			}
		}
	}
	return deps
}

func ScanDockerCompose(content string) []DiscoveredDependency {
	return NewEnvScanner(nil).ScanDockerCompose(content)
}

func (s *EnvScanner) ScanKubernetesManifest(content string) []DiscoveredDependency {
	var deps []DiscoveredDependency
	scanner := bufio.NewScanner(strings.NewReader(content))

	var currentName string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		}

		// In k8s, we often see:
		// - name: VAR_NAME
		//   value: "url"

		if strings.HasPrefix(line, "name: ") {
			currentName = strings.Trim(strings.TrimSpace(line[6:]), "\"'")
		} else if strings.HasPrefix(line, "value: ") && currentName != "" {
			value := strings.Trim(strings.TrimSpace(line[7:]), "\"'")

			if !secretRegex.MatchString(currentName) && s.regex.MatchString(currentName) {
				deps = append(deps, DiscoveredDependency{
					VarName:         currentName,
					VarValue:        value,
					ConfidenceScore: 15,
				})
			}
			currentName = "" // reset
		} else if strings.HasPrefix(line, "- name: ") {
			currentName = strings.Trim(strings.TrimSpace(line[8:]), "\"'")
		}
	}

	return deps
}

func ScanKubernetesManifest(content string) []DiscoveredDependency {
	return NewEnvScanner(nil).ScanKubernetesManifest(content)
}
