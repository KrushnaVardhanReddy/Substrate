# Phase 20 — Task 01: OTel/APM Ingest Receiver

## 1. Overview

Implement a lightweight OpenTelemetry (OTel) metric and span receiver in the Substrate
Go backend. The receiver accepts OTLP/HTTP payloads from any OTel-compatible APM tool
(Datadog Agent, Grafana Agent, Prometheus, OpenTelemetry Collector) and maps incoming
traffic data to Substrate's known API routes and schemas.

This enriches the blast radius calculation with **real traffic weight**, turning
Substrate's "which services theoretically depend on this schema" into "which services
are *actively calling this endpoint* with N req/min and P99 latency of Xms".

## 2. Background & Motivation

Today Substrate's Blast Radius is binary: a consumer repo either declares a dependency
or it doesn't. This produces false positives (zombie paths that nobody calls) and
underweights high-traffic paths. Enterprise customers need to know: "If I break
`GET /api/v1/payments`, which consumers have > 1000 req/min against it?"

OTel is the de-facto standard telemetry protocol. By implementing an OTLP/HTTP receiver,
Substrate becomes compatible with the *entire* observability ecosystem without requiring
customers to add any new infrastructure — they just redirect their existing OTel collector
export target to point at Substrate.

## 3. Technical Requirements

### 3.1 Database Schema (`api/internal/db/schema/`)

```sql
-- 20260802000001_otel_traffic.sql

CREATE TABLE otel_route_traffic (
    id           BIGSERIAL PRIMARY KEY,
    org          TEXT NOT NULL,
    route        TEXT NOT NULL,      -- e.g. "GET /api/v1/payments"
    service      TEXT NOT NULL,      -- calling service name (from OTel resource attrs)
    req_count    BIGINT NOT NULL DEFAULT 0,
    error_count  BIGINT NOT NULL DEFAULT 0,
    p99_ms       DOUBLE PRECISION,   -- p99 latency
    window_start TIMESTAMPTZ NOT NULL,
    window_end   TIMESTAMPTZ NOT NULL,
    recorded_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_otel_route_traffic_org_route
    ON otel_route_traffic (org, route, window_end DESC);

CREATE TABLE otel_sources (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org        TEXT NOT NULL UNIQUE,
    token_hash TEXT NOT NULL,   -- SHA-256 of the ingestion token (never store plaintext)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### 3.2 OTLP/HTTP Receiver Endpoint

```
POST /api/v1/otel/v1/metrics
POST /api/v1/otel/v1/traces
```

Accept `Content-Type: application/x-protobuf` (OTLP binary) **and**
`Content-Type: application/json` (OTLP JSON — simpler for testing).

Authentication: Bearer token validated against `otel_sources.token_hash`.

**Extraction logic for metrics:**
- Look for `http.route`, `http.method`, `http.status_code` attributes on spans/metrics.
- Aggregate into 1-minute windows.
- Upsert into `otel_route_traffic`.

**Extraction logic for traces:**
- Parse `span.kind = SERVER` spans.
- Extract `service.name` from resource attributes.
- Extract `http.route` + `http.method` from span attributes.

### 3.3 Traffic-Enriched Blast Radius API

Extend `GET /api/v1/impact/{org}/{repo}` to join `otel_route_traffic`:

```json
{
  "blast_radius": [
    {
      "consumer": "checkout-service",
      "route": "GET /api/v1/payments",
      "traffic_weight": {
        "req_per_min": 1420,
        "p99_ms": 87,
        "error_rate_pct": 0.3
      },
      "risk": "CRITICAL"
    }
  ]
}
```

Add `risk` field: `CRITICAL` if `req_per_min > 100`, `HIGH` if `> 10`, else `MEDIUM`.

### 3.4 Traffic Heatmap SSE Stream

```
GET /api/v1/org/{org}/otel/traffic/stream   (text/event-stream)
```

Broadcast real-time traffic updates as OTel data arrives. Dashboard subscribes to
colour-code graph nodes by live traffic volume.

### 3.5 OTel Source Management

```
POST   /api/v1/org/{org}/otel/sources         — provision ingestion token
GET    /api/v1/org/{org}/otel/sources         — list sources + last seen
DELETE /api/v1/org/{org}/otel/sources         — revoke token
GET    /api/v1/org/{org}/otel/traffic         — paginated traffic history
```

## 4. Hexagonal Architecture Constraints

- Port interface: `api/internal/ports/otel.go` → `OtelStore`.
- sqlc adapter: `api/internal/adapters/otel_store.go`.
- OTLP protobuf parsing: use `go.opentelemetry.io/proto/otlp` — the official Go proto
  bindings. Do NOT shell out to `otelcol` binary.
- HTTP handlers call only `OtelStore` port interface.

## 5. Rules

- Never store raw spans or full trace payloads — only the aggregated route-level metrics.
  This keeps storage bounded and avoids PII from request bodies leaking into Substrate.
- Token hashing: SHA-256 only, never store plaintext. Pattern mirrors `otel_sources.token_hash`.
- Upsert logic: use `INSERT ... ON CONFLICT DO UPDATE` for the 1-minute window aggregation.
- The OTLP JSON fallback path is required for E2E tests (no protobuf marshaling in tests).

## 6. Deliverables

1. `api/internal/db/schema/20260802000001_otel_traffic.sql` (CREATE)
2. `api/internal/db/queries/otel.sql` (CREATE)
3. `api/internal/ports/otel.go` (CREATE — port interface)
4. `api/internal/adapters/otel_store.go` (CREATE — sqlc adapter)
5. `api/internal/handlers/otel_handler.go` (CREATE — OTLP receiver + source CRUD)
6. `api/internal/handlers/otel_handler_test.go` (CREATE)
7. MODIFY: `api/internal/handlers/impact_handler.go` — join traffic weight
8. MODIFY: `api/cmd/server/main.go` — register new routes
9. Run `sqlc generate` and `go vet ./...`.
