package sandbox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func RunCode(code string) (string, string, error) {
	tmpDir, err := os.MkdirTemp("", "substrate-sandbox-*")
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFilePath := filepath.Join(tmpDir, "test.js")
	if err := os.WriteFile(testFilePath, []byte(code), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write test file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use node as a substitute for deno if deno isn't available
	// The problem statement says "wrapping Deno or Node"
	cmd := exec.CommandContext(ctx, "node", testFilePath)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	// Error could just mean the test failed, which is expected for Chaos tests showing breakage
	return stdout.String(), stderr.String(), err
}
