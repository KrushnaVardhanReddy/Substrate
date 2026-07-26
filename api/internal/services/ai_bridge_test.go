package services

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/spf13/viper"
)

func TestExecuteTool_GetBreakingChangeHistory(t *testing.T) {
	// Create a mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/history/myorg/myrepo" {
			if r.Header.Get("Authorization") != "Bearer test-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			limit := r.URL.Query().Get("limit")
			if limit != "5" && limit != "10" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"id":"123"}]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	// Set up viper configuration
	viper.Set("REGISTRY_API_URL", ts.URL)
	viper.Set("REGISTRY_API_TOKEN", "test-token")
	defer viper.Reset()

	tests := []struct {
		name      string
		args      string
		req       AIAnalyzeRequest
		want      string
		expectErr bool
	}{
		{
			name: "Valid request with limit",
			args: `{"repo": "myrepo", "limit": 5}`,
			req: AIAnalyzeRequest{
				Org: "myorg",
			},
			want:      `[{"id":"123"}]`,
			expectErr: false,
		},
		{
			name: "Valid request without limit",
			args: `{"repo": "myrepo"}`,
			req: AIAnalyzeRequest{
				Org: "myorg",
			},
			want:      `[{"id":"123"}]`,
			expectErr: false,
		},
		{
			name: "Missing repo",
			args: `{"limit": 5}`,
			req: AIAnalyzeRequest{
				Org: "myorg",
			},
			want:      "",
			expectErr: true,
		},
		{
			name: "Invalid JSON args",
			args: `{"repo": "myrepo"`,
			req: AIAnalyzeRequest{
				Org: "myorg",
			},
			want:      "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := executeTool("get_breaking_change_history", tt.args, tt.req)
			if (err != nil) != tt.expectErr {
				t.Errorf("executeTool() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if got != tt.want {
				t.Errorf("executeTool() got = %v, want %v", got, tt.want)
			}
		})
	}
}
