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
