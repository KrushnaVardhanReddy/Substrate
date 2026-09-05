# Substrate — Demo Guide
> **Format:** 1-hour live demo + discussion  
> **Stack for today:** REST / OpenAPI  
> **Environment:** Local stack (Postgres + API + Engine + Dashboard running)

---

## 0 · Quick-start Checklist (Run Before the Call)

```bash
# Terminal 1 — database
make postgres

# Terminal 2 — API (stateful registry)
make api

# Terminal 3 — diff engine (stateless)
make engine

# Terminal 4 — dashboard
make dashboard

# Seed realistic demo data
bash scripts/e2e/seed_via_api.sh

# Verify everything is healthy
curl -s http://localhost:8090/health   # → {"status":"ok"}
curl -s http://localhost:8080/health   # → {"status":"ok"}
# Open browser → http://localhost:5173
```

---

## 1 · What is Substrate?

**Substrate is an API Contract Governance Platform for microservice organizations.**

In plain language: it is the automated system that prevents one team from silently breaking another team's service — across every language, every schema format, and every CI/CD pipeline — without slowing developers down.

### The one-sentence pitch
> *"Substrate is the circuit breaker between services. When any team's change would break a downstream contract, Substrate catches it before it merges — and shows exactly who gets hurt and how badly."*

### What it is NOT
- ❌ Not just a linter (Spectral, Stoplight) — those only check one file in isolation.
- ❌ Not a service mesh (Istio, Linkerd) — those handle runtime traffic, not schema contracts.
- ❌ Not a documentation tool (Swagger UI, Redoc) — those are read-only viewers.
- ✅ Substrate is the **governance layer that sits between your CI/CD and your dependency graph**.

---

## 2 · The Problem We're Solving

### The silent break epidemic

In any organization running more than 3 microservices, this happens regularly:

```
Monday 9am:  Backend team merges a PR that renames
             `user_id` → `userId` in the Payments API.

Monday 2pm:  Mobile App crashes in production.

Tuesday:     4 engineers spend the day debugging.
             Root cause: a JSON field rename they
             didn't know 3 other services depended on.
```

**The real cost:**
- The backend developer had no idea anyone depended on that field name.
- The mobile developer had no way to know the contract changed.
- There was no CI gate. No alert. No governance.

### Why this is getting worse, not better

| Factor | Why it amplifies the problem |
|--------|------------------------------|
| More microservices | More cross-team dependencies, more blast radius |
| Faster release cycles | Less time for manual coordination |
| Remote / distributed teams | No hallway conversations |
| AI-generated code | Developers move faster, miss downstream impact |
| Polyglot architectures | REST + GraphQL + gRPC + SQL in one org — no unified view |

### The organizational gap nobody talks about

There is a **known-unknown gap**: the backend developer *knows* they changed the field. The mobile developer *doesn't know* anyone changed it. There is no system in between that connects these two facts — until Substrate.

---

## 3 · Architecture — Why Three Go Binaries?

