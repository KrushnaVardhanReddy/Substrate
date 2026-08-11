// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package testutils

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func SetupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping database test")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)

	// Create tables if they don't exist
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS partner_integrations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			vendor_name TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			webhook_url TEXT NOT NULL,
			webhook_secret TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		DELETE FROM partner_integrations;
	`)
	require.NoError(t, err)

	cleanup := func() {
		pool.Exec(ctx, "DELETE FROM partner_integrations;")
		pool.Close()
	}

	return pool, cleanup
}
