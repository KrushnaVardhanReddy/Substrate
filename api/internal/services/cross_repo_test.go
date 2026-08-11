// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPerformCrossRepoCheck_SLA(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		report := DiffReport{
			Breaking: []interface{}{
				map[string]interface{}{"rule_id": "FIELD_REMOVED", "severity": "BREAKING", "description": "some field removed"},
			},
			Summary: DiffReportSummary{BreakingCount: 1},
		}
		json.NewEncoder(w).Encode(report)
	}))
	defer mockServer.Close()
	os.Setenv("DIFF_ENGINE_URL", mockServer.URL)

	mockStore := &db.MockStore{
		GetContractsByProviderFullNameFunc: func(ctx context.Context, providerFullName string) ([]db.Contract, error) {
			return []db.Contract{
				{
					ID:         uuid.New(),
					SchemaType: "openapi",
				},
			}, nil
		},
		GetConsumersByProviderContractFunc: func(ctx context.Context, providerContractID uuid.UUID) ([]db.ConsumerDependency, error) {
			return []db.ConsumerDependency{
				{
					ConsumerFullName:   "acme/invoice-service",
					RequiredNoticeDays: 30,
				},
				{
					ConsumerFullName:   "acme/other-service",
					RequiredNoticeDays: 0,
				},
			}, nil
		},
		GetConsumerManifestsFunc: func(ctx context.Context, providerRepo, consumerRepo string) (json.RawMessage, error) {
			return nil, nil // No manifest pruning
		},
	}

	req := CrossRepoCheckRequest{
		ProviderRepo:      "acme/billing-api",
		SchemaType:        "openapi",
		HeadSchemaContent: "mock schema",
	}

	resp, err := PerformCrossRepoCheck(context.Background(), mockStore, req)
	assert.NoError(t, err)

	assert.False(t, resp.IsSafe)
	assert.Equal(t, 2, resp.TotalConsumers)
	assert.Equal(t, 2, resp.BrokenConsumers)

	assert.Len(t, resp.SLABreaches, 1)
	assert.Equal(t, "acme/invoice-service", resp.SLABreaches[0].Consumer)
	assert.Equal(t, 30, resp.SLABreaches[0].RequiredDays)
}
