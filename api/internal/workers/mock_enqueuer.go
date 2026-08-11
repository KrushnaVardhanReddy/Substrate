// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package workers

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

type MockJobEnqueuer struct {
	InsertFunc func(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error)
}

func (m *MockJobEnqueuer) Insert(ctx context.Context, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	if m.InsertFunc != nil {
		return m.InsertFunc(ctx, args, opts)
	}
	return &rivertype.JobInsertResult{}, nil
}

func (m *MockJobEnqueuer) InsertTx(ctx context.Context, tx pgx.Tx, args river.JobArgs, opts *river.InsertOpts) (*rivertype.JobInsertResult, error) {
	if m.InsertFunc != nil {
		return m.InsertFunc(ctx, args, opts)
	}
	return &rivertype.JobInsertResult{}, nil
}
