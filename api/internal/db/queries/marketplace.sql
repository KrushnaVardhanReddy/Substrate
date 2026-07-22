-- name: PublishPlugin :one
INSERT INTO marketplace_plugins (name, description, schema_content)
VALUES ($1, $2, $3)
ON CONFLICT (name) DO UPDATE
  SET description    = EXCLUDED.description,
      schema_content = EXCLUDED.schema_content,
      updated_at     = NOW()
RETURNING id;

-- name: ListPlugins :many
SELECT id, name, description, schema_content, created_at, updated_at
FROM marketplace_plugins
ORDER BY name;
