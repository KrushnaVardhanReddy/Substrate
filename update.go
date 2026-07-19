package db

import (
	"context"
	"encoding/json"
	"time"
)

type DiffReportRecord struct {
	ReportData json.RawMessage
	CreatedAt  time.Time
}
