# P12-T12 — Full End-to-End System Matrix Plan (ALL Product Features)

> **Goal:** Prove Substrate is production-ready. Every feature tested with real data, zero mocking.  
> **Coverage:** 7 demo repos × 3 paths + WASM engine + Impact API + MCP tools + AI Autofix + Enterprise dashboard + Telemetry

---

## Current State

| Item | Status |
|---|---|
| Forgejo running | ✅ `localhost:3000` |
| System webhook configured | ✅ Points to worker URL |
| Worker running with `GITEA_TOKEN` | ✅ Token set |
| Repos in Forgejo | ⚠️ Only `admin/microservices-demo` exists |
| `substrate.yaml` in demo repos | ⚠️ Only `microservices-demo` has it (wrong org — `substrate-demo-testing` instead of `admin`) |
| Other 6 repos (graphql, openai, stripe, realworld, slack, jaffle) | ❌ Not pushed to Forgejo yet |

---

## What "Full E2E" Means

```
Test pushes real file change
        │
        ▼
  Forgejo (localhost:3000)
        │  System Webhook (push event)
        ▼
  Cloudflare Worker (localhost:8787)
        │  fetches substrate.yaml from Forgejo API
        │  fetches provider schema file from Forgejo API
        │  calls Go Engine POST /diff
        │  calls Go API POST /api/v1/sync
        ▼
  Go API (localhost:8090)
        │  River queue processes SyncWebhookJob
        │  Upserts org, repos, dependency edges in Postgres
        │  SSE broadcast to connected clients
        ▼
  Svelte Dashboard (localhost:5173)
        │  Receives SSE update
        │  Graph re-renders with new state
        ▼
  Playwright assertion
```

---

## Setup Required (One-time, done in test beforeAll)

### Step 1 — Create all 7 repos in Forgejo via API (idempotent)

### Step 2 — Fix substrate.yaml org and push all repos

| Local Path | Forgejo Repo | Schema File | Type |
|---|---|---|---|
| `demo-repos/microservices-demo` | `admin/microservices-demo` | `protos/demo.proto` | protobuf |
| `demo-repos/graphql-schema` | `admin/graphql-schema` | `schema.graphql` | graphql |
| `demo-repos/openapi` (Stripe) | `admin/stripe-openapi` | `openapi/spec3.yaml` | openapi |
| `demo-repos/openai-openapi` | `admin/openai-openapi` | `openapi.yaml` | openapi |
| `demo-repos/realworld` | `admin/realworld` | `specs/api/openapi.yml` | openapi |
| `demo-repos/slack-api-specs` | `admin/slack-api-specs` | `events-api/slack_events_api_async_v1.json` | asyncapi |
| `demo-repos/jaffle_shop` | `admin/jaffle-shop` | `models/customers.sql` | sql |

### Step 3 — substrate.yaml format per repo

Each repo needs its own `substrate.yaml`. Example for microservices-demo:
```yaml
consumers:
  - name: frontend
    provider_repo: admin/microservices-demo
    schema_type: protobuf
    provider_spec_path: protos/demo.proto
    provider_branch: main
  - name: checkoutservice
    provider_repo: admin/microservices-demo
    schema_type: protobuf
    provider_spec_path: protos/demo.proto
    provider_branch: main
  - name: recommendationservice
    provider_repo: admin/microservices-demo
    schema_type: protobuf
    provider_spec_path: protos/demo.proto
    provider_branch: main
  - name: emailservice
    provider_repo: admin/microservices-demo
    schema_type: protobuf
    provider_spec_path: protos/demo.proto
    provider_branch: main
```

---

## Test Matrix — 3 Paths Per Repo

### 1. `microservices-demo` — Protobuf / gRPC

**Consumers:** `frontend`, `checkoutservice`, `recommendationservice`, `emailservice`
**Provider spec:** `protos/demo.proto`
**Key field:** `CartItem.product_id` (field 1)

