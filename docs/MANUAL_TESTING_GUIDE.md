# Substrate V1.0 Manual Testing Guide

This guide is designed for solo developers to manually verify the end-to-end functionality of the Substrate platform prior to a production release. It covers every major feature across all completed phases based strictly on the finalized specifications.

## How to Use This Guide
- Set up a clean environment (reset database, fresh terminal windows).
- Follow each phase sequentially.
- Check off each box `[x]` as you verify the behavior.

---

## 🏗️ Phase 1: Diff Engine & Parsers (Local CLI)

**Goal:** Ensure the core `substrate` engine accurately detects schema changes across all supported formats.

- [ ] **OpenAPI (REST):** Run `substrate diff` with a removed required field. Verify output shows `BREAKING_CHANGE`.
- [ ] **PostgreSQL (SQL):** Run `substrate diff` with a dropped column or changed data type. Verify output shows `BREAKING_CHANGE`.
- [ ] **GraphQL:** Run `substrate diff` with a removed field or an argument made required.
- [ ] **Protobuf / gRPC:** Run `substrate diff` removing a field number. Verify it flags as `BREAKING_CHANGE`.
- [ ] **AsyncAPI / Avro:** Run `substrate diff` with a removed channel or altered message property.
- [ ] **AI/ML Model Contracts:** Run `substrate diff` on a `.ml_model` schema with altered input tensor dimensions. Verify it triggers a break.
- [ ] **Enterprise Metadata (Salesforce/SOAP):** Run `substrate diff` on WSDL/XML dropping a `<CustomObject>`.
- [ ] **Terraform (IaC):** Verify resource recreation (e.g., `force_destroy = true` or immutable attribute change) is flagged correctly.
- [ ] **CLI Config Init:** Run `substrate init` in an empty repo. Verify it scaffolds a clean `substrate.yaml` config and GitHub Actions workflow.
- [ ] **Override Config (`substrate.yaml`):** Configure an override for a specific breaking change. Run `substrate diff`. Verify it passes with `✅ Acknowledged`.
- [ ] **Pending Acknowledgments:** Set an override status to `pending` with an `until: <date>` in the future. Verify it is permitted, but warns of expiration.

---

## 🐙 Phase 2: GitHub App Integration

**Goal:** Ensure PR comments and webhook statuses are triggered correctly on GitHub.

- [ ] **`substrate serve` Mode:** Run the local HTTP binary. Post a diff payload to `POST /diff`. Verify valid JSON response.
- [ ] **Webhook Receipt:** Open a PR with a breaking schema change. Verify the Cloudflare Worker receives the webhook.
- [ ] **PR Comment Posting:** Verify the bot posts a rich Markdown comment to the PR outlining the breaking changes.
- [ ] **Required Status Check:** Verify the GitHub Commit Status turns `🔴 Failed/Error` when a breaking change is unacknowledged.
- [ ] **Status Check Bypass:** Update `substrate.yaml` in the PR to acknowledge the break. Push. Verify the Commit Status turns `🟢 Success`.

---

## 🌐 Phase 3: Contract Registry & Dashboard

**Goal:** Verify cross-repo consumer checks, OAuth, and graphical UI.

- [ ] **GitHub OAuth Login:** Open the Dashboard UI in an incognito window. Click "Log in with GitHub". Verify the OAuth flow completes successfully.
- [ ] **Organization Management:** Verify that upon login, you can view the repositories specifically tied to your GitHub Organization (multi-tenant isolation).
- [ ] **Registry Sync:** Push a schema update to the `main` branch. Verify the Go API Server logs that the schema was snapshotted in PostgreSQL.
- [ ] **Cross-Repo Detection:** Open a breaking PR in a **provider** repository. Ensure the bot detects that a **consumer** repository relies on the broken field.
- [ ] **Dashboard Repositories:** Launch the Svelte UI (`localhost:5173`). Verify the repositories list correctly shows connected repos.
- [ ] **Dependency Graph:** Verify the visual node graph correctly links the provider and consumer repositories.
- [ ] **Compatibility Matrix:** View the matrix UI and ensure older consumer versions show red `❌` against the new provider breaking change.
- [ ] **Execution Modes:** Run the CLI with `--mode=legacy`. Verify that breaking changes only issue warnings (dry-run) and do not fail the CI exit code.
- [ ] **Rate Limiting:** Hit the API rapidly to trigger the Free Tier limit. Verify an HTTP 429 Too Many Requests is returned.

---

## 🧠 Phase 4: AI Intelligence Layer (MCP & Playground)

**Goal:** Ensure foundation models can reason over schema data and specialized modifiers.

