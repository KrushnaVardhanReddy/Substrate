// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRegisterAgentHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(m *db.MockStore)
		expectedStatus int
	}{
		{
			name: "success",
			requestBody: AgentRegistrationRequest{
				RepoName: "ai-agent",
				Owner:    "test-org",
				Tools: []AgentToolRequest{
					{
						ToolName:       "execute_sql",
						ParametersJSON: json.RawMessage(`{"database": "string"}`),
					},
				},
			},
			setupMock: func(m *db.MockStore) {
				m.RegisterAgentFunc = func(ctx context.Context, repoName, owner string, tools []db.AgentToolDependency) (uuid.UUID, error) {
					return uuid.New(), nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing required fields",
			requestBody: AgentRegistrationRequest{
				RepoName: "", // missing
				Owner:    "test-org",
			},
			setupMock:      func(m *db.MockStore) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing tool name",
			requestBody: AgentRegistrationRequest{
				RepoName: "ai-agent",
				Owner:    "test-org",
				Tools: []AgentToolRequest{
					{
						ToolName: "", // missing
					},
				},
			},
			setupMock:      func(m *db.MockStore) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid body",
			requestBody:    "not json",
			setupMock:      func(m *db.MockStore) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &db.MockStore{}
			tt.setupMock(m)

			handler := RegisterAgentHandler(m)

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/agents", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
