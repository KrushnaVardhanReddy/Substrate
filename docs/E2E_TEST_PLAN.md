# Substrate End-to-End (E2E) Live Testing Plan

To guarantee Substrate works in a real-world environment without any mocked data, we will perform a live, end-to-end test. This plan outlines the exact steps to spin up the local infrastructure and simulate real GitHub PRs across all 8 supported schema types — **including the Phase 3 cross-repo consumer check.**

---

## ✅ Prerequisites — Before You Start

Set these environment variables (copy from `.env.local`). The Worker will silently fail without them.

```bash
# GitHub App credentials
export GITHUB_APP_ID=<your_app_id>
export GITHUB_PRIVATE_KEY="$(cat path/to/private-key.pem)"
export GITHUB_WEBHOOK_SECRET=<your_webhook_secret>

# Registry API (must match INTERNAL_SERVICE_TOKEN on the api/ server)
export REGISTRY_API_URL=http://localhost:8090
export REGISTRY_API_TOKEN=local-dev-token
export INTERNAL_SERVICE_TOKEN=local-dev-token

# Database
export DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres
```

---

## 1. Local Infrastructure Setup

Run all three core components of Substrate locally in separate terminals.

### Terminal 1 — PostgreSQL Registry Database

```bash
docker run --name substrate-db \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=postgres \
  -d postgres

# Run database migrations (wait ~5s for postgres to start first)
cd api && go run cmd/migrate/main.go up
# Alternative if using golang-migrate CLI:
# migrate -path ./migrations -database "$DATABASE_URL" up
```

Verify: `psql $DATABASE_URL -c "\dt"` should list `organizations`, `repositories`, `contracts`, `dependencies` tables.

### Terminal 2 — Go Registry API Server

```bash
cd api && go run cmd/server/main.go
# Should log: "Registry API listening on :8090"
```

Verify: `curl http://localhost:8090/health` should return `{"status":"ok"}`.

### Terminal 3 — Go Diff Engine (Container Service)

```bash
cd engine && go run ./cmd/substrate serve
# Should log: "substrate serve: listening on :8080"
```

### Terminal 4 — SvelteKit Dashboard

```bash
cd dashboard && npm run dev
# Should log: "Local: http://localhost:3000"
```

Verify: Open `http://localhost:3000` — should show the Substrate dashboard.

---

## 2. Webhook Tunneling (The Cloudflare Worker)

We need GitHub to send real webhooks to our local Cloudflare Worker.

### Terminal 4 — ngrok tunnel

```bash
ngrok http 8787
# Note the forwarding URL: e.g. https://abc123.ngrok.io
```

### Terminal 5 — Cloudflare Worker (local dev)

```bash
cd github-app && npx wrangler dev
# Worker listens on http://localhost:8787
```

### Update GitHub App webhook URL

1. Go to **GitHub → Settings → Developer Settings → GitHub Apps → Substrate**
2. Set **Webhook URL** to your ngrok URL (e.g. `https://abc123.ngrok.io`)
3. Confirm webhook secret matches `GITHUB_WEBHOOK_SECRET`

---

## 3. "Test Labs" Setup (Real GitHub Repos)

Create two blank repositories on GitHub:
- `substrate-test-provider` — simulates the backend API team
- `substrate-test-consumer` — simulates the frontend/mobile team

Install the Substrate GitHub App on **both** repositories.

---

## 3b. Cross-Repo Consumer Registration ⭐ (The Phase 3 Killer Feature)

This is the most critical thing to validate. It proves that a breaking change in `provider` is detected on the `consumer` — without the consumer having an open PR at all.

### Step 1 — Register the consumer's dependency

In `substrate-test-consumer`, create a `substrate.yaml` declaring its dependency on the provider:

```yaml
# substrate-test-consumer/substrate.yaml
service: test-consumer
schema_type: openapi
spec_path: openapi.yaml

consumers:
  - name: "test-provider-api"
    provider_repo: "<your-github-org>/substrate-test-provider"
    schema_type: openapi
    provider_spec_path: "openapi.yaml"
    provider_branch: "main"
```

Also create a baseline `openapi.yaml` in `substrate-test-consumer/main`:

```yaml
openapi: "3.0.0"
info:
  title: Consumer App
  version: "1.0.0"
paths: {}
```

Push both files to `main`. **Watch the Worker logs (Terminal 5)** — you should see:
```
[push] Detected push to main on substrate-test-consumer
[registry-sync] Synced 1 dependency snapshot(s) to registry
```

Verify in the database:
```sql
SELECT r.full_name, c.spec_path, c.schema_type FROM contracts c
JOIN repositories r ON c.repo_id = r.id;
```
You should see the provider's `openapi.yaml` snapshot stored.

### Step 2 — Open a breaking PR on the provider

In `substrate-test-provider`, create `openapi.yaml` on `main`:

```yaml
openapi: "3.0.0"
info:
  title: Provider API
  version: "1.0.0"
paths:
  /users:
    get:
      summary: List users
      responses:
        "200":
          description: OK
  /users/{id}:
    get:
      summary: Get user by ID
      responses:
        "200":
          description: OK
```

