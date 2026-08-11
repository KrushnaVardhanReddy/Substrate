// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package archaeology

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
)

func TestEstimateCost(t *testing.T) {
	cost := EstimateCost(3)
	assert.Equal(t, 3, cost.Count)
	assert.Equal(t, 15000.0, cost.TotalCost)
}

func TestRun_NoRepo(t *testing.T) {
	err := Run("/non/existent/repo", time.Now().Add(-time.Hour), "openapi.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open repo")
}

func TestRun_Success(t *testing.T) {
	// Create a temporary git repo
	tempDir, err := os.MkdirTemp("", "archaeology_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	r, err := git.PlainInit(tempDir, false)
	assert.NoError(t, err)

	w, err := r.Worktree()
	assert.NoError(t, err)

	// Create an initial OpenAPI spec
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

	commit1, err := w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)

	// Create a breaking change
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

	// Ensure the archaeology test can successfully read these revisions.
	// We'll capture the stdout to avoid spamming the test output, but mainly we want to test if it errors out.
	err = Run(tempDir, time.Now().Add(-time.Hour*24), "openapi.yaml")
	assert.NoError(t, err)

	// Print commit hashes so we see them if we debug
	t.Logf("Commit 1: %s", commit1.String())
}
