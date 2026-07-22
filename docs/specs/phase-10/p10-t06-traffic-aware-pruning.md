# P10-T06: Traffic-Aware Pruning (Zombies)

## Overview
Correlate schema endpoints with live OTel/Datadog metrics to detect unused "zombie" APIs and auto-generate PRs to delete the dead code.

## Requirements
1. **Metrics Ingestion**: Accept OTel metric data (HTTP request counts per endpoint) via a webhook or polling endpoint.
2. **Zombie Detection**: Flag endpoints with zero traffic over the past 30 days as "zombies".
3. **Dashboard View**: Add a "Zombie APIs" tab in the dashboard listing all detected unused endpoints.
4. **Auto-PR**: For confirmed zombies (org-approved), open a PR in the provider repo deleting the dead endpoint code and its spec entry.
