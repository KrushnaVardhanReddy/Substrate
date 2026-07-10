# Substrate — AsyncAPI & Apache Avro Adapter Specification (Phase 1e)

> **Status:** APPROVED ✅
> **Spec-First Gate:** Jules MUST NOT implement Phase 1e until this document is approved.
> **Scope:** Add AsyncAPI 2.x/3.x and Apache Avro schema breaking-change detection to the Substrate diff engine.
> **Depends on:** `docs/specs/phase-1/diff-report-schema.md` (the DiffReport contract)

---

## Problem Statement

Event-driven architectures (Kafka, NATS, RabbitMQ) rely on AsyncAPI to describe message channels and payloads. Avro is the dominant serialisation format for Kafka at scale. Unlike REST APIs, breaking changes here are silent — a producer sends a message with a renamed field and every consumer silently starts receiving `null` until someone files a bug report two hours later.

Substrate Phase 1e adds protection for two formats:
1. **AsyncAPI 2.x/3.x** — uses `asyncapi/parser-go` to parse, custom rules for diff
2. **Apache Avro** — delegates compatibility checking to the Confluent/Apicurio Schema Registry API; Substrate only wraps the JSON response

---

## Part A: AsyncAPI

### Library Decision: `asyncapi/parser-go` ✅

**Library:** `github.com/asyncapi/parser-go` — Apache 2.0 license.

**Rationale:**
- Official Go AsyncAPI parser maintained by the AsyncAPI Initiative.
- Supports AsyncAPI 2.x and 3.x spec versions.
- Returns a structured `Document` object — no hand-rolled YAML parsing needed.
- Single binary promise maintained — pure Go library, no subprocess.
- No GPL or commercial license concerns.

**Install:** `go get github.com/asyncapi/parser-go`

---

### AsyncAPI Internal Representation (IR)

Before diffing, both base and head specs are parsed into this IR:

```go
type AsyncAPISpec struct {
    Version  string              // "2.6.0", "3.0.0"
    Channels map[string]Channel  // key = channel address e.g. "user/signedup"
    Servers  map[string]Server   // key = server name
}

type Channel struct {
    Address    string
    Publish    *Operation // AsyncAPI 2.x
    Subscribe  *Operation // AsyncAPI 2.x
    Send       *Operation // AsyncAPI 3.x
    Receive    *Operation // AsyncAPI 3.x
    Bindings   map[string]any
}

type Operation struct {
    Message *Message
}

type Message struct {
    Name        string
    ContentType string         // e.g. "application/json"
    Payload     map[string]any // raw JSON Schema object
}

type Server struct {
    URL      string
    Protocol string // kafka, mqtt, amqp, nats, etc.
}
```

---

### AsyncAPI Breaking Change Rules

#### Rule Table (18 rules: 10 BREAKING, 5 WARNING, 3 SAFE)

