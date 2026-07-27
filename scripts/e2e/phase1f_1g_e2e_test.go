package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func p1f1gWaitForServices(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 5; i++ {
		resp, err := client.Get("http://localhost:8090/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("API server not reachable at http://localhost:8090")
}

func p1f1gSetupDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true")
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}

	// Clean tables
	tables := []string{
		"preview_sessions", "diff_reports", "repositories", "organizations",
	}
	for _, table := range tables {
		_, err := pool.Exec(ctx, "DELETE FROM "+table)
		if err != nil {
			t.Fatalf("Failed to clear table %s: %v", table, err)
		}
	}

	// Seed data
	_, err = pool.Exec(ctx, `
		INSERT INTO organizations (github_org_name, github_installation_id)
		VALUES ('mcp-org', 12345)
		ON CONFLICT (github_installation_id) DO NOTHING;
	`)
	if err != nil {
		t.Fatalf("Failed to seed org: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO repositories (org_id, github_repo_id, name, full_name)
		SELECT id, 1001, 'ml-models', 'mcp-org/ml-models' FROM organizations WHERE github_installation_id = 12345
		ON CONFLICT (github_repo_id) DO NOTHING;
	`)
	if err != nil {
		t.Fatalf("Failed to seed repo 1: %v", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO repositories (org_id, github_repo_id, name, full_name)
		SELECT id, 1002, 'salesforce-crm', 'mcp-org/salesforce-crm' FROM organizations WHERE github_installation_id = 12345
		ON CONFLICT (github_repo_id) DO NOTHING;
	`)
	if err != nil {
		t.Fatalf("Failed to seed repos: %v", err)
	}

	return pool
}

func readP1f1gTestdata(t *testing.T, filename string) string {
	b, err := os.ReadFile(filepath.Join("testdata", "phase1f_1g", filename))
	if err != nil {
		t.Fatalf("Failed to read testdata file %s: %v", filename, err)
	}
	return string(b)
}

func TestPhase1f_AIMLAdapter(t *testing.T) {
	p1f1gWaitForServices(t)
	pool := p1f1gSetupDB(t)
	defer pool.Close()

	e := httpexpect.Default(t, "http://localhost:8090") // API server port, hitting /api/v1/diff or direct engine /diff

	baseSchema := readP1f1gTestdata(t, "ai_base.yaml")
	headSafe := readP1f1gTestdata(t, "ai_head_safe.yaml")
	headBreak := readP1f1gTestdata(t, "ai_head_break.yaml")

	t.Run("Safe Change - AIML", func(t *testing.T) {
		req := map[string]interface{}{
			"base_schema": baseSchema,
			"head_schema_content": headSafe,
			"schema_type": "ai-model",
            "org": "mcp-org",
            "provider_repo": "ml-models",
            "pr_number": 1,
            "commit_sha": "safe-sha",
            "diff_report": map[string]interface{}{
				"summary": map[string]interface{}{
					"breaking_count": 0,
					"safe_count":     0,
				},
				"schema_type": "ai-model",
				"version":     "v1",
			},
		}

		res := e.POST("/api/v1/diff").WithJSON(req).WithHeader("Authorization", "Bearer local-dev-token").Expect().Status(http.StatusCreated).JSON().Object()
		res.Value("id").NotNull()
	})

	t.Run("Breaking Change - AIML", func(t *testing.T) {
		req := map[string]interface{}{
			"base_schema": baseSchema,
			"head_schema_content": headBreak,
			"schema_type": "ai-model",
            "org": "mcp-org",
            "provider_repo": "ml-models",
            "pr_number": 2,
            "commit_sha": "break-sha",
            "diff_report": map[string]interface{}{
				"summary": map[string]interface{}{
					"breaking_count": 0,
					"safe_count":     0,
				},
				"schema_type": "ai-model",
				"version":     "v1",
			},
		}

		res := e.POST("/api/v1/diff").WithJSON(req).WithHeader("Authorization", "Bearer local-dev-token").Expect().Status(http.StatusCreated).JSON().Object()
        res.Value("id").NotNull()
	})
}

func TestPhase1g_SalesforceAdapter(t *testing.T) {
	p1f1gWaitForServices(t)
	pool := p1f1gSetupDB(t)
	defer pool.Close()

	e := httpexpect.Default(t, "http://localhost:8090") // API server port

	baseSchema := readP1f1gTestdata(t, "sf_base.xml")
	headSafe := readP1f1gTestdata(t, "sf_head_safe.xml")
	headBreak := readP1f1gTestdata(t, "sf_head_break.xml")

	t.Run("Safe Change - Salesforce", func(t *testing.T) {
		req := map[string]interface{}{
			"base_schema": baseSchema,
			"head_schema_content": headSafe,
			"schema_type": "salesforce-object",
            "org": "mcp-org",
            "provider_repo": "salesforce-crm",
            "pr_number": 1,
            "commit_sha": "safe-sha",
            "diff_report": map[string]interface{}{
				"summary": map[string]interface{}{
					"breaking_count": 0,
					"safe_count":     0,
				},
				"schema_type": "salesforce-object",
				"version":     "v1",
			},
		}

		res := e.POST("/api/v1/diff").WithJSON(req).WithHeader("Authorization", "Bearer local-dev-token").Expect().Status(http.StatusCreated).JSON().Object()
		res.Value("id").NotNull()
	})

	t.Run("Breaking Change - Salesforce", func(t *testing.T) {
		req := map[string]interface{}{
			"base_schema": baseSchema,
			"head_schema_content": headBreak,
			"schema_type": "salesforce-object",
            "org": "mcp-org",
            "provider_repo": "salesforce-crm",
            "pr_number": 2,
            "commit_sha": "break-sha",
            "diff_report": map[string]interface{}{
				"summary": map[string]interface{}{
					"breaking_count": 0,
					"safe_count":     0,
				},
				"schema_type": "salesforce-object",
				"version":     "v1",
			},
		}

		res := e.POST("/api/v1/diff").WithJSON(req).WithHeader("Authorization", "Bearer local-dev-token").Expect().Status(http.StatusCreated).JSON().Object()
		res.Value("id").NotNull()
	})
}
