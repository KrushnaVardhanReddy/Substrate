// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

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

func TestGetGraphAndGetSchema(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "substrate-cache-test-methods")
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

	// Insert test data
	_, err = cache.DB.Exec("INSERT INTO dependencies (provider, consumer, status) VALUES (?, ?, ?)", "org/repo1", "org/repo2", "SAFE")
	if err != nil {
		t.Fatal(err)
	}
	_, err = cache.DB.Exec("INSERT INTO dependencies (provider, consumer, status) VALUES (?, ?, ?)", "org/repo1", "org/repo3", "WARNING")
	if err != nil {
		t.Fatal(err)
	}
	_, err = cache.DB.Exec("INSERT INTO dependencies (provider, consumer, status) VALUES (?, ?, ?)", "otherorg/repo1", "otherorg/repo2", "SAFE")
	if err != nil {
		t.Fatal(err)
	}

	_, err = cache.DB.Exec("INSERT INTO schemas (repo, content) VALUES (?, ?)", "org/repo1", "schema_content")
	if err != nil {
		t.Fatal(err)
	}

	// Test GetGraph
	edges, err := cache.GetGraph("org")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(edges) != 2 {
		t.Fatalf("expected 2 edges for 'org', got %d", len(edges))
	}

	// Check content
	found1 := false
	found2 := false
	for _, r := range edges {
		if r.Provider == "org/repo1" && r.Consumer == "org/repo2" && r.Status == "SAFE" {
			found1 = true
		}
		if r.Provider == "org/repo1" && r.Consumer == "org/repo3" && r.Status == "WARNING" {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Errorf("missing expected edges")
	}

	// Test GetSchema
	schema, err := cache.GetSchema("org/repo1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(schema) != "schema_content" {
		t.Fatalf("expected 'schema_content', got %s", string(schema))
	}

	// Test GetSchema miss
	schema, err = cache.GetSchema("org/missing")
	if err != nil {
		t.Fatalf("expected no error on miss, got %v", err)
	}
	if schema != nil {
		t.Fatalf("expected nil for missing schema, got %v", string(schema))
	}

	// Test GetBreakingChangeHistory
	history, err := cache.GetBreakingChangeHistory("org/repo1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if history != nil {
		t.Fatalf("expected nil for history, got %v", history)
	}

	GlobalCache = nil
}
