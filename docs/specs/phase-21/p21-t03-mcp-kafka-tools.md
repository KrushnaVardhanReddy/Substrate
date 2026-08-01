# Phase 21 — Task 03: MCP Tools for Kafka Schema Governance

## 1. Overview

Three MCP Tools for headless Kafka schema governance — allowing AI agents to query
subject history, trigger manual diff comparisons, and list breaking events.

## 2. Tools

### `list_kafka_subjects`
```json
{
  "name": "list_kafka_subjects",
  "description": "List all tracked Kafka schema subjects for an org",
  "inputSchema": {
    "type": "object",
    "properties": {
      "org": { "type": "string" }
    },
    "required": ["org"]
  }
}
```

### `get_kafka_schema_diff`
```json
{
  "name": "get_kafka_schema_diff",
  "description": "Diff two versions of a Kafka schema subject for breaking changes",
  "inputSchema": {
    "type": "object",
    "properties": {
      "org":          { "type": "string" },
      "subject":      { "type": "string" },
      "from_version": { "type": "integer" },
      "to_version":   { "type": "integer" }
    },
    "required": ["org", "subject", "from_version", "to_version"]
  }
}
```

### `get_kafka_breaking_events`
```json
{
  "name": "get_kafka_breaking_events",
  "description": "Get recent Kafka schema breaking change events for an org",
  "inputSchema": {
    "type": "object",
    "properties": {
      "org":   { "type": "string" },
      "limit": { "type": "integer", "default": 20 }
    },
    "required": ["org"]
  }
}
```

## 3. Rules

- All handlers call `KafkaSchemaStore` port only.
- `get_kafka_schema_diff` reuses the existing diff engine — call the same adapter functions used by P1d/P1e, do not duplicate.
- Register inside the existing `api/internal/mcp/server.go`.

## 4. Deliverables

1. MODIFY: `api/internal/mcp/server.go` — add 3 new tools + KafkaSchemaStore dependency
2. `api/internal/mcp/kafka_tools_test.go` (CREATE)
3. Run `go test ./api/internal/mcp/...`.
