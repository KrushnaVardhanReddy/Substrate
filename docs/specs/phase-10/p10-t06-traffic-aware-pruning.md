# P10-T06: Traffic-Aware Pruning (Zombies)

## Overview
Correlate schema endpoints with live OTel/Datadog metrics to detect unused "zombie" APIs and auto-generate PRs to delete the dead code.

## Requirements
1. **Metrics Ingestion**: Accept OTel metric data (HTTP request counts per endpoint) via a webhook or polling endpoint.
2. **Zombie Detection**: Flag endpoints with zero traffic over the past 30 days as "zombies".
3. **Dashboard View**: Add a "Zombie APIs" tab in the dashboard listing all detected unused endpoints.
4. **Auto-PR**: For confirmed zombies (org-approved), open a PR in the provider repo deleting the dead endpoint code and its spec entry.

## Implementation Status: ✅ Implemented

### Database Schema
Migration `0048_endpoint_traffic.up.sql` creates the `endpoint_traffic` table:

```sql
CREATE TABLE IF NOT EXISTS endpoint_traffic (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    method VARCHAR(10) NOT NULL,
    path VARCHAR(255) NOT NULL,
    last_seen_at TIMESTAMP WITH TIME ZONE,
    request_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(repo_id, method, path)
);
```

### API Endpoints (`api/internal/handlers/otel_webhook.go`)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/v1/otel/webhook` | Service Token | Ingest OTel spans; upserts `endpoint_traffic` rows |
| `GET` | `/api/v1/org/{org}/zombies` | Service Token | Returns endpoints with zero traffic in last 30 days |
| `POST` | `/api/v1/org/{org}/zombies/pr` | Service Token | Creates a draft PR to prune a confirmed zombie endpoint |

### Data Layer (`api/internal/db/pgstore.go`)
- `UpsertEndpointTraffic(ctx, repoID, method, path, timestamp)` — inserts or increments `request_count`, updates `last_seen_at` using `GREATEST()`.
- `GetZeroTrafficEndpoints(ctx, orgName, since)` — joins `endpoint_traffic → repositories → organizations` filtering by `o.github_org_name` and `last_seen_at < since`.

### SQL Queries (`api/internal/db/queries/telemetry.sql`)
```sql
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
```

### E2E Validation
- Covered by `TestPhase10SystemE2E` in `scripts/e2e/phase10_e2e_test.go`:
  - **Scenario 1**: OTel webhook ingestion → expects `202 Accepted`
  - **Scenario 2**: Zombie detection query → expects `200 OK` with valid response
  - **Scenario 4**: Auto-SDK generation trigger → expects `202 Accepted`
