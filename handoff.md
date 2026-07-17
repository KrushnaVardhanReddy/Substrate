# Substrate Handoff

## Current Status (End of Day)
* **Goal Achieved:** Successfully stood up the local Forgejo Git Server (via `docker-compose.forgejo.yml`) to replace mock GitHub webhooks. Network connectivity works, webhooks are firing, and the local Cloudflare worker successfully intercepts and validates the payload (using the `testsecret` HMAC signature).
* **Blocker Identified:** The worker is currently returning `Ignored` because `github-client.ts` is hardcoded to call `api.github.com` instead of dynamically calling the Git server that triggered the webhook.
* **Tasks Delegated:** 
  * Delegated **P12-T10** to Jules to refactor the webhook worker using an Adapter Pattern to make it VCS-Agnostic (supporting GitHub, GitLab, and Gitea/Forgejo).
  * Delegated **P12-T11** to Jules to update the SvelteKit onboarding UI to let users select their VCS provider (Cloud vs Self-Hosted).

## Next Steps for Tomorrow
1. **Review & Merge Jules PRs:** Verify the VCS-Agnostic Adapter and the UI changes in `feature/dev`.
2. **End-to-End Forgejo Test:** Once the adapter is merged, test pushing to the local `microservices-demo` repository in Forgejo again. Verify that the worker successfully fetches `substrate.yaml` from Forgejo and updates the Svelte dashboard.
3. **P12-T08 (Enterprise VPC Deployment):** Transition focus to the final V2.0 production cutover, bundling the static assets and WASM binary into the Go single-binary deployment.

## Quick Start Reminders
* **Start Forgejo locally:** Run `make forgejo` (Starts the container on port 3000).
* **Start Backend Stack:** Run `make start-bg`.
* **View Graph:** Go to `http://localhost:5173/org/<forgejo-username>/graph`.

## 🧪 Comprehensive Demo Testing Guide

To fully validate Substrate across all 7 demo repositories (Stripe, Microservices, RealWorld, OpenAI, GraphQL, Slack, and Jaffle Shop), test along the following three axes:

### Axis 1: Breaking vs. Non-Breaking Changes
For each repository, test both a schema breaking change and a safe (non-breaking) addition to verify the diff engine logic.
* **Microservices (Protobuf):**
  * *Breaking:* Remove `CartItem.product_id`
  * *Safe:* Add `string notes = 3`
* **GitHub GraphQL:**
  * *Breaking:* Remove `User.email`
  * *Safe:* Add `User.age: Int`
* **Stripe (OpenAPI):**
  * *Breaking:* Remove `/v1/charges` endpoint
  * *Safe:* Add `/v2/beta/charges`
* **OpenAI (OpenAPI):**
  * *Breaking:* Remove `function_call` property
  * *Safe:* Add `metadata: Map` property
* **Jaffle Shop (SQL):**
  * *Breaking:* Drop `customer_lifetime_value` column
  * *Safe:* Add `age INTEGER` column
* **Slack (AsyncAPI):**
  * *Breaking:* Remove `channel_id` from payload
  * *Safe:* Add `thread_ts` to payload
* **RealWorld (OpenAPI):**
  * *Breaking:* Remove `/api/articles`
  * *Safe:* Add `/api/tags`

### Axis 2: Configuration Modality
* **Manual Yaml Creation:** Push a `substrate.yaml` file to the repository. Verify that the Cloudflare Worker webhook correctly parses the consumers block and explicitly draws the specified dependency edges in the graph.
* **Automated Discovery (Phase 5):** Remove the `substrate.yaml` file and rely on the Go API's internal discovery engine (`POST /discovery`). Verify that it auto-detects dependencies by scanning for URLs, SDK imports, or env vars, and draws the graph dynamically without explicit configuration.

### Axis 3: Execution Environment (Diffing Engine)
* **Cloud (Go API / CI/CD):** Push the change to the Git server (Forgejo). Ensure the Cloudflare Worker intercepts the webhook, routes the schemas to the Go Diff Engine container running on port `8080`, and correctly fails/passes the CI check.
* **WASM (In-Browser):** Open the Substrate Dashboard Visual API Studio. Make the breaking changes directly in the browser's schema editor. Verify that the WebAssembly-compiled diff engine (`engine.wasm`) catches the breakage instantly (0ms latency) entirely on the client side, without making any network requests to the Go API.

### Axis 4: Governance & Overrides (The "Yellow Path")
* **Intentional Breakage:** Push a breaking change (e.g. dropping a column), but include an `overrides` block in the `substrate.yaml` with a valid `approved_by` email and future `expires` date.
* **Verification:** Ensure that the CI check *passes* (with a warning) instead of failing, and the UI marks the change as "Acknowledged".

### Axis 5: Scale & Resilience
* **1,000-Node Stress Test:** Generate a massive mock graph and verify that the Svelte Flow layout calculation completes in under 2 seconds and the UI maintains 60fps during zooming/panning.
* **Network Drop Simulation:** Terminate the Go API process briefly. Verify that the UI displays a "Reconnecting..." toast and successfully re-establishes the Server-Sent Events (SSE) connection without losing graph state when the API comes back online.

### 🤖 Playwright E2E Automation Note
Since the entire stack (including the Forgejo Git server) is now fully local, this entire 5-axis test matrix can and will be automated via **Playwright**. 
We can write a script (`dashboard/tests/e2e/system-matrix.spec.ts`) that programmatically pushes commits to the local Forgejo container, triggers the webhook, and asserts that the Substrate Dashboard DOM updates with the correct blast radius and CI status.

> **Important Architectural Note for the E2E Tests:** 
> You will notice the Playwright script dynamically generates lightweight "stub" schema files (e.g., a 10-line `openapi.yaml`) in a `/tmp` folder instead of copying the real `demo-repos/` from disk. 
> *Why did we do this?* Because massive repos like `stripe/openapi` are 1.2GB and are explicitly `.gitignore`d. If Playwright relied on them, the E2E tests would instantly fail in CI environments where those ignored files don't exist. Generating stubs guarantees the test is lightning-fast and 100% reproducible on any machine.
