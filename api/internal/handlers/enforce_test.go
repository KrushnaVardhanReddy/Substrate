package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestEnforceGlobalHandler(t *testing.T) {
	mockGitHub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/orgs/myorg/repos") {
			if r.URL.Query().Get("page") == "1" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[{"name": "repo1", "default_branch": "main"}, {"name": "repo2", "default_branch": "master"}]`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
			return
		}

		if strings.HasPrefix(r.URL.Path, "/repos/myorg/") && strings.HasSuffix(r.URL.Path, "/protection") {
			if r.Method != "PUT" {
				t.Errorf("expected PUT, got %s", r.Method)
			}
			auth := r.Header.Get("Authorization")
			if auth != "Bearer dummy-token" {
				t.Errorf("expected auth 'Bearer dummy-token', got %s", auth)
			}

			var payload map[string]interface{}
			json.NewDecoder(r.Body).Decode(&payload)

			checks, ok := payload["required_status_checks"].(map[string]interface{})
			if !ok {
				t.Errorf("missing required_status_checks")
			} else {
				contexts, ok := checks["contexts"].([]interface{})
				if !ok || len(contexts) == 0 || contexts[0] != "substrate" {
					t.Errorf("missing substrate context")
				}
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/repos/myorg/") && strings.HasSuffix(r.URL.Path, "/contexts") {
			if r.Method != "DELETE" {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			auth := r.Header.Get("Authorization")
			if auth != "Bearer dummy-token" {
				t.Errorf("expected auth 'Bearer dummy-token', got %s", auth)
			}

			var payload []string
			json.NewDecoder(r.Body).Decode(&payload)
			if len(payload) != 1 || payload[0] != "substrate" {
				t.Errorf("expected [\"substrate\"], got %v", payload)
			}

			w.WriteHeader(http.StatusOK)
			return
		}

		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockGitHub.Close()

	// Override global github API url for tests
	originalGitHubUrl := githubApiBaseUrl
	githubApiBaseUrl = mockGitHub.URL
	defer func() { githubApiBaseUrl = originalGitHubUrl }()

	r := chi.NewRouter()
	r.Post("/api/v1/org/{org}/enforce", EnforceGlobalHandler())

	t.Run("Enforce True", func(t *testing.T) {
		reqBody := `{"enforce": true}`
		req := httptest.NewRequest("POST", "/api/v1/org/myorg/enforce", strings.NewReader(reqBody))
		req.Header.Set("Authorization", "Bearer dummy-token")
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", rr.Code)
		}
	})

	t.Run("Enforce False", func(t *testing.T) {
		reqBody := `{"enforce": false}`
		req := httptest.NewRequest("POST", "/api/v1/org/myorg/enforce", strings.NewReader(reqBody))
		req.Header.Set("Authorization", "Bearer dummy-token")
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", rr.Code)
		}
	})
}