Substrate is deliberately split into three separable binaries because different deployment contexts have different requirements:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           SUBSTRATE ARCHITECTURE                         │
│                                                                          │
│  ┌──────────────────┐     ┌──────────────────┐    ┌──────────────────┐  │
│  │  substrate CLI   │     │   substrate API   │    │  substrate-mcp   │  │
│  │  (engine binary) │     │   (api binary)    │    │  (mcp binary)    │  │
│  │                  │     │                   │    │                  │  │
│  │  Stateless.      │     │  Stateful.        │    │  Bridges the     │  │
│  │  No database.    │     │  Owns Postgres.   │    │  engine to AI    │  │
│  │  Runs in GitHub  │     │  Stores the live  │    │  IDEs (Cursor,   │  │
│  │  Actions, local  │     │  dependency graph,│    │  Claude, VS Code)│  │
│  │  terminal, or as │     │  org memberships, │    │  over stdio via  │  │
│  │  an HTTP service │     │  audit history,   │    │  the MCP JSON-   │  │
│  │  on port 8080.   │     │  and all events.  │    │  RPC protocol.   │  │
│  │                  │     │  Port 8090.       │    │                  │  │
│  └──────────────────┘     └──────────────────┘    └──────────────────┘  │
│          │                         │                       │             │
│          └─────────────────────────┴───────────────────────┘             │
│                                    │                                     │
│                         ┌──────────────────┐                            │
│                         │  SvelteKit UI    │                            │
│                         │  Port 5173       │                            │
│                         │  Live dependency │                            │
│                         │  graph, timeline,│                            │
│                         │  blast radius,   │                            │
│                         │  override mgmt.  │                            │
│                         └──────────────────┘                            │
└─────────────────────────────────────────────────────────────────────────┘
```

### How the stateless engine actually communicates

This is the key architectural question for any technical audience: **if the engine is stateless, how does it connect to the UI or the API?**

The answer: **it doesn't — by design.** The engine only speaks when spoken to, and through two strictly defined channels:

```
┌─────────────────────────────────────────────────────────────────┐
│                      ENGINE COMMUNICATION MODEL                   │
│                                                                   │
│  MODE 1 — CLI (exit code contract)                               │
│  ─────────────────────────────────                               │
│  GitHub Actions / Dev Terminal                                    │
│         │                                                         │
│         │  $ substrate diff base.yaml head.yaml                  │
│         ▼                                                         │
│  ┌─────────────┐   stdin/stdout + exit code only                 │
│  │  engine CLI  │   0 = SAFE  1 = WARNING  2 = BREAKING          │
│  └─────────────┘   ← never opens a socket, never calls the API  │
│                                                                   │
│  MODE 2 — HTTP serve (compute offload for the API)              │
│  ───────────────────────────────────────────────────             │
│  GitHub App Webhook                                               │
│         │                                                         │
│         │  POST /api/v1/sync  (schema payload)                   │
│         ▼                                                         │
│  ┌──────────────────┐   POST http://localhost:8080/diff          │
│  │  API  (:8090)    │ ──────────────────────────────►  engine    │
│  │  (stateful brain)│ ◄──────────────────────────────  (:8080)  │
│  │                  │   { breaking_changes: [...] }              │
│  │  Stores result   │                                            │
│  │  in Postgres     │                                            │
│  └──────────────────┘                                            │
│         │                                                         │
│         │  SSE / REST                                             │
│         ▼                                                         │
│  ┌──────────────────┐                                            │
│  │  Dashboard UI    │  ← ONLY talks to the API, never engine    │
│  │  (:5173)         │                                            │
│  └──────────────────┘                                            │
└─────────────────────────────────────────────────────────────────┘
```

**The key points:**

1. **The UI never touches the engine.** It only calls the API over REST and SSE. The engine is invisible to the frontend.

2. **The API uses the engine as a compute worker.** When a sync webhook arrives, the API calls `POST http://localhost:8080/diff` on the engine's HTTP serve mode, gets the diff result back as JSON, then stores it in Postgres. The engine is a pure function: input schemas in, breaking changes out.

3. **The CLI mode has zero runtime dependencies.** `substrate diff` reads two files, writes to stdout, exits. No sockets. No database. No network. This is what runs in GitHub Actions — a single static binary dropped into the CI environment.

4. **The MCP binary is a stdio bridge.** It wraps the engine's diff logic and exposes it over stdin/stdout using JSON-RPC (the Model Context Protocol). AI IDEs like Cursor and Claude Desktop spawn it as a subprocess — again, no ports, no HTTP server needed.

### The engine has two modes — same binary, different interfaces

```bash
# Mode 1: CLI — runs once, exits
substrate diff base.yaml head.yaml --format=json
# → JSON on stdout, exit 0/1/2

# Mode 2: HTTP serve — long-running worker, accepts POST /diff
substrate serve
# → Listening on :8080
# → POST /diff { "base": "...", "head": "..." }
# → returns { "breaking_changes": [...] }
```

Same compiled binary. The `serve` subcommand switches it from a one-shot CLI tool into a persistent HTTP microservice. The API starts it in serve mode and calls it over localhost.

### Why not one binary?

| Deployment scenario | What you need |
|--------------------|---------------|
| GitHub Actions CI gate | Only the CLI — stateless, zero runtime deps |
| Local developer ("does this break anything?") | Only the CLI — offline, instant |
| Organization-wide graph + dashboard | API + database |
| AI IDE assistant (Cursor/Claude) | MCP binary over stdio — no ports, no HTTP |
| Air-gapped enterprise | All three self-hosted; no external calls |

**The engine is independently valuable.** A team can use `substrate diff` in CI *today* — no database, no account, no signup — and get immediate value. The API and dashboard are the upgrade path.

### Why Go?

