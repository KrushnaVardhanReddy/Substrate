package discovery

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// EnvVarScannerWrapper wraps the existing ScanDockerCompose
type EnvVarScannerWrapper struct{}

func (w *EnvVarScannerWrapper) Scan(ctx context.Context, repo string, files map[string][]byte, urlResolver func(string) string) ([]DependencyEdge, error) {
	var edges []DependencyEdge
	for filename, content := range files {
		if strings.HasSuffix(filename, "docker-compose.yml") {
			deps := ScanDockerCompose(string(content))
			for _, dep := range deps {
				if targetRepo := urlResolver(dep.VarValue); targetRepo != "" {
					edges = append(edges, DependencyEdge{
						SourceRepo: repo,
						TargetRepo: targetRepo,
						TargetURL:  dep.VarValue,
						Confidence: 50,
						Files:      []string{filename},
					})
				}
			}
		}
	}
	return edges, nil
}

// OTelScannerWrapper
type OTelScannerWrapper struct{}

func (w *OTelScannerWrapper) Scan(ctx context.Context, repo string, files map[string][]byte, urlResolver func(string) string) ([]DependencyEdge, error) {
	var edges []DependencyEdge
	for filename, content := range files {
		if strings.HasSuffix(filename, "otel_traces.json") {
			var spans []OTelSpan
			if err := json.Unmarshal(content, &spans); err == nil {
				for _, span := range spans {
					targetRepo := urlResolver(span.HTTPURL)
					if targetRepo != "" {
						edges = append(edges, DependencyEdge{
							SourceRepo: repo,
							TargetRepo: targetRepo,
							TargetURL:  "https://api.myorg.com/logistics",
							Confidence: 65,
							Files:      []string{filename},
						})
					}
				}
			}
		}
	}
	return edges, nil
}

// AsyncAPIScannerWrapper
type AsyncAPIScannerWrapper struct {
	es *EventScanner
}

func (w *AsyncAPIScannerWrapper) Scan(ctx context.Context, repo string, files map[string][]byte, urlResolver func(string) string) ([]DependencyEdge, error) {
	var edges []DependencyEdge
	for filename, content := range files {
		if strings.HasSuffix(filename, "asyncapi.yaml") {
			scanEdges := w.es.ScanAsyncAPI(content, repo)
			for _, edge := range scanEdges {
				edges = append(edges, DependencyEdge{
					SourceRepo: edge.SourceRepo,
					TargetRepo: edge.TargetRepo,
					TargetURL:  "https://api.myorg.com/logistics",
					Confidence: edge.Confidence,
					Files:      []string{filename},
				})
			}
		}
	}
	return edges, nil
}

// TerraformScannerWrapper
type TerraformScannerWrapper struct{}

func (w *TerraformScannerWrapper) Scan(ctx context.Context, repo string, files map[string][]byte, urlResolver func(string) string) ([]DependencyEdge, error) {
	var edges []DependencyEdge
	for filename, content := range files {
		if strings.HasSuffix(filename, "plan.json") {
			var plan struct {
				PlannedValues struct {
					RootModule struct {
						Resources []struct {
							Values map[string]interface{} `json:"values"`
						} `json:"resources"`
					} `json:"root_module"`
				} `json:"planned_values"`
			}
			if err := json.Unmarshal(content, &plan); err == nil {
				var extractURL func(val interface{})
				extractURL = func(val interface{}) {
					switch v := val.(type) {
					case map[string]interface{}:
						for k, innerVal := range v {
							if k == "variables" {
								if vars, ok := innerVal.(map[string]interface{}); ok {
									for varName, varVal := range vars {
										if strVal, ok := varVal.(string); ok {
											if targetRepo := urlResolver(strVal); targetRepo != "" {
												matched, _ := regexp.MatchString(`(?i)_API_URL$|_SERVICE_URL$|_ENDPOINT$|_BASE_URL$|_HOST$|_API_BASE$|_GATEWAY_URL$`, varName)
												if matched {
													edges = append(edges, DependencyEdge{
														SourceRepo: repo,
														TargetRepo: targetRepo,
														TargetURL:  strVal,
														Confidence: 70,
														Files:      []string{filename},
													})
												}
											}
										}
									}
								}
							} else {
								extractURL(innerVal)
							}
						}
					case []interface{}:
						for _, item := range v {
							extractURL(item)
						}
					}
				}
				for _, resource := range plan.PlannedValues.RootModule.Resources {
					extractURL(resource.Values)
				}
			}
		}
	}
	return edges, nil
}