| Path | What we push | Expected dashboard |
|---|---|---|
| 🔴 Breaking | Remove `string product_id = 1` from `CartItem` | 4 consumer nodes turn red |
| 🟢 Safe | Add `string notes = 3` to `CartItem` (additive) | All nodes stay green |
| 🟡 Warning | Add `reserved 1;` reservation to `CartItem` | Warning badge on 4 consumers |

---

### 2. `graphql-schema` — GraphQL (GitHub schema)

**Consumer:** `admin/graphql-consumer`
**Provider spec:** `schema.graphql`
**Key type:** `User { login, email, ... }`

| Path | What we push | Expected dashboard |
|---|---|---|
| 🔴 Breaking | Remove a field from `User` type (e.g. `email`) | `graphql-consumer` node red |
| 🟢 Safe | Add `bio: String` to `User` type | Node stays green |
| 🟡 Warning | Add `@deprecated` to an existing field | Warning indicator on consumer |

---

### 3. `stripe-openapi` — OpenAPI 3.0 (Stripe — enterprise scale)

**Consumer:** `admin/stripe-consumer`
**Provider spec:** `openapi/spec3.yaml` (~60,000 lines — real shock-factor demo)
**Key paths:** `/v1/charges`, `/v1/customers`

| Path | What we push | Expected dashboard |
|---|---|---|
| 🔴 Breaking | Remove `/v1/charges` path | `stripe-consumer` node red |
| 🟢 Safe | Add `/v2/beta/charges` path | Node stays green |
| 🟡 Warning | Change response property from required → optional | Warning on consumer |

---

### 4. `openai-openapi` — OpenAPI 3.0 (OpenAI)

**Consumer:** `admin/ai-chatbot`
**Provider spec:** `openapi.yaml`
**Key endpoint:** `POST /v1/chat/completions` — `function_call` field

| Path | What we push | Expected dashboard |
|---|---|---|
| 🔴 Breaking | Remove `function_call` from request body | `ai-chatbot` node red — "AI integration will break" |
| 🟢 Safe | Add `metadata` optional field | Node stays green |
| 🟡 Warning | Deprecate `functions` parameter | Warning badge on consumer |

---

### 5. `realworld` — OpenAPI 3.1 (RealWorld / Conduit)

**Consumer:** `admin/conduit-mobile-app`
**Provider spec:** `specs/api/openapi.yml`
**Key paths:** `/articles`, `/articles/{slug}`, `/tags`

| Path | What we push | Expected dashboard |
|---|---|---|
| 🔴 Breaking | Remove `/articles` GET endpoint | `conduit-mobile-app` node red |
| 🟢 Safe | Add `/articles/{slug}/reactions` new endpoint | Node stays green |
| 🟡 Warning | Make `author` field required in response | Warning on consumer |

---

### 6. `slack-api-specs` — AsyncAPI (Slack Events API)

**Consumer:** `admin/slack-bot`
**Provider spec:** `events-api/slack_events_api_async_v1.json`
**Key field:** `channel_id` in message event payload

| Path | What we push | Expected dashboard |
|---|---|---|
| 🔴 Breaking | Remove `channel_id` from message payload | `slack-bot` node red — "Webhook bot will break" |
| 🟢 Safe | Add `thread_ts` field to message payload | Node stays green |
| 🟡 Warning | Mark `channel_id` as deprecated | Warning indicator |

---

### 7. `jaffle-shop` — SQL / dbt (Data Engineering)

**Consumers:** `admin/teradata-etl`, `admin/bi-dashboard`
**Provider spec:** `models/customers.sql`
**Key column:** `customer_lifetime_value`

| Path | What we push | Expected dashboard |
|---|---|---|
| 🔴 Breaking | Drop `customer_lifetime_value` column | Both `teradata-etl` + `bi-dashboard` red — "ETL pipeline will fail" |
| 🟢 Safe | Add `age INTEGER` column (additive) | Both nodes green |
| 🟡 Warning | Rename column (breaking alias) | Warning on both consumers |

