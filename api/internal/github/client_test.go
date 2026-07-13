package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRESTClient_SearchCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"items": [{"path": "config/substrate.yaml"}]}`))
	}))
	defer ts.Close()

	client := &RESTClient{
		client: ts.Client(),
		apiURL: ts.URL,
	}

	path, err := client.SearchCode(context.Background(), "owner", "repo", "query")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if path != "config/substrate.yaml" {
		t.Errorf("expected path 'config/substrate.yaml', got '%s'", path)
	}
}

func TestRESTClient_GetFileContent(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// "aGVsbG8=" is "hello" in base64
		w.Write([]byte(`{"content": "aGVsbG8=", "encoding": "base64"}`))
	}))
	defer ts.Close()

	client := &RESTClient{
		client: ts.Client(),
		apiURL: ts.URL,
	}

	content, err := client.GetFileContent(context.Background(), "owner", "repo", "path")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if content != "hello" {
		t.Errorf("expected content 'hello', got '%s'", content)
	}
}

func TestRESTClient_CreateDraftPR(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"object": {"sha": "dummy-sha"}}`))
			return
		}
		w.WriteHeader(http.StatusCreated)
		if strings.HasSuffix(r.URL.Path, "pulls") {
			w.Write([]byte(`{"html_url": "https://github.com/owner/repo/pull/1"}`))
			return
		}
		w.Write([]byte(`{"sha": "dummy-sha"}`))
	}))
	defer ts.Close()

	client := &RESTClient{
		client: ts.Client(),
		apiURL: ts.URL,
	}

	url, err := client.CreateDraftPR(context.Background(), "owner", "repo", "branch", "patch", "title", "body")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if url != "https://github.com/owner/repo/pull/1" {
		t.Errorf("expected url 'https://github.com/owner/repo/pull/1', got '%s'", url)
	}
}