Then open a PR that **removes** `/users/{id}`:

```yaml
# PR branch — openapi.yaml (breaking change)
paths:
  /users:
    get:
      summary: List users
      responses:
        "200":
          description: OK
  # /users/{id} has been removed — BREAKING
```

### Step 3 — Verify the cross-repo PR comment

On the provider's PR, Substrate should post **two sections** in the comment:

```
## ❌ Breaking Changes (1)

ENDPOINT_REMOVED
Path: GET /users/{id}
...

---

## 🌐 Cross-Repo Impact

This change affects 1 registered consumer(s):

| Consumer | Status | Breaking Changes |
|---|---|---|
| substrate-test-consumer | ❌ BREAKING | GET /users/{id} removed |

> ⚠️ Action required: Coordinate with the substrate-test-consumer team before merging.
```

**The `substrate/breaking-changes` status check must FAIL on the provider's PR.**

This validates: webhook → Worker → Registry cross-repo check → PR comment → status check. ✅

---

## 4. The Live E2E Matrix (Testing All 8 Adapters)

With the infrastructure running and cross-repo check validated, test the full diff engine by pushing different file types to `substrate-test-provider` and opening PRs that break them.

| Adapter | Test File | Safe Change (Merged) | Breaking Change PR (Blocked) | Expected Rule |
| :--- | :--- | :--- | :--- | :--- |
| **OpenAPI** | `openapi.yaml` | Add a new `/health` endpoint | Delete an existing `/users` endpoint | `ENDPOINT_REMOVED` |
| **GraphQL** | `schema.graphql` | Add a new `type Post` | Remove a field from `type User` | `GQL_FIELD_REMOVED` |
| **Protobuf** | `user.proto` | Add a new optional field | Change a field's type from `int32` to `string` | `PROTO_FIELD_TYPE_CHANGED` |
| **Avro** | `user.avsc` | Add a field with a default | Remove a required field | `AVRO_FIELD_REMOVED` |
| **SQL (PG)** | `schema.sql` | `CREATE TABLE logs...` | `ALTER TABLE users DROP COLUMN id;` | `COLUMN_REMOVED` |
| **Terraform** | `main.tf` | Add an S3 bucket | Change the provider version / delete an output | `TF_OUTPUT_REMOVED` |
| **AI/ML** | `model.yaml` | Add a new tag | Change the required input tensor shape | `AIML_INPUT_SHAPE_CHANGED` |
| **Enterprise** | `Account.object` | Add a custom field | Change field type from `Text` to `Number` | `SFDC_FIELD_TYPE_CHANGED` |

For each adapter:
1. ✅ Push the **safe change** to `main` → CI should pass (exit 0 or 1)
2. ❌ Open a **breaking change PR** → CI should fail (exit 2), PR comment should appear

---

## 5. Testing the MCP Server (AI IDE Integration)

> **Scope note:** This test validates the `check_compatibility` diff tool (which runs the real diff engine). The registry lookup tools (`get_dependency_graph`, `get_breaking_change_history`) still return mock data until the MCP tools are wired to the live registry in a future task.

### Step 1 — Build the binary

```bash
cd engine && go build -o substrate-mcp ./cmd/substrate-mcp/
```

### Step 2 — Configure Claude Desktop

Edit `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "substrate": {
      "command": "/absolute/path/to/engine/substrate-mcp"
    }
  }
}
```

Restart Claude Desktop.

### Step 3 — Run the compatibility check prompt

In Claude Desktop, send:

> *"What happens if I merge a PR that deletes the `/users` endpoint in my OpenAPI spec?"*

Provide the breaking OpenAPI YAML as context. Claude will autonomously call `check_compatibility` with `schema_type: openapi` and report back the exact breaking changes and blast radius.

**Expected behaviour:** Claude calls the MCP tool, tool runs the real diff engine, returns `ENDPOINT_REMOVED` in the response, Claude explains the impact in plain English.

---

## 6. Dashboard Verification

With real data in the registry from Steps 3b and 4:

1. Open `http://localhost:3000`
2. Verify the repos list shows `substrate-test-provider` and `substrate-test-consumer`
3. Verify the dependency edge `consumer → provider` is visible in the graph view
4. Verify the breaking change from Step 3b appears in the history/audit section

---

## Pass Criteria Summary

| Check | Pass Condition |
|---|---|
| Infra boots cleanly | PostgreSQL, API server, dashboard, Worker all start with no errors |
| Push event triggers sync | Consumer's `substrate.yaml` push → registry snapshot stored in DB |
| Provider PR → single-repo check | PR comment appears with `ENDPOINT_REMOVED` |
| Provider PR → cross-repo check | PR comment includes 🌐 section listing consumer as BREAKING |
| Status check blocks merge | `substrate/breaking-changes` = ❌ failure on breaking PR |
| All 8 adapters pass safe change | CI exits 0 or 1 for each safe change |
| All 8 adapters block breaking change | CI exits 2 for each breaking PR |
| MCP `check_compatibility` works | Claude correctly identifies the breaking change |
| Dashboard shows live data | Repos, dependency edge, and breaking change history visible |
