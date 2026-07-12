package mockserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestStartMockServer(t *testing.T) {
	tests := []struct {
		name       string
		mockStatus int
		mockResp   string
		expectCode int
		reqPath    string
		expectResp string
	}{
		{
			name:       "success with schema",
			mockStatus: http.StatusOK,
			mockResp:   `{"schema": "openapi: 3.0.0\npaths:\n  /test:\n    get:\n      responses:\n        '200':\n          description: ok\n          content:\n            application/json:\n              schema:\n                type: object", "schema_type": "openapi"}`,
			expectCode: http.StatusOK,
			reqPath:    "/test",
			expectResp: "{}",
		},
		{
			name:       "success with schema array",
			mockStatus: http.StatusOK,
			mockResp:   `{"schema": "openapi: 3.0.0\npaths:\n  /test:\n    get:\n      responses:\n        '200':\n          description: ok\n          content:\n            application/json:\n              schema:\n                type: array", "schema_type": "openapi"}`,
			expectCode: http.StatusOK,
			reqPath:    "/test",
			expectResp: "[]",
		},
		{
			name:       "fallback if not found",
			mockStatus: http.StatusNotFound,
			mockResp:   `{"error": "not found"}`,
			expectCode: http.StatusOK, // Even on error the server still starts and replies with fallback
			reqPath:    "/",
			expectResp: `{"message": "Mock server response for 2026-07-01"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mockRegistry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Assert timestamp query param
				if r.URL.Query().Get("timestamp") != "2026-07-01" {
					t.Errorf("expected timestamp param 2026-07-01, got %s", r.URL.Query().Get("timestamp"))
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				w.Write([]byte(tt.mockResp))
			}))
			defer mockRegistry.Close()

			os.Setenv("REGISTRY_API_URL", mockRegistry.URL)
			os.Setenv("SUBSTRATE_OWNER", "testorg")
			os.Setenv("SUBSTRATE_REPO", "testrepo")

			// Choose unique port per test
			port := 8090
			if tt.name == "success with schema" {
				port = 8092
			} else if tt.name == "success with schema array" {
				port = 8093
			} else {
				port = 8094
			}

			go func() {
				err := StartMockServer(ctx, "2026-07-01", port)
				if err != nil && err != http.ErrServerClosed {
					t.Errorf("expected no error or ErrServerClosed, got %v", err)
				}
			}()

			time.Sleep(100 * time.Millisecond)

			// nolint: noctx
			resp, err := http.Get(func() string {
				if port == 8092 {
					return "http://localhost:8092" + tt.reqPath
				} else if port == 8093 {
					return "http://localhost:8093" + tt.reqPath
				}
				return "http://localhost:8094" + tt.reqPath
			}())
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectCode {
				t.Errorf("expected status %d, got %d", tt.expectCode, resp.StatusCode)
			}
		})
	}
}
