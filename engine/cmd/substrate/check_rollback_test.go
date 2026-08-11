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
	"testing"
)

func TestCheckRollbackCmd_Safe(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/registry/can-rollback" {
			t.Errorf("expected /api/v1/registry/can-rollback, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("org") != "testorg" || r.URL.Query().Get("repo") != "testrepo" || r.URL.Query().Get("target_sha") != "abcdef" {
			t.Errorf("missing or incorrect query params")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"can_rollback": true,
			"message":      "Safe to rollback",
		})
	}))
	defer ts.Close()

	os.Setenv("REGISTRY_API_TOKEN", "testtoken")
	os.Setenv("SUBSTRATE_API_URL", ts.URL)
	os.Setenv("SUBSTRATE_DISABLE_CACHE_SYNC", "1")
	defer os.Unsetenv("REGISTRY_API_TOKEN")
	defer os.Unsetenv("SUBSTRATE_API_URL")
	defer os.Unsetenv("SUBSTRATE_DISABLE_CACHE_SYNC")

	cmd := exec.Command("go", "build", "-o", "substrate_bin")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %s", out)
	}
	defer os.Remove("substrate_bin")

	runCmd := exec.Command("./substrate_bin", "check-rollback", "--repo=testorg/testrepo", "--target-sha=abcdef")
	runOut, runErr := runCmd.CombinedOutput()

	if runErr != nil {
		t.Fatalf("cmd failed with %v: %s", runErr, string(runOut))
	}
}

func TestCheckRollbackCmd_Blocked(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"can_rollback": false,
			"message":      "Blocked",
			"blocked_by":   []string{"org/consumer1"},
		})
	}))
	defer ts.Close()

	os.Setenv("REGISTRY_API_TOKEN", "testtoken")
	os.Setenv("SUBSTRATE_API_URL", ts.URL)
	os.Setenv("SUBSTRATE_DISABLE_CACHE_SYNC", "1")
	defer os.Unsetenv("REGISTRY_API_TOKEN")
	defer os.Unsetenv("SUBSTRATE_API_URL")
	defer os.Unsetenv("SUBSTRATE_DISABLE_CACHE_SYNC")

	cmd := exec.Command("go", "build", "-o", "substrate_bin")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build failed: %s", out)
	}
	defer os.Remove("substrate_bin")

	runCmd := exec.Command("./substrate_bin", "check-rollback", "--repo=testorg/testrepo", "--target-sha=abcdef")
	runOut, runErr := runCmd.CombinedOutput()

	if runErr == nil {
		t.Fatalf("expected command to fail, but it succeeded: %s", string(runOut))
	}
}
