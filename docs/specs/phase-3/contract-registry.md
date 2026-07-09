# Phase 3 Spec: Cross-Repo Contract Registry

> **Status:** 🔄 IN PROGRESS  
> **Depends on:** Phase 2 GitHub App ✅ Complete  
> **Goal:** Substrate becomes a cross-repo orchestration platform — blocking backend PRs that break frontend, mobile, or any downstream consumer automatically.

---

## The Core Problem

A backend engineer deletes a required field from `openapi.yaml` and opens a PR.  
The frontend, mobile app, and analytics pipeline all depend on that field.  
**Nobody manually checks downstream repos. Production breaks.**

Substrate Phase 3 closes this gap entirely — automatically.

---

## Priority Tiers

| Priority | Tasks | Why |
|---|---|---|
| 🔴 **P1 — Core (Ship First)** | P3-T01, P3-T02, P3-T02b, P3-T02c | The killer feature. Zero value without this. |
| 🟡 **P2 — Auth + Access** | P3-T06, P3-T02d | Required for multi-org use and safe rollout. |
| 🟢 **P3 — Dashboard** | P3-T03, P3-T04, P3-T05, P3-T11 | Visualisation layer. Valuable but backend is prerequisite. |
| 🔵 **P4 — Integrations** | P3-T08, P3-T09, P3-T10 | MCP / AI layer. Ship after dashboard is proven. |
| ⚪ **P5 — Ops** | P3-T07 | Production hardening + free tier limits. |

---

## 🔴 P1 — Core Registry (MUST SHIP FIRST)

These tasks ARE the product. Nothing else matters until this works.

### P3-T01: PostgreSQL Schema + Go API Server
**Owner:** Jules  
**Spec File:** `docs/specs/contract-registry.md` (this file)  
**Blocked by:** Nothing  

**What to build:**
- New Go service in `api/` directory (separate from the diff engine in `engine/`)
- HTTP server using `net/http` (no Gin/Echo dependency — keep it stdlib to honour the single-binary philosophy)
- PostgreSQL connection using `pgx/v5` (no heavy ORM — raw SQL with `sqlc` or `pgx` directly)
- DB migrations using `golang-migrate/migrate`

