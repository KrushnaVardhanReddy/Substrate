# P7-T02: Targeted Notifications (Slack/Teams)

## Objective
When a breaking change is merged or attempted, the specific downstream teams that are impacted must be notified immediately in their collaboration tools. Instead of noisy global channels, Substrate must send targeted alerts to specific Slack channels or MS Teams webhooks based on the `CODEOWNERS` mapping.

## Requirements

### 1. Notification Routing Config
Organizations can map GitHub CODEOWNERS teams (e.g., `@acme/billing`) to specific Slack channels or MS Teams incoming webhooks via the Substrate Dashboard.
- **Database:** A new `notification_routes` table linking `org_id`, `github_team`, and `destination_url`.

### 2. Formatted Messaging
Substrate must format a human-readable payload specifically for Slack (Block Kit) or Teams (Adaptive Cards).
- The message should include:
  - The upstream repo that caused the break.
  - The exact breaking change (e.g., `FIELD_REMOVED: email`).
  - A direct link to the Substrate Blast Radius visualization.
  - An `@mention` mapping to alert the specific team on call.

### 3. API Integration
- Target: `api/internal/notifications/slack.go` and `teams.go`
- This should run asynchronously after the Cross-Repo check identifies the `BrokenConsumers` array.
- For each broken consumer, look up their `CODEOWNERS` in the `notification_routes` table and dispatch the message.
