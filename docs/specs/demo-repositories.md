# Substrate Demo Repositories

This document outlines the recommended open-source repositories to fork and use for Substrate demonstrations, stress-testing, and public V2.0 launch materials.

## 1. The "Enterprise Scale" Demo: Stripe API
*   **Repository:** [stripe/openapi](https://github.com/stripe/openapi)
*   **Purpose:** Extreme stress-testing and "shock factor" for the Svelte Flow UI.
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
