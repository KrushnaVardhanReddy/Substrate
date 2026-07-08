# Phase 3: Contract Registry Architecture Spec

## 1. Overview
The Contract Registry is the core of Substrate's cross-repo orchestration. It solves the downstream breakage problem by storing a continuous snapshot of consumer schemas and validating provider pull requests against all registered consumers.

### Goals:
- Store consumer contracts (e.g. `openapi.yaml`, `schema.graphql`) reliably.
- Sync schemas automatically from `main` branch pushes on consumer repositories.
- On a provider PR, fetch all registered consumer snapshots and run `DiffSchemas()` across boundaries.

---

## 2. PostgreSQL Schema

The registry uses a relational model to link providers (the source of truth) with consumers (the dependents).

### `organizations`
- `id` (UUID, PK)
- `github_installation_id` (BigInt)
- `name` (String)

### `repositories`
- `id` (UUID, PK)
- `org_id` (UUID, FK)
- `github_repo_id` (BigInt)
- `name` (String)

### `contracts` (The Provider's Spec)
- `id` (UUID, PK)
- `repo_id` (UUID, FK)
- `schema_type` (Enum: openapi, graphql, sql, etc)
- `path` (String) - e.g. `api/openapi.yaml`
- `latest_commit_sha` (String)
- `raw_content` (Text)

### `dependencies` (The Consumer's Link)
- `id` (UUID, PK)
- `consumer_repo_id` (UUID, FK)
- `provider_contract_id` (UUID, FK)
- `status` (Enum: active, broken)

---

## 3. Core Workflows

### Flow A: Consumer Sync (Main Branch Push)
When a repository merges a PR to `main`:
1. Substrate Worker receives `push` webhook.
2. Worker reads `substrate.yaml` to find `consumers` block.
3. For each consumer dependency, Worker fetches the latest consumer schema from GitHub.
4. Worker POSTs to Registry API: `POST /api/v1/sync`.
5. Registry API upserts the consumer schema snapshot into the `contracts` table and updates `dependencies`.

### Flow B: Cross-Repo Validation (Provider PR Opened)
When a provider repository opens a PR:
1. Substrate Worker receives `pull_request` webhook.
2. Worker fetches base and head schemas from GitHub.
3. Worker calls Container Service `POST /diff` (Single Repo check).
4. Worker POSTs to Registry API: `POST /api/v1/cross-repo-check`.
5. Registry API identifies all `dependencies` pointing to this provider.
6. For each consumer, Registry API runs the Diff Engine using the Provider's `head_schema` against the Consumer's stored snapshot.
7. Registry API aggregates results and returns a matrix of broken/safe consumers.
8. Worker posts PR comment highlighting downstream breaks.

---

## 4. API Endpoints (Go Server)

### `POST /api/v1/sync`
**Auth**: Internal Service Token
**Payload**:
```json
{
  "org": "myorg",
  "repo": "frontend",
  "commit_sha": "abc1234",
  "dependencies": [
    {
      "provider_repo": "backend-api",
      "schema_type": "openapi",
      "consumer_schema_content": "..."
    }
  ]
}
```

### `POST /api/v1/cross-repo-check`
**Auth**: Internal Service Token
**Payload**:
```json
{
  "org": "myorg",
  "provider_repo": "backend-api",
  "head_schema_content": "...",
  "schema_type": "openapi"
}
```
**Response**:
```json
{
  "total_consumers": 3,
  "broken_consumers": 1,
  "results": [
    {
      "consumer_repo": "frontend",
      "status": "broken",
      "breaking_changes": ["ENDPOINT_REMOVED"]
    }
  ]
}
```

---

## 5. Next Steps for Implementation
1. **P3-T01**: Scaffold the Go API server in `api/` directory with Gin/Echo and setup GORM or sqlc for PostgreSQL.
2. **P3-T02**: Implement the synchronization logic to populate the registry.
3. **P3-T02c**: Hook the API server into the `substrate serve` diff engine to run the cross-repo boundary checks.
