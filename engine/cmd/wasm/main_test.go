//go:build js && wasm

package main

import (
	"os"
	"syscall/js"
	"testing"
	"strings"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/diff"
)

func TestSubstrateDiffWrapper_MissingArgs(t *testing.T) {
	// Create dummy args with length < 2
	args := []js.Value{js.ValueOf("base")}

	res := substrateDiffWrapper(js.Undefined(), args)

	if resStr := res.(js.Value).String(); !strings.Contains(resStr, "expected 2 arguments") {
		t.Errorf("Expected error message for missing args, got %v", resStr)
	}
}

func TestSubstrateDiffWrapper_InvalidJSON(t *testing.T) {
	args := []js.Value{js.ValueOf("{ invalid: base }"), js.ValueOf("{ invalid: head }")}

	res := substrateDiffWrapper(js.Undefined(), args)

	if resStr := res.(js.Value).String(); !strings.Contains(resStr, "error") {
		t.Errorf("Expected error message for invalid input, got %v", resStr)
	}
}


// A WASM-specific test suite for the engine logic in GOOS=js GOARCH=wasm.
func TestWASMBridgeResilience(t *testing.T) {
	// P12-T05: Write automated JS-to-WASM bridge tests to ensure massive
	// malformed strings don't panic the diff engine.
	// We'll create a very large garbage string (simulating 5MB, here we use around 3MB)
	// to test the boundary and ensure it doesn't crash the WASM process but instead
	// returns a handled error instead of a fatal WebAssembly panic.
	massiveGarbage := strings.Repeat("this is some extremely broken YAML \n\t: { [ ] ] ] ", 100000)

	// Since diff.CompareOpenAPI currently reads from a file path instead of string,
	// and we are constrained to test the boundary resilience of the engine parsing logic,
	// we will write the payload to a temporary file.

	baseFile, err := os.CreateTemp("", "base-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(baseFile.Name())

	revFile, err := os.CreateTemp("", "rev-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(revFile.Name())

	_, err = baseFile.WriteString(massiveGarbage)
	if err != nil {
		t.Fatalf("Failed to write to base temp file: %v", err)
	}
	baseFile.Close()

	_, err = revFile.WriteString(massiveGarbage)
	if err != nil {
		t.Fatalf("Failed to write to rev temp file: %v", err)
	}
	revFile.Close()

	// Use a defer func with recover() to catch any panics that might occur.
	// We want to fail the test if it panics instead of letting the process crash.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Engine paniced on massive malformed string: %v", r)
		}
	}()

	// Invoke the core engine diffing function.
	_, err = diff.CompareOpenAPI(baseFile.Name(), revFile.Name(), true, nil)

	// We expect an error to be returned due to the invalid spec format, not a panic.
	if err == nil {
		t.Fatalf("Expected an error for massive malformed string, but got nil")
	}

	// Verify the error message contains expected failure text rather than a panic
	if !strings.Contains(err.Error(), "invalid character") && !strings.Contains(err.Error(), "yaml error") && !strings.Contains(err.Error(), "failed to load base spec") {
		t.Errorf("Expected an unmarshal/load error, got: %v", err)
	}
}
