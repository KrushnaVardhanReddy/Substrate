// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestROIHandler(t *testing.T) {
	tests := []struct {
		name           string
		orgName        string
		mockROIMetrics db.ROIMetrics
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:    "successful retrieval",
			orgName: "acme-corp",
			mockROIMetrics: db.ROIMetrics{
				TotalPreventedOutages:      5,
				TotalUndocumentedEndpoints: 2,
				HoursSaved:                 20,
				EstimatedDollarValueSaved:  2000,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"total_prevented_outages":5,"total_undocumented_endpoints":2,"hours_saved":20,"estimated_dollar_value_saved":2000}`,
		},
		{
			name:           "missing org name",
			orgName:        "",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "organization is required\n",
		},
		{
			name:           "internal error",
			orgName:        "acme-corp",
			mockError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &db.MockStore{
				GetROIMetricsFunc: func(ctx context.Context, orgID string) (db.ROIMetrics, error) {
					return tt.mockROIMetrics, tt.mockError
				},
			}

			req := httptest.NewRequest("GET", "/api/v1/telemetry/roi/"+tt.orgName, nil)

			// Set up chi router context for URL params
			rctx := chi.NewRouteContext()
			if tt.orgName != "" {
				rctx.URLParams.Add("org", tt.orgName)
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler := ROIHandler(mockStore)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if rr.Code == http.StatusOK {
				var resp db.ROIMetrics
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.mockROIMetrics, resp)
			} else {
				assert.Equal(t, tt.expectedBody, rr.Body.String())
			}
		})
	}
}
