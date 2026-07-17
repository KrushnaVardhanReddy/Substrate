package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/go-chi/chi/v5"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestExportDocsHandler(t *testing.T) {
	mockStore := &db.MockStore{
		GetDependencyGraphFunc: func(ctx context.Context, orgName string) ([]db.DependencyEdge, error) {
			if orgName == "testorg" {
				return []db.DependencyEdge{
					{ConsumerFullName: "testorg/consumer-a", ProviderFullName: "testorg/provider-b"},
					{ConsumerFullName: "testorg/consumer-a", ProviderFullName: "testorg/provider-c"},
				}, nil
			}
			return []db.DependencyEdge{}, nil
		},
	}

	handler := HandleExportDocs(mockStore)

	t.Run("Valid Org", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/export/docs/testorg", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("org", "testorg")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "# Substrate Architecture")
		assert.Contains(t, rr.Body.String(), "```mermaid")
		assert.Contains(t, rr.Body.String(), "graph TD")
		assert.Contains(t, rr.Body.String(), "\"testorg/consumer-a\" --> \"testorg/provider-b\"")
		assert.Contains(t, rr.Body.String(), "\"testorg/consumer-a\" --> \"testorg/provider-c\"")
	})

	t.Run("Missing Org", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/export/docs/", nil)
		rctx2 := chi.NewRouteContext()
		rctx2.URLParams.Add("org", "")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx2))

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
