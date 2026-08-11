// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

//go:build wasip1
// +build wasip1

package sql

import (
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
)

func DiffSchemas(base, head *SQLSchema) *report.DiffReport {
	return &report.DiffReport{
		SchemaType:      "sql",
		BreakingChanges: []report.Change{},
	}
}
