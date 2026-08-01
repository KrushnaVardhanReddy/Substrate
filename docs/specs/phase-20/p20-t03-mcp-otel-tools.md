# Phase 20 — Task 03: MCP Tools for OTel Traffic Queries

## 1. Overview

Expose three MCP Tools so AI agents (Claude, Cursor) can query live traffic data
during impact analysis, making blast radius assessments quantitative and actionable.

## 2. Tools

### `get_route_traffic`
Returns traffic metrics for a specific route over the last N minutes.

```json
{
  "name": "get_route_traffic",
  "description": "Get live traffic metrics (req/min, p99, error rate) for a specific API route",
  "inputSchema": {
    "type": "object",
    "properties": {
      "org":        { "type": "string" },
      "route":      { "type": "string", "description": "e.g. GET /api/v1/payments" },
      "window_min": { "type": "integer", "default": 60 }
    },
    "required": ["org", "route"]
  }
}
```

### `get_top_traffic_routes`
Returns the top N routes by request volume for an org (useful for blast radius triage).

```json
{
  "name": "get_top_traffic_routes",
  "description": "List the highest-traffic API routes for an org",
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

### `get_traffic_enriched_blast_radius`
Combines the existing `get_blast_radius` tool with live traffic data — returns each
consumer with its traffic weight so the AI can prioritise which consumers to notify first.

```json
{
  "name": "get_traffic_enriched_blast_radius",
  "description": "Get blast radius enriched with live OTel traffic weight per consumer",
  "inputSchema": {
    "type": "object",
    "properties": {
      "org":    { "type": "string" },
      "repo":   { "type": "string" },
      "change": { "type": "string", "description": "brief description of the schema change" }
    },
    "required": ["org", "repo"]
  }
}
```

## 3. Rules

- All tool handlers call `OtelStore` port — never raw DB.
- `get_traffic_enriched_blast_radius` must call both the existing `ImpactStore` and
  the new `OtelStore` and merge results in-memory.
- Register inside the existing `api/internal/mcp/server.go` — no new MCP file.

## 4. Deliverables

1. MODIFY: `api/internal/mcp/server.go` — add 3 new tools + OtelStore dependency
2. `api/internal/mcp/otel_tools_test.go` (CREATE)
3. Run `go test ./api/internal/mcp/...`.
