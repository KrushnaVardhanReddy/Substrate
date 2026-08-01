# Phase 19 — Task 04: Airbyte Sync Status Dashboard

## 1. Overview

Build a Svelte 5 dashboard panel at `/org/{org}/settings/integrations` showing the
real-time status of all Airbyte ingestion pipelines, record counts, and any Schema
Insurance violations detected from incoming streams.

## 2. UI Requirements

### 2.1 Route

`dashboard/src/routes/(app)/org/[org]/settings/integrations/+page.svelte`

### 2.2 Layout

The page has three sections:

#### Section A — Connected Sources

A card grid. Each card shows:
- Connector icon (use the first letter of the connector name in a coloured avatar)
- Source name (e.g. "HubSpot CRM")
- Connector ID (e.g. `source-hubspot`)
- Last sync time (relative, e.g. "3 hours ago")
- Record count badge
- Status pill: `Active` (green) / `Degraded` (amber — has violations) / `Error` (red)
- Actions: **Sync Now** button, **Delete** icon

#### Section B — Violation Feed

A live-updating violation log table (SSE-connected). Columns:
- Timestamp
- Stream name
- Field name
- Violation type (missing_required / type_mismatch / mrr_null)
- Severity badge (HIGH / MEDIUM / LOW)

SSE endpoint: `GET /api/v1/org/{org}/airbyte/violations/stream` (returns
`text/event-stream` with JSON violation events).

#### Section C — Add New Source

A collapsible form with fields:
- Source Name (text input)
- Connector ID (select with common presets: HubSpot, Salesforce, Stripe, Postgres, Snowflake, Custom)
- Config JSON (a `<textarea>` rendered as a JSON editor with basic syntax highlighting)
- Submit → calls `POST /api/v1/org/{org}/airbyte/sources`

### 2.3 Svelte 5 Constraints

- All reactive state via `$state` and `$derived` runes — no legacy stores.
- Use `$effect` for the SSE connection lifecycle (open on mount, close on destroy).
- Icons via direct Lucide imports: `import Plug from 'lucide-svelte/icons/plug'`, etc.
- No barrel imports from `lucide-svelte`.

### 2.4 API Integration

| UI action | API call |
|---|---|
| Page load | `GET /api/v1/org/{org}/airbyte/sources` |
| Sync Now | `POST /api/v1/org/{org}/airbyte/sources/{id}/sync` |
| Add source | `POST /api/v1/org/{org}/airbyte/sources` |
| Delete source | `DELETE /api/v1/org/{org}/airbyte/sources/{id}` |
| Live violations | SSE: `GET /api/v1/org/{org}/airbyte/violations/stream` |

### 2.5 Navigation

Add an **"Integrations"** link to the existing org settings sidebar
(`dashboard/src/lib/components/SettingsSidebar.svelte` or equivalent).

## 3. Rules

- The SSE connection must be cleanly closed in `$effect`'s cleanup function to prevent
  memory leaks.
- The violation table must be paginated client-side (show latest 50; button to load more).
- Config JSON textarea must validate JSON on blur and show an inline error if invalid.
- The **Sync Now** button must be disabled and show a spinner while the sync is in progress.

## 4. Deliverables

1. `dashboard/src/routes/(app)/org/[org]/settings/integrations/+page.svelte` (CREATE)
2. `dashboard/src/routes/(app)/org/[org]/settings/integrations/+page.ts` (CREATE — load function)
3. `dashboard/src/lib/components/AirbyteSourceCard.svelte` (CREATE)
4. `dashboard/src/lib/components/ViolationFeed.svelte` (CREATE — SSE-connected)
5. MODIFY: Settings sidebar — add Integrations nav link
6. `dashboard/src/lib/components/AirbyteSourceCard.test.ts` (CREATE — Vitest unit test)
7. `dashboard/src/lib/components/ViolationFeed.test.ts` (CREATE — Vitest unit test)

Commit: `"jules: P19-T04 add Airbyte sync status dashboard"`
Target branch: `feature/p19-t04-sync-dashboard`