---

## Playwright Test Architecture

```
tests/e2e/system-matrix-full.spec.ts
│
├── beforeAll (global setup — runs once)
│   ├── Create 7 repos in Forgejo via REST API (idempotent)
│   ├── Write substrate.yaml + baseline schema → git push to each repo
│   └── waitForPipelineSync() — poll graph API until ≥N edges appear
│
├── describe.serial: microservices-demo
│   ├── test: Red — push breaking proto, waitForSync, assertNodeRed(checkoutservice)
│   ├── test: Green — push safe proto, waitForSync, assertNodeGreen(all consumers)
│   └── test: Warning — push reserved field, waitForSync, assertNodeWarning
│
├── describe.serial: graphql-schema  (same pattern)
├── describe.serial: stripe-openapi  (same pattern)
├── describe.serial: openai-openapi  (same pattern)
├── describe.serial: realworld       (same pattern)
├── describe.serial: slack-api-specs (same pattern)
└── describe.serial: jaffle-shop     (same pattern)
```

**Key helpers to implement:**
```typescript
// Create repo in Forgejo if it doesn't exist
createForgejoRepo(name: string)

// Push real schema + substrate.yaml to a Forgejo repo
gitPushToForgejo(repoName: string, files: Record<string, string>)

// Wait for full pipeline: webhook → worker → API → River → graph updated
waitForPipelineSync(minNewEdges: number, timeoutMs: number)

// Playwright: open graph, search, assert node colour/badge
assertNodeState(page, searchTerm: string, state: 'red' | 'green' | 'warning' | 'any')
```

---

## What's 100% Real (Zero Mocking)

| Component | Real |
|---|---|
| Git push | ✅ `execSync('git push')` with real Forgejo creds |
| Forgejo webhook | ✅ Real HTTP POST from Forgejo to Worker |
| Schema file fetch | ✅ Worker fetches real file from Forgejo API using token |
| Go Engine diff | ✅ Real WASM engine with real schema content from actual forked repos |
| Go API sync | ✅ Real `POST /api/v1/sync` with real River queue |
| Dashboard UI | ✅ Real Playwright browser on live dashboard |

---

## Open Questions (need your answers before we build)

> [!IMPORTANT]
> Please confirm the following before we write the test code:

1. **Forgejo webhook URL:** Is it currently `http://host.containers.internal:8787` or something else? (Podman networking)
2. **`ENABLE_PUSH_CREATE`:** Should the test auto-create repos by pushing, or should it use the Forgejo REST API to create them first?
3. **SSE live updates:** Does the dashboard `/org/admin/graph` page auto-update when new sync data arrives via SSE, or does it need a page refresh?
4. **Cleanup:** After each test run, should we reset repos back to baseline, or accumulate state across runs?
5. **Warning path:** Does the Go Engine currently return a `warning` severity level, or only `breaking`/`safe`?

---

## Additional Feature Suites (WASM + MCP + Enterprise)

### Suite 8 — WASM Engine (In-Browser Diffing, P11-T08)

The Go diff engine is compiled to WASM and runs directly in the browser. The Playground page (`/playground`) and Visual Studio (`/studio`) use it for zero-latency, zero-backend schema diffing.

| Test | What it does | Expected |
|---|---|---|
| WASM loads without error | Navigate to `/playground`, wait for WASM init | No JS error, diff output renders |
| Real protobuf diff in-browser | Paste `CartItem` base + breaking head into playground | Diff result shows `BREAKING: field removed` |
| Real OpenAPI diff in-browser | Paste Stripe spec base + head (remove `/v1/charges`) | Diff result shows breaking change |
| Large payload (60k lines) | Paste full Stripe `spec3.yaml` into WASM engine | Completes in <5s, no crash/panic |
| Malformed YAML input | Paste invalid YAML | Returns safe error, no browser crash |

