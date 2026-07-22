# P14-T01: AI Contract Negotiation

## Objective
Implement an async team negotiation workflow. When a breaking change is detected on a PR, Substrate posts a structured GitHub comment tagging all affected consumer leads. It tracks their approval/rejection reactions, and only turns the provider PR green when all consumers have acknowledged and approved the breaking change.

## Requirements

### 1. `SCHEMAOWNERS` Parser
- Build a parser that reads a `SCHEMAOWNERS` file from the root of a consumer repository (similar to `CODEOWNERS`).
- Format: `frontend/src/api/*.ts @frontend-lead`

### 2. GitHub Comment Upgrades
- Update `api/internal/webhook/webhook.go` or `api/internal/github` to create a dedicated negotiation comment when a breaking change is detected.
- The comment must specifically `@mention` the GitHub usernames mapped from the `SCHEMAOWNERS` file of the impacted consumers.
- Add an interactive section: "React with 👍 to acknowledge and approve this breaking change, or 👎 to block."

### 3. GitHub Webhook Listener (Issue Comment Reactions)
- Add a new webhook listener in `api/internal/webhook/` for GitHub `issue_comment` and `issue_comment_reaction` events.
- Track when an `@mentioned` user reacts with 👍.

### 4. Async Check Suite Updates
- Initially, set the GitHub Check Suite status to `pending` with the title "Waiting for Consumer Approval".
- Once all tagged consumers have reacted with 👍, dynamically update the Check Suite status to `success`.

## Acceptance Criteria
- Unit tests verify `SCHEMAOWNERS` parsing.
- The GitHub App successfully sets a `pending` check suite on breaking changes if consumers are impacted.
- Reacting with 👍 from a consumer lead turns the check suite green.
