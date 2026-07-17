# Substrate

## The Living Map of Your Engineering Ecosystem

> **Status:** Phase 12 🧪 — **V2.0 Quality Assurance.** Phase 11 (Advanced Graph Visualization) has been completely merged, including the core UI/UX overhaul, Cascading Blast Radius, Team Neighborhoods, and Time-Travel Scrubber. We are currently executing the Phase 12 Playwright and Go E2E tests, with Wave 1 successfully merged. Wave 2 is currently executing (Live sessions: Stitch [`3443016659672603780`, `14129042895366541864`], Jules [`9405238502927674003`, `13216611699123662740`, `8622645150866572603`]).

Substrate is a CI/CD-integrated data contract and dependency intelligence platform that prevents downstream data failures before they reach production.

It acts as a proactive firewall for engineering teams by analyzing schema changes, understanding repository dependencies, identifying impacted systems, and blocking unsafe changes before they break applications, analytics, dashboards, and data pipelines.

---

# Quick Start — 3 Minutes to Protect Your API

## Step 1 — Add the GitHub Action

Create `.github/workflows/substrate.yml` in your repository:

```yaml
name: Substrate API Contract Guard

on:
  pull_request:
    branches: [main]

jobs:
  check-api-contracts:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # needed to access base branch files

      - name: Check API Breaking Changes
        uses: KrushnaVardhanReddy/Substrate@v0.1.0
        with:
          base_schema: openapi.yaml        # path to your OpenAPI spec on main
          head_schema: openapi.yaml        # path to your OpenAPI spec in the PR
          config: substrate.yaml          # optional — see Step 2
```

> **How it works:** When a PR is opened, Substrate fetches the spec from `main` (base) and compares it to the spec in the PR branch (head). If a breaking change is detected, the CI check fails and the merge is blocked.

---

## Step 2 — Add `substrate.yaml` (optional but recommended)

Place `substrate.yaml` in the **root of your repository**:

```yaml
# substrate.yaml — Substrate configuration for this repository

service: your-service-name          # human-readable name shown in CI output
schema_type: openapi                # openapi | sql | graphql | protobuf
spec_path: openapi.yaml             # path to your spec file from repo root

owners:
  - team: platform-team
    contact: platform@yourcompany.com
```

If you skip this file, Substrate runs with defaults — no config required for basic usage.

---

## Step 3 — Open a PR with a Breaking Change

Remove an endpoint or a required field from your OpenAPI spec and open a PR. You should see:

```
❌ BREAKING CHANGES (1)
  ENDPOINT_REMOVED
  Path: paths./customers/{id}
  Description: Endpoint /customers/{id} was removed.
  Recommendation: Add 'deprecated: true' before removing endpoints.

Exit code: 2  ← CI fails, merge is blocked
```

---

## Acknowledging a Breaking Change (Override)

If a breaking change is intentional and all consumers have been migrated, you can acknowledge it in `substrate.yaml` to unblock the merge:

```yaml
overrides:
  - rule_id: ENDPOINT_REMOVED
    path: paths./customers/{id}
    reason: "Legacy endpoint removed after all 3 consumers migrated to v2. See RFC-1042."
    approved_by: you@yourcompany.com
    expires: 2026-12-31
```

The CI check will pass and show `✅ Acknowledged (override active until 2026-12-31)` instead of failing.

---

## Exit Codes

| Code | Meaning | CI Result |
|---|---|---|
| `0` | No changes, or all changes are SAFE | ✅ Pass |
| `1` | WARNING-level changes detected | ✅ Pass (with warning) |
| `2` | BREAKING changes detected | ❌ Fail — blocks merge |
| `3` | Spec is invalid (parse error) | ❌ Fail |

---

## 📚 Comprehensive Documentation

For a deep dive into configuration options (like integrating Prometheus for zero-traffic breaking change downgrades), setting up the AI Autofix features, and integrating the VSCode extension, please refer to the official [Substrate User Guide (docs/USER_GUIDE.md)](docs/USER_GUIDE.md).

---

# The Problem

Modern engineering organizations have hundreds of interconnected systems:

- Microservices
- APIs
- Databases
- Event streams
- Data warehouses
- BI dashboards
- ML pipelines

A small schema change in one repository can silently break multiple downstream systems.

Example:

A developer changes:

```sql
ALTER TABLE customers
DROP COLUMN email;
````

The developer sees a successful deployment.

But downstream systems break:

```
Customer Service
        |
        |
 -------------------------
 |           |            |
Billing   Marketing   Analytics
                     |
                  Tableau
