# P10-T02: API Gateway Auto-Sync (Cloud Integrations)

## Objective
Develop a synchronization agent that automatically imports and updates API definitions from popular cloud API Gateways (e.g., AWS API Gateway, Kong, Apigee) directly into the Substrate registry.

## Context
Organizations often define their external-facing APIs directly within an API Gateway, bypassing version control for schema definitions. To ensure Substrate has a complete and accurate map of the real-world infrastructure, we need to pull these definitions dynamically from the source of truth in the cloud.

## Requirements

### 1. Integration Adapters
- Implement modular Go adapters for connecting to:
  - AWS API Gateway (via AWS SDK).
  - Kong (via Kong Admin API).
- The adapters should authenticate securely (using stored credentials or IAM roles) and fetch the active OpenAPI/Swagger definitions.

### 2. Auto-Sync Worker
- Create a background worker process in the Go Engine that periodically polls the configured API Gateways.
- When changes are detected, trigger the standard Substrate ingestion pipeline to generate a new schema version and calculate blast radius.
- Support webhooks from the gateways (if supported) for real-time updates.

### 3. Substrate.yaml Configuration
- Allow users to define gateway integrations in their `substrate.yaml` or project settings.

## Acceptance Criteria
1. The Go backend can successfully authenticate with AWS API Gateway and download an OpenAPI spec.
2. The auto-sync worker updates the internal Substrate registry when the gateway schema changes, triggering an impact analysis.
3. The UI indicates that the schema was sourced from a cloud provider rather than a git repository.
