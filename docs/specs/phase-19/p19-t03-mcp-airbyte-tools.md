# Phase 19 — Task 03: MCP Tooling for Airbyte Configs

## 1. Overview

Expose Substrate's Airbyte ingestion management surface as MCP Tools, enabling FDAIEs
and AI agents (Claude, Cursor) to configure enterprise data integrations completely
headlessly without touching any UI or CLI.

## 2. Motivation

Substrate's headless MCP server (P15-T14) gives AI agents full control over API
governance. Phase 19 adds data ingestion to that surface. An FDAIE should be able to
say *"Connect our HubSpot CRM to Substrate so blast radius includes real MRR"* in
natural language and have Claude invoke the MCP tools end-to-end.

## 3. MCP Tools to Register

Add the following to `api/internal/mcp/server.go`:

### Tool: `configure_airbyte_source`

```json
{
  "name": "configure_airbyte_source",
  "description": "Registers a new Airbyte data source for an org. Use this to connect a CRM, database, or SaaS app so Substrate can use real revenue data for blast radius calculations.",
  "inputSchema": {
    "type": "object",
    "required": ["org", "name", "connector", "config"],
    "properties": {
      "org":       { "type": "string", "description": "Substrate org slug" },
      "name":      { "type": "string", "description": "Human-readable label, e.g. 'HubSpot CRM'" },
      "connector": { "type": "string", "description": "Airbyte connector ID, e.g. 'source-hubspot'" },
      "config":    { "type": "object", "description": "Connector-specific config (API key, credentials, etc.)" }
    }
  }
}
```

**Handler behaviour:** Calls `POST /api/v1/org/{org}/airbyte/sources` internally and
returns the created source ID and name.

---

### Tool: `trigger_airbyte_sync`

```json
{
  "name": "trigger_airbyte_sync",
  "description": "Triggers an immediate Airbyte sync for a configured source. Returns the number of records ingested and any schema violations detected.",
  "inputSchema": {
    "type": "object",
    "required": ["org", "source_id"],
    "properties": {
      "org":       { "type": "string" },
      "source_id": { "type": "string", "description": "UUID of the configured Airbyte source" },
      "streams":   { "type": "array", "items": { "type": "string" }, "description": "Optional list of stream names to sync. Defaults to all streams." }
    }
  }
}
```

**Handler behaviour:** Issues an HTTP callback to the Airbyte source's configured
webhook trigger URL (stored in the source config). Polls the ingest endpoint for
completion and returns a summary: `{ "records_ingested": 1420, "violations": 3 }`.

---

### Tool: `list_airbyte_sources`

```json
{
  "name": "list_airbyte_sources",
  "description": "Lists all configured Airbyte data sources for an org, including their last sync time and violation count.",
  "inputSchema": {
    "type": "object",
    "required": ["org"],
    "properties": {
      "org": { "type": "string" }
    }
  }
}
```

---

### Tool: `get_airbyte_violations`

```json
{
  "name": "get_airbyte_violations",
  "description": "Returns recent stream schema violations from Airbyte ingestion for a given org and stream. Use before blast radius calculations to verify data quality.",
  "inputSchema": {
    "type": "object",
    "required": ["org"],
    "properties": {
      "org":    { "type": "string" },
      "stream": { "type": "string", "description": "Filter by stream name (optional)" },
      "limit":  { "type": "integer", "default": 20 }
    }
  }
}
```

## 4. Architecture Constraints

- All MCP tool handlers must call the existing Go service/store interfaces via
  `ports.AirbyteStore` — never directly query the DB from the MCP handler.
- Sensitive config fields (API keys) must be redacted in all tool responses:
  replace values for fields named `api_key`, `token`, `secret`, `password` with `"[REDACTED]"`.
- Register all four tools in the existing `mcp.Server.registerTools()` method.

## 5. Deliverables

1. MODIFY: `api/internal/mcp/server.go` — add 4 new tools
2. MODIFY: `api/internal/mcp/server.go` — add `AirbyteStore` dependency injection
3. `api/internal/mcp/airbyte_tools_test.go` (CREATE — unit tests for each tool handler)
4. Run `go test ./api/internal/mcp/...` before committing.

Commit: `"jules: P19-T03 add MCP tools for Airbyte source management"`
Target branch: `feature/p19-t03-mcp-airbyte-tools`
