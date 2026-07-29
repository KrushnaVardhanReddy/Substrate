-- name: InsertEcosystemEvent :one
INSERT INTO ecosystem_events (
  org, repo, event_type, description, event_time
) VALUES (
  $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetEcosystemEventsByOrg :many
SELECT * FROM ecosystem_events
WHERE org = $1
AND event_time >= $2
AND event_time <= $3
ORDER BY event_time DESC;
