-- name: GetPrunedSchemaCache :one
SELECT *
FROM schema_prune_cache
WHERE org = $1 AND repo = $2 AND intent_hash = $3 AND computed_at > NOW() - INTERVAL '30 minutes';

-- name: UpsertPrunedSchemaCache :exec
INSERT INTO schema_prune_cache (org, repo, intent_hash, pruned_schema, computed_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (org, repo, intent_hash) DO UPDATE
SET pruned_schema = EXCLUDED.pruned_schema,
    computed_at = EXCLUDED.computed_at;
