package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockShadowAnalyzer struct {
	CapturedLogs []string
	CoverageMap  map[string]bool
}

func (m *MockShadowAnalyzer) Analyze(logs []string) map[string]bool {
	m.CapturedLogs = logs
	// Simulate parsing logic
	coverage := make(map[string]bool)
	for _, log := range logs {
		if strings.Contains(log, "POST /users") {
			coverage["#/paths/~1users/post"] = true
		}
	}
	m.CoverageMap = coverage
	return coverage
}

type MockSyncClient struct {
	SyncedSchema string
	SyncTriggered bool
}

func (m *MockSyncClient) Sync(schema string) error {
	m.SyncedSchema = schema
	m.SyncTriggered = true
	return nil
}

func TestPhase6QAFeedbackLoop(t *testing.T) {
	cancel := func(){}
	defer cancel()

	// 1. Build the engine binary (substrate CLI)
	engineDir, err := filepath.Abs("../../engine")
	require.NoError(t, err)

	binPath := filepath.Join(engineDir, "substrate_test_bin")
	cmdBuild := exec.Command("go", "build", "-o", "substrate_test_bin", "./cmd/substrate")
	cmdBuild.Dir = engineDir
	err = cmdBuild.Run()
	require.NoError(t, err, "Failed to compile the substrate CLI")

	defer os.Remove(binPath) // Cleanup

	// 2. Setup Fixtures & Registry Mock
	schemaBytes, err := os.ReadFile("testdata/phase6/mock_api.yaml")
	require.NoError(t, err, "Failed to read fixture schema")
	schemaStr := string(schemaBytes)

	schemaJSONBytes, _ := json.Marshal(schemaStr)
	schemaJSON := string(schemaJSONBytes)

	mockRegistry := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/schema/testorg/testrepo" {
			timestamp := r.URL.Query().Get("timestamp")
			assert.Equal(t, "2026-07-01", timestamp)

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"schema": ` + schemaJSON + `, "schema_type": "openapi"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockRegistry.Close()

	os.Setenv("REGISTRY_API_URL", mockRegistry.URL)
	os.Setenv("SUBSTRATE_OWNER", "testorg")
	os.Setenv("SUBSTRATE_REPO", "testrepo")

	// Step 1: Mock Server Orchestration (P6-T04)
	mockPort := "8185"
	cmdMock := exec.Command(binPath, "mock", "--timestamp", "2026-07-01", "--port", mockPort)
	cmdMock.Dir = engineDir
	err = cmdMock.Start()
	require.NoError(t, err, "Failed to start mock server")

	defer func() {
		if cmdMock.Process != nil {
			cmdMock.Process.Kill()
		}
	}()

	time.Sleep(500 * time.Millisecond) // wait for server to bind

	// Step 2: Auto-Test Generation (P6-T03)
	cmdGen := exec.Command(binPath, "generate-tests", "../scripts/e2e/testdata/phase6/mock_api.yaml", "--url", "http://localhost:"+mockPort)
	cmdGen.Dir = engineDir

	var genOut bytes.Buffer
	cmdGen.Stdout = &genOut
	cmdGen.Stderr = os.Stderr

	err = cmdGen.Run()
	require.NoError(t, err, "Failed to run generate-tests")

	generatedCode := genOut.String()
	assert.Contains(t, generatedCode, "exceeds maxLength", "Should test maxLength")
	assert.Contains(t, generatedCode, "missing required", "Should test missing required")

	// Step 3: Fuzzing the Mock Server
	validPayload := `{"name": "Alice", "email": "alice@example.com"}`
	resp, err := http.Post("http://localhost:"+mockPort+"/users", "application/json", strings.NewReader(validPayload))
	require.NoError(t, err, "Failed to POST to mock server")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Mock server should respond with 200 OK")

	// Step 4: Shadow API Coverage (P6-T02)
	mockAnalyzer := &MockShadowAnalyzer{}
	trafficLogs := []string{"POST /users HTTP/1.1", "GET /unknown HTTP/1.1"}
	coverage := mockAnalyzer.Analyze(trafficLogs)

	assert.True(t, coverage["#/paths/~1users/post"], "POST /users should be marked as covered")
	assert.False(t, coverage["#/paths/~1unknown/get"], "GET /unknown should not be covered")

	// Step 5: Postman Sync Webhook (P6-T01)
	mockSyncClient := &MockSyncClient{}

	// Simulate webhook payload processing
	err = mockSyncClient.Sync(schemaStr)
	require.NoError(t, err, "Failed to sync schema")

	assert.True(t, mockSyncClient.SyncTriggered, "Sync should have been triggered")
	assert.Equal(t, schemaStr, mockSyncClient.SyncedSchema, "Synced schema should match the original fixture schema")
}
