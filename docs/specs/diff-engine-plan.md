# Substrate — Go Diff Engine: Architecture & Development Plan

> **Status:** Planning / Pre-Code  
> **Spec-First Gate:** No Go code is written until the specs in this document are finalized and agreed upon.

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

| Priority | Format | Why |
|---|---|---|
| 1 | **OpenAPI 3.x YAML/JSON** | Most REST APIs, highest pain point, largest market |
| 2 | **SQL Migrations** | Postgres ALTER TABLE, extremely common |
| 3 | **GraphQL SDL** | Large market, especially product companies |
| 4 | **Protobuf & gRPC** | Enterprise/microservice-heavy, Google-style shops |
| 5 | **AsyncAPI & Apache Avro** | Kafka-heavy data engineering & event-driven architectures |
| 6 | **AI/ML Model Contracts** | Preventing silent data-drift and ML training failures |
| 7 | **Enterprise Metadata** | Salesforce Custom Objects & SOAP WSDLs for legacy integrations |

**Rule:** Start with OpenAPI only. Ship that. Then add SQL. Do not try to support everything on day one.

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

### Milestone 1: The Engine (Weeks 1–4)
> A Go CLI that takes two OpenAPI YAML files and outputs a JSON `DiffReport` of breaking vs. non-breaking changes. All tests pass.

### Milestone 2: The Hook (Weeks 5–8)
> A GitHub App that listens for PR webhook events, clones the repo, runs the diff engine, and posts an impact comment on the PR.

### Milestone 3: The MVP (Weeks 9–16)
> Multi-repo support, a basic SvelteKit dashboard showing connected repos and their schemas, manual `substrate.yaml` dependency declarations, and a free tier ready for beta users.

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
