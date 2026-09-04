package ports

import (
	"context"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
	"github.com/google/uuid"
)

type AirbyteStore interface {
	KMSStore
	UpsertAirbyteSource(ctx context.Context, arg sqlcgen.UpsertAirbyteSourceParams) (sqlcgen.AirbyteSource, error)
	ListAirbyteSources(ctx context.Context, org string) ([]sqlcgen.AirbyteSource, error)
	GetAirbyteSource(ctx context.Context, arg sqlcgen.GetAirbyteSourceParams) (sqlcgen.AirbyteSource, error)
	GetAirbyteSourceByID(ctx context.Context, arg sqlcgen.GetAirbyteSourceByIDParams) (sqlcgen.AirbyteSource, error)
	DeleteAirbyteSource(ctx context.Context, arg sqlcgen.DeleteAirbyteSourceParams) error
	BulkInsertAirbyteRecords(ctx context.Context, org string, sourceID uuid.UUID, stream string, records [][]byte) (int, error)
}
