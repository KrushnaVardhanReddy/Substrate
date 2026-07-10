# Substrate — Go Diff Engine: Architecture & Development Plan

> **Status:** Active — Engine live in production (v0.1.0). Phase 1b (SQL) in progress.
> **Spec-First Gate:** No Go code is written until the specs in this document are finalized and agreed upon.
> **Strategy Decision (2026-07-07):** Phase 1c–1g schema formats are deliberately deferred until Phase 2 (GitHub App) ships. See [Execution Order](#development-milestones) below.

---

## What Is the Diff Engine?

The Go Diff Engine is the **beating heart of Substrate**. It is a standalone Go module (and CLI binary) that takes two versions of a schema file and produces a structured, machine-readable report of every change — classified by severity.

```
Input:   schema_before.yaml  +  schema_after.yaml
Output:  DiffReport (JSON)
```

Everything else in the Substrate platform — the GitHub App, the dashboard, the dependency graph — is a consumer of this engine. If this engine is wrong, the entire product is wrong. **Trust is the entire product.**

---

## Technology Decisions

### Why Go?

- Compiles to a single static binary — easy to distribute and run in any CI environment
- Excellent performance for file I/O and JSON processing
- Strong standard library — no framework needed
- Easy to embed into a GitHub App sidecar later

### Why PostgreSQL (not a Graph DB)?

This is a question worth addressing upfront.

**The case for a Graph DB (Neo4j, etc.):** The core data structure IS a graph. Graph DBs are optimized for traversal queries like *"find all nodes impacted within 3 hops of this change."*

**Why PostgreSQL wins right now:**

1. **Operational simplicity.** Graph DBs are expensive to self-host, complex to operate, and require specialized knowledge. A small team does not want to debug Neo4j cluster issues at 2am.
2. **PostgreSQL can traverse graphs via recursive CTEs:**
   ```sql
   WITH RECURSIVE dependents AS (
     SELECT consumer_id FROM dependencies WHERE provider_id = 'customer-service'
     UNION ALL
     SELECT d.consumer_id FROM dependencies d
     INNER JOIN dependents dep ON d.provider_id = dep.consumer_id
   )
   SELECT * FROM dependents;
   ```
   For 500 services and a few thousand edges, this is fast enough.
3. **Schema evolution flexibility.** As the product matures and the graph model gets more nuanced (field-level dependencies, version-aware edges, historical snapshots), a relational model gives more flexibility than a rigid graph DB schema.

**Long-term:** If Substrate reaches enterprise scale with 1000+ repo organizations, we revisit a dedicated graph DB or the `Apache AGE` PostgreSQL extension. That is a good problem to have.

---

## Supported Schema Formats (Prioritized)

Formats are delivered in two waves. Wave 1 must ship before Phase 2 (GitHub App). Wave 2 resumes after Phase 2 is live, ordered by real user demand.

| Priority | Wave | Format | Status | Why |
|---|---|---|---|---|
| 1 | Wave 1 | **OpenAPI 3.x YAML/JSON** | ✅ shipped | Most REST APIs, highest pain point, largest market |
| 2 | Wave 1 | **SQL Migrations** (PostgreSQL DDL) | 🔄 in progress | Databases are everywhere; completes the core contract engine |
| 3 | Wave 2 | **GraphQL SDL** | 🔒 post-Phase 2 | Large market, especially product companies |
| 4 | Wave 2 | **Protobuf & gRPC** | 🔒 post-Phase 2 | Enterprise/microservice-heavy, Google-style shops |
| 5 | Wave 2 | **AsyncAPI & Apache Avro** | 🔒 post-Phase 2 | Kafka-heavy data engineering & event-driven architectures |
| 6 | Wave 2 | **AI/ML Model Contracts** | 🔒 post-Phase 2 | Preventing silent data-drift and ML training failures |
| 7 | Wave 2 | **Enterprise Metadata** | 🔒 post-Phase 2 | Salesforce Custom Objects & SOAP WSDLs for legacy integrations |

**Rationale for the Wave split:** Wave 2 formats are deferred not because they are unimportant, but because shipping Phase 2 (GitHub App, PR comments, user accounts) before them is the right product move. Distribution unlocks monetization. Monetization funds Wave 2 development. Building 1c–1g before Phase 2 is expanding coverage for users we don't yet have.

**Rule:** Start with OpenAPI only. Ship that. Then add SQL. Then ship the GitHub App. Do not expand schema formats before the platform exists.

---

## Internal Architecture

```
CLI Input (two file paths)
        ↓
File Parser (auto-detects schema type)
        ↓
Schema Normalizer (converts to Internal Representation — IR)
        ↓
Diff Comparator (produces raw list of changes)
        ↓
Rule Engine (classifies each change by severity)
        ↓
Report Generator (outputs DiffReport as JSON or human-readable text)
```

### The Key Insight: Internal Representation (IR)

Instead of writing a separate diff algorithm for each schema format (OpenAPI, GraphQL, Protobuf, SQL), we parse all of them into **one common tree structure** first, then run a **single diff algorithm** on that tree.

This is how production-grade tools like `buf` (Protobuf) and `oasdiff` (OpenAPI) work internally.

### Parser Strategy: Use Libraries, Own the Rules

| Concern | Approach |
|---|---|
| **Parsing schema files** | Use existing open-source Go libraries (solved problem) |
| **Diff & classification logic** | Build in-house — this is our actual IP |

Recommended Go parsing libraries:

- OpenAPI → `github.com/pb33f/libopenapi`
- SQL → `github.com/pganalyze/pg_query_go`
- GraphQL → `github.com/graphql-go/graphql`
- Protobuf → `google.golang.org/protobuf`

---

## Development Milestones

### Milestone 1: Core Contract Engine (Phase 1a + 1b) ✅
> OpenAPI 3.x diff engine shipped as v0.1.0 GitHub Action. SQL (PostgreSQL DDL) engine in active development. Both use the same DiffReport output contract and rule engine pattern.

### Milestone 2: The Distribution Platform (Phase 2) ⏳ NEXT
> GitHub App: 1-click org install, PR webhook listener, automated PR comments, merge blocking, user accounts and org management. Deployed on Cloudflare Workers. This is the product monetization unlock — it ships immediately after Phase 1b SQL is complete.

### Milestone 3: The Intelligence Layer (Phase 3) 💡
> Multi-repo dependency graph. SvelteKit dashboard showing connected repos, schema history, and impact analysis. Manual `substrate.yaml` dependency declarations. MCP server for IDE integration. Free tier live for beta users.

### Milestone 4: Extended Contract Formats (Phase 1c–1g, Wave 2) 💡
> Resume schema format expansion in demand order: GraphQL SDL → Protobuf/gRPC → AsyncAPI/Avro → AI/ML Model Contracts → Enterprise Metadata. Ordering is driven by user requests from Phase 2 installs, not predetermined.

### Milestone 5: AI Engineering Assistant (Phase 4) 💡
> AI layer on top of the dependency graph: system explanations, auto-generated migration code, impact predictions, and multi-agent handoff memory.

---

## Spec-First Gate for This Module

Before any Go code is written for the diff engine, the following two spec documents must be completed and agreed upon:

| Order | Document | Why First |
|---|---|---|
| **1st** | `docs/specs/diff-report-schema.md` | Defines the output contract — the JSON structure of a `DiffReport`. This is the ground truth for all consumers (GitHub App, Dashboard, CLI output). |
| **2nd** | `docs/specs/breaking-change-rules.md` | Defines what constitutes a "breaking" vs "non-breaking" change. This feeds the Rule Engine. It depends on knowing the output format first. |

---

## Why DiffReport Schema Comes First

The `DiffReport` JSON schema is the **public contract** of the entire engine. It is what:
- The GitHub App parses to write PR comments
- The Dashboard renders to show impact reports  
- The CLI prints to the terminal
- Future integrations (Slack, Jira, PagerDuty) consume

The breaking change rule table defines internal classification logic, but it is meaningless without knowing where those classifications go. **You design the output first, then you design the rules that produce it.** This is the spec-first principle applied recursively.

> Next step → Create `docs/specs/diff-report-schema.md`
