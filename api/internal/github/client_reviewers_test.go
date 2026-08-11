// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package github

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"
)

func TestRESTClient_RequestReviewers(t *testing.T) {
	var requestedUsers []string
	var requestedTeams []string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/pulls/1/requested_reviewers" {
			t.Errorf("expected path /repos/owner/repo/pulls/1/requested_reviewers, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected method POST, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var payload map[string][]string
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		requestedUsers = payload["reviewers"]
		requestedTeams = payload["team_reviewers"]

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"html_url": "https://github.com/owner/repo/pull/1"}`))
	}))
	defer ts.Close()

	client := &RESTClient{
		client: ts.Client(),
		apiURL: ts.URL,
	}

	err := client.RequestReviewers(context.Background(), "owner", "repo", 1, []string{"@myorg/api-platform", "@user1", "user2", "org/team2"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedUsers := []string{"user1", "user2"}
	expectedTeams := []string{"api-platform", "team2"}

	sort.Strings(requestedUsers)
	sort.Strings(expectedUsers)
	sort.Strings(requestedTeams)
	sort.Strings(expectedTeams)

	if !reflect.DeepEqual(requestedUsers, expectedUsers) {
		t.Errorf("expected users %v, got %v", expectedUsers, requestedUsers)
	}
	if !reflect.DeepEqual(requestedTeams, expectedTeams) {
		t.Errorf("expected teams %v, got %v", expectedTeams, requestedTeams)
	}
}