---

### Suite 9 — Impact API + MCP Tools (P9-T15)

`GET /api/v1/impact/{org}/{repo}` is the public API enabling CI/CD pipelines and AI agents (Cursor/Claude via MCP) to block merges based on blast radius.

| Test | What it does | Expected |
|---|---|---|
| Impact API — no consumers | `GET /impact/admin/orphan-repo` | Returns `{"consumers": [], "risk_score": 0}` |
| Impact API — `microservices-demo` | After seeding, `GET /impact/admin/microservices-demo` | Returns 4 consumers with risk score > 0 |
| Impact API — breaking change | After pushing breaking proto, re-query impact | Response shows `broken_consumers > 0`, `is_safe: false` |
| Impact API — safe change | After pushing safe proto, re-query impact | `is_safe: true`, no broken consumers |
| Can-Deploy gate | `GET /registry/can-deploy?org=admin&repo=microservices-demo` | Returns `false` while breaking change is live |
| Can-Rollback gate | `GET /registry/can-rollback?org=admin&repo=microservices-demo` | Returns `true` for prior safe commit |

---

### Suite 10 — AI Autofix (P4 Intelligence Layer)

`POST /api/v1/ai/analyze` and `POST /api/v1/ai/autofix` power the AI-driven fix suggestions in PR comments.

| Test | What it does | Expected |
|---|---|---|
| AI Analyze — breaking change | POST breaking protobuf schemas to `/ai/analyze` | Returns explanation of what broke and why |
| AI Autofix — protobuf | POST breaking + base proto to `/ai/autofix` | Returns a `safe_patch` that re-adds the field |
| AI Autofix — OpenAPI | POST breaking Stripe spec to `/ai/autofix` | Returns a patch restoring the removed path |
| Autofix patch is valid | Apply the returned `safe_patch` to engine | Engine returns `breaking_count: 0` |

---

### Suite 11 — Enterprise Dashboard Pages (All Routes)

Every dashboard page is a product feature. Each needs a real E2E assertion.

| Page | URL | Test |
|---|---|---|
| Home / Org overview | `/org/admin` | Shows repo list with real synced repos |
| Dependency Graph | `/org/admin/graph` | Renders Svelte Flow canvas with all 7 provider nodes (via search) |
| Impact Matrix | `/org/admin/matrix` | Shows provider × consumer matrix with risk scores |
| Repo Catalog | `/org/admin/catalog` | Lists all registered repos with schema types |
| Repo Detail | `/org/admin/catalog/microservices-demo` | Shows schema history, diff viewer, consumer list |
| Diff Viewer | `/diff/{id}` | After a real sync, opens a diff and shows side-by-side comparison |
| Visual API Studio | `/studio` | Opens with blank canvas, WASM loads, drag-drop a schema |
| Playground | `/playground` | Paste schemas, WASM diffs them in-browser, result displays |
| Settings / Webhooks | `/org/admin/settings` | Shows registered egress webhooks, add/remove works |
| Time-Travel | Graph page with scrubber | Scrubber shows historical snapshots, clicking past date re-renders graph |
| Command Palette | `Cmd+K` on graph page | Opens, types repo name, navigates to that node |

---

### Suite 12 — SSE Real-Time Updates (P11-T09)

| Test | What it does | Expected |
|---|---|---|
| SSE connection established | Navigate to `/org/admin/graph`, check EventSource | Connection opens to `/api/v1/events` |
| Graph auto-updates without refresh | Push schema change while graph page is open | Node state changes within 30s without page refresh |
| Multiple concurrent SSE clients | Open 3 browser tabs on graph page, push change | All 3 update simultaneously |
| SSE reconnects after drop | Kill and restart the API server, re-open dashboard | SSE reconnects, graph still works |

---

### Suite 13 — Cross-Repo Blast Radius (Full Pipeline)

This is the flagship demo — one provider change rippling to many consumers.

