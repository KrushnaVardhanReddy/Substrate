# P17-T02: Legacy API FinOps Translation

## Objective
Expand the existing "Zombie API" detection system to calculate and display the exact estimated dollar amount saved by sunsetting legacy endpoints. By assigning a financial value to unused API endpoints, Engineering Managers can justify prioritization of API deprecations.

## Architecture

### 1. FinOps Estimation Engine
- **Logic:** Each "zombie" (0 requests in 30 days) carries a hidden maintenance burden (CI/CD compute, security scanning, infrastructure footprint). 
- **Calculation:** We will use an industry-standard heuristic for MVP: **$500 per month** per deprecated REST/GraphQL endpoint in cloud compute and maintenance savings.
- **Implementation:** Create a new struct `ZombieFinOpsReport` that wraps the existing `ZeroTrafficEndpoint` list and provides a `TotalMonthlySavings` integer.

### 2. Backend API Updates
- **Endpoint:** Modify the existing `GET /api/v1/org/{org}/zombies` handler to return a structured response containing both the array of zombies and the aggregate FinOps data.
```go
type ZombieFinOpsResponse struct {
    Zombies              []sqlcgen.ZeroTrafficEndpoint `json:"zombies"`
    TotalMonthlySavings  int                           `json:"total_monthly_savings"`
    SavingsPerEndpoint   int                           `json:"savings_per_endpoint"`
}
```

### 3. Dashboard UI Updates
- **File:** `dashboard/src/routes/(app)/org/[org]/zombies/+page.svelte`
- **Changes:**
  - Update the page state and API fetch to handle the new `ZombieFinOpsResponse` JSON structure.
  - Render a prominent "FinOps Savings" banner/widget at the top of the page displaying the `total_monthly_savings` in a large, green font (e.g. `$X,XXX / month`).
  - Add a "Potential Savings" column to the table, showing the $500 savings per endpoint row.
- **Constraints:** Must use Svelte 5 runes (`$state`, `$effect`). Must use direct icon imports (`import DollarSign from 'lucide-svelte/icons/dollar-sign'`).

## Deliverables
1. Update `api/internal/handlers/otel_webhook.go` (`GetZombiesHandler`) to return the FinOps struct.
2. Update unit tests in `api/internal/handlers/otel_webhook_test.go`.
3. Update `dashboard/src/routes/(app)/org/[org]/zombies/+page.svelte` UI to reflect the new structure and render the FinOps widget.
4. Update `dashboard/tests/e2e/zombies.spec.ts` if the mocked API response payload needs adjustment.
