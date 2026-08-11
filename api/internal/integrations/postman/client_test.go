// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package postman

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_SyncSchema(t *testing.T) {
	tests := []struct {
		name         string
		workspaceID  string
		schema       string
		mockStatus   int
		expectDelete bool
		wantErr      bool
	}{
		{
			name:         "success with existing collection in workspace",
			workspaceID:  "ws-1",
			schema:       `{"openapi":"3.0.0","info":{"title":"Test API"}}`,
			mockStatus:   http.StatusOK,
			expectDelete: true,
			wantErr:      false,
		},
		{
			name:         "success without workspace",
			workspaceID:  "",
			schema:       `{"openapi":"3.0.0"}`,
			mockStatus:   http.StatusOK,
			expectDelete: false,
			wantErr:      false,
		},
		{
			name:         "api error on workspace get",
			workspaceID:  "ws-error",
			schema:       `{}`,
			mockStatus:   http.StatusInternalServerError,
			expectDelete: false,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleted := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/workspaces/"+tt.workspaceID {
					if tt.mockStatus != http.StatusOK {
						w.WriteHeader(tt.mockStatus)
						return
					}
					// Return a mock workspace containing the target collection
					resp := getWorkspaceResponse{}
					resp.Workspace.Collections = []collectionInfo{
						{ID: "coll-1", Name: "[Substrate] Test API"},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(resp)
					return
				}

				if r.Method == http.MethodDelete && r.URL.Path == "/collections/coll-1" {
					deleted = true
					w.WriteHeader(http.StatusOK)
					return
				}

				if r.Method == http.MethodPost && r.URL.Path == "/import/openapi" {
					if tt.workspaceID != "" && r.URL.Query().Get("workspace") != tt.workspaceID {
						t.Errorf("Expected workspace %s, got %s", tt.workspaceID, r.URL.Query().Get("workspace"))
					}
					w.WriteHeader(http.StatusOK)
					return
				}

				t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer server.Close()

			client := &Client{
				HTTPClient:  server.Client(),
				BaseURL:     server.URL,
				APIKey:      "test-key",
				WorkspaceID: tt.workspaceID,
			}

			err := client.SyncSchema(context.Background(), tt.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.SyncSchema() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.expectDelete && !deleted {
				t.Error("Expected an existing collection to be deleted, but it wasn't")
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Error("Expected NewClient to return a non-nil client")
	}
}

func TestExtractTitleFromSchema(t *testing.T) {
	tests := []struct {
		schema string
		want   string
	}{
		{`{"info":{"title":"Test JSON"}}`, "Test JSON"},
		{`{"info":{"title":""}}`, ""},
		{`{"other":"field"}`, ""},
		{"openapi: 3.0.0\ninfo:\n  title: Test YAML\n", "Test YAML"},
		{"openapi: 3.0.0\ninfo:\n  title: \"Test YAML Quotes\"\n", "Test YAML Quotes"},
		{"openapi: 3.0.0\n", ""},
	}
	for _, tt := range tests {
		got := extractTitleFromSchema(tt.schema)
		if got != tt.want {
			t.Errorf("extractTitleFromSchema(%q) = %q, want %q", tt.schema, got, tt.want)
		}
	}
}

func TestOverrideSchemaTitle(t *testing.T) {
	jsonSchema := `{"info":{"title":"Old JSON"}}`
	wantJSON := `{"info":{"title":"New JSON"}}`
	if got := overrideSchemaTitle(jsonSchema, "New JSON"); got != wantJSON {
		t.Errorf("overrideSchemaTitle JSON = %q, want %q", got, wantJSON)
	}

	yamlSchema := "openapi: 3.0.0\ninfo:\n  title: Old YAML\n  version: 1.0"
	wantYAML := "openapi: 3.0.0\ninfo:\n  title: New YAML\n  version: 1.0"
	if got := overrideSchemaTitle(yamlSchema, "New YAML"); got != wantYAML {
		t.Errorf("overrideSchemaTitle YAML = %q, want %q", got, wantYAML)
	}
}