| Test | What it does | Expected |
|---|---|---|
| Blast radius depth 1 | `microservices-demo` breaks → 4 direct consumers red | 4 red nodes in graph |
| Blast radius depth 2 | If a consumer of `microservices-demo` is also a provider to another service | Nth-degree consumers also shown as impacted |
| Blast radius cleared | Restore `microservices-demo` to safe | All consumers back to green |
| Multi-repo simultaneous break | Break `stripe-openapi` AND `openai-openapi` at same time | Both consumer graphs simultaneously red |
| Override / Yellow path | Push breaking change + override YAML | Node shows yellow "Acknowledged" badge, not red |

---

### Suite 14 — Telemetry & ROI (P12-T07)

| Test | What it does | Expected |
|---|---|---|
| ROI endpoint | `GET /api/v1/telemetry/roi/admin` | Returns `{ breaking_changes_caught, estimated_hours_saved }` |
| Drift telemetry | `POST /telemetry/drift` with a drift event | 202 Accepted |
| Trace telemetry | `POST /telemetry/traces` with a trace | 202 Accepted |

---

## Complete Test File Structure

```
dashboard/tests/e2e/
├── system-matrix-full.spec.ts       ← 7 demo repos × 3 paths (git push → webhook → UI)
├── wasm-engine.spec.ts              ← Suite 8: WASM in-browser diffing
├── impact-api.spec.ts               ← Suite 9: Impact API + CI/CD gates
├── ai-autofix.spec.ts               ← Suite 10: AI analyze + autofix
├── dashboard-pages.spec.ts          ← Suite 11: All dashboard routes
├── sse-realtime.spec.ts             ← Suite 12: SSE live updates
├── blast-radius.spec.ts             ← Suite 13: Cross-repo blast radius
└── telemetry.spec.ts                ← Suite 14: Telemetry + ROI
```

## Summary — What This Covers

| Category | Feature | Mocked? |
|---|---|---|
| Core pipeline | git push → Forgejo → webhook → worker → engine → API → UI | ❌ Zero |
| Schema types | Protobuf, GraphQL, OpenAPI×3, AsyncAPI, SQL | ❌ Zero |
| Demo repos | All 7 real forked repos from GitHub | ❌ Zero |
| WASM engine | In-browser diffing with real schemas | ❌ Zero |
| Impact API | CI/CD gate queries | ❌ Zero |
| AI Autofix | Fix suggestion generation | ❌ Zero |
| Dashboard | All 10 pages | ❌ Zero |
| SSE | Real-time UI updates | ❌ Zero |
| Blast radius | Multi-depth consumer impact | ❌ Zero |
| Telemetry | ROI and trace tracking | ❌ Zero |

**Total estimated tests: ~80 across 8 spec files.**


## Backend State Sync Fix (Red Path UI Readiness)

During the E2E matrix execution, we discovered a gap in the Go API (`api/internal/services/push.go`): the Engine correctly evaluates `BREAKING` changes and queues AI Autofix PRs, but it does NOT update the `status` column in the `dependencies` PostgreSQL table. This causes the UI to render the node as `SAFE` even when broken.

**Specification to fix:**
1. **DB Layer:** Add `UpdateDependencyStatus(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, status string) error` to `api/internal/db/store.go` and `pgstore.go`.
2. **Implementation:** In `api/internal/db/queries.go`, implement the SQL: `UPDATE dependencies SET status = $3 WHERE consumer_repo_id = $1 AND provider_contract_id = $2`.
3. **Service Layer:** In `api/internal/services/push.go`, when `diffReport.Summary.BreakingCount > 0`, call `UpdateDependencyStatus(ctx, consumer.ConsumerRepoID, contract.ID, "BREAKING")`.
4. **Recovery:** Also add logic so that if `BreakingCount == 0`, we reset the status to `"SAFE"`, enabling the Green Path tests to pass when a safe schema is subsequently pushed.