| Rule ID | Severity | Pattern | Rationale |
|---|---|---|---|
| `ASYNCAPI_CHANNEL_REMOVED` | BREAKING | A channel key is removed from `channels` | All producers/consumers subscribed to that channel lose their contract. |
| `ASYNCAPI_CHANNEL_ADDRESS_CHANGED` | BREAKING | The `address` field of a channel changes | Topic/queue name changes break all existing subscriptions. Wire-level break. |
| `ASYNCAPI_OPERATION_REMOVED` | BREAKING | A `publish`/`subscribe` (v2) or `send`/`receive` (v3) operation is removed from a channel | Consumers subscribed to that operation lose their binding. |
| `ASYNCAPI_MESSAGE_PAYLOAD_FIELD_REMOVED` | BREAKING | A field is removed from a message's JSON Schema payload | Consumers deserialising this field will receive `null` or throw. |
| `ASYNCAPI_MESSAGE_PAYLOAD_FIELD_TYPE_CHANGED` | BREAKING | A field's `type` changes in a message payload | Type mismatch at deserialisation — consumers will fail or corrupt data. |
| `ASYNCAPI_MESSAGE_PAYLOAD_REQUIRED_ADDED` | BREAKING | A field is added to the `required` array of a message payload | Existing producers that omit this field will now produce invalid messages. |
| `ASYNCAPI_MESSAGE_PAYLOAD_FIELD_RENAMED` | BREAKING | A field key is removed and a similar key appears (similarity > 0.7 Levenshtein) | Consumer code referencing the old name breaks. Treat as removal + addition. |
| `ASYNCAPI_MESSAGE_CONTENT_TYPE_CHANGED` | BREAKING | The `contentType` of a message changes (e.g. `application/json` → `avro/binary`) | Consumer deserializers are format-specific — incompatible change. |
| `ASYNCAPI_SERVER_REMOVED` | BREAKING | A server entry is removed from `servers` | Clients configured for that server URL will fail to connect. |
| `ASYNCAPI_SERVER_PROTOCOL_CHANGED` | BREAKING | The `protocol` of a server changes (e.g. `kafka` → `mqtt`) | Client libraries are protocol-specific. Existing connections break. |
| `ASYNCAPI_CHANNEL_ADDED` | SAFE | A new channel is added to `channels` | Additive. No existing consumer is affected. |
| `ASYNCAPI_OPERATION_ADDED` | SAFE | A new operation is added to an existing channel | Additive. No existing consumer is affected. |
| `ASYNCAPI_MESSAGE_PAYLOAD_FIELD_ADDED_OPTIONAL` | SAFE | A new field is added to a message payload without adding it to `required` | Consumers using strict deserializers may warn, but they won't crash. |
| `ASYNCAPI_MESSAGE_PAYLOAD_REQUIRED_REMOVED` | WARNING | A field is removed from the `required` array (field still exists) | Producers may start omitting this field. Strict consumers expecting it may break. |
| `ASYNCAPI_MESSAGE_PAYLOAD_ENUM_VALUE_REMOVED` | BREAKING | A value is removed from an enum in a message payload | Producers sending that value will produce invalid messages; consumers crash. |
| `ASYNCAPI_MESSAGE_PAYLOAD_ENUM_VALUE_ADDED` | WARNING | A new value is added to an enum in a message payload | Strict consumer `switch` statements will hit `default` unexpectedly. |
| `ASYNCAPI_CHANNEL_DEPRECATED` | WARNING | A channel gains `x-deprecated: true` or `deprecated: true` | Not immediately breaking; consumers should be notified. |
| `ASYNCAPI_SERVER_URL_CHANGED` | WARNING | The URL of a server changes | Clients must reconfigure; old URL may still work during migration. |

---

### AsyncAPI Implementation: `engine/internal/diff/asyncapi.go`

```go
// CompareAsyncAPI compares two AsyncAPI spec files and returns a DiffReport.
// baseFile and headFile are paths to AsyncAPI YAML or JSON files.
func CompareAsyncAPI(baseFile, headFile string) (*report.DiffReport, error)
```

**Implementation flow:**
1. Read and parse `baseFile` and `headFile` using `asyncapi/parser-go`
2. Extract IR (channels, servers, messages) from both documents
3. Diff channels: check for removed/added/modified channels
4. Diff operations: check for removed/added operations per channel
5. Diff message payloads: apply JSON Schema diff rules
6. Diff servers: check for removed/changed servers
7. Build `report.DiffReport{SchemaType: "asyncapi"}` and return

---

### AsyncAPI Test Requirements

**File:** `engine/internal/diff/asyncapi_test.go` (package `diff_test`)

All 12 mandatory test cases (table-driven, skip if parse fails gracefully):

| # | Test Case | Expected |
|---|---|---|
| 1 | No changes — identical specs | 0 BREAKING, 0 WARNING |
| 2 | Channel removed | 1 BREAKING: `ASYNCAPI_CHANNEL_REMOVED` |
| 3 | Channel address changed | 1 BREAKING: `ASYNCAPI_CHANNEL_ADDRESS_CHANGED` |
| 4 | Message payload field removed | 1 BREAKING: `ASYNCAPI_MESSAGE_PAYLOAD_FIELD_REMOVED` |
| 5 | Message payload field type changed | 1 BREAKING: `ASYNCAPI_MESSAGE_PAYLOAD_FIELD_TYPE_CHANGED` |
| 6 | Required field added to payload | 1 BREAKING: `ASYNCAPI_MESSAGE_PAYLOAD_REQUIRED_ADDED` |
| 7 | Enum value removed from payload | 1 BREAKING: `ASYNCAPI_MESSAGE_PAYLOAD_ENUM_VALUE_REMOVED` |
| 8 | New channel added | 0 BREAKING (safe) |
| 9 | New optional field added to payload | 0 BREAKING (safe) |
| 10 | Server removed | 1 BREAKING: `ASYNCAPI_SERVER_REMOVED` |
| 11 | Server protocol changed | 1 BREAKING: `ASYNCAPI_SERVER_PROTOCOL_CHANGED` |
| 12 | Channel deprecated | 0 BREAKING, 1 WARNING: `ASYNCAPI_CHANNEL_DEPRECATED` |

