package db

import (
	"context"
	"fmt"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *PGStore) UpsertAirbyteSource(ctx context.Context, arg sqlcgen.UpsertAirbyteSourceParams) (sqlcgen.AirbyteSource, error) {
	q := sqlcgen.New(s.pool)
	return q.UpsertAirbyteSource(ctx, arg)
}

func (s *PGStore) ListAirbyteSources(ctx context.Context, org string) ([]sqlcgen.AirbyteSource, error) {
	q := sqlcgen.New(s.pool)
	return q.ListAirbyteSources(ctx, org)
}

func (s *PGStore) GetAirbyteSource(ctx context.Context, arg sqlcgen.GetAirbyteSourceParams) (sqlcgen.AirbyteSource, error) {
	q := sqlcgen.New(s.pool)
	return q.GetAirbyteSource(ctx, arg)
}

func (s *PGStore) GetAirbyteSourceByID(ctx context.Context, arg sqlcgen.GetAirbyteSourceByIDParams) (sqlcgen.AirbyteSource, error) {
	q := sqlcgen.New(s.pool)
	return q.GetAirbyteSourceByID(ctx, arg)
}

func (s *PGStore) DeleteAirbyteSource(ctx context.Context, arg sqlcgen.DeleteAirbyteSourceParams) error {
	q := sqlcgen.New(s.pool)
	return q.DeleteAirbyteSource(ctx, arg)
}

func (s *PGStore) BulkInsertAirbyteRecords(ctx context.Context, org string, sourceID uuid.UUID, stream string, records [][]byte) (int, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := sqlcgen.New(tx)
	var ingested int
	for _, rec := range records {
		_, err := q.InsertAirbyteStagedRecord(ctx, sqlcgen.InsertAirbyteStagedRecordParams{
			Org: org,
			SourceID: pgtype.UUID{
				Bytes: sourceID,
				Valid: true,
			},
			Stream: stream,
			Data:   rec,
		})
		if err != nil {
			return 0, fmt.Errorf("failed to insert record: %w", err)
		}
		ingested++
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("failed to commit tx: %w", err)
	}
	return ingested, nil
}