- **Single static binary** — drop it anywhere. No JVM, no Node, no Python runtime.
- **Native WASM compilation** — the entire diff engine compiles to WebAssembly and runs in the browser. The diff happens client-side with zero server round-trip.
- **Race-condition safe** — the concurrent SSE broadcaster and the dependency graph writer use proper sync.RWMutex — no goroutine leaks.
- **Sub-millisecond diffs** — a 2,000-field OpenAPI spec diffs in under 1ms. No async queues needed for the hot path.

---


---



## 4 · What Substrate Supports

| Schema Format | Examples | Status |
|--------------|---------|--------|
| OpenAPI 3.x / Swagger 2.x | REST APIs | Full |
| Protocol Buffers | gRPC services | Full |
| GraphQL SDL | Apollo, Hasura | Full |
| AsyncAPI | Kafka, webhooks, event streams | Full |
| SQL DDL | Postgres, MySQL migrations | Full |
| AI/ML Model Cards | OpenAI spec format | Full |
| Terraform HCL | Infrastructure schemas | Full |

One governance platform. One dependency graph. Seven schema languages.

---

## 5 · The Demo — A Real 5-Service E-Commerce Scenario

We are going to walk through a **realistic, production-like scenario** using the live running stack. No toy examples. No `FooBar` APIs.

### The system under demo

```
mcp-org/frontend  ────────────────────────────────────┐
      │                                                │
      │ depends on                                     │
      ▼                                                │
mcp-org/backend  ◄──── mcp-org/discovery-test         │
      │                                                │
      │ depends on                            depends on
      ▼                                                │
mcp-org/postgres-db          mcp-org/core-repo ◄───────┘

testorg/api-gateway ──► testorg/users-api
testorg/payments-api ──► testorg/users-api

(All seeded and live in your Postgres right now)
```

This is your **real dependency graph** — not a mock. The data is already seeded. Open the dashboard now.

---

### Step 1 — Show the dependency graph

**Navigate to:** http://localhost:5173 → log in → select mcp-org

**What to point out:**
- Every node is a real repository registered in the contract registry.
- Edges represent live dependency contracts — not just code imports, but *schema contracts*.
- Click any node → the right panel shows the full blast radius: who depends on it, how many consumers, and what schema they agreed on.
- The graph is **live** — it updates in real-time via SSE (Server-Sent Events) as new pushes register.

```bash
# Show the raw data behind the graph
curl -s http://localhost:8090/api/v1/graph/mcp-org \
  -H "Authorization: Bearer local-dev-token" | jq '.'
```

---

### Step 2 — Initialize Substrate in a new service

This is how a developer onboards. Run it live in terminal:

```bash
# Build the CLI first
cd engine && go build -o substrate ./cmd/substrate/ && cd ..

# Scaffold a new API design with the AI assistant
mkdir /tmp/payments-api-demo && cd /tmp/payments-api-demo
./engine/substrate init --design
# Substrate asks: "What kind of API are you building?"
# Type: "A payments API with endpoints to create a charge,
#        list charges, and refund a charge. All require auth."
# Substrate generates openapi.yaml instantly.
```

**Show the generated `substrate.yaml`:**

```yaml
# substrate.yaml — auto-generated by `substrate init`
service: payments-api
schema_type: openapi
spec_path: openapi.yaml

owners:
  - team: payments-team
    contact: payments@company.com

overrides: []
```

---

### Step 3 — The stateless CLI in a CI/CD pipeline

This is the **zero-dependency gate** — no database, no account, just the binary.

```bash
# A SAFE change — exit code 0, pipeline passes
substrate diff base.yaml head.yaml --format=text

# A BREAKING change — exit code 2, pipeline BLOCKED
substrate diff base.yaml head_breaking.yaml --format=json
```

**Breaking change output:**

```json
{
  "substrate_version": "0.1.0",
  "schema_type": "openapi",
  "summary": {
    "total_changes": 1,
    "breaking_count": 1,
    "overall_severity": "BREAKING"
  },
  "breaking_changes": [
    {
      "rule_id": "FIELD_REMOVED",
      "severity": "BREAKING",
      "path": "GET /api/articles → response.body.articles[].author",
      "description": "Required response field removed",
      "recommendation": "Add a deprecation period or version the endpoint"
    }
  ]
}
```

**Exit code 2 = pipeline BLOCKED.**

### The GitHub Actions integration

