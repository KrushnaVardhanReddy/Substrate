# Spec: P12-T10 — Local Demo Repository Testing Suite

## 1. Overview
Validate that all 7 demo repositories can be tested entirely locally — without GitHub webhooks, GitHub App tokens, or ngrok tunnels. The test suite bypasses the Cloudflare Worker and hits the Go API + Go Engine directly, proving the full sync → diff → cross-repo-check pipeline works for every supported schema type.

## 2. Problem Statement
The current E2E test suite (`scripts/e2e/`) requires a live GitHub token, creates real repos, and opens real PRs. This makes it slow, flaky, and impossible to run during a conference demo without internet access. A local-only test suite enables:
- Instant feedback during development
- Demo rehearsal without GitHub dependency
- CI/CD pipeline testing in air-gapped environments

## 3. Architecture

### Bypass Strategy
```
┌─────────────┐     ┌──────────────┐     ┌───────────────┐
│  GitHub App  │     │   Go API     │     │  Go Engine    │
│  (Worker)    │────▶│  :8090       │────▶│  :8080        │
└─────────────┘     └──────────────┘     └───────────────┘
       │                    ▲                    ▲
       │                    │                    │
       ▼                    │                    │
  ┌─────────┐        ┌──────┴───────┐    ┌──────┴──────┐
  │ GitHub   │        │ local-test.sh│    │ Engine API  │
  │ Webhooks │        │ (bypasses    │    │ /diff       │
  └─────────┘        │  worker)     │    └─────────────┘
                     └──────────────┘
```

The test suite makes direct HTTP calls:
1. `POST http://localhost:8080/diff` — Engine diff (schema comparison)
2. `POST http://localhost:8090/api/v1/sync` — API sync (dependency graph seeding)
3. `POST http://localhost:8090/api/v1/cross-repo-check` — Cross-repo blast radius
4. `POST http://localhost:8090/api/v1/diff` — Diff report storage
5. `GET http://localhost:8090/api/v1/graph/{org}` — Graph verification

## 4. Test Matrix (7 Demo Repos × 5 Schema Types)

| # | Demo Repo | Schema Type | Spec File | Breaking Scenario | Safe Scenario |
|---|-----------|-------------|-----------|-------------------|---------------|
| 1 | microservices-demo | protobuf | `protos/demo.proto` | Remove `CartItem.product_id` field | Add `string notes = 3` to `CartItem` |
| 2 | graphql-schema | graphql | `schema.graphql` | Remove `email` field from `User` | Add `age: Int` to `User` |
| 3 | openapi (Stripe) | openapi | `openapi/spec3.yaml` | Remove `/v1/charges` endpoint | Add `/v2/beta/charges` endpoint |
| 4 | openai-openapi | openapi | `openapi.yaml` | Remove `function_call` from `ChatCompletionRequest` | Add `metadata: Map` to `ChatCompletionRequest` |
| 5 | jaffle_shop | sql | `models/customers.sql` | Drop `customer_lifetime_value` column | Add `age INTEGER` nullable column |
| 6 | slack-api-specs | asyncapi | `events-api/slack_events_api_async_v1.json` | Remove `channel_id` from event payload | Add `thread_ts` optional field |
| 7 | realworld | openapi | `specs/` (any OpenAPI spec) | Remove `/api/articles` endpoint | Add `/api/tags` endpoint |

## 5. Files Created

### Scripts
- `scripts/local-test.sh` — Main orchestrator (runs all 3 layers)
- `scripts/test-engine.sh` — Layer 1: Engine diff tests for all schema types
- `scripts/test-sync.sh` — Layer 2: API sync + dependency graph seeding
- `scripts/test-cross-repo.sh` — Layer 3: Cross-repo blast radius verification

