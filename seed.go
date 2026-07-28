package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func main() {
	payload := map[string]interface{}{
		"installation_id": 9999,
		"org":             "mcp-org",
		"repo":            "mcp-org/frontend",
		"github_repo_id":  102,
		"commit_sha":      "sha-frontend-123",
		"files": []map[string]interface{}{
			{
				"path":    ".env",
				"content": "API_URL=http://backend:8080",
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "http://localhost:8090/api/v1/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer local-dev-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	fmt.Println("Frontend Webhook Status:", resp.StatusCode)

	// Now for the backend
	payload2 := map[string]interface{}{
		"installation_id": 9999,
		"org":             "mcp-org",
		"repo":            "mcp-org/backend",
		"github_repo_id":  101,
		"commit_sha":      "sha-backend-123",
		"files": []map[string]interface{}{
			{
				"path":    "openapi.yaml",
				"content": "openapi: 3.0.0\ninfo:\n  title: Backend\n  version: 1.0.0\npaths:\n  /users:\n    get:\n      responses:\n        '200':\n          description: OK",
			},
			{
				"path":    "substrate.yaml",
				"content": "project: backend\nteam: core\ncomponent_type: backend\n",
			},
		},
	}

	body2, _ := json.Marshal(payload2)
	req2, _ := http.NewRequest("POST", "http://localhost:8090/api/v1/webhook", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer local-dev-token")

	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp2.Body.Close()
	fmt.Println("Backend Webhook Status:", resp2.StatusCode)
}