```yaml
# .github/workflows/substrate.yml
name: API Contract Gate
on: [pull_request]

jobs:
  contract-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Install Substrate CLI
        run: |
          curl -L https://github.com/KrushnaVardhanReddy/substrate/releases/latest/download/substrate-linux-amd64 \
            -o /usr/local/bin/substrate && chmod +x /usr/local/bin/substrate

      - name: Run API Contract Check
        run: |
          substrate diff \
            origin/main:openapi.yaml \
            HEAD:openapi.yaml \
            --config=substrate.yaml \
            --format=json
```

---

### Step 4 — The blast radius

Now we escalate from "one API changed" to "how many teams does this break?"

```bash
# Register the breaking change via the live API
# (simulating the GitHub App webhook firing)
curl -s -X POST http://localhost:8090/api/v1/sync \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer local-dev-token" \
  -d '{
    "installation_id": 123456,
    "org": "mcp-org",
    "consumer_repo": "mcp-org/frontend",
    "consumer_github_repo_id": 10102,
    "commit_sha": "sha-breaking-demo",
    "dependencies": [{
      "provider_repo": "mcp-org/backend",
      "provider_github_repo_id": 10101,
      "schema_type": "openapi",
      "spec_path": "openapi.yaml",
      "branch": "main",
      "raw_content": "openapi: 3.0.0\ninfo:\n  title: Backend API\n  version: 2.0.0\npaths:\n  /users:\n    get:\n      responses:\n        \"200\":\n          description: OK"
    }]
  }'

# Query the blast radius
curl -s "http://localhost:8090/api/v1/impact/mcp-org/backend" \
  -H "Authorization: Bearer local-dev-token" | jq '.'
```

**In the dashboard → click the `backend` node:**
- Sidebar shows: **3 consumers will break**
- `frontend`, `discovery-test-repo`, and `data-warehouse` highlighted in blast radius
- Timeline panel shows the exact commit, timestamp, and detected change

This is the **"Aha" moment**: the backend developer can see *before merging* that their change cascades 3 levels deep.

---

### Step 5 — Acknowledge a breaking change (the governance loop)

Sometimes breaking changes are intentional. Substrate has a formal approval workflow.

**Edit `substrate.yaml` to add an override:**

```yaml
service: backend-api
schema_type: openapi
spec_path: openapi.yaml

overrides:
  - rule_id: ENDPOINT_REMOVED
    path: GET /users
    reason: "Deprecated in favour of GET /v2/users. Migration guide: docs/migration-v2.md"
    approved_by: jane.smith@company.com
    expires: 2027-01-01
```

```bash
# Run again with the approved override
substrate diff base.yaml head.yaml --config=substrate.yaml --format=json
# Output: overall_severity: "SAFE" — exit code 0
```

**What this gives you organizationally:**
- Paper trail — who approved, when, and why. Auditable.
- Expiry date — the waiver auto-expires. No silent permanent technical debt.
- Git history — the approval is committed to the repo. Rollback = revert the commit.
- Dashboard view — overrides appear in the UI with expiry dates highlighted.

---

### Step 6 — The MCP integration (AI IDE awareness)

```bash
# Build the MCP binary
cd engine && go build -o substrate-mcp ./cmd/substrate-mcp/ && cd ..
```

Add to `claude_desktop_config.json` or Cursor MCP settings:

```json
{
  "mcpServers": {
    "substrate": {
      "command": "/path/to/engine/substrate-mcp",
      "args": []
    }
  }
}
```

**In Claude or Cursor, ask:**
> "What happens to our system if I delete the email field from the Users API?"

Claude calls `substrate.get_blast_radius` → gets the live dependency graph → answers:

> "Deleting `email` from `users-api` will break 3 downstream services: `payments-api` (uses it for receipts), `notifications-service` (uses it for alerts), and `mobile-app` (displays it in the profile screen). I recommend adding a 60-day deprecation notice first."

The developer **never left their IDE**. The break was caught before a single line of code changed.

---

## 6 · Execution Modes — Gradual Enterprise Adoption

| Mode | Exit code behaviour | Use case |
|------|---------------------|----------|
| `--mode=audit` | Always exits 0. Logs everything. | 90-day proof of value trial — silent monitoring, ROI dashboard fills up |
| `--mode=default` | Blocks on BREAKING. Warns on WARNING. | Standard enforcement |
| `--mode=strict` | Blocks on BREAKING and WARNING. | New microservices, Spec-First greenfield |
| `--mode=legacy` | Only blocks catastrophic changes. | Brownfield legacy monolith migration |

---

## 7 · Competitive Positioning

