# Substrate Demo Repositories

This document outlines the recommended open-source repositories to fork and use for Substrate demonstrations, stress-testing, and public V2.0 launch materials.

## 1. The "Enterprise Scale" Demo: Stripe API
*   **Repository:** [stripe/openapi](https://github.com/stripe/openapi)
*   **Purpose:** Extreme stress-testing and "shock factor" for the Cytoscape UI.
*   **Why it works:** Stripe maintains one of the largest and most complex OpenAPI specifications in existence. 
*   **Demo Script:** 
    1. Import the Stripe repo into Substrate.
    2. Show the massive rendering capabilities of the WASM diff engine.
    3. Intentionally remove a deeply nested, critical field (e.g., `charge.amount`).
    4. Demonstrate how Substrate catches the breaking change instantly in the Visual API Studio.

## 2. The "Cross-Service Blast Radius" Demo: Google Cloud Microservices
*   **Repository:** [GoogleCloudPlatform/microservices-demo](https://github.com/GoogleCloudPlatform/microservices-demo)
*   **Purpose:** Demonstrating the true value of Substrate's cross-repo dependency graph.
*   **Why it works:** This is a realistic 10-tier e-commerce microservices application (Checkout, Payment, Email, Currency, etc.) communicating via gRPC/Protobufs and REST.
*   **Demo Script:**
    1. Map the individual microservice folders as separate consumers/providers in `substrate.yaml`.
    2. Open the Substrate Dashboard to reveal the complex 10-node dependency graph.
    3. Introduce a breaking change in the `payment-service`.
    4. Watch the UI's "Blast Radius" feature dynamically highlight the `checkout-service` in red, proving that Substrate detects downstream cascading failures.

## 3. The "Standard SaaS" Demo: RealWorld (Conduit)
*   **Repository:** [gothinkster/realworld](https://github.com/gothinkster/realworld)
*   **Purpose:** The standard "Zero-to-One" onboarding demo for new users.
*   **Why it works:** RealWorld is the modern "TodoMVC". It has dozens of backend implementations that all strictly adhere to a single standardized OpenAPI spec for a Medium.com clone.
*   **Demo Script:**
    1. Go through the Substrate Onboarding Wizard using a RealWorld fork.
    2. Drop in a `substrate.yaml` configuration.
    3. Open a GitHub Pull Request that changes the `/api/articles` endpoint response format.
    4. Show the Substrate GitHub App automatically commenting on the PR and blocking the merge.

## 4. The "AI/ML Integration" Demo: OpenAI API Specs
*   **Repository:** [openai/openai-openapi](https://github.com/openai/openai-openapi)
*   **Purpose:** Proving that Substrate can govern LLM/AI integrations and prevent downstream AI application crashes.
*   **Why it works:** AI startups depend entirely on upstream API stability from providers like OpenAI, Anthropic, or local inference servers like Ollama. 
*   **Demo Script:**
    1. Create a graph where `openai-openapi` is the upstream provider, and a downstream consumer is a mock "AI Chatbot" repo.
    2. Introduce a breaking change to the `ChatCompletionRequestMessage` (e.g., removing the `function_call` field).
    3. Substrate highlights the downstream Chatbot as "Broken", demonstrating how AI agent frameworks can be protected from upstream LLM provider API drift.

## 5. The "Federated GraphQL" Demo: GitHub GraphQL Schema
*   **Repository:** [octokit/graphql-schema](https://github.com/octokit/graphql-schema) (or Apollo Odyssey Voyage)
*   **Purpose:** Showing that Substrate goes beyond REST and supports GraphQL AST schema diffing.
*   **Why it works:** GraphQL APIs are notoriously difficult to version without breaking consumers. GitHub maintains a massive, heavily-typed GraphQL schema.
*   **Demo Script:**
    1. Import the GitHub GraphQL schema.
    2. Remove a highly-utilized query (e.g., `user.repositories`).
    3. Substrate's diff engine (configured for GQL) catches the missing field and blocks the PR, preventing frontend React/Apollo apps from crashing on missing data.

## 6. The "Async Event / Webhook" Demo: Slack API Specs
*   **Repository:** [slackapi/slack-api-specs](https://github.com/slackapi/slack-api-specs)
*   **Purpose:** Proving Substrate handles asynchronous Event-Driven Architecture (AsyncAPI / Webhooks).
*   **Why it works:** Slack's entire ecosystem runs on webhooks and the Events API. Changing a payload shape instantly breaks downstream bots.
*   **Demo Script:**
    1. Import the Slack AsyncAPI/OpenAPI webhook definitions.
    2. Drop the `channel_id` field from a `message.channels` event payload.
    3. Show Substrate protecting a downstream Slack Bot repository from the broken payload contract.

## 7. The "Data Engineering / ETL" Demo: Database to Teradata/Snowflake
*   **Repository:** [dbt-labs/jaffle_shop](https://github.com/dbt-labs/jaffle_shop) (or any SQL migration repo)
*   **Purpose:** Showing that Substrate protects data pipelines, BI dashboards, and enterprise data warehouses (like Teradata).
*   **Why it works:** Data teams constantly deal with upstream engineers accidentally renaming or dropping database columns, which silently breaks ETL jobs and executive dashboards.
*   **Demo Script:**
    1. Import a repository containing SQL schemas (or dbt models) that pipe data into Teradata.
    2. Have a backend engineer open a PR dropping the `customer_lifetime_value` column from the Postgres production database.
    3. Show Substrate catching the schema drift and warning that the downstream "Teradata ETL Pipeline" and "Executive BI Dashboard" will fail if the PR is merged.

## Local Testing Strategy (No GitHub Required)

All 7 demo repos can be validated entirely locally by bypassing the Cloudflare Worker and hitting the Go API + Engine directly. This enables offline demo rehearsal, instant feedback during development, and CI/CD testing in air-gapped environments.

### Architecture: Direct API Testing
```
┌──────────────────────────────────────────────────────────┐
│                   Local Test Scripts                      │
│  scripts/local-test.sh → test-engine.sh → test-sync.sh  │
└────────────┬─────────────────────┬───────────────────────┘
             │                     │
             ▼                     ▼
   ┌─────────────────┐   ┌─────────────────┐
   │  Go Engine       │   │  Go API          │
   │  POST /diff      │   │  POST /sync      │
   │  :8080           │   │  POST /cross-repo│
   └─────────────────┘   │  :8090           │
                          └─────────────────┘
```

### Test Matrix

| Demo Repo | Schema Type | Breaking Test | Safe Test |
|-----------|-------------|---------------|-----------|
| microservices-demo | protobuf | Remove `CartItem.product_id` | Add `string notes = 3` |
| graphql-schema | graphql | Remove `User.email` | Add `User.age: Int` |
| openapi (Stripe) | openapi | Remove `/v1/charges` | Add `/v2/beta/charges` |
| openai-openapi | openapi | Remove `function_call` | Add `metadata: Map` |
| jaffle_shop | sql | Drop `customer_lifetime_value` | Add `age INTEGER` |
| slack-api-specs | asyncapi | Remove `channel_id` | Add `thread_ts` |
| realworld | openapi | Remove `/api/articles` | Add `/api/tags` |

### Running Locally
```bash
# Start all services
make start-bg

# Run full local test suite (no GitHub token needed)
./scripts/local-test.sh

# Run specific schema type
./scripts/local-test.sh --schema=protobuf

# Run specific layer only
./scripts/local-test.sh --layer=1  # Engine diff only
./scripts/local-test.sh --layer=2  # Sync + graph only
./scripts/local-test.sh --layer=3  # Cross-repo blast radius only

# Reset database and re-test
./scripts/local-test.sh --reset
```

### Prerequisites
- `make start-bg` running (Postgres, API, Engine, Worker, Dashboard)
- `curl` and `jq` installed
- No GitHub token required

See `docs/specs/phase-12/p12-t10-local-demo-testing.md` for the full specification.

---

## Implementation Notes (Phase 12 Demo Adjustments)

To support the above demos (specifically the Monorepo Microservices Demo), several architectural adjustments were implemented to the local environment and webhooks:

1. **Multi-Consumer Monorepo Sync:** The GitHub App Webhook Worker (`github-app/src/index.ts`) was updated to loop over all parsed consumers in `substrate.yaml` and sync them individually to the Registry API (using `${owner}/${entry.name}` as the `consumer_repo`). This prevents all microservices from collapsing into a single repository node in the graph.
2. **Explicit Consumer Mapping:** The `demo-repos/microservices-demo/substrate.yaml` was updated to explicitly list every microservice consumer (e.g., `frontend`, `checkoutservice`) to satisfy the TypeScript worker's strict parsing regex.
3. **Pseudo GitHub Repo IDs (Upcoming Fix):** To bypass the Postgres unique constraint on `github_repo_id` (which was blocking multiple microservices from the same repository), the TS webhook worker will be updated to hash the `consumer_repo` full name into a pseudo-integer ID, mimicking the Go backend's Phase 5 Auto-Discovery logic.
4. **Go Backend JSON Tags:** The Go API database models (`Repository` and `Contract`) were updated with explicit lowercase JSON struct tags (`json:"name"`, `json:"full_name"`) to match the SvelteKit dashboard UI expectations, resolving the "Blank Repository Name" bug.
5. **Internal Service Authentication:** Fixed a bug in the local development `Makefile` where the Go API was rejecting sync requests (`401 Unauthorized`) because `INTERNAL_SERVICE_TOKEN` was missing from the `make api` environment.
6. **Wrangler Webhook Caching:** Updated the TS Webhook worker to gracefully fall back to a personal `GITHUB_TOKEN` to bypass Wrangler `.dev.vars` caching issues when testing webhooks locally, and added `GITHUB_TOKEN` to `github-app/src/types.ts`.
7. **Tier Limits Middleware Bypass:** The Go API enforces a 3-repo maximum on the free tier. Because the microservices demo requires 5 repositories, the `TierLimitsMiddleware` was updated in `api/internal/server/limits.go` to unconditionally bypass billing limits when `ENVIRONMENT=development` is provided, and the `Makefile` was updated to inject this variable.
