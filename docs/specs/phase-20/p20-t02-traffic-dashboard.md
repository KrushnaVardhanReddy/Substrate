# Phase 20 — Task 02: Traffic-Weight Dashboard

## 1. Overview

Extend the Substrate Svelte 5 dashboard to visualize OTel-derived traffic data:
- Colour-code dependency graph nodes by live traffic volume.
- Surface traffic weight inline in the Blast Radius modal.
- Provide a dedicated Traffic page with route-level histograms.

## 2. Technical Requirements

### 2.1 Graph Node Traffic Colouring

Modify `dashboard/src/lib/components/DependencyGraph.svelte`:

- Fetch `GET /api/v1/org/{org}/otel/traffic` on mount.
- Derive per-node `req_per_min` by summing all routes for that service.
- Apply a CSS custom property `--node-traffic-intensity` (0–1) based on a log scale.
- Node border glow: `box-shadow: 0 0 {intensity * 20}px hsl(217, 91%, 60%)`.
- Add a "Traffic Mode" toggle in the graph toolbar — off by default, does not break
  existing heatmap modes.

### 2.2 Blast Radius Modal Enhancement

Extend `BlastRadiusModal.svelte`:
- Add a `Traffic` column to the consumer table showing `req/min`, `p99`, `error %`.
- Sort consumers by `req_per_min DESC` by default (highest risk first).
- Badge: CRITICAL (red), HIGH (amber), MEDIUM (slate) based on server-side `risk` field.

### 2.3 Dedicated Traffic Page

New route: `/org/{org}/observability`

Components:
- `OtelSourceCard.svelte` — shows each registered OTel source, last-seen timestamp, revoke button.
- `RouteTrafficChart.svelte` — bar chart (use `chart.js` already in deps if available, else
  render a pure SVG bar chart — do NOT add a new charting library dependency).
- `TrafficFeed.svelte` — SSE-connected live feed of `otel/traffic/stream` events, newest first.

### 2.4 OTel Token Provisioning UI

On the Traffic page: "Connect APM" modal with:
- Generate token button → `POST /api/v1/org/{org}/otel/sources`.
- Display generated token once in a masked input field with copy button.
- Show the exact `otelcol` YAML config snippet to paste into customer's collector:

```yaml
exporters:
  otlphttp:
    endpoint: "https://substrate.example.com/api/v1/otel"
    headers:
      Authorization: "Bearer <TOKEN>"
```

## 3. Svelte 5 Rules

- All state via `$state`, `$derived`, `$effect` runes — no legacy stores.
- Lucide icons: direct imports only. Example:
    `import Activity from 'lucide-svelte/icons/activity'`
- SSE in `$effect` with cleanup: `const es = new EventSource(url); return () => es.close()`.
- Add Observability link to the org sidebar navigation.

## 4. Deliverables

1. `dashboard/src/routes/(app)/org/[org]/observability/+page.svelte` (CREATE)
2. `dashboard/src/routes/(app)/org/[org]/observability/+page.ts` (CREATE)
3. `dashboard/src/lib/components/OtelSourceCard.svelte` (CREATE)
4. `dashboard/src/lib/components/RouteTrafficChart.svelte` (CREATE)
5. `dashboard/src/lib/components/TrafficFeed.svelte` (CREATE)
6. MODIFY: `DependencyGraph.svelte` — traffic colour mode
7. MODIFY: `BlastRadiusModal.svelte` — traffic weight columns
8. MODIFY: org sidebar — add Observability nav item
9. `dashboard/src/lib/components/OtelSourceCard.test.ts` (CREATE)
10. Run `npm run check` and `npx vitest run`.