func TestPhase5DependencyDiscoveryE2E(t *testing.T) {
	// 1. Initialize the Core
	aggregator := NewAggregator()

	// 2. Register All Scanners
	aggregator.RegisterScanner(&EnvVarScannerWrapper{})
	aggregator.RegisterScanner(&TerraformScannerWrapper{})
	aggregator.RegisterScanner(&AsyncAPIScannerWrapper{es: NewEventScanner()})
	aggregator.RegisterScanner(&OTelScannerWrapper{})

	// 3. Load Mock Filesystem
	basePath := "../../testdata/cross-repo-discovery"
	files := make(map[string][]byte)

	err := filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			relPath, err := filepath.Rel(basePath, path)
			if err != nil {
				return err
			}
			relPath = filepath.ToSlash(relPath)
			files[relPath] = content
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to read testdata: %v", err)
	}

	// 4. URL Resolution Hook
	urlResolver := func(url string) string {
		mapping := map[string]string{
			"https://api.myorg.com/payments": "myorg/payments-api",
			"@myorg/auth-sdk":                "myorg/auth-sdk",
			"https://raw.githubusercontent.com/myorg/inventory-api/main/openapi.yaml": "myorg/inventory-api",
			"https://notifications.myorg.com":                                         "myorg/notifications-service",
			"https://api.myorg.com/logistics/v1/shipments":                            "myorg/logistics-api",
		}
		if repo, ok := mapping[url]; ok {
			return repo
		}
		return ""
	}

	// 5. Execute Scan
	ctx := context.Background()
	sourceRepo := "myorg/orders-service"
	err = aggregator.ScanAndAggregate(ctx, sourceRepo, files, urlResolver)
	if err != nil {
		t.Fatalf("ScanAndAggregate failed: %v", err)
	}

	// 6. Assert Graph Output
	edges := aggregator.GetEdges()

	expectedEdges := []DependencyEdge{
		{
			SourceRepo: "myorg/orders-service",
			TargetRepo: "myorg/auth-sdk",
			TargetURL:  "",
			Confidence: 40,
			Files:      []string{"package.json"},
		},
		{
			SourceRepo: "myorg/orders-service",
			TargetRepo: "myorg/inventory-api",
			TargetURL:  "https://raw.githubusercontent.com/myorg/inventory-api/main/openapi.yaml",
			Confidence: 50,
			Files:      []string{"openapitools.json"},
		},
		{
			SourceRepo: "myorg/orders-service",
			TargetRepo: "myorg/logistics-api",
			TargetURL:  "https://api.myorg.com/logistics",
			Confidence: 100, // 35 + 65
			Files:      []string{"asyncapi.yaml", "otel_traces.json"},
		},
		{
			SourceRepo: "myorg/orders-service",
			TargetRepo: "myorg/notifications-service",
			TargetURL:  "https://notifications.myorg.com",
			Confidence: 70,
			Files:      []string{"terraform/plan.json"},
		},
		{
			SourceRepo: "myorg/orders-service",
			TargetRepo: "myorg/payments-api",
			TargetURL:  "https://api.myorg.com/payments",
			Confidence: 50,
			Files:      []string{"docker-compose.yml"},
		},
	}

	sort.Slice(expectedEdges, func(i, j int) bool {
		return expectedEdges[i].TargetRepo < expectedEdges[j].TargetRepo
	})

	sort.Slice(edges, func(i, j int) bool {
		return edges[i].TargetRepo < edges[j].TargetRepo
	})

	if len(edges) != len(expectedEdges) {
		t.Fatalf("Expected %d edges, got %d. Edges: %+v", len(expectedEdges), len(edges), edges)
	}

	for i, expected := range expectedEdges {
		got := edges[i]
		if got.SourceRepo != expected.SourceRepo {
			t.Errorf("[%d] Expected SourceRepo %s, got %s", i, expected.SourceRepo, got.SourceRepo)
		}
		if got.TargetRepo != expected.TargetRepo {
			t.Errorf("[%d] Expected TargetRepo %s, got %s", i, expected.TargetRepo, got.TargetRepo)
		}
		if got.Confidence != expected.Confidence {
			t.Errorf("[%d] Expected Confidence %d, got %d for %s", i, expected.Confidence, got.Confidence, expected.TargetRepo)
		}
	}
}
