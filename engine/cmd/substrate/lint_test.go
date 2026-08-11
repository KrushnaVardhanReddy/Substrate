// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestLintCommand(t *testing.T) {
	// Create a dummy schema
	f, err := os.CreateTemp("", "schema*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	_, _ = f.WriteString("openapi: 3.0.0\ninfo:\n  title: Mock API\n")
	f.Close()

	// Clear env vars to force fallback mode
	os.Unsetenv("SUBSTRATE_AI_BASE_URL")
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("SUBSTRATE_AI_API_KEY")

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	lintSchemaPath = f.Name()
	err = lintCmd.RunE(lintCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("lintCmd.RunE failed: %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "FALLBACK_MODE") {
		t.Errorf("expected FALLBACK_MODE in output, got: %s", output)
	}
	if !strings.Contains(output, "\"score\": 90") {
		t.Errorf("expected score 90 in output, got: %s", output)
	}
}
