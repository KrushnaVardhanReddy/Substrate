// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package main

import (
	"context"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestE2EWatchCommand(t *testing.T) {
	// Build the substrate binary
	cmd := exec.Command("go", "build", "-o", "substrate", ".")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("failed to build substrate binary: %v", err)
	}
	defer os.Remove("./substrate")

	// Create a temporary directory to watch
	dir := t.TempDir()

	// Start substrate watch in the background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	watchCmd := exec.CommandContext(ctx, filepath.Join(viper.GetString("PWD"), "substrate"), "watch")
	watchCmd.Dir = dir

	// We need to capture the output to verify it prints "spec updated"
	// However, we just need to verify it processes the file modification.
	// We will write a dummy openapi.yaml to check if it gets created.

	// Because it uses AI by default and might fail without keys,
	// we will rely on the deterministic fallback by unsetting the Base URL.
	watchCmd.Env = append(os.Environ(), "SUBSTRATE_AI_BASE_URL=")

	// We run it in a way that we can read its output
	outputBytes := &strings.Builder{}
	watchCmd.Stdout = outputBytes
	watchCmd.Stderr = outputBytes

	err = watchCmd.Start()
	if err != nil {
		t.Fatalf("failed to start watch command: %v", err)
	}

	// Give the watcher a moment to initialize
	time.Sleep(500 * time.Millisecond)

	// Create a .go file to trigger the watcher
	testFile := filepath.Join(dir, "main.go")
	err = os.WriteFile(testFile, []byte("package main"), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Wait for the debounce window (500ms) + processing time
	time.Sleep(1500 * time.Millisecond)

	cancel()
	_ = watchCmd.Wait()

	output := outputBytes.String()

	if !strings.Contains(output, "watch started") {
		t.Errorf("expected 'watch started' in output, got: %s", output)
	}

	if !strings.Contains(output, "spec updated") {
		t.Errorf("expected 'spec updated' in output, got: %s", output)
	}

	// Verify openapi.yaml was created with the fallback mock
	yamlContent, err := os.ReadFile(filepath.Join(dir, "openapi.yaml"))
	if err != nil {
		t.Fatalf("expected openapi.yaml to be created: %v", err)
	}

	if !strings.Contains(string(yamlContent), "Mock Watch API") {
		t.Errorf("expected mock yaml content, got: %s", string(yamlContent))
	}
}
