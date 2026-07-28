-- name: InsertSchemaValidationGap :exec
INSERT INTO schema_validation_gaps (method, path, payload, issue, severity)
VALUES ($1, $2, $3, $4, $5);

-- name: GetSchemaValidationGaps :many
SELECT * FROM schema_validation_gaps ORDER BY created_at DESC;