```

Problems:

* Dashboards show incorrect data
* ETL pipelines fail
* APIs break
* Teams discover failures after production deployment
* Engineers waste hours finding impacted systems

Current solutions mostly detect problems after they happen.

Substrate prevents them before they happen.

---

# The Vision

## From reactive debugging to proactive prevention

Today:

```
Developer makes change

        ↓

Production failure

        ↓

Teams investigate

        ↓

Emergency fixes
```

With Substrate:

```
Developer creates PR

        ↓

Substrate analyzes change

        ↓

Dependency graph identifies impact

        ↓

Unsafe changes are blocked

        ↓

Teams negotiate before merge
```

---

# Core Philosophy: 100% Spec-First Approach

Substrate enforces a **strict spec-first development model**. 

* **The Rule:** Any structural change to the system (API modifications, database schema updates, event payload changes) **must first be defined and validated in the specification (OpenAPI, GraphQL, Protobuf, SQL migrations, etc.)** before the underlying code is allowed to change.
* **The Gate:** Substrate acts as the CI/CD gatekeeper. If a PR contains code changes that alter an implicit schema without a corresponding, validated update to the explicit specification, **the build fails**.
* **The Result:** The specification is never out of date. It remains the absolute, single source of truth for the entire engineering organization, and downstream consumers can rely on it with 100% confidence.

---

# Core Product

## 1. Data Contract Protection

Substrate integrates directly into CI/CD pipelines.

When a developer creates a pull request:

```
GitHub PR

    ↓

Substrate GitHub App

    ↓

Schema Analysis

    ↓

Contract Validation

    ↓

Approve or Block
```

Example:

Before:

```json
{
  "customer_id": "string",
  "email": "string"
}
```

Developer changes to:

```json
{
  "customer_id": "string"
}
```

Substrate detects:

```
❌ Breaking Change Detected

Removed field:
customer.email

Affected consumers:

🔴 billing-service
🔴 marketing-service
🟡 analytics-pipeline

Recommendation:
Create migration before removing field
```

## 2. Handling ORMs and Live Databases (Spec-First)

Substrate requires a text-based schema definition (e.g., `schema.sql`) to analyze databases. If your team uses an ORM (like Prisma or Hibernate) and does not maintain a static `.sql` file in the repository, you can generate it on-the-fly in your CI/CD pipeline before running Substrate:

```yaml
- name: Spin up test database & apply ORM migrations
  run: |
    docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=pass postgres:15
    npm run prisma:migrate
    pg_dump --schema-only postgres://postgres:pass@localhost:5432/db > schema.sql

- name: Check SQL Breaking Changes
  uses: KrushnaVardhanReddy/Substrate@v0.1.0
  with:
    base_schema: schema.sql # Dump from main branch
    head_schema: schema.sql # Dump from PR branch
    schema_type: sql
```

This enforces our **Spec-First** philosophy without requiring direct connections to your staging or production databases.

---

# Architecture

```
                         Organization

        Repo A              Repo B              Repo C
   (Customer API)      (Billing Service)   (Analytics)

          |                  |                   |
          |                  |                   |
          ----------------------------------------
                          |
                  Substrate Platform
                          |
        ---------------------------------
        |               |               |
  Schema Engine    Rules Engine    Dependency Engine
        |               |               |
        ---------------------------------
                          |
                 Dependency Graph Database
                          |
                 SvelteKit Dashboard
```

---

# Technology Stack

## Frontend

```
SvelteKit
TypeScript
Tailwind CSS
```

Responsibilities:

* Dashboard
* Dependency visualization
* Schema diff viewer
* Impact analysis
* Documentation explorer

---

## Backend

```
Go
PostgreSQL
Redis (future)
```

Responsibilities:

* API server
* Schema analysis
* Contract engine
* Dependency processing
* Organization management

---

## Developer Integrations

```
TypeScript GitHub App
Go CLI
```

Supported future integrations:

* GitHub
* GitLab
* Bitbucket
* Jenkins
* CircleCI

---

# Multi Repository Intelligence

Substrate connects multiple repositories inside an organization.

Example:

```
Company

├── customer-service
├── billing-api
├── marketing-service
├── analytics-pipeline
├── frontend-app
└── data-platform
```

Substrate understands relationships between them.

---

# Dependency Discovery

Substrate builds a living dependency graph using:

## 1. Code Analysis

Detect:

* API calls
* Database connections
* Kafka topics
* Event consumers
* GraphQL queries

Example:

```
billing-api

calls:

customer-service
```

Relationship:

```
billing-api → customer-service
```

---

## 2. Schema Detection

Analyze:

* OpenAPI specifications
* GraphQL schemas
* Protobuf files
* Avro schemas
* Database migrations
* dbt models

---

## 3. CI/CD Metadata

Understand:

* What changed
* Which repository changed
* Who owns the service
* What systems consume it

---

## 4. Developer Configuration

Repositories can define:

```
substrate.yaml
```

Example:

```yaml
service: billing-api

