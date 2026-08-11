// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
)

func TestArchaeologyCmd(t *testing.T) {
	assert.NotNil(t, archaeologyCmd)
	assert.Equal(t, "archaeology", archaeologyCmd.Use)
}

func TestParseCustomDuration(t *testing.T) {
	d, err := ParseCustomDuration("2-years")
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(2)*24*365*time.Hour, d)

	d, err = ParseCustomDuration("3-months")
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(3)*24*30*time.Hour, d)

	d, err = ParseCustomDuration("14-days")
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(14)*24*time.Hour, d)

	d, err = ParseCustomDuration("24h")
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(24)*time.Hour, d)
}

func TestArchaeologyCLI_E2E(t *testing.T) {
	// Compile the CLI
	buildCmd := exec.Command("go", "build", "-o", "substrate", ".")
	err := buildCmd.Run()
	assert.NoError(t, err)

	// Create a temporary git repo to run real assertions against
	tempDir, err := os.MkdirTemp("", "archaeology_e2e")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	r, err := git.PlainInit(tempDir, false)
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	specPath := filepath.Join(tempDir, "openapi.yaml")
	specContent1 := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      responses:
        '200':
          description: OK`
	err = os.WriteFile(specPath, []byte(specContent1), 0644)
	assert.NoError(t, err)

	_, err = w.Add("openapi.yaml")
	assert.NoError(t, err)

	_, err = w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)

	// Breaking change
	specContent2 := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      responses:
        '201':
          description: Created`
	err = os.WriteFile(specPath, []byte(specContent2), 0644)
	assert.NoError(t, err)

	_, err = w.Add("openapi.yaml")
	assert.NoError(t, err)

	_, err = w.Commit("Breaking commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)

	absBinary, _ := filepath.Abs("substrate")

	// Run archaeology command.
	runCmd := exec.Command(absBinary, "archaeology", "--repo", tempDir, "--since", "1-days")
	var out bytes.Buffer
	runCmd.Stdout = &out
	runCmd.Stderr = &out // also capture stderr just in case
	err = runCmd.Run()
	assert.NoError(t, err)

	outStr := out.String()
	assert.Contains(t, outStr, "Retroactive Dependency Archaeology Report")
	assert.Contains(t, outStr, "**Total Breaking Changes:** 1")
	assert.Contains(t, outStr, "**Total Estimated Cost:** $5000.00")
}
