# Phase 16: Business Metadata & Node Branding (P16-T02 & P16-T05)

## Overview
Currently, `substrate.yaml` configures technical details for API consumer/provider relationships. However, to act as a true Developer Portal (SSOT), we need to display organizational business context directly on the API docs page and within the Dependency Graph.

This task introduces an optional `metadata` block to `substrate.yaml` and propagates it to both the API Documentation UI and the Cytoscape/Svelte Flow graph nodes.

## Technical Requirements

### 1. Config Parser Update
Update `api/internal/config/config.go` to parse the new `metadata` block in `substrate.yaml`.
```go
type Config struct {
    Metadata *BusinessMetadata `yaml:"metadata,omitempty"`
    // ... existing fields ...
}

type BusinessMetadata struct {
    Owner        string `yaml:"owner,omitempty" json:"owner,omitempty"`
    SlackChannel string `yaml:"slack_channel,omitempty" json:"slack_channel,omitempty"`
    PagerDuty    string `yaml:"pagerduty,omitempty" json:"pagerduty,omitempty"`
    PM           string `yaml:"pm,omitempty" json:"pm,omitempty"`
    SLATier      string `yaml:"sla_tier,omitempty" json:"sla_tier,omitempty"`
    NodeColor    string `yaml:"node_color,omitempty" json:"node_color,omitempty"`
}
```

### 2. Database Sync
Update the database schema if necessary, or simply ensure that when `repositories` (or equivalent) are synced via `api/internal/webhook/push.go`, the `BusinessMetadata` is extracted and saved as JSONB in a new `metadata` column in the `repositories` table.
- Create migration to add `metadata JSONB` to `repositories`.
- Update `api/internal/db/queries/repositories.sql` to include `metadata` in inserts/updates and select it.
- Re-run `sqlc generate`.

### 3. API Payload
Ensure endpoints like `GET /api/v1/repos/{org}/{repo}` and `GET /api/v1/graph/{org}` return the `metadata` object in the JSON response for nodes.

### 4. SvelteKit UI Updates
- **API Documentation Page:** Display a nicely styled "Business Context" card or header section showing Owner, Slack, PagerDuty, PM, and SLA Tier if they exist.
- **Dependency Graph (Svelte Flow):** Update the custom node component to use `data.metadata.node_color` as a border or accent color if provided, falling back to the default theme color.

## Rules
- No Mocks: Ensure the real Go backend parses and serves the data to the UI.
- Use `sqlc` for database access.
- Strictly adhere to `Go` and `SvelteKit` project conventions.
