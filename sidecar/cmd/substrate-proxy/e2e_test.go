package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestE2E_Sidecar(t *testing.T) {
	mockSchema := `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /known:
    get:
      responses:
        '200':
          description: OK
`

	anomalyChan := make(chan struct{})

	// Mock Substrate server (Registry + Telemetry)
	substrateServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/schema/") {
			w.Header().Set("Content-Type", "application/yaml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockSchema))
			return
		}

		if r.URL.Path == "/api/v1/telemetry/drift" {
			var payload struct {
				Path string `json:"path"`
			}
			json.NewDecoder(r.Body).Decode(&payload)

			if payload.Path == "/unknown" {
				close(anomalyChan)
			}
			w.WriteHeader(http.StatusAccepted)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer substrateServer.Close()

	// Mock Target server
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK from target"))
	}))
	defer targetServer.Close()

	// Build the sidecar binary
	cmdBuild := exec.Command("go", "build", "-o", "substrate_proxy_test", ".")
	if err := cmdBuild.Run(); err != nil {
		t.Fatalf("failed to build sidecar binary: %v", err)
	}
	defer exec.Command("rm", "substrate_proxy_test").Run()

	// Run the sidecar binary
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmdRun := exec.CommandContext(ctx, "./substrate_proxy_test",
		"-listen=:8082",
		"-target="+targetServer.URL,
		"-substrate-url="+substrateServer.URL,
		"-org=test-org",
		"-repo=test-repo",
		"-token=test-token",
		"-sample-rate=1.0",
	)

	var stderr bytes.Buffer
	cmdRun.Stderr = &stderr

	if err := cmdRun.Start(); err != nil {
		t.Fatalf("failed to start sidecar binary: %v", err)
	}

	// Wait for server to start and schema to be fetched
	time.Sleep(2 * time.Second)

	// Send request to known endpoint (should not report anomaly)
	resp, err := http.Get("http://localhost:8082/known")
	if err != nil {
		t.Fatalf("failed to send request to known endpoint: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %d", resp.StatusCode)
	}

	// Send request to unknown endpoint (should report anomaly)
	resp2, err := http.Get("http://localhost:8082/unknown")
	if err != nil {
		t.Fatalf("failed to send request to unknown endpoint: %v", err)
	}
	resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %d", resp2.StatusCode)
	}

	// Wait for anomaly to be reported
	select {
	case <-anomalyChan:
		// Success
	case <-time.After(3 * time.Second):
		t.Error("timeout waiting for anomaly report")
		t.Logf("Sidecar stderr: %s", stderr.String())
	}
}
