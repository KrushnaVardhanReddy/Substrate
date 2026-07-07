# Substrate

## The Living Map of Your Engineering Ecosystem

Substrate is a CI/CD-integrated data contract and dependency intelligence platform that prevents downstream data failures before they reach production.

It acts as a proactive firewall for engineering teams by analyzing schema changes, understanding repository dependencies, identifying impacted systems, and blocking unsafe changes before they break applications, analytics, dashboards, and data pipelines.

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

## Public/Private Repo Split

Substrate follows the standard developer tool playbook (like Vercel, Datadog, Stripe) by splitting the product into a public engine and a private SaaS platform:

- **Public Repository (`substrate-engine`)**: Contains the Go CLI, GitHub Action, and core diff logic. This builds trust, allows community audits, and acts as the distribution wedge via GitHub Marketplace.
- **Private Repository (`substrate-app`)**: Contains the GitHub App webhook handler, API server, billing, and SvelteKit dashboard. This protects the proprietary business logic and SaaS platform.

## Distribution & Go-to-Market (GTM)



1. **GitHub Marketplace (The MVP Wedge) 🏆**: The immediate path to users is shipping the CLI as a GitHub Action. Developers search "API breaking changes", install the action, and get CI failures on breaking changes. This is the highest ROI acquisition channel.
2. **"Optic Alternative" SEO**: With Optic sunset in 2026, teams are actively searching for replacements. Publishing content positioning Substrate as the modern alternative captures this high-intent traffic.
3. **Developer Communities (HN, Reddit)**: Sharing the engineering journey (e.g., "Building an API contract checker in Go using oasdiff") generates organic awareness.
4. **Direct Outreach**: Finding public repos with API specs and offering the tool directly to engineering managers to secure the first 10 design partners.

---

# Product Evolution

## Phase 1: Breaking Change Prevention

Goal:

Stop production failures across every schema type your engineering organization uses.

Phase 1 ships incrementally as schema format support is added to the same diff engine:

* **1a — OpenAPI 3.x** (REST APIs) ← *ships first*
* **1b — SQL Migrations** (PostgreSQL DDL)
* **1c — GraphQL SDL**
* **1d — Protobuf & gRPC** (Microservices)
* **1e — AsyncAPI & Apache Avro** (Kafka / Event-Driven Architectures)
* **1f — AI/ML Model Contracts** (Model inputs, dataset schemas)
* **1g — Enterprise Metadata** (Salesforce Custom Objects, SOAP WSDLs)

Core features in all sub-phases:

* GitHub integration
* Schema diff detection
* Contract validation
* PR comments
* Merge blocking

### Phase 1e: AI & ML Contract Protection

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
* **Automated Dependency Discovery:** Eliminate manual YAML configuration by automatically mapping the dependency graph through:
  * **Distributed Tracing:** Native ingestion of OpenTelemetry, Datadog APM, or New Relic traces.
  * **Network Layer (eBPF & Service Mesh):** Direct integration with Istio, Linkerd, or eBPF network logs to map service-to-service communication.
  * **Static Code Analysis:** AST scanning in CI to detect SDK imports and API calls.
* **Service Ownership:** Map every discovered node to a team and an alert channel.

---

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

# Business Model & Monetization

Substrate scales in value as an organization's complexity grows. The proposed model is a Product-Led Growth (PLG) approach featuring three tiers:

## 1. Free / Starter Tier (PLG Wedge)
* **Target:** Individual developers, small startups, side projects.
* **Limits:** Up to 3 connected repositories, max 5 developers, 30 days of schema change history.
* **Goal:** Create a frictionless 2-minute GitHub App installation that delivers an immediate "Aha!" moment when it catches the first breaking change.

## 2. Pro / Team Tier (Active Seat-Based)
* **Target:** Mid-market companies and scale-ups (50 to 300+ engineers).
* **Model:** Per-seat pricing (e.g., ~$25/user/month) for every active code contributor.
* **Why it works:** More engineers = more communication silos = higher risk of downstream breakages. 
* **Key strategy:** "View-only" seats are free. Product Managers, Data Analysts, and QA can view the dependency graph and architecture docs at no cost, allowing the tool to virally spread through the organization.

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
1. Best **API Contract** Platform (Phase 1a)
   ↓
2. Best **Universal Contract** Platform (Phase 1b - 1g)
   ↓
3. Best **Dependency** Platform (Phase 2)
   ↓
4. Best **Engineering Knowledge** Platform (Phase 3 & 4)

One step at a time. Laser focus on Phase 1a.
