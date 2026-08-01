# Phase 19 — Task 02: Dynamic Schema Insurance for Ingested Streams

## 1. Overview

Hook Substrate's existing Schema Insurance rules engine (P15-T10) into incoming Airbyte
data streams so that every ingested record is automatically validated against declared
semantic contracts. Violations are surfaced in real-time as Schema Insurance claims,
denominated in MRR impact using the CRM Blast Radius engine.

## 2. Motivation

Schema Insurance today validates *API schema changes* at PR-time. This task extends it
to validate *data records* at ingestion-time. When Airbyte syncs Salesforce contacts into
Substrate and a required `mrr` field is null, that is a schema violation with a calculable
blast radius — it means Substrate's blast radius calculations for that org are corrupted.
This feature makes that failure explicit and actionable.

## 3. Technical Requirements

### 3.1 Stream Contract Declaration

Extend `substrate.yaml` to allow declaring a stream contract alongside API schemas:

```yaml
airbyte_streams:
  - name: contacts
    source: source-salesforce
    required_fields:
      - customer_id
      - email
      - mrr
    field_types:
      mrr: number
      customer_id: string
```

### 3.2 Stream Contract Store

Add a `stream_contracts` table:

```sql
CREATE TABLE stream_contracts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org         TEXT NOT NULL,
    stream      TEXT NOT NULL,
    contract    JSONB NOT NULL,   -- parsed substrate.yaml airbyte_streams entry
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org, stream)
);
```

### 3.3 Validation Engine (`api/internal/services/stream_validator.go`)

```go
type StreamValidator struct {
    store ports.AirbyteStore
    insurance ports.InsuranceStore
}

func (v *StreamValidator) ValidateRecord(
    ctx context.Context,
    org, stream string,
    record map[string]any,
) ([]StreamViolation, error)
```

**Validation rules applied per record:**
1. **Required field presence** — any field declared in `required_fields` must be non-null.
2. **Type conformance** — field values must match declared `field_types`.
3. **MRR null-check** — if the stream contains an `mrr` field, a null value generates a
   `HIGH` severity violation (it will corrupt blast radius calculations).

### 3.4 Integration with the Ingest Handler

Modify `POST /api/v1/airbyte/ingest` (from P19-T01) to:

1. After bulk-inserting records, asynchronously dispatch each record to `StreamValidator`.
2. On any violations, insert a row into the existing `insurance_claims` table (reuse
   P15-T10 schema) with:
   - `claim_type = "stream_violation"`
   - `source = "airbyte:<stream>"`
   - `severity` mapped from violation severity
   - `details` containing the offending record ID and field name.
3. Trigger the existing SSE broadcast so the UI shows violations in real-time.

### 3.5 API Endpoint

```
GET /api/v1/org/{org}/airbyte/violations?stream=contacts&since=2026-08-01
```

Returns paginated stream violation claims, joining `insurance_claims` with
`airbyte_staged_records` to include the raw offending record.

## 4. Rules

- Validation must be asynchronous — it must NOT block the 202 Accepted response from the
  ingest endpoint. Use a goroutine + channel pattern (not a full job queue).
- Do NOT re-implement the Insurance claim schema — reuse `insurance_claims` table and
  `ports.InsuranceStore` exactly as defined in P15-T10.
- The validator must be unit-testable in isolation (pure function logic, no DB calls inside
  the core validation logic).

## 5. Deliverables

1. `api/internal/db/schema/20260801000002_stream_contracts.sql` (CREATE)
2. `api/internal/db/queries/stream_contracts.sql` (CREATE)
3. `api/internal/services/stream_validator.go` (CREATE)
4. `api/internal/services/stream_validator_test.go` (CREATE — table-driven, no DB)
5. MODIFY: `api/internal/handlers/airbyte_handler.go` — add async validation dispatch
6. MODIFY: `api/internal/handlers/airbyte_handler.go` — add `GET .../violations` endpoint
7. Run `go test ./api/internal/services/...` before committing.
