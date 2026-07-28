# Phase 17 - Task 06: Deployment & Incident Graph Overlay

## 1. Goal
Provide a root-cause analysis visualization tool by ingesting deployment and incident webhooks (e.g., from GitHub Actions, Datadog) and overlaying them chronologically onto the Cytoscape Dependency Graph.

## 2. Requirements
- Add `POST /api/v1/events` endpoint to ingest arbitrary ecosystem events: `{ org, repo, event_type: "deployment" | "incident", description, timestamp }`.
- Add `GET /api/v1/events/{org}` to fetch events for a specific timeframe.
- In the `dashboard/src/routes/(app)/org/[org]/graph/+page.svelte` dependency graph view, add an "Event Timeline" panel.
- When an event is selected from the timeline, visually highlight the corresponding repo node in the Cytoscape graph (e.g., flash blue for deployments, red for incidents).

## 3. Database Schema
```sql
CREATE TABLE ecosystem_events (
  id          SERIAL PRIMARY KEY,
  org         TEXT NOT NULL,
  repo        TEXT NOT NULL,
  event_type  TEXT NOT NULL,
  description TEXT,
  event_time  TIMESTAMPTZ NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```
