package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
)

type McpAuditLog = sqlcgen.McpAuditLog

func (s *PGStore) InsertMCPAuditLog(ctx context.Context, agentID *int, toolName string, req, res []byte) (int32, error) {
	var aID pgtype.Int4
	if agentID != nil {
		aID.Int32 = int32(*agentID)
		aID.Valid = true
	}

	queries := sqlcgen.New(s.pool)
	return queries.InsertMCPAuditLog(ctx, sqlcgen.InsertMCPAuditLogParams{
		AgentID:         aID,
		ToolName:        toolName,
		RequestPayload:  req,
		ResponsePayload: res,
	})
}

func (s *PGStore) ListMCPAuditLogs(ctx context.Context) ([]McpAuditLog, error) {
	queries := sqlcgen.New(s.pool)
	return queries.ListMCPAuditLogs(ctx)
}
