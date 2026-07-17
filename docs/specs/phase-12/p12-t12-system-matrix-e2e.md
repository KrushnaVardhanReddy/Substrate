# P12-T12: System Matrix E2E Test (Playwright + Forgejo)

## Objective
Create a master Playwright End-to-End test suite (`system-matrix.spec.ts`) that validates the entire Substrate architecture from a real `git push` to the SvelteKit dashboard rendering. This suite must iterate through all 7 core demo repository schemas, validating the Red (Breaking), Green (Safe), and Yellow (Override) paths natively via the local Forgejo Git server.

## Architecture & Automation Strategy
The flow to automate for EACH test case:
1. Since real demo repositories (like Stripe) are gigabytes in size and not fully committed to our Git tree, your Playwright script must dynamically generate minimal "stub" schema files (e.g. a simple `openapi.yaml`, `schema.graphql`) into a temporary `/tmp/<repo-name>-test` directory using Node's `fs.writeFileSync`.
2. Playwright uses Node's `child_process.execSync` to run `git init`, `git add .`, `git commit`, and configure the remote to `http://admin:admin@localhost:3000/admin/<repo-name>.git`.
3. Playwright modifies a schema file, commits the change, and runs `git push -f` to `http://admin:admin@localhost:3000/admin/<repo-name>.git`.
4. Forgejo fires a webhook to the Cloudflare Worker (`localhost:8787`).
5. Playwright observes the Svelte Dashboard (`localhost:5173`) and asserts that the UI updates via Server-Sent Events (SSE).

## 🧪 The Matrix (Test Cases)

You must implement parameterized testing (e.g., using a matrix array and `for...of` loop in Playwright) to execute the following matrix for all 7 repo types. 

For each schema type, run the following 3 sub-tests sequentially:

### 1. The Red Path (Breaking Change)
* Action: Remove a critical field/endpoint from the schema and push.
* Assertion: Wait for SSE to update the DOM. Assert that `page.locator('.blast-radius-alert')` is visible, the node turns RED, and the commit status shows `Exit code: 2`.

### 2. The Green Path (Safe Change)
* Action: Revert the break. Add a safe, non-breaking field/endpoint to the schema and push.
* Assertion: Assert that the graph remains GREEN, no blast radius alerts appear, and the commit status shows `Exit code: 0`.

### 3. The Yellow Path (Governance Override)
* Action: Re-apply the breaking change, but write an `overrides` block in `substrate.yaml` with a valid `approved_by` email and `expires` date. Push the commit.
* Assertion: Assert that the node is marked YELLOW (Acknowledged), and the commit status passes with a warning.

### The 7 Repository Schemas to Test

| Repo Name | Schema Type | Breaking Change | Safe Change |
|-----------|-------------|-----------------|-------------|
| `microservices-demo` | Protobuf (`.proto`) | Remove `CartItem.product_id` | Add `string notes = 3` |
| `stripe-api` | OpenAPI (`.yaml`) | Remove `/v1/charges` endpoint | Add `/v2/beta/charges` endpoint |
| `realworld-api` | OpenAPI (`.yaml`) | Remove `/api/articles` | Add `/api/tags` |
| `openai-api` | OpenAPI (`.yaml`) | Remove `function_call` property | Add `metadata: Map` property |
| `github-graphql` | GraphQL (`.graphql`) | Remove `User.email` | Add `User.age: Int` |
| `slack-webhooks` | AsyncAPI (`.yaml`) | Remove `channel_id` from payload | Add `thread_ts` to payload |
| `jaffle-shop-db` | SQL (`.sql`) | Drop `customer_lifetime_value` | Add `age INTEGER` |

## Additional Configuration Modality Tests
Outside the core matrix, add two specific tests for dependency mapping:
1. **Manual YAML Mapping:** Push a `substrate.yaml` with explicit `depends_on` blocks and verify the edges are drawn correctly in the UI.
2. **Phase 5 Auto-Discovery:** Remove `substrate.yaml`, push a code file with a hardcoded URL (e.g., `fetch('http://api.stripe.com/v1')`), and assert the Go API auto-discovers and draws the edge.

## Constraints
* Ensure the tests use Playwright's `test.describe.serial` since they modify a shared local git repository sequentially.
* Do NOT mock the API requests (`page.route`). This must test the real webhook pipeline.
* Use `execSync` safely to execute git commands in temporary local folders.
* Set appropriate timeouts (`test.setTimeout(120000)`), as the webhook pipeline + Go diffing engine might take several seconds to process.
