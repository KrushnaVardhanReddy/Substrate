package diff_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/config"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/diff"
)

func TestCompareAvro(t *testing.T) {
	tempDir := t.TempDir()
	headFile := filepath.Join(tempDir, "head.avsc")
	sampleAvro := `{"type": "record", "name": "Payment", "fields": [{"name": "id", "type": "string"}]}`
	if err := os.WriteFile(headFile, []byte(sampleAvro), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		setupServer  func() *httptest.Server
		cfgFactory   func(url string) *config.SubstrateConfig
		wantBreaking int
		wantWarning  int
		wantRule     string
		descContains string
	}{
		{
			name:        "no registry URL configured",
			setupServer: func() *httptest.Server { return nil },
			cfgFactory: func(url string) *config.SubstrateConfig {
				return &config.SubstrateConfig{}
			},
			wantBreaking: 0,
			wantWarning:  1,
			wantRule:     "AVRO_NO_REGISTRY_CONFIGURED",
		},
		{
			name: "registry returns compatible",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"is_compatible": true}`))
				}))
			},
			cfgFactory: func(url string) *config.SubstrateConfig {
				return &config.SubstrateConfig{
					Service: "payments",
					Avro: &config.AvroConfig{
						SchemaRegistryURL: url,
					},
				}
			},
			wantBreaking: 0,
			wantWarning:  0,
		},
		{
			name: "registry returns incompatible",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"is_compatible": false, "messages": ["removed field 'id'"]}`))
				}))
			},
			cfgFactory: func(url string) *config.SubstrateConfig {
				return &config.SubstrateConfig{
					Service: "payments",
					Avro: &config.AvroConfig{
						SchemaRegistryURL: url,
					},
				}
			},
			wantBreaking: 1,
			wantWarning:  0,
			wantRule:     "AVRO_INCOMPATIBLE",
			descContains: "removed field",
		},
		{
			name: "registry returns HTTP 500",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error": "internal server error"}`))
				}))
			},
			cfgFactory: func(url string) *config.SubstrateConfig {
				return &config.SubstrateConfig{
					Service: "payments",
					Avro: &config.AvroConfig{
						SchemaRegistryURL: url,
					},
				}
			},
			wantBreaking: 0,
			wantWarning:  1,
			wantRule:     "AVRO_REGISTRY_UNAVAILABLE",
		},
		{
			name: "registry connection refused",
			setupServer: func() *httptest.Server {
				return nil
			},
			cfgFactory: func(url string) *config.SubstrateConfig {
				return &config.SubstrateConfig{
					Service: "payments",
					Avro: &config.AvroConfig{
						SchemaRegistryURL: "http://127.0.0.1:1", // guaranteed to fail
					},
				}
			},
			wantBreaking: 0,
			wantWarning:  1,
			wantRule:     "AVRO_REGISTRY_UNAVAILABLE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serverURL string
			ts := tt.setupServer()
			if ts != nil {
				serverURL = ts.URL
				defer ts.Close()
			}

			cfg := tt.cfgFactory(serverURL)
			rep, err := diff.CompareAvro("ignored_base.avsc", headFile, cfg)
			if err != nil {
				t.Fatalf("CompareAvro returned error: %v", err)
			}

			if rep.SchemaType != "avro" {
				t.Errorf("Expected SchemaType 'avro', got '%s'", rep.SchemaType)
			}

			if len(rep.BreakingChanges) != tt.wantBreaking {
				t.Errorf("Expected %d breaking changes, got %d", tt.wantBreaking, len(rep.BreakingChanges))
			}

			if len(rep.Warnings) != tt.wantWarning {
				t.Errorf("Expected %d warnings, got %d", tt.wantWarning, len(rep.Warnings))
			}

			if tt.wantBreaking > 0 {
				if rep.BreakingChanges[0].RuleID != tt.wantRule {
					t.Errorf("Expected rule %s, got %s", tt.wantRule, rep.BreakingChanges[0].RuleID)
				}
				if tt.descContains != "" && !strings.Contains(rep.BreakingChanges[0].Description, tt.descContains) {
					t.Errorf("Expected description to contain %q, got %q", tt.descContains, rep.BreakingChanges[0].Description)
				}
			}

			if tt.wantWarning > 0 {
				if rep.Warnings[0].RuleID != tt.wantRule {
					t.Errorf("Expected rule %s, got %s", tt.wantRule, rep.Warnings[0].RuleID)
				}
			}
		})
	}
}
