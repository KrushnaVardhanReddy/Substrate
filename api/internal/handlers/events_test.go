package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestEventBroker(t *testing.T) {
	broker := &EventBroker{
		clients: make(map[chan []byte]bool),
	}

	tests := []struct {
		name      string
		action    func()
		wantCount int
	}{
		{
			name: "Add client 1",
			action: func() {
				broker.AddClient(make(chan []byte))
			},
			wantCount: 1,
		},
		{
			name: "Add client 2",
			action: func() {
				broker.AddClient(make(chan []byte))
			},
			wantCount: 2,
		},
		{
			name: "Remove client 1 (mocking by clearing map)",
			action: func() {
				for k := range broker.clients {
					broker.RemoveClient(k)
					break
				}
			},
			wantCount: 1,
		},
		{
			name: "Remove client 2",
			action: func() {
				for k := range broker.clients {
					broker.RemoveClient(k)
					break
				}
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.action()
			assert.Equal(t, tt.wantCount, broker.GetClientCount())
		})
	}
}

func TestEventsHandler(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(EventsHandler))
	defer server.Close()

	tests := []struct {
		name        string
		timeout     time.Duration
		expectError bool
	}{
		{
			name:        "Connect and quick disconnect",
			timeout:     50 * time.Millisecond,
			expectError: true, // expect context canceled error from client.Do
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", server.URL, nil)
			assert.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()
			req = req.WithContext(ctx)

			client := server.Client()

			startCount := GlobalEventBroker.GetClientCount()

			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
			}

			time.Sleep(100 * time.Millisecond) // wait for server to unregister

			assert.Equal(t, startCount, GlobalEventBroker.GetClientCount(), "Broker count should return to start count after disconnect")
		})
	}
}

type dummyWriter struct {
	http.ResponseWriter
}

func TestEventsHandler_NotFlusher(t *testing.T) {
	tests := []struct {
		name         string
		writer       http.ResponseWriter
		expectedCode int
	}{
		{
			name:         "Unsupported flusher",
			writer:       dummyWriter{httptest.NewRecorder()},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/events", nil)
			EventsHandler(tt.writer, req)

			rec, ok := tt.writer.(dummyWriter).ResponseWriter.(*httptest.ResponseRecorder)
			if ok {
				assert.Equal(t, tt.expectedCode, rec.Code)
			}
		})
	}
}

func TestCreateEventHandler(t *testing.T) {
	mockStore := &db.MockStore{
		InsertEcosystemEventFunc: func(ctx context.Context, arg sqlcgen.InsertEcosystemEventParams) (sqlcgen.EcosystemEvent, error) {
			return sqlcgen.EcosystemEvent{
				Org:         arg.Org,
				Repo:        arg.Repo,
				EventType:   arg.EventType,
				Description: arg.Description,
				EventTime:   arg.EventTime,
			}, nil
		},
	}
	handler := CreateEventHandler(mockStore)

	t.Run("Valid deployment event", func(t *testing.T) {
		reqBody := EventRequest{
			Org:         "test-org",
			Repo:        "test-repo",
			EventType:   "deployment",
			Description: "v1.0.0",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/events", bytes.NewReader(bodyBytes))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("Invalid event type", func(t *testing.T) {
		reqBody := EventRequest{
			Org:       "test-org",
			Repo:      "test-repo",
			EventType: "invalid-type",
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/events", bytes.NewReader(bodyBytes))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "event_type must be deployment or incident")
	})
}

func TestGetEventsHandler(t *testing.T) {
	mockStore := &db.MockStore{
		GetEcosystemEventsByOrgFunc: func(ctx context.Context, arg sqlcgen.GetEcosystemEventsByOrgParams) ([]sqlcgen.EcosystemEvent, error) {
			return []sqlcgen.EcosystemEvent{
				{Org: "test-org", Repo: "repo1", EventType: "deployment"},
			}, nil
		},
	}
	handler := GetEventsHandler(mockStore)

	t.Run("Get events for org", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/events/test-org", nil)
		r := chi.NewRouter()
		r.Get("/api/v1/events/{org}", handler)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "test-org")
	})
}
