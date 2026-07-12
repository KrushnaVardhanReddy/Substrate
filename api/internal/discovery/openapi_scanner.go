package discovery

import (
	"context"
	"encoding/json"
	"regexp"

	"gopkg.in/yaml.v3"
)

// OpenAPIScanner implements Scanner for openapi-generator-config.yaml and openapitools.json
type OpenAPIScanner struct{}

func NewOpenAPIScanner() *OpenAPIScanner {
	return &OpenAPIScanner{}
}

func (s *OpenAPIScanner) Scan(ctx context.Context, repo string, files map[string][]byte, urlResolver func(string) string) ([]DependencyEdge, error) {
	var edges []DependencyEdge

	// Scan openapi-generator-config.yaml or .openapi-generator-config.yaml
	for filename, content := range files {
		if filename == "openapi-generator-config.yaml" || filename == ".openapi-generator-config.yaml" {
			edges = append(edges, s.scanYAMLConfig(repo, filename, content, urlResolver)...)
		} else if filename == "openapitools.json" {
			edges = append(edges, s.scanJSONTools(repo, filename, content, urlResolver)...)
		} else if filename == "package.json" {
		    edges = append(edges, s.scanPackageScripts(repo, filename, content, urlResolver)...)
		}
	}

	return edges, nil
}

func (s *OpenAPIScanner) scanYAMLConfig(repo, filename string, content []byte, urlResolver func(string) string) []DependencyEdge {
	var edges []DependencyEdge

	var config struct {
		InputSpec string `yaml:"inputSpec"`
	}

	if err := yaml.Unmarshal(content, &config); err != nil {
		return edges
	}

	if config.InputSpec != "" {
		if targetRepo := urlResolver(config.InputSpec); targetRepo != "" {
			edges = append(edges, DependencyEdge{
				SourceRepo: repo,
				TargetRepo: targetRepo,
				TargetURL:  config.InputSpec,
				Confidence: 50,
				Files:      []string{filename},
			})
		}
	}

	return edges
}

func (s *OpenAPIScanner) scanJSONTools(repo, filename string, content []byte, urlResolver func(string) string) []DependencyEdge {
	var edges []DependencyEdge

	var tools struct {
		GeneratorCLI struct {
			Generators map[string]struct {
				InputSpec string `json:"inputSpec"`
			} `json:"generators"`
		} `json:"generator-cli"`
	}

	if err := json.Unmarshal(content, &tools); err != nil {
		return edges
	}

	for _, gen := range tools.GeneratorCLI.Generators {
		if gen.InputSpec != "" {
			if targetRepo := urlResolver(gen.InputSpec); targetRepo != "" {
				edges = append(edges, DependencyEdge{
					SourceRepo: repo,
					TargetRepo: targetRepo,
					TargetURL:  gen.InputSpec,
					Confidence: 50,
					Files:      []string{filename},
				})
			}
		}
	}

	return edges
}

var openapiGeneratorRegex = regexp.MustCompile(`openapi-generator\s+generate\s+.*-i\s+([^\s]+)`)
var swaggerCodegenRegex = regexp.MustCompile(`swagger-codegen\s+generate\s+.*-i\s+([^\s]+)`)

func (s *OpenAPIScanner) scanPackageScripts(repo, filename string, content []byte, urlResolver func(string) string) []DependencyEdge {
    var edges []DependencyEdge

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}

	if err := json.Unmarshal(content, &pkg); err != nil {
		return edges // Skip on error
	}

	for _, script := range pkg.Scripts {
	    if matches := openapiGeneratorRegex.FindStringSubmatch(script); len(matches) > 1 {
	        if targetRepo := urlResolver(matches[1]); targetRepo != "" {
				edges = append(edges, DependencyEdge{
					SourceRepo: repo,
					TargetRepo: targetRepo,
					TargetURL:  matches[1],
					Confidence: 50,
					Files:      []string{filename},
				})
			}
	    }

	    if matches := swaggerCodegenRegex.FindStringSubmatch(script); len(matches) > 1 {
	        if targetRepo := urlResolver(matches[1]); targetRepo != "" {
				edges = append(edges, DependencyEdge{
					SourceRepo: repo,
					TargetRepo: targetRepo,
					TargetURL:  matches[1],
					Confidence: 50,
					Files:      []string{filename},
				})
			}
	    }
	}

	return edges
}