depends_on:
  - customer-service.customer
  - kafka.customer-events
```

---

# Dependency Graph

Substrate creates a complete map:

```
             customer-service

                    |
        -------------------------
        |           |           |
    billing     analytics    marketing

                    |
                 Snowflake

                    |
                 Tableau
```

The graph answers:

* Who depends on this service?
* What breaks if this changes?
* Who owns the impacted system?
* What migration is required?

---

# Auto Generated Documentation

The dependency graph becomes a living documentation system.

Instead of manually maintaining:

* Architecture diagrams
* Service catalogs
* API documentation
* Dependency maps

Substrate generates them automatically.

---

# Service Catalog

Example:

```
Customer Service

Repository:
github.com/company/customer-service

Owner:
Customer Platform Team

Language:
Go

Database:
PostgreSQL


Exposes:

GET /customers/{id}


Consumed By:

→ Billing Service
→ Marketing Service
→ Analytics Pipeline
```

---

# Architecture Documentation

Automatically generated:

```
             Customer API

                  |
        ---------------------
        |                   |
   Billing API        Marketing API
        |
     PostgreSQL
        |
     Snowflake
        |
    Tableau Reports
```

*Substrate allows instant export of these architectural visualizations to high-resolution PNGs for use in RFCs, Confluence, and SOC2 Audits.*

---

# API Documentation

Generated automatically:

```
Customer API

GET /customers/{id}

Used by:

- billing-service
- mobile-app
- reporting-service

Last Changed:
July 2026

Breaking Changes:
None
```

---

# Change History

Every schema element has history.

Example:

```
Customer.email

Created:
January 2024

Used By:
12 services

Last Modified:
March 2026

Breaking Risk:
HIGH
```

---

# AI Engineering Assistant

Future AI capabilities:

Developers can ask:

"What happens if I change customer data?"

Substrate answers:

```
Impact Analysis:

customer.email is used by:

1. Billing Service
2. Marketing Service
3. Analytics Pipeline
4. Tableau Dashboard


Recommended Action:

Create migration first.

Estimated affected teams:
3
```

---

Developer question:

"How does payment processing work?"

Substrate explains:

```
Payment Flow:

1. Checkout Service receives payment request

2. Creates PaymentCreated event

3. Payment Worker processes event

4. Transaction stored in PostgreSQL

5. Analytics pipeline sends data to Snowflake


Repositories involved:

- checkout-service
- payment-worker
- analytics-service
```

---

# Architecture & Go-to-Market Strategy

## Deploying the GitHub App yourself

The Substrate architecture consists of two deployed components:
1. **Container Service** (Go Diff Engine)
2. **Cloudflare Worker** (Webhook Receiver)

You can deploy the container using `fly deploy` (requires the Fly CLI and an account). This uses the `fly.toml` and `Dockerfile.serve` to spin up the engine.

The Cloudflare Worker is deployed automatically via GitHub Actions (see `.github/workflows/deploy-worker.yml`), provided you set `CF_API_TOKEN` and `CF_ACCOUNT_ID` in your GitHub repository secrets.

## Distribution Model — Private Repo + Docker Hub Public Image

Substrate keeps all source code, specs, and business logic in a **private repository**. The GitHub Action is distributed via a **public Docker Hub image** — users get a working, versioned binary with no source code exposed.

**How it works:**
- Every release tag (`v0.x.0`) triggers a CI workflow that builds the Docker image and pushes it to `kpakkiragari/substrate-engine` on Docker Hub (public).
- `action.yml` references the image directly: `docker://kpakkiragari/substrate-engine:v0.x.0`.
- Users install via `uses: KrushnaVardhanReddy/Substrate@v0.x.0` as usual. GitHub pulls the pre-built image — no source required.
- **Nothing proprietary is ever exposed.** The image is a compiled binary black box.

**Release flow:**
```
git tag v0.2.0 && git push origin v0.2.0
        ↓
GitHub Actions: build Dockerfile → push kpakkiragari/substrate-engine:v0.2.0 to Docker Hub
        ↓
auto-commit: action.yml pinned to v0.2.0
        ↓
GitHub Marketplace: users on @v0.2.0 get the update
```

## Distribution & Go-to-Market (GTM)

