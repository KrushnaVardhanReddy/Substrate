# Substrate Core Data Contracts

This folder contains the single source of truth for Substrate's core data models. These models are shared across the Go API, Diff Engine, Svelte Frontend, and AI Agents (MCP Tools).

## `DiffReport` (JSON)
Output by the CLI engine (`substrate diff --format json`) and returned by the API (`GET /api/v1/diff`).

```json
{
  "substrate_version": "0.1.0",
  "schema_type": "openapi",
  "compared_at": "2026-07-15T12:00:00Z",
  "mode": "strict",
  "summary": {
    "total_changes": 2,
    "breaking_count": 1,
    "warning_count": 0,
    "safe_count": 1,
    "overall_severity": "BREAKING"
  },
  "breaking_changes": [
    {
      "id": "openapi_field_removed_user_name",
      "rule_id": "FIELD_REMOVED",
      "severity": "BREAKING",
      "path": "User.name",
      "description": "Field 'name' was removed from 'User'",
      "consumer_impacts": [
        "frontend/src/api/user.ts"
      ]
    }
  ],
  "warnings": [],
  "safe_changes": []
}
```

## `GraphNode` & `GraphEdge` (JSON)
Returned by `GET /api/v1/graph`.

```json
{
  "nodes": [
    {
      "id": "github.com/myorg/backend-api",
      "name": "backend-api",
      "type": "provider",
      "status": "safe",
      "predictive_risk_score": 0.05
    }
  ],
  "edges": [
    {
      "source": "github.com/myorg/frontend",
      "target": "github.com/myorg/backend-api",
      "label": "consumes"
    }
  ]
}
```
