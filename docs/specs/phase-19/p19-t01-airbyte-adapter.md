# Phase 19 — Task 01: Airbyte Ingestion Adapter (sqlc)

## 1. Overview

Implement a new Go adapter in the Substrate API backend that orchestrates Airbyte's
Connector Development Kit (CDK) protocol to pull normalised record streams from external
data sources (CRMs, databases, third-party APIs) into Substrate's Postgres staging tables.

This adapter is the data-plane foundation for the "Universal CRM Blast Radius" feature:
it allows Substrate to denominate blast radius in real MRR and customer counts sourced
from any of Airbyte's 300+ connectors, not just the hard-coded Stripe and Salesforce
integrations.

## 2. Background & Motivation

Substrate's existing CRM Blast Radius engine (P10-T19) supports Stripe and Salesforce via
point integrations. Enterprise customers frequently use HubSpot, Chargebee, Snowflake, or
internal Postgres databases as their source of truth for revenue data. Building N custom
connectors is not sustainable. By wrapping the Airbyte protocol, Substrate inherits the
entire Airbyte connector ecosystem at the cost of one adapter.

**Architecture decision:** Substrate acts as an Airbyte *destination* (consumer), not as an
Airbyte *orchestration host*. Customers run their own Airbyte instance (or Airbyte Cloud).
Substrate exposes a lightweight HTTP endpoint that Airbyte's destination connector calls
to deliver normalised records. This avoids adding Airbyte's heavyweight Docker Compose
stack as a Substrate runtime dependency.

## 3. Technical Requirements

### 3.1 Database Schema (`api/internal/db/schema/`)

Add a new migration:

```sql
-- 20260801000001_airbyte_ingestion.sql

CREATE TABLE airbyte_sources (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org         TEXT NOT NULL,
    name        TEXT NOT NULL,                  -- human label, e.g. "HubSpot CRM"
    connector   TEXT NOT NULL,                  -- e.g. "source-hubspot"
    config_enc  BYTEA NOT NULL,                 -- AES-256-GCM encrypted JSON config
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org, name)
);

CREATE TABLE airbyte_staged_records (
    id          BIGSERIAL PRIMARY KEY,
    org         TEXT NOT NULL,
    source_id   UUID NOT NULL REFERENCES airbyte_sources(id) ON DELETE CASCADE,
    stream      TEXT NOT NULL,                  -- Airbyte stream name, e.g. "contacts"
    data        JSONB NOT NULL,
    ingested_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_airbyte_staged_records_org_stream
    ON airbyte_staged_records (org, stream, ingested_at DESC);
```

### 3.2 sqlc Queries

Add CRUD queries for `airbyte_sources` and insert/list queries for
`airbyte_staged_records` in `api/internal/db/queries/airbyte.sql`. Run `sqlc generate`
to regenerate type-safe Go methods.

### 3.3 Airbyte Destination Webhook Handler

Implement `api/internal/handlers/airbyte_handler.go`:

```
POST /api/v1/airbyte/ingest
```

**Request body (Airbyte Destination HTTP format):**
```json
{
  "org":    "acme",
  "source": "source-hubspot",
  "stream": "contacts",
  "records": [
    { "customer_id": "cus_123", "email": "bob@acme.com", "mrr": 500 }
  ]
}
```

**Behaviour:**
1. Validate the `org` via `PASETO` internal service token (`Authorization: Bearer <token>`).
2. Look up or create the `airbyte_sources` row for `(org, source)`.
3. Bulk-insert all records into `airbyte_staged_records` in a single transaction.
4. Return `202 Accepted` with `{ "ingested": <count> }`.

### 3.4 Source CRUD REST API

```
POST   /api/v1/org/{org}/airbyte/sources         — create source config
GET    /api/v1/org/{org}/airbyte/sources          — list sources
DELETE /api/v1/org/{org}/airbyte/sources/{id}     — delete source + staged data
```

All routes protected by `authzMW` (org-scoped admin role required for write operations).

### 3.5 Config Encryption

The `config_enc` column stores the connector JSON config encrypted with AES-256-GCM
using the org's KMS-derived key (reuse the BYOK key derivation from P15-T12). This
ensures Airbyte credentials (API keys, OAuth tokens) never rest in plaintext.

## 4. Hexagonal Architecture Constraints

- Define a `AirbyteStore` port interface in `api/internal/ports/airbyte.go`.
- The sqlc-backed implementation lives in `api/internal/adapters/airbyte_store.go`.
- HTTP handlers must only interact with the port interface — never directly with the DB package.

## 5. Rules

- Do NOT import any Airbyte Go SDK. The protocol is simple JSON over HTTP; implement it natively.
- Config encryption is mandatory — reject `POST /sources` if the KMS key is not available.
- All bulk inserts must be wrapped in a single DB transaction; partial ingestion is not acceptable.

## 6. Deliverables

1. `api/internal/db/schema/20260801000001_airbyte_ingestion.sql` (CREATE)
2. `api/internal/db/queries/airbyte.sql` (CREATE)
3. `api/internal/ports/airbyte.go` (CREATE — port interface)
4. `api/internal/adapters/airbyte_store.go` (CREATE — sqlc adapter)
5. `api/internal/handlers/airbyte_handler.go` (CREATE)
6. `api/internal/handlers/airbyte_handler_test.go` (CREATE — table-driven unit tests)
7. MODIFY: `api/cmd/server/main.go` — register new routes
8. Run `sqlc generate` and `go vet ./...` before committing.
