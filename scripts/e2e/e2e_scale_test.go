package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestScaleGeneratorE2E(t *testing.T) {
	// E2E test invokes the compiled binary to ensure it works end to end.
	buildCmd := exec.Command("go", "build", "-o", "substrate_scale_test_bin", "scale_generator.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}
	defer exec.Command("rm", "-f", "substrate_scale_test_bin").Run()

	cmd := exec.Command("./substrate_scale_test_bin", "--scale", "10", "--concurrency", "2", "--duration", "2s")
	cmd.Env = append(os.Environ(), "GITHUB_TOKEN=dummy")

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("execution failed: %v\nOutput: %s", err, string(out))
	}

	outStr := string(out)

	// Assertions based on expected report
	if !strings.Contains(outStr, "SCALE SIMULATION REPORT") {
		t.Errorf("Expected report header in output, got: %s", outStr)
	}

	if !strings.Contains(outStr, "Panics Caught:") {
		t.Errorf("Expected panics count in output, got: %s", outStr)
	}

	if !strings.Contains(outStr, "Graph Accuracy:") {
		t.Errorf("Expected graph accuracy in output, got: %s", outStr)
	}
}
