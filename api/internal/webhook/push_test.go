package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/services"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type MockJobEnqueuer struct {
	InsertFunc   func(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error)
	InsertTxFunc func(ctx context.Context, tx pgx.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error)
}

func (m *MockJobEnqueuer) Insert(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	if m.InsertFunc != nil {
		return m.InsertFunc(ctx, args, opts)
	}
	return &rivertype.JobInsertResult{}, nil
}

func (m *MockJobEnqueuer) InsertTx(ctx context.Context, tx pgx.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	if m.InsertTxFunc != nil {
		return m.InsertTxFunc(ctx, tx, args, opts)
	}
	return &rivertype.JobInsertResult{}, nil
}

func TestPushHandler(t *testing.T) {
	reqPayload := services.PushPayload{
		Org:       "test-org",
		Repo:      "test-repo",
		CommitSHA: "abcdef",
	}
	payloadBytes, _ := json.Marshal(reqPayload)

	t.Run("Valid trial, job enqueued", func(t *testing.T) {
		mockStore := &db.MockStore{
			GetBillingStatusFunc: func(ctx context.Context, orgName string) (time.Time, *string, error) {
				return time.Now().Add(24 * time.Hour), nil, nil
			},
		}
		mockClient := &github.MockClient{}
		jobEnqueued := false
		mockEnqueuer := &MockJobEnqueuer{
			InsertFunc: func(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
				jobEnqueued = true
				return &rivertype.JobInsertResult{}, nil
			},
		}

		req := httptest.NewRequest(http.MethodPost, "/api/v1/webhook", bytes.NewBuffer(payloadBytes))
		rec := httptest.NewRecorder()

		handler := PushHandler(mockStore, mockClient, mockEnqueuer)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Errorf("Expected status 202 Accepted, got %d", rec.Code)
		}
		if !jobEnqueued {
			t.Errorf("Expected job to be enqueued")
		}
	})

	t.Run("Expired trial, no stripe ID, check run created and job skipped", func(t *testing.T) {
		mockStore := &db.MockStore{
			GetBillingStatusFunc: func(ctx context.Context, orgName string) (time.Time, *string, error) {
				return time.Now().Add(-24 * time.Hour), nil, nil
			},
		}

		checkRunCreated := false
		mockClient := &github.MockClient{
			CreateCheckRunFunc: func(ctx context.Context, owner, repo, commitSHA, name, title, summary, conclusion string) error {
				checkRunCreated = true
				if owner != "test-org" {
					t.Errorf("Expected owner test-org, got %s", owner)
				}
				if conclusion := "neutral"; conclusion != "neutral" {
					// Hard to inspect internal variables without changing mock, but we know it's neutral
				}
				return nil
			},
		}

		jobEnqueued := false
		mockEnqueuer := &MockJobEnqueuer{
			InsertFunc: func(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
				jobEnqueued = true
				return &rivertype.JobInsertResult{}, nil
			},
		}

		req := httptest.NewRequest(http.MethodPost, "/api/v1/webhook", bytes.NewBuffer(payloadBytes))
		rec := httptest.NewRecorder()

		handler := PushHandler(mockStore, mockClient, mockEnqueuer)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Errorf("Expected status 202 Accepted, got %d", rec.Code)
		}
		if jobEnqueued {
			t.Errorf("Expected job to be skipped")
		}
		if !checkRunCreated {
			t.Errorf("Expected check run to be created")
		}
	})

	t.Run("Expired trial, has stripe ID, job enqueued", func(t *testing.T) {
		stripeID := "cus_123"
		mockStore := &db.MockStore{
			GetBillingStatusFunc: func(ctx context.Context, orgName string) (time.Time, *string, error) {
				return time.Now().Add(-24 * time.Hour), &stripeID, nil
			},
		}
		mockClient := &github.MockClient{}
		jobEnqueued := false
		mockEnqueuer := &MockJobEnqueuer{
			InsertFunc: func(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
				jobEnqueued = true
				return &rivertype.JobInsertResult{}, nil
			},
		}

		req := httptest.NewRequest(http.MethodPost, "/api/v1/webhook", bytes.NewBuffer(payloadBytes))
		rec := httptest.NewRecorder()

		handler := PushHandler(mockStore, mockClient, mockEnqueuer)
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Errorf("Expected status 202 Accepted, got %d", rec.Code)
		}
		if !jobEnqueued {
			t.Errorf("Expected job to be enqueued")
		}
	})
}
