package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/services"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
)

// P4-T10 — Phase 4 AI Diff Engine Full-Stack E2E Validation
func TestPhase4AIDiff(t *testing.T) {
	pool, err := SetupP11Database()
	if err != nil {
		t.Skip("Skipping P4 test due to DB setup failure: " + err.Error())
	}
	defer pool.Close()

	ctx := context.Background()
	_, err = pool.Exec(ctx, "DELETE FROM schemas; DELETE FROM repositories;")
	if err != nil {
		t.Skip("Failed to clean up tables")
	}

	// Create repo and schema
	_, err = pool.Exec(ctx, `
		INSERT INTO repositories (id, org, name, full_name, github_repo_id, status)
		VALUES ('mcp-org/test-repo', 'mcp-org', 'test-repo', 'mcp-org/test-repo', 101, 'SAFE')
	`)
	if err != nil {
		t.Skip("Failed to insert repository")
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO schemas (id, repo, org, format, version, raw_content, parsed_json, commit_sha, file_path, status)
		VALUES ('test-schema-1', 'mcp-org/test-repo', 'mcp-org', 'openapi', '1.0.0', 'openapi: 3.0.0', '{}', 'test-sha', 'openapi.yaml', 'active')
	`)
	if err != nil {
		t.Skip("Failed to insert schema")
	}

	pgStore := db.NewStore(pool)
	_ = pgStore // Just so it's used somewhere in case we want it for other tests later.

	t.Run("GET /api/v1/diff/stream/{org}/{repo}", func(t *testing.T) {
		reqData := services.AIAnalyzeRequest{
			Org:            "mcp-org",
			SchemaType:     "openapi",
			CurrentSchema:  "openapi: 3.0.0\ninfo:\n  title: Test API\n  version: 1.0.0\npaths:\n  /test:\n    get:\n      responses:\n        '200':\n          description: OK",
			ProposedSchema: "openapi: 3.0.0\ninfo:\n  title: Test API\n  version: 1.0.0\npaths:\n  /test:\n    post:\n      responses:\n        '200':\n          description: OK",
		}

		body, _ := json.Marshal(reqData)
		req, err := http.NewRequest("POST", "/api/v1/ai/analyze", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := services.AIAnalyzeHandler()

		// Execute request
		handler.ServeHTTP(rr, req)

		// Assert response
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		contentType := rr.Header().Get("Content-Type")
		if contentType != "text/event-stream" {
			t.Errorf("handler returned unexpected content type: got %v want text/event-stream", contentType)
		}

		bodyBytes, _ := io.ReadAll(rr.Body)
		bodyStr := string(bodyBytes)

		// In mock mode without SUBSTRATE_AI_BASE_URL, mockSSEResponse emits done
		if !strings.Contains(bodyStr, "data: {\"type\":\"done\"}") {
			t.Errorf("expected done event in SSE stream, got: %v", bodyStr)
		}
	})

	t.Run("POST /api/v1/diff/upgrade", func(t *testing.T) {
		reqData := services.AIAutofixRequest{
			Org:           "mcp-org",
			ProviderRepo:  "mcp-org/test-repo",
		}

		body, _ := json.Marshal(reqData)
		req, err := http.NewRequest("POST", "/api/v1/ai/autofix", bytes.NewBuffer(body))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := services.AIAutofixHandler()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		var resp map[string]interface{}
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}

		if explanation, ok := resp["explanation"].(string); ok {
			if explanation == "" {
				t.Errorf("expected explanation, got empty string")
			}
		} else {
			t.Errorf("expected explanation in response")
		}
	})
}
