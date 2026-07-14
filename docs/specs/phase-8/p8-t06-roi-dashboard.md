# Spec: P8-T06 - Management ROI Dashboard

## 1. Overview
The most critical feature for selling Substrate to enterprise Engineering Directors and CTOs is the ability to empirically prove its value. The Management ROI Dashboard will consume the Phase 7 Audit Mode telemetry and Drift Detection anomalies to visualize "Prevented Disasters".

## 2. Requirements

### 2.1 Backend Aggregation Endpoint
- Create `GET /api/v1/telemetry/roi/{org}`.
- Query the database to calculate:
  - Total cross-repo breaking changes blocked (or shadowed in Audit Mode) in the last 30 days.
  - Total undocumented endpoints detected by the Sidecar Proxy.
  - Calculated Value: Assume each cross-repo breaking change that reaches production costs an average of 4 engineering hours to debug and fix. Return `Hours Saved` and `Estimated Dollar Value Saved` (assuming $100/hr loaded cost).

### 2.2 Svelte Dashboard Widget
- Add a new "Executive Summary" tab to the Svelte application.
- Render high-impact metric cards:
  - "Outages Prevented (30d): X"
  - "Engineering Hours Saved: Y"
  - "Drift Anomalies Detected: Z"
- Include a list of the top 3 most "volatile" repositories (repos that attempt the most breaking changes).

## 3. Implementation Steps
1. Implement the SQL aggregation queries in `api/internal/db/`.
2. Expose the HTTP handler in `api/internal/handlers/`.
3. Update the SvelteKit dashboard layout and create the `ROIWidget.svelte` component to fetch and display the data.
