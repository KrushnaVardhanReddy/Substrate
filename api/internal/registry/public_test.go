// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package registry_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/registry"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestHandlePublishSchema(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		schemaName     string
		version        string
		reqBody        registry.PublishRequest
		mockStoreErr   error
		expectedStatus int
	}{
		{
			name:       "valid request",
			namespace:  "stripe",
			schemaName: "api",
			version:    "v1",
			reqBody: registry.PublishRequest{
				SchemaType:    "openapi",
				SchemaContent: "openapi: 3.0.0",
			},
			mockStoreErr:   nil,
			expectedStatus: http.StatusCreated,
		},
		{
			name:       "missing schema type",
			namespace:  "stripe",
			schemaName: "api",
			version:    "v1",
			reqBody: registry.PublishRequest{
				SchemaContent: "openapi: 3.0.0",
			},
			mockStoreErr:   nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "store error",
			namespace:  "stripe",
			schemaName: "api",
			version:    "v1",
			reqBody: registry.PublishRequest{
				SchemaType:    "openapi",
				SchemaContent: "openapi: 3.0.0",
			},
			mockStoreErr:   errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &db.MockStore{
				PublishPublicSchemaFunc: func(ctx context.Context, namespace, name, version, schemaType, content string) error {
					return tt.mockStoreErr
				},
			}

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", tt.namespace)
			rctx.URLParams.Add("name", tt.schemaName)
			rctx.URLParams.Add("version", tt.version)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler := registry.HandlePublishSchema(store)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandleFetchSchema(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		schemaName     string
		version        string
		mockStoreRes   *db.PublicSchema
		mockStoreErr   error
		expectedStatus int
	}{
		{
			name:       "found",
			namespace:  "stripe",
			schemaName: "api",
			version:    "v1",
			mockStoreRes: &db.PublicSchema{
				NamespaceName: "stripe",
				Name:          "api",
				Version:       "v1",
				SchemaType:    "openapi",
				SchemaContent: "openapi: 3.0.0",
			},
			mockStoreErr:   nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "not found",
			namespace:      "stripe",
			schemaName:     "api",
			version:        "v1",
			mockStoreRes:   nil,
			mockStoreErr:   errors.New("not found"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &db.MockStore{
				GetPublicSchemaFunc: func(ctx context.Context, namespace, name, version string) (*db.PublicSchema, error) {
					return tt.mockStoreRes, tt.mockStoreErr
				},
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("namespace", tt.namespace)
			rctx.URLParams.Add("name", tt.schemaName)
			rctx.URLParams.Add("version", tt.version)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler := registry.HandleFetchSchema(store)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
