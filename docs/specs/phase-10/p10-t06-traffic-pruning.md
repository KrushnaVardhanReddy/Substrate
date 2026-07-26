# Phase 10 - Task 06: Traffic-Aware Pruning (Zombies)

## 1. Goal
Correlate schema endpoints with live Datadog/OTel metrics to detect unused "zombie" APIs and auto-generate PRs to delete the dead code.

## 2. Requirements
- Ingest OpenTelemetry/Datadog metrics.
- Overlay traffic data on the Cytoscape graph.
- Highlight endpoints with 0 traffic in the last 30 days.
