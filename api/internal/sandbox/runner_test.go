package sandbox

import (
	"strings"
	"testing"
)

func TestRunCode_Success(t *testing.T) {
	code := `console.log("hello world");`
	stdout, stderr, err := RunCode(code)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(stdout, "hello world") {
		t.Errorf("expected stdout to contain 'hello world', got %q", stdout)
	}
	if stderr != "" {
		t.Errorf("expected empty stderr, got %q", stderr)
	}
}

func TestRunCode_Failure(t *testing.T) {
	code := `throw new Error("breakage");`
	_, stderr, err := RunCode(code)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(stderr, "breakage") {
		t.Errorf("expected stderr to contain 'breakage', got %q", stderr)
	}
}
