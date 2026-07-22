-- name: RecordDriftAnomaly :exec
INSERT INTO drift_anomalies (org_name, repo_name, method, path, error_message)
VALUES ($1, $2, $3, $4, $5);

-- name: GetDriftAnomalies :many
SELECT id, org_name, repo_name, method, path, error_message, timestamp
FROM drift_anomalies
WHERE org_name = $1 AND repo_name = $2
ORDER BY timestamp DESC;

-- name: UpsertEndpointTraffic :exec
INSERT INTO endpoint_traffic (repo_id, method, path, last_seen_at, request_count)
VALUES ($1, $2, $3, $4, 1)
ON CONFLICT (repo_id, method, path) DO UPDATE
  SET last_seen_at  = GREATEST(endpoint_traffic.last_seen_at, EXCLUDED.last_seen_at),
      request_count = endpoint_traffic.request_count + 1;

-- name: GetZeroTrafficEndpoints :many
SELECT et.id, et.repo_id, et.method, et.path, et.last_seen_at, et.request_count
FROM endpoint_traffic et
JOIN repositories r ON et.repo_id = r.id
JOIN organizations o ON r.org_id = o.id
WHERE o.github_org_name = $1
  AND (et.last_seen_at IS NULL OR et.last_seen_at < $2);

-- name: GetROIMetrics :one
SELECT
  (SELECT COUNT(*) FROM webhook_events
   WHERE (is_audit_mode = true OR status = 'blocked')
     AND org_name = $1
     AND timestamp > NOW() - INTERVAL '30 days')          AS total_prevented_outages,
  (SELECT COUNT(*) FROM drift_anomalies
   WHERE org_name = $1
     AND timestamp > NOW() - INTERVAL '30 days')          AS total_undocumented_endpoints;
