package checker

import (
	"context"

	"go.opentelemetry.io/otel"
)

// ParseSchema is a helper to wrap parser logic with tracing
func ParseSchema(ctx context.Context, parserFunc func() error) error {
	_, span := otel.Tracer("engine").Start(ctx, "ParseSchema")
	defer span.End()

	return parserFunc()
}

// CalculateDiff is a helper to wrap diff logic with tracing
func CalculateDiff(ctx context.Context, diffFunc func() error) error {
	_, span := otel.Tracer("engine").Start(ctx, "CalculateDiff")
	defer span.End()

	return diffFunc()
}
