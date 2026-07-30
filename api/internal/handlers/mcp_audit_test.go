package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
)

type mockMCPAuditStore struct {
	*db.MockStore
	logs []sqlcgen.McpAuditLog
}

func (m *mockMCPAuditStore) ListMCPAuditLogs(ctx context.Context) ([]sqlcgen.McpAuditLog, error) {
	return m.logs, nil
}

func TestListMCPAuditLogs(t *testing.T) {
	now := time.Now()
	mockStore := &mockMCPAuditStore{
		MockStore: &db.MockStore{},
		logs: []sqlcgen.McpAuditLog{
			{
				ID: 1,
				ToolName: "test-tool",
				ExecutedAt: now,
			},
		},
	}

	handler := &MCPAuditHandler{Store: mockStore}

	req, _ := http.NewRequest("GET", "/api/v1/ai/audit", nil)
	rr := httptest.NewRecorder()

	handler.ListMCPAuditLogs(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "test-tool")
}
