// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildCheckDeployBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	binPath := filepath.Join(dir, "substrate")
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}
	return binPath
}

func TestCheckDeployCommand(t *testing.T) {
	binPath := buildCheckDeployBinary(t)

	tests := []struct {
		name           string
		repoArg        string
		commitArg      string
		mockStatus     int
		mockResponse   map[string]interface{}
		expectedExit   int
		expectedStdout string
		expectedStderr string
	}{
		{
			name:       "safe to deploy",
			repoArg:    "owner/repo",
			commitArg:  "a1b2c3d4",
			mockStatus: http.StatusOK,
			mockResponse: map[string]interface{}{
				"safe":    true,
				"message": "Safe to deploy",
			},
			expectedExit:   0,
			expectedStdout: "✅ SAFE TO DEPLOY\n\nProvider: owner/repo (a1b2c3d4)\n\nCompatibility Check:\n- consumers -> COMPATIBLE\n",
		},
		{
			name:       "deployment blocked",
			repoArg:    "owner/repo",
			commitArg:  "a1b2c3d4",
			mockStatus: http.StatusConflict,
			mockResponse: map[string]interface{}{
				"safe":       false,
				"message":    "Deployment blocked: required consumers have not updated to handle breaking change",
				"blocked_by": []string{"owner/consumer1", "owner/consumer2"},
			},
			expectedExit:   1,
			expectedStdout: "❌ DEPLOYMENT BLOCKED\n\nProvider: owner/repo (a1b2c3d4)\n\nThe following consumers are INCOMPATIBLE with this deployment:\n- owner/consumer1\n- owner/consumer2\n",
		},
		{
			name:           "missing repo",
			repoArg:        "",
			commitArg:      "a1b2c3d4",
			expectedExit:   3,
			expectedStderr: "Error: --repo flag is required\n",
		},
		{
			name:       "server error",
			repoArg:    "owner/repo",
			commitArg:  "a1b2c3d4",
			mockStatus: http.StatusInternalServerError,
			mockResponse: map[string]interface{}{
				"message": "Internal Server Error",
			},
			expectedExit:   3,
			expectedStderr: "Error from server: HTTP 500 Internal Server Error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/registry/can-deploy" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if r.Header.Get("Authorization") != "Bearer test-token" {
					t.Errorf("missing or incorrect authorization header: %s", r.Header.Get("Authorization"))
				}
				if r.URL.Query().Get("repo") != tt.repoArg {
					t.Errorf("unexpected repo query: %s", r.URL.Query().Get("repo"))
				}
				if r.URL.Query().Get("commit") != tt.commitArg {
					t.Errorf("unexpected commit query: %s", r.URL.Query().Get("commit"))
				}
				w.WriteHeader(tt.mockStatus)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			args := []string{"check-deploy"}
			if tt.repoArg != "" {
				args = append(args, "--repo", tt.repoArg)
			}
			if tt.commitArg != "" {
				args = append(args, "--commit", tt.commitArg)
			}

			cmd := exec.Command(binPath, args...)
			cmd.Env = append(os.Environ(),
				"SUBSTRATE_API_URL="+server.URL,
				"REGISTRY_API_TOKEN=test-token",
				"SUBSTRATE_DISABLE_CACHE_SYNC=1",
			)

			var stdout, stderr strings.Builder
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				} else {
					t.Fatalf("unexpected error running command: %v", err)
				}
			}

			if exitCode != tt.expectedExit {
				t.Errorf("expected exit code %d, got %d. stderr: %s", tt.expectedExit, exitCode, stderr.String())
			}

			if tt.expectedStdout != "" && stdout.String() != tt.expectedStdout {
				t.Errorf("expected stdout %q, got %q", tt.expectedStdout, stdout.String())
			}
			if tt.expectedStderr != "" && stderr.String() != tt.expectedStderr {
				t.Errorf("expected stderr %q, got %q", tt.expectedStderr, stderr.String())
			}
		})
	}
}