**PostgreSQL Schema (exact DDL):**

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE organizations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  github_installation_id BIGINT UNIQUE NOT NULL,
  github_org_name TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE repositories (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  org_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
  github_repo_id BIGINT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  full_name TEXT NOT NULL,  -- e.g. "myorg/backend-api"
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE contracts (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  repo_id UUID REFERENCES repositories(id) ON DELETE CASCADE,
  schema_type TEXT NOT NULL,       -- openapi | graphql | sql | protobuf
  spec_path TEXT NOT NULL,         -- e.g. "api/openapi.yaml"
  branch TEXT NOT NULL DEFAULT 'main',
  latest_commit_sha TEXT,
  raw_content TEXT NOT NULL,
  synced_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(repo_id, spec_path, branch)
);

CREATE TABLE dependencies (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  consumer_repo_id UUID REFERENCES repositories(id) ON DELETE CASCADE,
  provider_contract_id UUID REFERENCES contracts(id) ON DELETE CASCADE,
  status TEXT NOT NULL DEFAULT 'active',  -- active | broken | unknown
  last_checked_at TIMESTAMPTZ,
  UNIQUE(consumer_repo_id, provider_contract_id)
);
```

**Required API Endpoints:**

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/api/v1/sync` | Consumer syncs their schema snapshot to the registry |
| `POST` | `/api/v1/cross-repo-check` | Provider PR triggers cross-consumer compatibility check |
| `GET` | `/api/v1/repos/:org` | List all registered repos for an org |
| `GET` | `/api/v1/graph/:org` | Return dependency graph as JSON (for dashboard) |
| `GET` | `/health` | Health check (used by Fly.io / load balancer) |

---

### P3-T02: `substrate.yaml` Consumer Declaration Parser
**Owner:** Jules  
**Blocked by:** P3-T01  

**What to build:**
- Extend the existing `substrate.yaml` parser in the Go engine to support a new `consumers` block
- When the Cloudflare Worker processes a `push` to `main`, it reads this block and triggers `POST /api/v1/sync`

**New `substrate.yaml` format:**
```yaml
service: frontend
schema_type: openapi
spec_path: openapi.yaml

# NEW in Phase 3 — declare what this repo depends on
consumers:
  - name: "users-api"
    provider_repo: "myorg/backend-api"
    schema_type: openapi
    provider_spec_path: "api/openapi.yaml"
    provider_branch: "main"
  - name: "payments-api"
    provider_repo: "myorg/payments-service"
    schema_type: openapi
    provider_spec_path: "openapi.yaml"
    provider_branch: "main"
```

> **Note:** The naming is intentional. The `frontend` repo is declaring who it consumes. The backend repos need no changes at all — Substrate discovers them from the consumer side. This is the zero-config approach for providers.

---

### P3-T02b: Contract Registry Sync — Push to `main`
**Owner:** Jules  
**Blocked by:** P3-T01, P3-T02  

**What to build:**
- In the Cloudflare Worker (`github-app/src/index.ts`), handle the `push` GitHub webhook event
- On a push to `main`, read `substrate.yaml` from the pushed commit
- For each `consumers` entry, fetch the provider's current spec from GitHub using the GitHub App token
- POST the snapshot to `POST /api/v1/sync` on the Registry API server

**New Worker environment variables (add to `Env` interface in `types.ts`):**
```
REGISTRY_API_URL    — base URL of the Go Registry API (e.g. https://api.substrate.dev)
REGISTRY_API_TOKEN  — internal service token (matches INTERNAL_SERVICE_TOKEN on api/ server)
```

**Push event detection:**
- GitHub-Event header must equal `"push"`
- `ref` must be `"refs/heads/main"` — ignore feature branch and tag pushes
- `payload.deleted` must not be `true` — ignore branch deletion events

**New TypeScript types (add to `types.ts`):**
```ts
interface ConsumerEntry {
  name: string;
  provider_repo: string;        // "myorg/backend-api"
  schema_type: string;          // openapi | graphql | sql | protobuf
  provider_spec_path: string;   // "api/openapi.yaml"
  provider_branch: string;      // default "main"
}

interface SyncDependency {
  provider_repo: string;
  provider_github_repo_id: number;
  schema_type: string;
  spec_path: string;
  branch: string;
  raw_content: string;
}

interface SyncRequest {
  installation_id: number;
  org: string;
  consumer_repo: string;            // full_name "myorg/frontend"
  consumer_github_repo_id: number;
  commit_sha: string;
  dependencies: SyncDependency[];
}
```

**Full push handler flow:**
1. Validate webhook signature (same as PR handler)
2. Check `X-GitHub-Event: push` → if not push, fall through to PR handler
3. Parse push payload — if `ref !== "refs/heads/main"` or `deleted === true`, return 200 Ignored
4. Generate installation token using existing `generateInstallationToken()`
5. Fetch `substrate.yaml` from pushed commit SHA using existing `fetchFileContent()`
6. Parse `consumers` block — if empty or missing, return 200 Ignored
7. For each consumer entry (in parallel via `Promise.all`):
   a. Fetch provider repo metadata: `GET https://api.github.com/repos/{provider_repo}` → extract `.id`
   b. Fetch provider spec file: `fetchFileContent(token, owner, repo, entry.provider_spec_path, entry.provider_branch)`
   c. Build `SyncDependency` object
8. `POST {REGISTRY_API_URL}/api/v1/sync` with `Authorization: Bearer {REGISTRY_API_TOKEN}`
9. Return 200 with `{ synced: N }`

**Error handling:** Wrap entire push handler in `try/catch`. On any error: `console.error` + return 200 (never 5xx to GitHub).

**New files:**
- `github-app/src/registry-client.ts` — `parseConsumersFromYaml()` + `syncToRegistry()`

**YAML parsing approach:** Regex-based stub (same pattern as existing `parseYaml()` in `index.ts`).
Do NOT add a full YAML library — parse the `consumers:` block with targeted regex.

**Test requirements (Vitest):**
- `registry-client.test.ts`: 5 cases for `parseConsumersFromYaml()` (empty, no consumers, single, missing branch defaults to main, multiple)
- `registry-client.test.ts`: 2 cases for `syncToRegistry()` (200 success, non-2xx throws)
- `webhook.test.ts`: 5 new cases for push event parsing (main push, feature branch, tag, deleted, non-push)
- `index.test.ts`: 5 new integration cases for push handler (consumers found, no substrate.yaml, no consumers, non-main branch, existing PR tests must still pass)

**Files to create/modify:**
- CREATE: `github-app/src/registry-client.ts`
- CREATE: `github-app/src/registry-client.test.ts`
- MODIFY: `github-app/src/types.ts` (add Env fields + interfaces above)
- MODIFY: `github-app/src/webhook.ts` (add `parsePushEvent()`)
- MODIFY: `github-app/src/index.ts` (add push handler before PR handler)
- MODIFY: `github-app/src/webhook.test.ts` (add push test cases)
- MODIFY: `github-app/src/index.test.ts` (add push integration tests)

**Flow diagram:**
```
push to frontend/main
    ↓
Worker validates HMAC → reads substrate.yaml from commit SHA
    ↓
Worker parses consumers block → [{ provider_repo: "myorg/backend-api", ... }]
    ↓
Worker fetches provider spec from GitHub API (parallel for all consumers)
    ↓
Worker POSTs SyncRequest to Registry API (/api/v1/sync)
    ↓
Registry: UpsertOrg → UpsertRepo (consumer) → UpsertRepo (provider) →
          UpsertContract (provider spec snapshot) → UpsertDependency
    ↓
Worker returns 200 { synced: 1 }
```

---

### P3-T02c: Cross-Repo Compatibility Check — Provider PR
**Owner:** Antigravity (complex orchestration logic, done in this spec)  
**Blocked by:** P3-T01, P3-T02b  

**What to build:**
- In the Cloudflare Worker, after running the single-repo diff on a PR, also call `POST /api/v1/cross-repo-check`
- The Registry API fetches all registered consumers that depend on this provider
- For each consumer, runs `DiffSchemas()` using provider's PR head schema vs consumer's stored snapshot
- Returns aggregated `ConsumerCompatibilityMatrix`
- Worker appends a **cross-repo impact section** to the PR comment

**PR Comment Format (cross-repo section):**
```
## 🌐 Cross-Repo Impact

This change affects **2 registered consumers**:

| Consumer | Status | Breaking Changes |
|---|---|---|
| frontend | ❌ BREAKING | `GET /users/{id}` — `email` field removed |
| mobile-app | ✅ Safe | No breaking changes detected |

> ⚠️ **Action required:** Coordinate with the `frontend` team before merging.
> The `substrate/breaking-changes` check is now FAILING.
```

**Registry API — `POST /api/v1/cross-repo-check` full spec:**

Request:
```json
{
  "org": "myorg",
  "provider_repo": "myorg/backend-api",
  "head_schema_content": "openapi: 3.0.0 ...",
  "schema_type": "openapi"
}
```

Response:
```json
{
  "total_consumers": 2,
  "broken_consumers": 1,
  "is_safe": false,
  "results": [
    {
      "consumer_repo": "myorg/frontend",
      "status": "breaking",
      "diff_report": {
        "breaking": [{ "rule": "ENDPOINT_MODIFIED", "path": "GET /users/{id}", "message": "field email removed" }],
        "warning": [],
        "info": [],
        "summary": { "breaking_count": 1, "warning_count": 0, "info_count": 0 }
      }
    },
    {
      "consumer_repo": "myorg/mobile-app",
      "status": "safe",
      "diff_report": { "breaking": [], "warning": [], "info": [], "summary": { "breaking_count": 0, "warning_count": 0, "info_count": 0 } }
    }
  ]
}
```

---

## 🟡 P2 — Auth + Testing

### P3-T06: GitHub OAuth + Org Management
**Owner:** Jules  
**Blocked by:** P3-T01  
**Why P2:** Without auth, the registry is open to anyone. Must be in place before any public launch.

**What to build:**
- GitHub OAuth 2.0 flow — users log in with their GitHub account
- Org-level access control — users can only see/manage repos within their GitHub org/installation
- Internal service token for Worker → Registry API communication (separate from OAuth)
- JWT session management

---

### P3-T02d: Cross-Repo E2E Fixture Tests
**Owner:** Jules  
**Blocked by:** P3-T02c  
**Why P2:** Validates the most critical flow without needing real infra.

**What to build:**
- Create `engine/cmd/substrate/testdata/cross-repo/` directory
- Add fixture files simulating multi-repo scenario:
  - `provider_base.yaml` (backend-api before change)
  - `provider_head.yaml` (backend-api after breaking change)
  - `consumer_snapshot.yaml` (frontend's stored snapshot — matches provider_base)
- E2E test: assert that `DiffSchemas(provider_head, consumer_snapshot)` returns the expected breaking change
- This tests the cross-repo check logic before the full registry infrastructure is live

---

## 🟢 P3 — Dashboard

### P3-T03: SvelteKit Project Setup + Design System
**Owner:** Antigravity  
**Blocked by:** P3-T01 (needs API to wire to)  

**What to build:**
- New `dashboard/` directory at repo root
- SvelteKit + TypeScript + Vite setup
- Design system: dark theme, teal accent (#00BFA5), Inter font
- Core layout: sidebar (repos list) + main panel + top nav
- Auth integration with P3-T06 GitHub OAuth

---

### P3-T04: Connected Repos List + Schema Browser
**Owner:** Antigravity + Jules  
**Blocked by:** P3-T03, P3-T06  

**What to build:**
- Dashboard page showing all repos registered with Substrate for the org
- For each repo: name, schema type, last synced, status (healthy/broken)
- Schema browser: view the raw stored snapshot content for any repo + diff view between snapshots

---

### P3-T05: Dependency Graph Visualization
**Owner:** Antigravity  
**Blocked by:** P3-T04  

**What to build:**
- Interactive node graph using `d3.js` or `cytoscape.js`
- Nodes = repositories, edges = dependencies (with directional arrows)
- Click a node → show what it provides and what it consumes
- Colour coding: green = all safe, red = breaking dependencies exist
- Data source: `GET /api/v1/graph/:org`

---

### P3-T11: Compatibility Matrix Dashboard ⭐ Enterprise Feature
**Owner:** Antigravity  
**Blocked by:** P3-T05  

**What to build:**
- Table view: rows = provider versions, columns = consumer repos
- Each cell = green tick (safe) or red cross (breaking)
- Inspired by PactFlow's compatibility matrix — the #1 enterprise upsell feature
- Answers: "Which version of `backend-api` is safe to deploy to production given what `frontend` and `mobile-app` currently depend on?"

---

## 🔵 P4 — MCP Server

### P3-T08: MCP Server Spec
**Owner:** Antigravity  
**Blocked by:** P3-T04 (needs real data)  

Create `docs/specs/mcp-server.md` defining the MCP tools the server exposes.

**Core MCP Tools:**
- `get_dependency_graph(org)` — returns full graph as structured JSON
- `check_compatibility(provider_repo, head_schema_content)` — runs cross-repo check
- `get_schema_history(repo, path)` — returns schema evolution over time
- `get_breaking_change_history(repo)` — returns all historical breaking changes

---

### P3-T09: MCP Server Implementation
**Owner:** Jules  
**Blocked by:** P3-T08  

- Implement in Go, integrated into the existing `api/` server
- Expose at `POST /mcp` as JSON-RPC 2.0

---

### P3-T10: MCP Deployment + IDE Integration Docs
**Owner:** Antigravity  
**Blocked by:** P3-T09  

- Document how to configure Cursor, Claude, Copilot to use Substrate's MCP server
- Show how an AI can answer "what breaks if I change this?" using Substrate as a knowledge source

---

## ⚪ P5 — Production Ops

### P3-T07: Free Tier Limits + Production Deployment
**Owner:** Jules  
**Blocked by:** P3-T06  

**What to build:**
- Free tier: 3 repos, 5 dependencies, 30-day history
- Enforcement in the Registry API middleware
- Deploy the Go API server to Fly.io (separate from the existing diff-engine container)
- Add `DATABASE_URL` secret management for Fly.io

---

## Execution Order (Bottom-Up)

```
P3-T01 (DB + API Server)
    ↓
P3-T02 (substrate.yaml parser) + P3-T06 (OAuth)  ← run parallel
    ↓
P3-T02b (sync on push)
    ↓
P3-T02c (cross-repo check on PR) + P3-T02d (E2E tests)  ← run parallel
    ↓
P3-T03 (Dashboard scaffold)
    ↓
P3-T04 (Repos list) → P3-T05 (Graph) → P3-T11 (Matrix)
    ↓
P3-T08 (MCP Spec) → P3-T09 (Implementation) → P3-T10 (Docs)
    ↓
P3-T07 (Production hardening)
```

## Exit Criteria for Phase 3

- [ ] Backend engineer opens a PR that removes an endpoint used by the frontend
- [ ] Substrate automatically posts a PR comment listing the frontend as a broken consumer
- [ ] The `substrate/breaking-changes` status check fails on the backend PR
- [ ] The backend engineer is blocked from merging until the breaking change is resolved
- [ ] The dependency graph dashboard shows the frontend → backend edge as red
