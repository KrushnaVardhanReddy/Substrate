# Phase 12 - Task 14: System Matrix Overrides & Completion (Yellow Path)

## 1. Goal
Complete the "Yellow Path" (Overrides/Acknowledged) across the entire Substrate architecture. This ensures that when a developer pushes a breaking schema change but provides a valid `overrides:` block in their `substrate.yaml`, the backend downgrades the breaking change to an `ACKNOWLEDGED` status, and the frontend visually renders this as a non-breaking warning (Yellow/Orange) instead of a catastrophic failure (Red).

Finally, this task expands the E2E matrix in `system-matrix-full.spec.ts` to include the remaining schema types (SQL and AsyncAPI), testing Red, Green, and Yellow paths for all 7 implemented repositories (`microservices-demo`, `stripe-api`, `github-graphql`, `jaffle-shop-db`, `slack-webhooks`, `openai-openapi`, and `realworld`).

## 2. Requirements

### 2.1 Frontend UI (`dashboard/src/lib/components/ServiceNode.svelte`)
- The UI currently only renders `.safe` and `.breaking`.
- Update the `<div class="status-indicator">` binding to support a third state: `.warning` (or `.acknowledged`).
- Ensure this state maps to an `ACKNOWLEDGED` status returned from the backend.
- Update global CSS to render the `.warning` status indicator in a distinct color (e.g., `#f59e0b` / Amber).

### 2.2 Backend Logic (`api/internal/services/sync.go`)
- In `ProcessSync`, after the Diff Engine returns a `BreakingCount > 0`, the backend currently loops through the consumers and sets their dependency status to `BREAKING`.
- Modify this loop: parse the head `substrate.yaml`. If the consumer has an `overrides` array that explicitly matches the breaking change rule or consumer name, downgrade the status to `ACKNOWLEDGED`.
- Update the database query in `UpdateDependencyStatus` to accept `"ACKNOWLEDGED"` as a valid enum.

### 2.3 E2E Test Suite (`dashboard/tests/e2e/system-matrix-full.spec.ts`)
- **Retrofit existing suites:** Add a `Test C: Yellow Path: Push breaking change with override` to the existing `Protobuf`, `OpenAPI`, and `GraphQL` test blocks.
  - The test must push a schema that is mathematically breaking, but push a `substrate.yaml` containing the `overrides` key.
  - The test must assert that `page.locator('.status-indicator.warning')` (or `.acknowledged`) has count `1` instead of `.breaking`.
- **Implement remaining suites:**
  - Create a test block for **SQL** (`jaffle-shop-db`).
    - Use a tiny 10-line SQL stub for `BASE`, `BREAKING` (e.g., dropping a column), `SAFE` (adding a column), and `OVERRIDE`.
  - Create a test block for **AsyncAPI** (`slack-webhooks`).
    - Use a tiny 10-line AsyncAPI YAML stub for `BASE`, `BREAKING` (removing a payload property), `SAFE` (adding a payload property), and `OVERRIDE`.
  - Create a test block for **OpenAPI (openai)** (`openai-openapi`). Use a tiny OpenAPI stub with a `chat/completions` endpoint.
  - Create a test block for **OpenAPI (realworld)** (`realworld`). Use a tiny OpenAPI stub with an `/api/articles` endpoint.
## 3. Architecture Constraints
- **Do not push massive repos to Forgejo.** We must continue using the "stub generation" strategy (defining 10-line string variables inside the Playwright script and pushing them) to simulate repo changes.
- **Do not alter `pushToForgejo`**. The helper perfectly simulates standard `git init`, `commit`, and `push` via Forgejo API repository creation.

## 4. Deliverables
- [ ] Updated `ServiceNode.svelte` with Yellow path UI classes.
- [ ] Updated `sync.go` (and related DB layers) to support `ACKNOWLEDGED` status.
- [ ] Updated `system-matrix-full.spec.ts` containing Red/Green/Yellow paths for all 7 repositories.
