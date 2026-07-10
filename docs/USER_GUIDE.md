# Substrate User Guide

Welcome to the Substrate official documentation. Substrate is a proactive CI/CD gatekeeper that uses semantic diffing, cross-repo dependency graphs, and AI intelligence to prevent schema breaking changes from ever reaching production.

---

## 1. Getting Started

### What is Substrate?
Substrate analyzes your pull requests. If you alter an API endpoint, a GraphQL type, a SQL column, or a Kafka topic, Substrate identifies every downstream repository that relies on your code. If your change breaks them, Substrate blocks your PR and generates a safe AI-powered remediation patch.

### Installation

**1. GitHub App (Recommended)**
Install the [Substrate GitHub App](https://github.com/marketplace/actions/substrate-api-contract-guard) on your organization.

**2. CLI Binary (Local Dev)**
```bash
go install github.com/KrushnaVardhanReddy/substrate/engine/cmd/substrate@latest
```

---

## 2. Configuration Reference (`substrate.yaml`)

The `substrate.yaml` file lives in the root of your repository. It is entirely optional; without it, Substrate runs with sensible defaults.

```yaml
service: my-users-api           # Required: The canonical name of your service
schema_type: openapi            # Required: openapi | graphql | sql | protobuf | avro | asyncapi
spec_path: docs/openapi.yaml    # Required: Path to your schema file

# --- Phase 3: Consumers (Manual Declarations) ---
consumers:
  - name: frontend-dashboard
    provider_repo: myorg/frontend

# --- Phase 4: Traffic-Aware Diffing ---
traffic:
  provider: prometheus
  endpoint: "http://prometheus.internal:9090"
  lookback_days: 30
  downgrade_threshold: 0        # Changes to endpoints with 0 traffic are downgraded to WARNING

# --- Phase 4: Compliance Auditing ---
compliance:
  slack_webhook: "https://hooks.slack.com/services/..."
  security_channel: "#security-alerts"

# --- Intentional Breaking Changes ---
overrides:
  - rule_id: ENDPOINT_REMOVED
    path: paths./legacy/api
    reason: "Consumers migrated to v2"
    expires: 2026-12-31
```

---

## 3. CLI Reference

If you are running Substrate locally or in a custom pipeline, use the CLI:

### `substrate diff`
Compares two schema files and exits with code `2` if a breaking change is found.
```bash
substrate diff --base=main.yaml --head=pr.yaml --type=openapi
```

### `substrate serve`
Starts the Diff Engine in HTTP Server mode (used by the GitHub App worker).
```bash
substrate serve --port=8080
```

### `substrate init`
Scaffolds a `substrate.yaml` configuration file interactively.
```bash
substrate init
```

---

## 4. Contract Registry

The Substrate Contract Registry (Phase 3) enables **Cross-Repo Impact Analysis**.

1. When a consumer (e.g., `iOS-App`) pushes its `substrate.yaml` to `main`, Substrate records its dependency on your API.
2. When you open a PR on the API, Substrate checks the Registry.
3. If your change breaks the `iOS-App`, the GitHub App will print a **Compatibility Matrix** directly in your PR comment, explicitly listing `iOS-App` as a blocked consumer.

---

## 5. Rule Reference

Substrate supports over 100 semantic breaking change rules across 8 different schema adapters.

| Schema Type | Substrate Engine | Example Breaking Rule | Example Safe Rule |
|-------------|------------------|-----------------------|-------------------|
| **OpenAPI** | `oasdiff` | `ENDPOINT_REMOVED` | `ENDPOINT_ADDED` |
| **GraphQL** | `gqlparser` | `GQL_FIELD_REMOVED` | `GQL_TYPE_ADDED` |
| **SQL (PG)**| `pg_query_go` | `COLUMN_REMOVED` | `TABLE_CREATED` |
| **Protobuf**| `buf` | `PROTO_FIELD_TYPE_CHANGED` | `PROTO_FIELD_ADDED` |
| **AsyncAPI**| `parser-go` | `ASYNCAPI_CHANNEL_REMOVED` | `ASYNCAPI_CHANNEL_ADDED` |
| **Avro**    | Custom | `AVRO_INCOMPATIBLE` | `AVRO_COMPATIBLE` |

*(Note: AI/ML Models and Salesforce Enterprise metadata are also natively supported).*

---

## 6. AI & Shift-Left

Substrate is powered by a local, on-premise AI intelligence layer designed to keep your code private while providing cutting-edge remediation.

### AI PR Remediation
When Substrate blocks your PR, it doesn't just leave you guessing. The AI agent analyzes the exact rule violation and your schema context to generate an **AI Impact Analysis**. 
It will suggest a **Safe Remediation** code block (e.g., "Add `@deprecated: true` instead of deleting the field") that you can copy and paste to instantly resolve the CI failure.

### The AI Playground
The Substrate Dashboard features an interactive AI Playground. 
Paste your current schema and proposed schema, click "Analyze with Substrate AI", and watch the Server-Sent Events (SSE) stream print out the AI's thought process, the breaking findings, and the auto-generated code fix in real-time.

### Shift-Left VSCode Extension
Don't wait for CI. Install the Substrate VSCode Extension.
When you delete a required field in your `openapi.yaml` and hit save, the extension instantly places a red squiggly line under the code and displays a tooltip warning you of the blast radius before you even type `git commit`.
