# V1.0 System E2E Tests Specification

> **Status:** DRAFT
> **Phase:** V1.0 Pre-Flight (V1-T07)
> **Scope:** Defines the end-to-end orchestration tests for the Substrate V1.0 pipeline.

---

## 1. Objective

The V1.0 System E2E test suite validates the entire Substrate lifecycle from the perspective of an end-user organization adopting the platform. It ensures that the zero-touch onboarding, AI scaffolding, contract registry, interactive UI diffs, and deployment safety gates all integrate seamlessly.

## 2. Testing Architecture

Since Substrate consists of multiple components (GitHub Worker, Go API Registry, Svelte Dashboard, CLI), the E2E test will orchestrate these interactions using Go's `testing` package but must **avoid in-memory mocks for the database**. 

**CRITICAL: True System Integration**
The E2E suite MUST connect to the real, locally running PostgreSQL database and the real, locally running Go API server via HTTP. 
- **No Mock Drivers:** Do not use sqlite or mock DB drivers. Connect directly to `postgres://postgres:postgres@localhost:5432/substrate`.
- **Real HTTP:** Fire actual HTTP requests with `Authorization: Bearer local-dev-token` to `http://localhost:8090/api/v1/webhook`.
- **Mocks Allowed:** Only external systems like GitHub APIs and LLM (OpenAI) endpoints may be mocked via local HTTP test servers.

The test file should be located at: `scripts/e2e/v1_e2e_test.go`

## 3. The End-to-End Scenario Flow

The test must execute the following sequential steps, simulating a real-world developer workflow:

### Step 1: Zero-Touch Onboarding (V1-T01)
- **Action:** Simulate a GitHub Webhook payload for `installation_repositories` being sent to the Worker.
- **Assertion:** The Worker must scan the mock repository, detect an `openapi.yaml`, and successfully generate a payload to open a Pull Request adding `substrate.yaml`.

### Step 2: AI Architect Scaffolding (V1-T02)
- **Action:** Execute the `substrate init --design` CLI command programmatically.
- **Mock:** The OpenAI API endpoint must be mocked to return a valid OpenAPI 3.0 YAML string.
- **Assertion:** Verify that the output OpenAPI file and `substrate.yaml` are correctly written to the local filesystem.

### Step 3: Schema Break & Diff Generation (V1-T03)
- **Action:** Introduce a breaking change (e.g., remove a required field) to the OpenAPI spec. Trigger the Diff Engine.
- **Assertion:** The engine must detect the breaking change, generate a `DiffReport`, and store the JSON in the database, returning a valid `DiffID` for the Interactive UI.

### Step 4: Local Developer Remediation (V1-T05)
- **Action:** The developer uses the local CLI `substrate validate` to test a fix (restoring the field).
- **Assertion:** The CLI must return an exit code `0` (Safe) for the fixed schema.

### Step 5: Deployment Safety Gate (V1-T04)
- **Action:** Execute `substrate check-deploy` CLI command simulating a **provider** (e.g., `backend-api`) attempting to deploy a breaking change to production.
- **Mock State:** The registry must be seeded with a **consumer** (e.g., `frontend`) that is currently deployed in production and depends on the old schema.
- **Assertion:** The command must return exit code `1` (Blocked) and `HTTP 409 Conflict`, successfully preventing the provider from deploying and breaking the consumer.
