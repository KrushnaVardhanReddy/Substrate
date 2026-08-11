// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestE2E_PostMortem(t *testing.T) {
	// Compile the Substrate binary
	cmd := exec.Command("go", "build", "-o", "substrate")
	err := cmd.Run()
	assert.NoError(t, err)
	defer os.Remove("./substrate")

	// Start mock Substrate API
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/graph/default" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{}"))
			return
		}

		assert.Equal(t, "/api/v1/postmortem", r.URL.Path)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "text/markdown")
		w.Write([]byte("# Post-Mortem\n\n"))
		w.Write([]byte("Root cause: removed field id"))
	}))
	defer mockAPI.Close()

	// Execute postmortem command
	postmortemCmd := exec.Command("./substrate", "postmortem", "--incident", "2023-10-10T12:00:00Z")
	postmortemCmd.Env = append(os.Environ(),
		"REGISTRY_API_TOKEN=test-token",
		"SUBSTRATE_API_URL="+mockAPI.URL,
		"SUBSTRATE_DISABLE_CACHE_SYNC=1",
	)

	out, err := postmortemCmd.CombinedOutput()
	assert.NoError(t, err, string(out))

	output := string(out)
	assert.Contains(t, output, "# Post-Mortem")
	assert.Contains(t, output, "Root cause: removed field id")
}
