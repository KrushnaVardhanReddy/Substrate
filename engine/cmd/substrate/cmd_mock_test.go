package main

import (
	"bytes"
	"github.com/spf13/viper"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestMockCmdE2E(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "substrate_test_bin", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build test binary: %v", err)
	}
	defer func() {
		// Clean up test binary
		exec.Command("rm", "substrate_test_bin").Run()
	}()

	cmd := exec.Command("./substrate_test_bin", "mock", "--timestamp", "2026-07-01", "--port", "8091")

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start mock server: %v", err)
	}
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	}()

	time.Sleep(2000 * time.Millisecond)

	resp, err := http.Get("http://localhost:8091/")
	if err != nil {
		t.Fatalf("failed to make request: %v\nOutput: %s", err, out.String())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func init() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}