- [ ] **MCP Server - Dependency Graph:** Run an MCP Client (e.g., Claude Desktop). Use the `get_dependency_graph` tool and verify it returns correct JSON.
- [ ] **MCP Server - Compatibility Check:** Use the `check_compatibility` tool. Ensure it successfully executes the Go diff engine logic.
- [ ] **MCP Server - Live Registry Wire-up:** Verify tools like `get_schema_file` and `analyze_repository` fetch live data, not mock data.
- [ ] **AI Playground (Streaming):** Open the Playground UI in the Svelte dashboard. Select a schema type, intentionally break it, and click "Analyze". Verify the AI stream outputs findings and a suggested auto-fix.
- [ ] **VSCode Extension:** Open a schema in VSCode. Delete a required field. Verify inline warning squiggles appear.
- [ ] **Traffic-Aware Diffing:** Feed simulated Datadog/OTel traffic data indicating a field has 0 requests. Remove the field. Verify it downgrades from `BREAKING` to `WARNING (Unused)`.
- [ ] **Compliance Auditing:** Add a field named `ssn` or `medical_history`. Modify it. Verify the engine auto-tags the diff with `[HIPAA]` or `[PII]` warnings.

---

## 🧭 Phase 5: Automated Dependency Discovery

**Goal:** Verify that Substrate can automatically detect cross-repo dependencies without manual configuration.

- [ ] **Env Var Scanner:** Add a repository with `.env.example` containing `USERS_API_URL=https://api.myorg.com`. Verify the scanner maps this to the `users-api` repo.
- [ ] **Package Scanner:** Push a `package.json` with an internal `@myorg/auth` dependency. Verify it creates a link in the dependency graph.
- [ ] **Terraform Outputs:** Verify the Terraform scanner detects `output` URLs and correctly maps them to deployed API repos.
- [ ] **Event-Driven Discovery:** Map an AsyncAPI `$ref` URL or Kafka topic to its producer/consumer. Verify it appears in the graph.
- [ ] **Runtime Confirmation:** Feed mock OTel/eBPF traces matching static discovery. Verify the edge confidence score goes up.
- [ ] **Confidence Scoring UI:** View the dependency graph in the Svelte dashboard. Verify that edges show Confidence Scores (e.g., 85/100, High/Medium/Low).

---

## 🧪 Phase 6: QA & Automation Layer

**Goal:** Test the automated testing and validation integrations.

- [ ] **Postman Sync:** Trigger a schema update and verify Substrate correctly pushes the updated schema/folders to a Postman Collection.
- [ ] **Shadow API Coverage:** Run API traffic through a mock trace generator. View the Dashboard UI to ensure undocumented fields are highlighted as "Shadow Traffic".
- [ ] **Mock Server Time Machine:** Run `substrate mock start --repo users-api --commit HEAD~5`. Make a curl request to it and verify it returns a mock response matching the historical schema version.
- [ ] **Test Code Generation:** Provide a constrained schema. Trigger test generation and verify executable Playwright/Go tests are produced.

---

## 🛡️ V1.0 Pre-Flight Features (Final Polish)

**Goal:** Test the final deployment, onboarding, and validation gates.

- [ ] **Zero-Touch Onboarding (V1-T01):** Add a new repository to the GitHub App installation. Verify the bot automatically opens a PR with a `substrate.yaml` file.
- [ ] **CLI AI Architect (V1-T02):** Run `substrate init --design`. Provide a text prompt (e.g., "User login API"). Verify a valid OpenAPI file is generated interactively.
- [ ] **Interactive Diff Viewer (V1-T03):** Click the "Preview Link" generated in a GitHub PR comment. Verify the Dashboard opens a split-pane Diff Viewer showing Red/Green highlights.
- [ ] **Local Validation CLI (V1-T05):** Run `substrate validate` locally against a base branch. Verify the terminal outputs correct Red/Green formatting offline.
- [ ] **Deployment Safety Gate (V1-T04):** Simulate a CI/CD pipeline step running `substrate check-deploy --repo provider --env prod`.
    - **Test A:** Provider has a breaking change, consumer is *not* updated. Verify it exits with `1` (Blocked).
    - **Test B:** Provider has a breaking change, consumer *is* already updated. Verify it exits with `0` (Safe).
- [ ] **Legal & Licensing Audit (V1-T06):** Check that the `ATTRIBUTIONS.md` file exists. Open the Dashboard UI and click the Footer to ensure the "Credits / OSS Licenses" modal renders correctly.

---

## 📋 Final System Checks
- [ ] Run `make e2e-*` and ensure all programmatic Go tests pass across all adapters.
- [ ] Run `npm run test` in the `github-app` and `dashboard` directories.