### Test Payloads
- `scripts/test-payloads/protobuf/base.proto`
- `scripts/test-payloads/protobuf/head-breaking.proto`
- `scripts/test-payloads/protobuf/head-safe.proto`
- `scripts/test-payloads/graphql/base.graphql`
- `scripts/test-payloads/graphql/head-breaking.graphql`
- `scripts/test-payloads/graphql/head-safe.graphql`
- `scripts/test-payloads/openapi/base.yaml`
- `scripts/test-payloads/openapi/head-breaking.yaml`
- `scripts/test-payloads/openapi/head-safe.yaml`
- `scripts/test-payloads/sql/base.sql`
- `scripts/test-payloads/sql/head-breaking.sql`
- `scripts/test-payloads/sql/head-safe.sql`
- `scripts/test-payloads/asyncapi/base.json`
- `scripts/test-payloads/asyncapi/head-breaking.json`
- `scripts/test-payloads/asyncapi/head-safe.json`
- `scripts/test-payloads/microservices/substrate.yaml`

## 6. Script Requirements

### `local-test.sh` (Main Orchestrator)
```
Usage: ./scripts/local-test.sh [--layer=1|2|3|all] [--schema=openapi|protobuf|graphql|sql|asyncapi|all]

Options:
  --layer    Which test layer to run (default: all)
  --schema   Filter by schema type (default: all)
  --verbose  Print full curl output and response bodies
  --reset    Run `make reset-demo` before testing

Exit codes:
  0 = All tests passed
  1 = One or more tests failed
  2 = Prerequisites not met (services not running)
```

**Prerequisite checks (fail fast):**
1. `curl -s http://localhost:8090/health` must return `{"status":"ok"}`
2. `curl -s http://localhost:8080/` must return non-error (engine running)
3. Postgres container `substrate-postgres` must be running

### `test-engine.sh` (Layer 1: Engine Diff)
For each schema type, sends `POST /diff` to the engine with base + head schemas and validates:
- Response is HTTP 200
- Response JSON contains `breaking_changes` array
- Breaking scenario: `breaking_count > 0`
- Safe scenario: `breaking_count == 0`
- Correct `rule_id` is present (e.g., `ENDPOINT_REMOVED`, `PROTO_FIELD_TYPE_CHANGED`, `GQL_FIELD_REMOVED`, `COLUMN_REMOVED`, `ASYNCAPI_CHANNEL_REMOVED`)

### `test-sync.sh` (Layer 2: API Sync)
Seeds the dependency graph by calling `POST /api/v1/sync` for each demo repo:
- Creates org `local-demo-testing` with installation_id `99999`
- Creates provider repos (one per demo repo)
- Creates consumer repos (linked to providers via dependencies)
- Verifies via `GET /api/v1/graph/local-demo-testing` that:
  - All expected repos appear as nodes
  - All expected dependencies appear as edges
  - Node names are non-empty (validates JSON tags fix)

### `test-cross-repo.sh` (Layer 3: Cross-Repo Blast Radius)
After syncing, calls `POST /api/v1/cross-repo-check` with a breaking schema change and verifies:
- Response contains `broken_consumers > 0`
- Each broken consumer has correct `consumer_repo` and `rule_id`
- No 402 Payment Required (tier limits bypass in development mode)

## 7. Payload Design Rules
1. **Self-contained:** Each payload is a complete, valid schema file (not a fragment).
2. **Minimal:** Schemas contain only enough structure to trigger the relevant breaking change rule. No more than 30 lines per file.
3. **Deterministic:** Same input always produces same output. No timestamps, no randomness.
4. **Realistic:** Uses field names and types that mirror the actual demo repos (e.g., `CartItem`, `User`, `/users/{id}`).

## 8. Success Criteria
- `./scripts/local-test.sh` exits 0 with all tests passing.
- Each schema type produces at least 1 breaking change and 1 safe change.
- The dependency graph after sync contains all 7 demo repos with correct edges.
- Cross-repo check returns blast radius for microservices-demo (4 downstream consumers).
- Total runtime under 30 seconds (no network calls, no sleeps).
- No changes to any Go or TypeScript source code — only new scripts and payloads.
