package webhook

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

func TestReactionHandler(t *testing.T) {
	mockStore := &db.MockStore{}
	mockGh := &github.MockClient{}

	checkRunCalled := false
	mockGh.CreateCheckRunFunc = func(ctx context.Context, owner, repo, commitSHA, name, title, summary string) error {
		checkRunCalled = true
		return nil
	}
	mockGh.GetPullRequestHeadSHAFunc = func(ctx context.Context, owner, repo string, issueNumber int) (string, error) {
		return "mock_head_sha", nil
	}
	mockGh.GetIssueCommentReactionsFunc = func(ctx context.Context, owner, repo string, issueNumber int, commentID int64) ([]string, error) {
		return []string{"@alice"}, nil
	}

	handler := ReactionHandler(mockStore, mockGh)

	payload := `{
		"action": "created",
		"reaction": {"content": "+1"},
		"sender": {"login": "alice"},
		"repository": {"name": "repo", "owner": {"login": "org"}},
		"comment": {"id": 123, "body": "## Action Required\n@alice"}
	}`

	req := httptest.NewRequest("POST", "/api/v1/webhook/reaction", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("Expected status 202, got %d", w.Code)
	}

	if !checkRunCalled {
		t.Errorf("Expected check run to be created")
	}
}
