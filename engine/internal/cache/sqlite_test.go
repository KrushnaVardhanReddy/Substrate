package cache

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestInitCache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "substrate-cache-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "cache.db")
	cache, err := InitCache(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cache == nil {
		t.Fatal("expected cache to not be nil")
	}
	
	// Reset GlobalCache for other tests
	GlobalCache = nil
}

func TestSyncFromRemote(t *testing.T) {
	// Setup test database
	tmpDir, err := os.MkdirTemp("", "substrate-cache-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "cache.db")
	cache, err := InitCache(dbPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	defer cache.DB.Close()

	// Setup mock server
	mockEdges := []GraphEdge{
		{Provider: "org/repo1", Consumer: "org/repo2", Status: "SAFE"},
		{Provider: "org/repo1", Consumer: "org/repo3", Status: "WARNING"},
	}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/graph/org" {
			t.Errorf("expected path /api/v1/graph/org, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("expected token test-token, got %s", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockEdges)
	}))
	defer server.Close()

	err = cache.SyncFromRemote(context.Background(), server.URL, "test-token", "org")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify data inserted
	rows, err := cache.DB.Query("SELECT provider, consumer, status FROM dependencies")
	if err != nil {
		t.Fatalf("expected no error querying db, got %v", err)
	}
	defer rows.Close()

	var results []GraphEdge
	for rows.Next() {
		var e GraphEdge
		if err := rows.Scan(&e.Provider, &e.Consumer, &e.Status); err != nil {
			t.Fatal(err)
		}
		results = append(results, e)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	
	// Check content
	found1 := false
	found2 := false
	for _, r := range results {
		if r.Provider == "org/repo1" && r.Consumer == "org/repo2" && r.Status == "SAFE" {
			found1 = true
		}
		if r.Provider == "org/repo1" && r.Consumer == "org/repo3" && r.Status == "WARNING" {
			found2 = true
		}
	}
	if !found1 {
		t.Errorf("missing edge org/repo1 -> org/repo2")
	}
	if !found2 {
		t.Errorf("missing edge org/repo1 -> org/repo3")
	}

	// Reset GlobalCache for other tests
	GlobalCache = nil
}
