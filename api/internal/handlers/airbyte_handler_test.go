package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/crypto"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestAirbyteHandler_Ingest(t *testing.T) {
	orgName := "test-org"
	sourceName := "source-hubspot"

	validKeyARN := "arn:aws:kms:us-east-1:123456789012:key/mrk-12345"
	_, _ = crypto.NewKMSClient(validKeyARN)

	tests := []struct {
		name           string
		reqBody        AirbyteIngestRequest
		setupMock      func(*db.MockStore)
		expectedStatus int
	}{
		{
			name: "successful ingest with existing source",
			reqBody: AirbyteIngestRequest{
				Org:    orgName,
				Source: sourceName,
				Stream: "contacts",
				Records: []json.RawMessage{
					[]byte(`{"id": "1", "name": "bob"}`),
				},
			},
			setupMock: func(m *db.MockStore) {
				m.GetAirbyteSourceFunc = func(ctx context.Context, arg sqlcgen.GetAirbyteSourceParams) (sqlcgen.AirbyteSource, error) {
					return sqlcgen.AirbyteSource{
						ID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
					}, nil
				}
				m.BulkInsertAirbyteRecordsFunc = func(ctx context.Context, org string, sourceID uuid.UUID, stream string, records [][]byte) (int, error) {
					return 1, nil
				}
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "successful ingest creating new source",
			reqBody: AirbyteIngestRequest{
				Org:    orgName,
				Source: sourceName,
				Stream: "contacts",
				Records: []json.RawMessage{
					[]byte(`{"id": "1", "name": "bob"}`),
				},
			},
			setupMock: func(m *db.MockStore) {
				m.GetAirbyteSourceFunc = func(ctx context.Context, arg sqlcgen.GetAirbyteSourceParams) (sqlcgen.AirbyteSource, error) {
					return sqlcgen.AirbyteSource{}, errors.New("not found")
				}
				m.GetOrgKMSConfigFunc = func(ctx context.Context, orgName string) (string, string, error) {
					return "aws", validKeyARN, nil
				}
				m.UpsertAirbyteSourceFunc = func(ctx context.Context, arg sqlcgen.UpsertAirbyteSourceParams) (sqlcgen.AirbyteSource, error) {
					return sqlcgen.AirbyteSource{
						ID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
					}, nil
				}
				m.BulkInsertAirbyteRecordsFunc = func(ctx context.Context, org string, sourceID uuid.UUID, stream string, records [][]byte) (int, error) {
					return 1, nil
				}
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "fails when no KMS config",
			reqBody: AirbyteIngestRequest{
				Org:    orgName,
				Source: sourceName,
				Stream: "contacts",
				Records: []json.RawMessage{},
			},
			setupMock: func(m *db.MockStore) {
				m.GetAirbyteSourceFunc = func(ctx context.Context, arg sqlcgen.GetAirbyteSourceParams) (sqlcgen.AirbyteSource, error) {
					return sqlcgen.AirbyteSource{}, errors.New("not found")
				}
				m.GetOrgKMSConfigFunc = func(ctx context.Context, orgName string) (string, string, error) {
					return "", "", errors.New("no kms key")
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing required fields",
			reqBody: AirbyteIngestRequest{
				Org: "",
			},
			setupMock: func(m *db.MockStore) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &db.MockStore{}
			tt.setupMock(mockStore)
			handler := NewAirbyteHandler(mockStore)

			bodyBytes, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/airbyte/ingest", bytes.NewReader(bodyBytes))
			rr := httptest.NewRecorder()

			handler.Ingest(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestAirbyteHandler_CreateSource(t *testing.T) {
	orgName := "test-org"
	validKeyARN := "arn:aws:kms:us-east-1:123456789012:key/mrk-12345"
	_, _ = crypto.NewKMSClient(validKeyARN)

	tests := []struct {
		name           string
		reqBody        AirbyteSourceRequest
		setupMock      func(*db.MockStore)
		expectedStatus int
	}{
		{
			name: "successful create",
			reqBody: AirbyteSourceRequest{
				Name:      "hubspot",
				Connector: "source-hubspot",
				Config:    []byte(`{}`),
			},
			setupMock: func(m *db.MockStore) {
				m.GetOrgKMSConfigFunc = func(ctx context.Context, orgName string) (string, string, error) {
					return "aws", validKeyARN, nil
				}
				m.UpsertAirbyteSourceFunc = func(ctx context.Context, arg sqlcgen.UpsertAirbyteSourceParams) (sqlcgen.AirbyteSource, error) {
					return sqlcgen.AirbyteSource{
						ID: pgtype.UUID{Bytes: uuid.New(), Valid: true},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "fails when no KMS config",
			reqBody: AirbyteSourceRequest{
				Name:      "hubspot",
				Connector: "source-hubspot",
				Config:    []byte(`{}`),
			},
			setupMock: func(m *db.MockStore) {
				m.GetOrgKMSConfigFunc = func(ctx context.Context, orgName string) (string, string, error) {
					return "", "", errors.New("no kms key")
				}
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &db.MockStore{}
			tt.setupMock(mockStore)
			handler := NewAirbyteHandler(mockStore)

			bodyBytes, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/org/"+orgName+"/airbyte/sources", bytes.NewReader(bodyBytes))

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("org", orgName)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.CreateSource(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestAirbyteHandler_ListSources(t *testing.T) {
	mockStore := &db.MockStore{}
	mockStore.ListAirbyteSourcesFunc = func(ctx context.Context, org string) ([]sqlcgen.AirbyteSource, error) {
		return []sqlcgen.AirbyteSource{}, nil
	}
	handler := NewAirbyteHandler(mockStore)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/org/test-org/airbyte/sources", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("org", "test-org")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.ListSources(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestAirbyteHandler_DeleteSource(t *testing.T) {
	mockStore := &db.MockStore{}
	mockStore.DeleteAirbyteSourceFunc = func(ctx context.Context, arg sqlcgen.DeleteAirbyteSourceParams) error {
		return nil
	}
	handler := NewAirbyteHandler(mockStore)

	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/org/test-org/airbyte/sources/"+id, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("org", "test-org")
	rctx.URLParams.Add("id", id)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.DeleteSource(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rr.Code)
	}
}