**Testdata:** `engine/internal/diff/testdata/asyncapi/`
Use AsyncAPI 2.6.0 format for all fixtures. Use `user/signedup` as the canonical test channel.

Sample base fixture `base_no_change/asyncapi.yaml`:
```yaml
asyncapi: "2.6.0"
info:
  title: User Events
  version: "1.0.0"
channels:
  user/signedup:
    subscribe:
      message:
        payload:
          type: object
          required: [id, email]
          properties:
            id:
              type: string
            email:
              type: string
```

---

## Part B: Apache Avro

### Library Decision: Schema Registry Compatibility API ✅

**Approach:** Do NOT build a custom Avro schema parser or compatibility checker.
Call the Confluent Schema Registry (or Apicurio Registry) **REST compatibility API**.

**Rationale:**
- Avro schema evolution rules (union promotion, field defaults, ordering, type widening) are extremely complex. The Schema Registry has been implementing and testing these rules for years.
- Building a custom Avro checker would require deep knowledge of Avro's full compatibility matrix (`BACKWARD`, `FORWARD`, `FULL`, `NONE`) and would take months to get right.
- The Schema Registry REST API (`/compatibility/subjects/{subject}/versions/latest`) is the industry standard for Avro compatibility checks.
- **Single-binary promise:** No subprocess, no new Go library. Pure HTTP call using `net/http`.
- If no Schema Registry URL is configured, Substrate skips Avro compatibility checking and returns an informational message.

**Registry API used:**
```
POST /compatibility/subjects/{subject}/versions/latest
Authorization: Basic <user>:<password>  (optional)
Content-Type: application/vnd.schemaregistry.v1+json

{"schema": "<avro_schema_json_string>"}
```

Response:
```json
{"is_compatible": true}
```
or
```json
{"is_compatible": false, "messages": ["...reason..."]}
```

---

### Avro Breaking Change Rule

| Rule ID | Severity | Pattern | Rationale |
|---|---|---|---|
| `AVRO_INCOMPATIBLE` | BREAKING | Schema Registry returns `{"is_compatible": false}` | The Avro compatibility level (BACKWARD/FORWARD/FULL) is violated. |
| `AVRO_REGISTRY_UNAVAILABLE` | WARNING | Cannot reach Schema Registry URL | Registry may be temporarily down — warn but don't block. |
| `AVRO_NO_REGISTRY_CONFIGURED` | WARNING | `schema_registry_url` not set in `substrate.yaml` | User has not configured the registry; Avro check is skipped. |

---

### Avro `substrate.yaml` Config

Add a new optional config key (extend the existing `SubstrateConfig` struct):

```yaml
service: payments
schema_type: avro
spec_path: schemas/payment.avsc   # path to Avro schema file (.avsc or .json)

# New optional section for Avro
avro:
  schema_registry_url: https://registry.mycompany.com
  subject: payments-value           # Schema Registry subject name
  username: substrateci             # optional Basic Auth
  password: ${SCHEMA_REGISTRY_TOKEN} # env var expansion not needed — read from env
```

**If `avro.schema_registry_url` is not set:**
- Substrate returns a `DiffReport` with 1 WARNING: `AVRO_NO_REGISTRY_CONFIGURED`
- Message: `"Avro compatibility checking requires schema_registry_url in substrate.yaml"`
- Does NOT fail the build (WARNING does not block merge by default)

---

### Avro Implementation: `engine/internal/diff/avro.go`

```go
// CompareAvro checks Avro schema compatibility against a configured Schema Registry.
// baseFile is ignored (the registry holds the current version).
// headFile is the new .avsc schema file to check for compatibility.
// cfg is the parsed substrate.yaml config (may contain avro.schema_registry_url).
func CompareAvro(baseFile, headFile string, cfg *config.SubstrateConfig) (*report.DiffReport, error)
```

**Implementation flow:**
1. Read `cfg.Avro.SchemaRegistryURL` — if empty, return WARNING report `AVRO_NO_REGISTRY_CONFIGURED`
2. Read the head Avro schema from `headFile` (raw JSON string — .avsc is JSON)
3. Construct the subject: use `cfg.Avro.Subject` if set, else derive from `cfg.Service + "-value"`
4. POST to `{registry_url}/compatibility/subjects/{subject}/versions/latest`
   - Body: `{"schema": "<escaped head schema json>"}`
   - Optional Basic Auth if username/password set
   - 10 second timeout
