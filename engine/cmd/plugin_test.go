package cmd

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestPublishCmd(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/marketplace/publish" {
			w.WriteHeader(http.StatusCreated)
		}
	}))
	defer ts.Close()

	os.Setenv("SUBSTRATE_API_URL", ts.URL)
	defer os.Unsetenv("SUBSTRATE_API_URL")

	tmpDir := t.TempDir()
	pluginFile := filepath.Join(tmpDir, "plugin.json")
	os.WriteFile(pluginFile, []byte(`{"name": "test-plugin", "description": "test", "schema": {"rules": []}}`), 0644)

	publishCmd.SetArgs([]string{pluginFile})
	if err := publishCmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestInstallCmd(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/marketplace/plugins" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"name": "test-plugin", "schema_content": {"rules": [{"id": "test-rule"}]}}]`))
		}
	}))
	defer ts.Close()

	os.Setenv("SUBSTRATE_API_URL", ts.URL)
	defer os.Unsetenv("SUBSTRATE_API_URL")

	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	// Create a dummy substrate.yaml so config.LoadConfig does not fail or handles it gracefully
	os.WriteFile("substrate.yaml", []byte("version: \"1.0\"\nservice: test"), 0644)

	installCmd.SetArgs([]string{"test-plugin"})
	if err := installCmd.Execute(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	data, err := os.ReadFile("substrate.yaml")
	if err != nil {
		t.Fatalf("expected substrate.yaml to be created, got error: %v", err)
	}

	if string(data) == "" {
		t.Fatal("expected substrate.yaml to have content")
	}
}
