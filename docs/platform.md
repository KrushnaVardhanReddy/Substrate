# Substrate Platform Documentation

Substrate is a CI/CD-integrated data contract and dependency intelligence platform. It acts as a proactive firewall for engineering teams by analyzing schema changes, mapping cross-repository dependencies, and blocking unsafe changes before they break downstream consumers.

---

## 1. Getting Started

### What is Substrate?
In a microservices architecture, managing API contracts across different repositories is difficult. Substrate automatically tracks the "provider" (the service hosting an API) and the "consumers" (the services calling the API). 

When a provider opens a Pull Request that introduces a breaking change (e.g., deleting a column in a SQL database, or removing a field in an OpenAPI spec), Substrate intercepts the PR, identifies all registered consumers, calculates the semantic diff, and blocks the PR if it breaks consumer contracts.

### Installation

**Homebrew (macOS/Linux):**
```bash
brew install krushnavardhanreddy/tap/substrate
```

**Docker:**
```bash
docker pull krushnavardhanreddy/substrate:latest
```

**Go Install:**
```bash
go install github.com/krushnavardhanreddy/substrate/engine/cmd/substrate@latest
```

### CI/CD Quickstart (GitHub Actions)

Add this to your repository's `.github/workflows/substrate.yml`:

```yaml
name: Substrate Contract Guard
on: [pull_request]

jobs:
  check-contracts:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - name: Run Substrate
        uses: KrushnaVardhanReddy/Substrate@v1.0.0
        with:
          base_schema: openapi.yaml
          head_schema: openapi.yaml
          config: substrate.yaml
```

---

## 2. Configuration Reference

Substrate uses a `substrate.yaml` file located in the root of your repository to understand what schemas to track and validate.

### Provider Configuration
For a repository that exposes an API (e.g., `backend-api`):

```yaml
# substrate.yaml
service: backend-api
schema_type: openapi          # Supported: openapi, sql, graphql, protobuf, asyncapi, avro, terraform-plan, ai-model
spec_path: api/openapi.yaml   # Relative path to your schema
on_breaking_change: block     # block (fail PR) or warn (comment only)

owners:
  - team: platform-team
    contact: platform@company.com
```

### Consumer Configuration
For a repository that consumes an API (e.g., `frontend-dashboard`):

```yaml
# substrate.yaml
service: frontend-dashboard

consumers:
  - name: "users-api"
    schema_type: openapi
    provider_repo: "myorg/backend-api"      # The GitHub repo of the provider
    provider_spec_path: "api/openapi.yaml"
    provider_branch: "main"                 # Tracked branch for baseline
```

### Breaking Change Overrides
If you need to intentionally introduce a breaking change and have migrated all clients, you can override the rule:

```yaml
overrides:
  - rule_id: ENDPOINT_REMOVED
    path: "GET /api/v1/users/{id}"
    reason: "Endpoint deprecated. All clients migrated to v2."
    approved_by: "platform-team"
    expires: "2026-12-31"
```

---

## 3. CLI Reference

The Substrate CLI (`substrate`) allows you to test changes locally or run the engine in a pipeline.

### `substrate diff`
Compare two schema versions locally.
```bash
substrate diff --base=old.yaml --head=new.yaml --type=openapi
```

### `substrate check`
Check the current workspace against the contract registry.
```bash
substrate check --config=substrate.yaml
```

### `substrate serve`
Run Substrate as an HTTP server (used by the GitHub App orchestration layer).
```bash
substrate serve --port=8080 --mode=strict
```
_Flags:_
- `--mode=legacy` (Dry-run mode, will not block PRs)
- `--mode=strict` (Enforces blocking PRs on breaking changes)

### `substrate init`
Scaffold a new `substrate.yaml` file.
```bash
substrate init --type=openapi --path=api.yaml
```

---

## 4. Contract Registry (Phase 3)

The Contract Registry is the heart of Substrate's cross-repo intelligence. It stores snapshots of all consumer contracts.

**How it works:**
1. **Push-to-Main Sync:** When a consumer (e.g., `frontend`) merges a PR to `main`, a webhook notifies Substrate. Substrate reads the `substrate.yaml`, downloads the provider's current schema, and stores the relationship in the PostgreSQL registry.
2. **PR Cross-Repo Check:** When a provider (e.g., `backend-api`) opens a PR, Substrate asks the registry: "Who consumes this API?". It then diffs the PR's proposed schema against the stored consumer requirements.
3. **Compatibility Matrix:** The Substrate Dashboard (`http://localhost:5173/org/[org]/matrix`) provides a visual grid showing exactly which versions of providers are compatible with which consumers.

---

## 5. Rule Reference (Linter Rules)

Substrate uses semantic rules to determine if a change is breaking.

### OpenAPI Rules
| Rule ID | Severity | Description | Recommendation |
|---------|----------|-------------|----------------|
| `ENDPOINT_REMOVED` | BREAKING | An endpoint path was removed. | Use `@deprecated` instead of deleting. |
| `REQUEST_PARAM_REMOVED` | BREAKING | A required query or path parameter was removed. | Keep the parameter but make it optional. |
| `FIELD_REMOVED` | BREAKING | A property was removed from a response payload. | Deprecate the field and retain it in the schema. |
| `TYPE_CHANGED` | BREAKING | A field's data type changed (e.g., `integer` to `string`). | Create a new field with the new type. |
| `NEW_REQUIRED_PARAM` | BREAKING | A new required parameter was added. | Make the new parameter optional. |

### SQL Rules
| Rule ID | Severity | Description | Recommendation |
|---------|----------|-------------|----------------|
| `COLUMN_DROPPED` | BREAKING | A database column was dropped. | Do not drop columns until all consumers are updated. |
| `TABLE_DROPPED` | BREAKING | A database table was dropped. | Deprecate the table at the application layer. |
| `COLUMN_TYPE_CHANGED` | BREAKING | A column data type was modified. | Ensure the change is backward-compatible (e.g., `VARCHAR(50)` to `VARCHAR(100)`). |

### GraphQL Rules
| Rule ID | Severity | Description | Recommendation |
|---------|----------|-------------|----------------|
| `FIELD_REMOVED` | BREAKING | A field was removed from an object type. | Use the `@deprecated` directive. |
| `ARGUMENT_REMOVED` | BREAKING | An argument was removed from a field. | Keep the argument but ignore it in the resolver. |
| `NON_NULL_ADDED` | BREAKING | A previously nullable field is now non-null. | Keep it nullable and validate in business logic. |
