# P17-T04: PagerDuty Blast Radius Injection

## Overview
Integrate with Datadog/PagerDuty. On incident creation, Substrate injects the visual Cytoscape Dependency Graph into the incident description to instantly show downstream blast radius.

## Goals
1. Add an API endpoint `POST /api/v1/integrations/pagerduty/webhook` to receive incident payloads.
2. Query the dependency graph (`GET /api/v1/dependencies/{org}`) for the affected service.
3. Call the PagerDuty API to update the incident notes with a link to the Substrate dashboard and a list of affected downstream services.

## Architecture
- **DB:** Expand `org_integrations` table to store PagerDuty API keys securely.
- **Go Backend:** Add a new handler in `api/internal/handlers/pagerduty.go`.
- **Worker:** (Optional) Use a goroutine to send the update asynchronously.
