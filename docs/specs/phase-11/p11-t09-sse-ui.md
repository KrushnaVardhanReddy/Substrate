# Spec: P11-T09 - Server-Sent Events (SSE) Real-Time UI

## 1. Overview
Stream real-time cross-repo diff results from the River queue directly to the dashboard via SSE and Go channels, eliminating UI polling.

## 2. Requirements
- Update the Go API (`api/`) to include a new route `GET /api/v1/events`.
- This route must set headers `Content-Type: text/event-stream` and write a heartbeat every 5 seconds.
- Create a channel registry to broadcast events to connected clients.

## 3. Implementation Steps

### Phase A: Backend SSE Broker (Completed via PR 108)
- ✅ Update the Go API (`api/`) to include a new route `GET /api/v1/events`.
- ✅ This route sets headers `Content-Type: text/event-stream` and writes a heartbeat every 5 seconds.
- ✅ Create a channel registry to broadcast events to connected clients.

### Phase B: Frontend SSE Client (Pending)
- Remove the 5-second polling interval (`setInterval(fetchGraph, 5000)`) in the Dependency Graph (`dashboard/src/routes/(app)/org/[org]/graph/+page.svelte`).
- Initialize a native `EventSource` connection to `/api/v1/events` during the `onMount` lifecycle block.
- Add event listeners to automatically call `fetchGraph()` (or directly ingest the streamed JSON) whenever a real-time event is pushed by the backend.
- Ensure the `EventSource` connection is properly closed in the `onMount` cleanup function.

## 4. Stitch & Jules Workflow
- **Jules:** Implemented the Go SSE handler and the broadcast mechanism (✅ Done).
- **Stitch:** Implement the Frontend EventSource logic and remove polling in `+page.svelte` (⏳ Next).