1. **GitHub Marketplace (The MVP Wedge) 🏆**: The immediate path to users is shipping the CLI as a GitHub Action. Developers search "API breaking changes", install the action, and get CI failures on breaking changes. This is the highest ROI acquisition channel.
2. **Developer Communities (HN, Reddit)**: Sharing the engineering journey generates organic awareness. Timed for post-Phase 2 when the full platform story is compelling.
3. **Direct Outreach**: Finding public repos with API specs and offering the tool directly to engineering managers to secure the first 10 design partners.
4. **Platform Launch**: Full public announcement after Phase 2 (GitHub App) is live — the story is stronger when users can install a GitHub App, not just a CLI.

---

# Product Evolution

## Phase 1: Breaking Change Prevention

Goal:

Stop production failures across every schema type your engineering organization uses.

Phase 1 ships incrementally as schema format support is added to the same diff engine.

**Wave 1 — Core Contract Formats (ships before Phase 2):**

| Format | Status | Library Strategy |
|---|---|---|
| **1a — OpenAPI 3.x** (REST APIs) | ✅ *v0.1.0 shipped* | `oasdiff` (Apache 2.0) — wrap, don't build |
| **1b — SQL Migrations** (PostgreSQL DDL) | 🔄 *in progress* | `pg_query_go` (MIT, Postgres's own parser) + spike `stripe/pg-schema-diff` (Apache 2.0) |

**Wave 2 — Extended Contract Formats (resumes after Phase 2):**

> These are deferred until Phase 2 (GitHub App + distribution) is live. Once Phase 2 is live, real user demand — not guesswork — will drive which format ships next. **Execution order within Wave 2 follows effort (easiest first):** 1d → 1e → 1h → 1c → 1f → 1g.

| Format | Effort | Status | Library Strategy |
|---|---|---|---|
| **1d — Protobuf & gRPC** | 🟢 Lowest | ✅ *shipped* | **`bufbuild/buf`** (Apache 2.0) — the `oasdiff` of Protobuf 🏆 same adapter pattern |
| **1e — AsyncAPI & Apache Avro** | 🟢 Lowest | ✅ *shipped* | `asyncapi/parser-go` (Apache 2.0) + Confluent Schema Registry API for Avro compat |
| **1h — Infrastructure as Code (Terraform)** | 🟢 Lowest | ✅ *shipped* | Go stdlib `encoding/json` parsing of `terraform plan` output. |
| **1c — GraphQL SDL** | 🟡 Medium | ✅ *shipped* | `vektah/gqlparser` (MIT) for parsing; custom rules (~15–20); pure Go, no subprocess |
| **1f — AI/ML Model Contracts** | 🟡 Medium | ✅ *shipped* | `yaml.v3` + `jsonschema` — both **already in `go.mod`**, no new deps |
| **1g — Enterprise Metadata** | 🔴 Highest | ✅ *shipped* | Go stdlib `encoding/xml` for both WSDL and Salesforce metadata XML snapshots — pure single-binary |

Core features in all sub-phases:

* GitHub integration
* Schema diff detection
* Contract validation
* PR comments
* Merge blocking

### Phase 1f: AI & ML Contract Protection

The AI problem is identical to the API problem, but failures are more dangerous because they are **silent**.

A broken REST API throws a `500`. A broken ML pipeline quietly trains on wrong features for weeks.

**Dataset Schema Contracts:**

```
❌ BREAKING: Column 'annual_revenue' renamed/removed

Affected consumers:
🔴 churn-model (feature: annual_revenue)
🔴 ltv-prediction-pipeline

Recommendation:
Update downstream feature code before applying migration.
```

**ML Model Input/Output Contracts** via `substrate.yaml`:

```yaml
model: churn-predictor
version: 2.1.0
inputs:
  - name: tenure_months
    type: float
  - name: monthly_charges
    type: float
outputs:
  - name: churn_probability
    type: float
    range: [0.0, 1.0]
```

If a new model version removes `tenure_months` → Substrate blocks the deployment.

**LLM Structured Output Contracts:**

Every team building on LLMs with structured JSON output (function calling, JSON mode) has schema contracts. If a prompt template change alters the expected output structure, downstream parsers silently break. Substrate version-controls these schemas and catches breaking changes in CI.

---

## Phase 2: GitHub App & Dependency Intelligence

Goal: Move from CLI-only to a fully automated CI/CD bot.

Features:
* **1-click GitHub App installation:** Automated PR comments and merge blocking.
* **Automated Dependency Discovery:** Eliminate manual `substrate.yaml` configuration by automatically building the full dependency graph from signals already present in your codebase. Multi-signal, confidence-scored. Full spec: `docs/specs/phase-5/dependency-discovery.md`.
  * **Environment Variable Scanning ⭐:** Scans `.env.example`, `docker-compose.yml`, K8s manifests, GitHub Actions `env:` blocks, `fly.toml`, and `Dockerfile` for `*_API_URL` / `*_ENDPOINT` patterns. Resolves URL values against a URL→Repo registry (auto-populated via GitHub Deployments API). **Zero infrastructure required.**
  * **Package Manifest Analysis:** Detects internal SDK imports (`@myorg/users-sdk` in `package.json`, internal modules in `go.mod`) as direct schema contract dependencies. The SDK import IS the dependency.
  * **OpenAPI Generator Config:** Scans `openapitools.json` and `.openapi-generator-config.yaml` for `inputSpec` URLs — the most deterministic signal possible (100% explicit reference to the provider's schema).
  * **Docker Compose / Kubernetes / Helm:** Extracts `depends_on` blocks and environment variable URL values; resolves Kubernetes DNS patterns (`users-service.default.svc.cluster.local`) against the service registry.
  * **Infrastructure as Code (Terraform):** Parses `terraform_remote_state` data sources and environment variable injections from resource references (e.g. AWS ECS task definitions wiring services together).
  * **Message Queue / Event-Driven:** AsyncAPI `$ref` cross-repo URLs and Kafka consumer group → topic → producer mappings for event-driven architectures.
  * **Distributed Tracing (Runtime Confirmation):** OTel, Datadog APM, New Relic — confirms statically-discovered dependencies with live traffic evidence.
  * **Network Layer (Enterprise):** eBPF and Istio/Envoy service mesh logs for kernel-level dependency mapping in Kubernetes environments.
  * **Confidence Scoring:** All signals are combined into a 0–100 confidence score per dependency edge. High (80+), Medium (50–79), Low (<50) — shown as solid/dashed/hidden edges in the dependency graph.
* **Service Ownership:** Map every discovered node to a team and an alert channel.

---


### Enabling the Merge Blocker (Branch Protection)

Substrate emits a commit status context called `substrate/breaking-changes` on every PR. For Substrate to actually block a merge, the repository administrator must configure this status context as a **Required Status Check**.

To enable this:

1. Go to your repository **Settings** > **Branches**.
2. Edit the branch protection rule for your default branch (e.g., `main`).
3. Enable **Require status checks to pass before merging** and search for `substrate/breaking-changes` to add it to the list.

Once enabled, any PR with unacknowledged breaking changes (where `on_breaking_change: block`) will have its merge button disabled by GitHub.

## Phase 3: The Intelligence Layer & Enterprise Integrations

Goal: Turn Substrate into the central registry of truth for the entire engineering ecosystem.

Features:
* **Developer Dashboard & Impact Analysis:** Cross-repo dependency graphs. "If I change this OpenAPI schema, which 3 repos will break?"
* **Audit Trails:** Centralized view of all overridden breaking changes across the company.
* **CI/CD Agnosticism:** Native integrations for Jenkins, GitLab CI, and Bitbucket (beyond GitHub) to capture the on-premise enterprise market.
* **API Gateway & Event Registry Sync:** Once a schema is marked safe and merged, Substrate acts as the single source of truth and pushes the approved schema via a generic plugin layer to:
1. **API Gateways:** Kong, AWS API Gateway, Apigee.
2. **Schema Registries:** Confluent (Kafka), Apollo Studio.
3. **Data Catalogs:** Collibra, Alation, Snowflake.
4. **Developer Portals:** Spotify Backstage.
5. **Observability (APM):** Datadog, Sentry, New Relic (Inject deployment markers when schemas change so runtime errors can be instantly correlated to API diffs).
6. **ITSM & Deployment Gates:** ServiceNow, Jira Service Management (Act as an automated deployment gate that intercepts production Change Requests. If Substrate detects an unapproved contract breakage, it automatically rejects the Change Request and halts the deployment pipeline).
7. **Communication:** Slack, MS Teams (Route breaking change override requests to `#platform-engineering` for one-click approval).
* **Automated SDK Generation:** Webhook triggers to automatically regenerate downstream TypeScript/Python SDKs and open PRs on consumer repositories when a safe backend API change is merged.
* **Automated Test Generation:** Avoid building competing testing tools; instead, generate Postman Collections and MuleSoft APIkit test suites directly from the approved schema. Substrate pushes updates via Pull Requests or Cloud APIs, ensuring testing teams are always testing the correct contract without overwriting their custom local scripts.
---

## Phase 4: AI Engineering Assistant

Goal:

Create an AI layer on top of engineering knowledge.

Features:

* **System Explanations:** "How does payment processing work across these 3 repos?"
* **Migration Recommendations:** Auto-generate migration code when a breaking change is detected.
* **Impact Predictions:** "If I change customer data, which downstream teams do I need to notify?"
* **AI State Registry & Agent Handoffs:** Serve as the permanent memory bank for multi-agent workflows (Jules, Cursor, Copilot). Agents can store their conversational context, current blockers, and task states in Substrate, allowing seamless handoffs between human developers and AI assistants without losing context.

---

# Final Vision

Substrate becomes:

"The living map of your engineering ecosystem."

A platform that understands:

* What systems exist
* How they connect
* Who owns them
* What changes affect them
* How to safely evolve them

The first wedge:

Prevent breaking schema changes in REST APIs.

The long-term platform:

Contract protection for every data boundary in engineering — APIs, databases, data pipelines, ML models, and AI systems.

---

## 🛠️ Local Development Architecture (The 5 Terminals)

Substrate is built as a highly decoupled, microservice-like architecture to allow maximum flexibility (CLI-mode vs Cloud-mode vs IDE-mode).

To run Substrate locally, you need 5 terminals running simultaneously:

1. **Database** (`make postgres`)
   - Runs a local PostgreSQL 15 container to store the cross-repo graph.
2. **Registry API** (`cd api && go run ./cmd/server`)
   - The stateful brain. Connects to Postgres, maps dependencies, and exposes the graph to the Dashboard and MCP.
3. **Diff Engine** (`cd engine && go run ./cmd/substrate serve`)
   - The stateless worker. It only takes two schemas, compares them, and returns the breaking changes. Exposed over port 8080.
4. **GitHub Worker** (`cd github-app && npm run dev`)
   - The Cloudflare Worker that listens to GitHub Webhooks, fetches the PR files, and orchestrates the Registry API and Diff Engine.
5. **Svelte Dashboard** (`cd dashboard && npm run dev`)
   - The visualizer UI to see the live graph and audit history.

**Why multiple Golang binaries?**
- `engine/cmd/substrate`: A stateless CLI tool that can be run in Github Actions or as a microservice (`serve`).
- `api/cmd/server`: A stateful API that requires Postgres. Separated from the engine so the engine can be used purely locally/offline.
- `engine/cmd/substrate-mcp`: A specialized wrapper for AI IDEs (Cursor, Claude) that provides tools via JSON-RPC over stdio. It operates statelessly locally, but makes HTTP requests to the centralized `api/cmd/server` to fetch the global cross-repo dependency graph.

---

# 🚀 WASM Micro-Tools (Product-Led Growth Engine)

Because Substrate's core Go diffing engine is compiled to a portable WebAssembly (`engine.wasm`) binary, we have a massive competitive advantage for lead generation: we can run enterprise-grade schema analysis entirely in the user's browser for free, with zero server costs.

As part of Phase 13 (Go-To-Market), we will launch a standalone, single-page hub at **`tools.substrate.dev`** housing a suite of "Micro-Tools" that act as lead-generation magnets for the core Substrate CI/CD platform.

### Target Personas & Tools

**1. For Backend & Frontend Developers**
* **The "Zero-Trust" OpenAPI Diff Studio:** Users paste `v1.yaml` on the left and `v2.yaml` on the right. The WASM engine highlights breaking changes instantly. (Sell: "100% Secure. Your proprietary APIs never leave your laptop.")
* **VS Code / Cursor Extension:** The WASM engine runs on every keystroke inside the IDE, providing instant red squiggly lines if an edit breaks an API contract, without hitting a server.
* **Instant SDK Compiler:** Paste a GraphQL or OpenAPI spec, and the WASM engine compiles it into a strongly-typed TypeScript/Python SDK instantly.

**2. For QA & Testing Teams**
* **In-Browser Mock Server:** Paste an OpenAPI spec, and the WASM engine uses a Service Worker to intercept network requests, returning fake JSON data that matches the schema perfectly. A local mock API in seconds.
* **Instant Postman/Playwright Generator:** Paste an API contract to generate a fully populated Postman Collection or Playwright E2E script covering all endpoints.

**3. For DevOps & Platform Engineers**
* **Terraform / Helm Blast Radius Visualizer:** Paste a `terraform plan` output. The WASM engine parses the IaC and draws a dependency graph showing exactly what resources will be destroyed.

**4. For Data Engineers**
* **The "Will it Break?" SQL Migration Tester:** Paste `schema.sql` and `migration.sql` to instantly verify if dropping a column breaks downstream ETL contracts.
* **dbt DAG Visualizer:** Paste massive `models.sql` files to instantly draw the Directed Acyclic Graph (DAG) of table flows without spinning up a data warehouse.

### Architecture Strategy
* **Repository:** We will create a new `/tools-site/` directory inside this monorepo. This allows it to effortlessly consume the `engine.wasm` output from the `Makefile`.
* **Deployment:** Hosted on Cloudflare Pages for $0/month.
* **The Hook:** Every tool solves a daily pain point for free, but features a call-to-action: *"Want to automate this in your CI pipeline? Install the Substrate GitHub App."*

### UI/UX Reference Architectures (The Blueprint)
To ensure these tools are massively successful, we will heavily model their UX after the "Hall of Fame" developer micro-tools:
1. **Transform.tools** (Clean Split-Screen): Our SDK Compilers will mimic their beautifully simple left/right pastebin interface.
2. **Regex101.com** (Zero-Latency Feedback): Our SQL Migration Tester must feel like Regex101—instant red/green DOM updates on every keystroke.
3. **JWT.io** (The PLG Funnel): We will copy Auth0's playbook, using a highly useful free tool to subtly funnel enterprise teams into our paid CI/CD product.
4. **CyberChef** (Complex Client-Side Pipelines): Proves that heavy computations (like our WASM AST Diffing) can be trusted to run 100% securely on the client.
5. **AST Explorer** (Deep Debugging): We will expose a simplified view of our Go Engine's AST parser so developers can visually debug why their schemas are failing.

---

# 🚀 Phase 14: Moonshots & Enterprise Platform Evolution

While Phase 13 focuses on Developer Acquisition (WASM micro-tools), Phase 14 focuses on converting massive enterprise accounts by evolving Substrate from a "schema checker" into an **Autonomous Platform**.

**1. The AI "Auto-Fix" PR (Downstream Remediation)**
* **The Problem:** Blocking a breaking change creates a standoff between Team A (who made the change) and Team B (who consumes it).
* **The Moonshot:** When Substrate detects a break, it uses our AI Intelligence Layer to *automatically open a PR against Team B's repository* with the exact code changes needed to adapt to Team A's new schema. We don't just block the break; we write the code to fix it.

**2. The Zero-Config Developer Portal (The "Backstage" Killer)**
* **The Problem:** Enterprises spend millions maintaining Spotify Backstage, but the `catalog-info.yaml` files go stale instantly because humans have to update them.
* **The Moonshot:** Because Substrate automatically parses IaC, Docker files, and APIs to draw the dependency graph, Substrate *is* an auto-generating Developer Portal. We will add a "Catalog" UI that lists every microservice, its owner, and its API docs—100% automatically generated from code.

**3. The Architecture "Time Machine"**
* **The Idea:** Since Substrate processes every webhook and stores the state of the graph at every commit, we will add a slider to the bottom of the Dashboard. Architects can drag it backward in time to see exactly how their microservice architecture evolved over the last 12 months.

**4. Shadow API Detection (Static vs. Runtime)**
* **The Idea:** Integrate Substrate with Datadog/OpenTelemetry. Substrate compares the "Static Contract" (the OpenAPI file in GitHub) against the "Runtime Reality" (the actual HTTP traffic). If traffic hits `/api/v1/hidden` but it's not in the schema, Substrate flags a **Shadow API Security Alert**.

**5. "Cost of Breakage" Analytics**
* **The Idea:** Assign an estimated engineering-hour cost to every node in the graph based on commit frequency. The GitHub PR comment doesn't just say *"You are breaking 3 services."* It says, *"Warning: This breaking change will require an estimated 45 engineering hours to fix across 3 teams. Are you sure?"*

**6. The GitOps Architecture Wiki**
* **The Problem:** Dependency graphs tell you *what* connects to *what*, but they don't explain *why*. Standalone wikis rot because they are disconnected from the code.
* **The Moonshot:** We introduce a `SUBSTRATE.md` standard. Teams write their architecture decisions, runbooks, and mermaid diagrams directly in their repo. The Substrate Webhook auto-fetches this file and attaches it to the node in the Service Catalog. The Substrate Dashboard becomes a living, auto-updating engineering wiki that is reviewed in the exact same PRs as the code changes.

---

# Business Model & Monetization

Substrate scales in value as an organization's complexity grows. The proposed model is a Product-Led Growth (PLG) approach utilizing a **Usage-Based Pricing Model (Per-Repository)** to eliminate seat-based friction:

## 1. Community Tier (PLG Wedge)
* **Target:** Individual developers, small startups, side projects.
* **Price:** Free forever.
* **Limits:** Up to **3 connected repositories**.
* **Goal:** Create a frictionless 2-minute GitHub App installation. By allowing 3 repos, teams can connect a backend API to a frontend consumer and experience the cross-repo "Aha!" moment for free. The moment they want to roll it out to their wider architecture (4+ repos), they hit the paywall.

## 2. Pro / Scale Tier (Usage-Based)
* **Target:** Mid-market companies and scale-ups moving to microservices.
* **Model:** Per-Repository pricing (e.g., $49/mo for up to 5 connected repos; $299/mo for up to 25 repos).
* **Why it works:** CI/CD tools suffer from "per-seat" friction because CI runs for the whole team automatically. Charging by repository scales effortlessly as their architecture grows, without haggling over which engineer needs a paid "seat".

## 3. Enterprise Tier (Custom Pricing)
* **Target:** Large enterprises, Fintech, Healthcare ($50k+/year).
* **Features:**
  * Self-hosted / VPC deployment (data privacy for financial/health data)
  * SSO / SAML integration (Okta, etc.)
  * **Policy-as-Code Engine:** Move beyond hardcoded "breaking changes" to custom enterprise governance rules. Examples:
    * `if pii_fields_changed -> require: compliance-team-approval`
    * `if affected_services > 10 -> reject_change_automatically`
    * `if service == payments -> require: security-review`
  * Compliance reporting and audit logs (SOC2)

---

# Execution Strategy & The Biggest Risk

The biggest risk to Substrate is **Scope Creep**.

The risk is trying to become OpenAPI, SQL, Kafka, GraphQL, AI, SDK Generator, Backstage, Datadog, Snowflake, and Collibra all on day one.

Trying to be "everything everywhere all at once" is how startups die.

**The Execution Plan:**

```
1. Best API Contract Platform          Phase 1a ✅  shipped v0.1.0
        ↓
2. All 8 Schema Adapters               Phase 1b–1g ✅  SQL, GraphQL, Protobuf,
                                                        AsyncAPI, Avro, Terraform,
                                                        AI/ML, Enterprise (Salesforce)
        ↓
3. Best Contract + Distribution        Phase 2  ✅  GitHub App live. Cloudflare
   Platform                                         Worker + Container wired.
        ↓
4. Best Intelligence Platform          Phase 3  ✅  Registry ✅, Dashboard ✅,
                                                    MCP Server ✅, AI Playground ✅
        ↓
5. V1.0 Release + Marketplace          v1.0  ✅  Docs complete. Platform ready.
        ↓
6. Best AI Engineering Platform        Phase 4  ✅  AI Reasoning Bridge →
                                                     Streaming SSE → Schema
                                                     Patch Generator (Jules
                                                     sessions active)
        ↓
7. Zero-Config Dependency Graph        Phase 5  ✅  Env Scanners → Terraform →
                                                     Message Queues → Graph UI
        ↓
8. QA & Automation Layer               Phase 6  ✅  Test Gen → Postman Sync →
                                                     Coverage → Mock Servers
        ↓
9. V1.0 General Availability           v1.0  🚀  GitHub App Auto-Discovery,
                                                     CLI AI Architect,
                                                     Interactive Diff Viewer,
                                                     Deployment Safety Gate
```

## Upcoming V1.0 Capabilities
We are preparing for V1.0 General Availability! The following exciting new features are coming:
* **GitHub App Auto-Discovery:** Zero-touch onboarding by automatically discovering API contracts across repositories.
* **CLI AI Architect:** (`substrate init --design`) Interactive LLM-powered CLI to design OpenAPI contracts before writing code.
* **Interactive Diff Viewer:** Vercel-style preview UI for a side-by-side visual comparison of schema breaks right in PR comments.
* **Deployment Safety Gate:** (`substrate check-deploy`) Blocks provider microservices from deploying breaking changes in CI/CD before their downstream consumers are updated and ready.

## Future Vision & Roadmap
For details on Phase 7 (Enterprise/Scale), Phase 8 (Governance), and Phase 9 (Security/Ecosystem) beyond V1.0, please see our [Roadmap](docs/ROADMAP.md).

**Why Phase 2 before Phase 1c–1g?**

Phase 2 is the distribution unlock — it turns a CLI tool into a product with users, accounts, and a monetization funnel. Building GraphQL or Protobuf support before Phase 2 is expanding coverage for users we don't have yet. Once Phase 2 is live, real usage data — not guesses — will determine which schema format ships next in Wave 2. This mirrors how every successful developer tool grows: nail the wedge, build the platform, then expand features from a position of distribution.

One step at a time.

---

# Credits & License

Substrate is proprietary software. All source code, specifications, and internal tooling are confidential.

The Phase 1a OpenAPI core engine relies on the incredible work done by the [oasdiff](https://github.com/Tufin/oasdiff) community (Apache 2.0). Full attribution is provided in the `NOTICES` file at the repository root, as required by the Apache 2.0 license.