5. Parse response:
   - `is_compatible: true` → return DiffReport with 0 breaking changes
   - `is_compatible: false` → return DiffReport with 1 BREAKING change: `AVRO_INCOMPATIBLE`
   - HTTP error / timeout → return DiffReport with 1 WARNING: `AVRO_REGISTRY_UNAVAILABLE`

---

### Avro Config Extension: `engine/internal/config/config.go`

Add to `SubstrateConfig`:

```go
type AvroConfig struct {
    SchemaRegistryURL string `yaml:"schema_registry_url"`
    Subject           string `yaml:"subject"`            // optional
    Username          string `yaml:"username"`           // optional Basic Auth
    Password          string `yaml:"password"`           // optional Basic Auth
}

type SubstrateConfig struct {
    // ... existing fields ...
    Avro *AvroConfig `yaml:"avro,omitempty"`
}
```

---

### Avro Test Requirements

**File:** `engine/internal/diff/avro_test.go` (package `diff_test`)

All tests mock the HTTP call — do NOT make real network requests.

| # | Test Case | Mock Response | Expected |
|---|---|---|---|
| 1 | No registry URL configured | (no HTTP) | 1 WARNING: `AVRO_NO_REGISTRY_CONFIGURED` |
| 2 | Registry returns compatible | `{"is_compatible": true}` | 0 BREAKING, DiffReport.SchemaType = "avro" |
| 3 | Registry returns incompatible | `{"is_compatible": false, "messages": ["removed field"]}` | 1 BREAKING: `AVRO_INCOMPATIBLE` |
| 4 | Registry HTTP 500 | server error | 1 WARNING: `AVRO_REGISTRY_UNAVAILABLE` |
| 5 | Registry connection refused | network error | 1 WARNING: `AVRO_REGISTRY_UNAVAILABLE` |

**Strategy for mocking:** Use `httptest.NewServer()` to spin up a local mock server in each test. Pass its URL as `cfg.Avro.SchemaRegistryURL`.

---

## CLI Integration

### `substrate` CLI — New dispatch cases

In `engine/cmd/substrate/main.go`, add to the schema_type dispatch:

```go
case "asyncapi":
    diffReport, err = diff.CompareAsyncAPI(baseSchema, headSchema)

case "avro":
    diffReport, err = diff.CompareAvro(baseSchema, headSchema, cfg)
```

### HTTP Serve Mode

In `engine/cmd/substrate/serve.go` `/diff` endpoint, add the same two cases.
Also update the validation guard to accept `"asyncapi"` and `"avro"` as valid schema types.

---

## Files to Create / Modify

| File | Action |
|---|---|
| `docs/specs/phase-1/asyncapi-avro-adapter.md` | ✅ This file (spec) |
| `engine/internal/diff/asyncapi.go` | CREATE — AsyncAPI adapter |
| `engine/internal/diff/asyncapi_test.go` | CREATE — 12 test cases |
| `engine/internal/diff/testdata/asyncapi/*/asyncapi.yaml` | CREATE — fixture files |
| `engine/internal/diff/avro.go` | CREATE — Avro adapter |
| `engine/internal/diff/avro_test.go` | CREATE — 5 test cases |
| `engine/internal/config/config.go` | MODIFY — add AvroConfig struct + **add `"asyncapi"` and `"avro"` to schema_type validator** |
| `engine/internal/config/config_test.go` | MODIFY — add Avro config parse test |
| `engine/cmd/substrate/main.go` | MODIFY — add asyncapi + avro cases |
| `engine/cmd/substrate/serve.go` | MODIFY — add asyncapi + avro cases |
| `engine/go.mod` | MODIFY — add `github.com/asyncapi/parser-go` |

---

## Out of Scope for Phase 1e

- AsyncAPI 3.x server bindings (protocol-specific config) — future task
- Avro `FORWARD_TRANSITIVE` / `FULL_TRANSITIVE` compatibility levels — registry handles this, no Substrate change needed
- Avro schema inference from code annotations — future AI task
- Kafka Connect transforms — separate schema format
- Protobuf-based AsyncAPI message payloads — handled by Phase 1d

---

## Jules Prompt Locations

- `prompts/phase-1e-asyncapi-avro/t01_asyncapi_adapter.txt` — AsyncAPI parser + rules
- `prompts/phase-1e-asyncapi-avro/t02_avro_adapter.txt` — Avro Schema Registry adapter

> These two prompts can be submitted to Jules in parallel — they modify different files.
> T01 adds `engine/internal/diff/asyncapi.go`.
> T02 adds `engine/internal/diff/avro.go` and modifies `engine/internal/config/config.go`.
