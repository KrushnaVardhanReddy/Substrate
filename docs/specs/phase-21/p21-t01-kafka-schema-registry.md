# Phase 21 — Task 01: Kafka / Confluent Schema Registry Live Webhook

## 1. Overview

Integrate Substrate with Confluent Schema Registry (and compatible registries:
Apicurio, AWS Glue Schema Registry) via a webhook receiver. When a Kafka schema
evolves, the registry fires a webhook to Substrate, which runs breaking change
detection + blast radius analysis automatically — no PR needed.

This opens a completely new market segment: event-driven teams using Kafka as their
primary service bus. No competitor governs this today.

## 2. Background & Motivation

Substrate already supports Avro and AsyncAPI schema formats (P1e). What's missing
is *live event-driven triggering* of that analysis. Today a team must push to a VCS
PR to trigger Substrate. Kafka teams evolve schemas via the Schema Registry API —
they never open a PR for schema changes.

By receiving Schema Registry webhooks, Substrate becomes the API governance layer
for the **entire modern data stack**: HTTP APIs (existing) + event streams (Phase 21).

**Architecture decision:** Substrate acts as a webhook *receiver*. Customers configure
their Confluent / Apicurio instance to emit schema change events to
`POST /api/v1/kafka/schema-event`. Substrate does not connect to Kafka brokers directly.

## 3. Technical Requirements

### 3.1 Database Schema

```sql
-- 20260803000001_kafka_schemas.sql

CREATE TABLE kafka_schema_sources (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org          TEXT NOT NULL,
    name         TEXT NOT NULL,           -- e.g. "Confluent Cloud Prod"
    registry_url TEXT NOT NULL,           -- e.g. "https://psrc-xxxx.us-east-1.aws.confluent.cloud"
    token_hash   TEXT NOT NULL,           -- SHA-256 of the inbound webhook token
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org, name)
);

CREATE TABLE kafka_schema_versions (
    id          BIGSERIAL PRIMARY KEY,
    org         TEXT NOT NULL,
    source_id   UUID NOT NULL REFERENCES kafka_schema_sources(id) ON DELETE CASCADE,
    subject     TEXT NOT NULL,     -- Confluent subject, e.g. "payments-value"
    version     INT NOT NULL,
    schema_type TEXT NOT NULL,     -- "AVRO", "PROTOBUF", "JSON"
    schema_raw  TEXT NOT NULL,     -- the raw schema string
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org, subject, version)
);
```

### 3.2 Webhook Receiver Endpoint

```
POST /api/v1/kafka/schema-event
```

**Supported payload formats:**

Confluent Schema Registry webhook:
```json
{
  "subject": "payments-value",
  "version": 5,
  "schemaType": "AVRO",
  "schema": "{\"type\":\"record\",\"name\":\"Payment\",...}"
}
```

Apicurio Registry webhook (CloudEvents format):
```json
{
  "specversion": "1.0",
  "type": "io.apicur.registry.artifact.updated",
  "data": { "artifactId": "payments", "version": "5", "type": "AVRO" }
}
```

Auth: `Authorization: Bearer <WEBHOOK_TOKEN>` matched against `token_hash`.

**Behaviour:**
1. Validate token → look up `kafka_schema_sources` row.
2. Persist the new schema version to `kafka_schema_versions`.
3. Fetch the previous version from DB (version - 1).
4. If previous version exists: run Substrate diff engine (reuse existing Avro/Protobuf
   adapter) against `(prev_schema, new_schema)`.
5. If breaking changes detected: create an `events` row (same table as P17-T06
   deployment events) with `type: "kafka_breaking_change"` + full DiffReport JSON.
6. Broadcast SSE event to any connected dashboard clients.
7. Return `202 Accepted`.

### 3.3 Schema Source Management

```
POST   /api/v1/org/{org}/kafka/sources         — register a schema registry
GET    /api/v1/org/{org}/kafka/sources         — list sources
DELETE /api/v1/org/{org}/kafka/sources/{id}    — remove source
GET    /api/v1/org/{org}/kafka/schemas         — list all tracked subjects + versions
GET    /api/v1/org/{org}/kafka/schemas/{subject}/diff — diff two versions of a subject
```

### 3.4 Kafka Blast Radius

For a detected breaking Kafka schema change, blast radius = which microservices consume
this Kafka topic. Substrate infers this from:
1. Declared consumers in `substrate.yaml` (new field `kafka_consumers`).
2. Existing OTel traffic data if Phase 20 is deployed (consumers emitting spans with
   `messaging.kafka.topic` attribute).

## 4. Hexagonal Architecture Constraints

- Port interface: `api/internal/ports/kafka.go` → `KafkaSchemaStore`.
- sqlc adapter: `api/internal/adapters/kafka_store.go`.
- Schema diff reuses existing `api/internal/diff/` package (Avro + Protobuf adapters).
- HTTP handlers call only `KafkaSchemaStore` port.

## 5. Rules

- Do NOT connect to Kafka brokers. Webhook-only.
- Support at minimum Confluent Schema Registry and Apicurio (CloudEvents) payload formats.
- Breaking change detection reuses existing diff engine — do not duplicate logic.
- Token auth follows same SHA-256 hash pattern as `otel_sources`.

## 6. Deliverables

1. `api/internal/db/schema/20260803000001_kafka_schemas.sql` (CREATE)
2. `api/internal/db/queries/kafka.sql` (CREATE)
3. `api/internal/ports/kafka.go` (CREATE — port interface)
4. `api/internal/adapters/kafka_store.go` (CREATE — sqlc adapter)
5. `api/internal/handlers/kafka_handler.go` (CREATE)
6. `api/internal/handlers/kafka_handler_test.go` (CREATE)
7. MODIFY: `api/cmd/server/main.go` — register new routes
8. Run `sqlc generate` and `go vet ./...`.
