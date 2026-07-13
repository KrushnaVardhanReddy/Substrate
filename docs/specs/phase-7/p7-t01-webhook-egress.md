# P7-T01: Enterprise Webhook & Event Egress

## Objective
Rather than hardcoding integrations with specific ITSM tools like ServiceNow or Jira, Substrate will adopt an Event-Driven Architecture (EDA). Substrate will emit standardized JSON events (webhooks) whenever a contract is broken or overridden. This allows enterprise platform teams to ingest these events into their own existing infrastructure (ServiceNow Event Management, AWS EventBridge, Datadog, Zapier) to drive their bespoke Change Advisory Board (CAB) and notification workflows.

## Requirements

### 1. Webhook Registry Configuration
Organizations must be able to configure outbound webhook destinations for their Substrate events.
- **Config:** A new `webhooks` array in the database at the Organization level, or globally via environment variables for on-prem deployments.
- **Endpoint:** `POST /api/v1/org/{org}/webhooks` to register a new egress URL and an optional `secret` for HMAC signing.

### 2. Standardized Event Payload
Substrate will emit events using the CloudEvents format or a standard JSON envelope.

**Event Type:** `substrate.breaking_change.detected`
**Payload Example:**
```json
{
  "event_id": "evt_987654321",
  "event_type": "substrate.breaking_change.detected",
  "timestamp": "2026-07-13T12:00:00Z",
  "data": {
    "organization": "acme-corp",
    "provider_repo": "acme-corp/billing-api",
    "pr_number": 42,
    "commit_sha": "abc123def",
    "breaking_count": 2,
    "broken_consumers": [
      {
        "repo": "acme-corp/invoice-service",
        "codeowners": ["@acme-corp/billing-team"]
      }
    ],
    "diff_url": "https://substrate.acme.corp/diff/diff_12345"
  }
}
```

### 3. Egress Dispatcher (Worker)
- Target: `api/internal/webhook/egress.go`
- After a `POST /api/v1/diff` completes and identifies a breaking change (and executes the cross-repo check to gather impacted consumers), it will drop a job onto an asynchronous channel or queue.
- The `egress` worker will iterate over registered webhooks for the organization, calculate the `X-Hub-Signature-256` HMAC (if a secret is provided), and POST the JSON payload to the customer's endpoint.
- Include a retry mechanism with exponential backoff for failed deliveries.

### 4. Integration with API Flow
- The cross-repo check logic currently separated in the GitHub App must be made available to the Egress Dispatcher. The API should ideally resolve the impacted consumers *before* emitting the webhook, so the event contains the full Blast Radius context for the ITSM systems to route tickets to the correct teams.
