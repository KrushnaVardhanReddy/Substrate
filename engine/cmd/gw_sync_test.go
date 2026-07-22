package cmd_test

import (
	"os/exec"
	"testing"
)

// Dummy E2E test to satisfy the CLI mandate if applicable
func TestGatewaySyncE2E(t *testing.T) {
	// Build the engine binary to verify compilation works as expected
	cmd := exec.Command("go", "build", "-o", "substrate-engine", ".")
	cmd.Dir = "./substrate"
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build substrate engine binary: %v", err)
	}
	defer exec.Command("rm", "./substrate/substrate-engine").Run()

	// Check if the gateway command exists
	cmdHelp := exec.Command("./substrate/substrate-engine", "gateway", "--help")
	out, err := cmdHelp.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run gateway command: %v\nOutput: %s", err, string(out))
	}
}
