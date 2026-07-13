# P7-T01: ServiceNow/Jira Dynamic CAB (Change Advisory Board)

## Objective
Enterprise organizations require formal tracking of breaking changes via ITSM tools like ServiceNow or Jira. Substrate must automatically create an Incident or Change Request ticket when a breaking change is detected in a PR, and dynamically assign the Tech Leads (`CODEOWNERS`) of the downstream consumer repositories as mandatory approvers.

## Requirements

### 1. ITSM Adapter Interface
Substrate API must introduce an `itsm` package with a generic `Adapter` interface:
```go
type TicketRequest struct {
    ProviderRepo     string
    PRNumber         int
    BreakingCount    int
    ConsumersBroken  []string
    Approvers        []string // Extracted from downstream CODEOWNERS
}

type Adapter interface {
    CreateTicket(ctx context.Context, req TicketRequest) (ticketID string, ticketURL string, err error)
}
```

### 2. Implementations
- **ServiceNow Adapter:** Uses SNOW Table API (`/api/now/table/change_request`) to create a normal change ticket.
- **Jira Adapter:** Uses Jira Cloud REST API (`/rest/api/3/issue`) to create an Issue of type "Change".

### 3. Substrate API Integration
- Target: `api/internal/handlers/diff.go` or a new webhook dispatcher.
- When `POST /api/v1/diff` receives a report with `breaking_count > 0` AND the configuration has `itsm.enabled: true`, the API will:
  1. Identify all broken consumers.
  2. Fetch the `CODEOWNERS` file from GitHub for those consumer repositories.
  3. Extract the GitHub handles/emails.
  4. Call the configured ITSM Adapter to generate the ticket.
  5. Return the `TicketURL` back to the GitHub App to include in the PR comment.

### 4. GitHub App Comment Update
- The GitHub App PR comment formatter must be updated to conditionally display:
  `🎫 **CAB Ticket Created:** [CHG001234](https://company.service-now.com/...)`
- This ensures developers know an approval process has been initiated.
