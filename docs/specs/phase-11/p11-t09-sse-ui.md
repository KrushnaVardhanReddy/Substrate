# Spec: P11-T09 - Server-Sent Events (SSE) Real-Time UI

## 1. Overview
Stream real-time cross-repo diff results from the River queue directly to the dashboard via SSE and Go channels, eliminating UI polling.

## 2. Requirements
- Update the Go API (`api/`) to include a new route `GET /api/v1/events`.
- This route must set headers `Content-Type: text/event-stream` and write a heartbeat every 5 seconds.
- Create a channel registry to broadcast events to connected clients.

## 3. Stitch & Jules Workflow
- **Jules:** Implement the Go SSE handler and the broadcast mechanism.