| Tool | What it does | Why Substrate wins |
|------|-------------|-------------------|
| Spectral | Lints a single OpenAPI file | No cross-repo graph. No blast radius. |
| Optic | API change detection for one repo | No dependency graph. No organization-wide view. |
| Backstage | Service catalog | Catalog is read-only. Backstage describes; Substrate enforces. |
| Kong / Apigee | API gateway (runtime) | Runtime-only. Cannot catch a schema change before it deploys. |
| SonarQube | Code quality | Does not understand API contracts or cross-service dependencies. |

**The unique position:** Substrate is the only tool that combines a multi-schema diff engine, a live cross-repo dependency graph, a CI enforcement gate, and an AI IDE integration in a single deployable system.

---

## 8 · Deployment Options

### SaaS (instant, zero-ops)
```
Install GitHub App → Substrate hosted → No infra required
```

### Self-hosted (enterprise, air-gapped)
```bash
docker-compose up

# Or Kubernetes via Helm
helm install substrate ./helm-chart \
  --set postgres.url="postgresql://..." \
  --set ai.provider="azure-openai" \
  --set ai.endpoint="https://your-azure-openai.openai.azure.com/"
```

**Enterprise data guarantee:** When self-hosted, schemas never leave your network. The diff engine runs locally. AI features point at your own Azure OpenAI or Ollama endpoint.

---

## 9 · Common Architect Questions

**Q: How does this handle monorepos?**
A: The `substrate.yaml` supports multiple `consumers` blocks. Each microservice is declared as a separate consumer. The sync endpoint maps them individually into the graph.

**Q: What if teams don't maintain their `openapi.yaml`?**
A: `substrate init --ai` scans existing route handlers (Gin, Express, FastAPI, Spring) and auto-generates the initial spec.

**Q: How does override approval work in practice?**
A: The `substrate.yaml` override block is committed to Git. A CODEOWNERS rule on `substrate.yaml` requires architecture team approval. The expiry date is enforced at diff time — expired overrides do not suppress failures.

**Q: What is the performance at scale?**
A: The diff engine processes a 2,000-field OpenAPI spec in under 1ms. The Cytoscape graph renders 1,000 nodes and 3,000 edges without degradation (tested in the stress-test suite). Postgres query times are sub-5ms at 10k repos.

**Q: GDPR / schema data sovereignty?**
A: Self-hosted deployment means zero data leaves your VPC. The GitHub App is OAuth-scoped to read only the schema files it needs. It never reads application code or environment variables.

---

## 10 · The 60-Minute Demo Script

| Time | Section | What to show |
|------|---------|--------------|
| 0:00–0:05 | The problem | Tell the story — the silent break, the 4-engineer day |
| 0:05–0:15 | Dashboard | Live dependency graph, blast radius on node click |
| 0:15–0:25 | `substrate init` | AI scaffold generates openapi.yaml + substrate.yaml live |
| 0:25–0:35 | CLI + CI gate | Safe diff (exit 0) then breaking diff (exit 2) then GitHub Actions config |
| 0:35–0:45 | Blast radius | Sync a breaking change, watch graph update, show impacted nodes |
| 0:45–0:50 | Override flow | Add override with expiry, diff passes, audit trail in dashboard |
| 0:50–0:55 | MCP in IDE | Ask Claude about blast radius, get real answer from live graph |
| 0:55–1:00 | Q&A | Have competitive positioning table ready |

---

## 11 · Live Commands Reference Card

```bash
# Engine (stateless)
substrate diff base.yaml head.yaml
substrate diff base.yaml head.yaml --config=substrate.yaml
substrate diff base.yaml head.yaml --format=json
substrate diff base.yaml head.yaml --format=changelog
substrate init
substrate init --design

# API (stateful) — show graph state
curl http://localhost:8090/health
curl http://localhost:8090/api/v1/graph/mcp-org -H "Authorization: Bearer local-dev-token"
curl http://localhost:8090/api/v1/impact/mcp-org/backend -H "Authorization: Bearer local-dev-token"
curl http://localhost:8090/api/v1/repos/mcp-org -H "Authorization: Bearer local-dev-token"
curl http://localhost:8090/api/v1/events/mcp-org -H "Authorization: Bearer local-dev-token"

# Dashboard
open http://localhost:5173
# mcp-org: frontend → backend → postgres-db
# testorg: payments-api → users-api, api-gateway → users-api

# Reset demo to clean state
make reset-demo
bash scripts/e2e/seed_via_api.sh
```

---

*Generated from the live Substrate codebase. All commands tested against the running local stack.*
