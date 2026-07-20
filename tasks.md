# Substrate — Task Tracker

> Last updated: 2026-07-15 (Phase 11 🎨 Core UI/UX Overhaul tasks T10, T11, T14, T15 merged ✅)
> Tracking all development phases, tasks, and their current status.

---

## Legend

| Symbol | Status |
|---|---|
| 🔄 | In Progress |
| ⏳ | Ready to Start (all dependencies met) |
| 🔒 | Blocked (waiting on dependency) |
| 💡 | Planned (not yet started) |

---


> **Note:** Tasks for Phases 0 through 6 have been archived to [completed_tasks.md](./completed_tasks.md)

## 🔐 Phase 9: Compliance, IDEs & Developer Experience

**Goal:** Provide compliance auditing, IDE-level developer experience, and governance mapping.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P9-T02** | 🟡 P2 | **Continuous AI Sync (`watch`)** — Background daemon that monitors code changes in the IDE and updates the local OpenAPI spec in real-time. | Unassigned | 💡 Backlog | `docs/specs/go-to-market-strategy.md` |
| **P9-T03** | 🟢 P3 | **Compliance Mapping** — Auto-tag schemas with SOC2/GDPR/HIPAA warnings when fields like `ssn` or `medical_history` are detected. | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |
| **P9-T04** | 🔵 P4 | **Quality Gates (SonarQube-style)** — Allow setting different failure thresholds based on service tier (e.g., Tier 1 allows 0 warnings, Beta allows breakages). | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |
| **P9-T05** | ⚪ P5 | **Hexagonal Architecture & `sqlc` Refactor** — Formally isolate engines from HTTP transports, migrate raw `pgx` queries to `sqlc` for type-safe DB layer generation. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T06** | ⚪ P6 | **Configuration Management (Viper)** — Migrate `os.Getenv` calls to Viper for robust `.env`, CLI flag, and YAML configuration loading. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T07** | 🔴 P1 | **AI Migration Planner** — Upgrade AI Autofix to generate safe, multi-step migration plans for complex schema/database changes with minimal downtime. | Jules | ✅ Complete | `docs/specs/phase-9/p9-t07-ai-migration-planner.md` |
| **P9-T08** | 🟡 P2 | **Deployment Risk Scoring** — Synthesize breaking change data, infrastructure changes, and downstream blast radius into a holistic "Deployment Risk Score". | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T09** | 🟢 P3 | **AI Impact Analysis Summaries** — Pass cross-repo blast radius checks to the AI handler to generate a plain-English impact summary on PRs. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T11** | 🚀 P1 | **Automated Deprecation Campaigns** — Track sunsetting endpoints, auto-open issues in downstream consumer repos, and nag them until 0% usage is reached. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T13** | 🟡 P2 | **Embedded Mermaid Blast Radius** — Upgrade the GitHub PR comment bot to render a visual Mermaid.js flowchart of the exact blast radius directly inside the PR, eliminating the need to click away. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T14** | 🟢 P3 | **Ephemeral API Preview URLs** — Generate temporary, shareable Substrate dashboard URLs for PRs so engineers can share proposed schema changes and interactive diffs with frontend teams before merging. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T16** | 🚀 P1 | **WASM Git Pre-Commit Hooks** — Blazing fast local Git hooks that run `substrate diff` in 0.02s before code ever leaves the developer's laptop. | Jules | ✅ Complete | `docs/specs/phase-9/p9-t16-wasm-hooks.md` |
| **P9-T17** | 🔴 P1 | **Phase 9 E2E Validation (No Mocks)** — Comprehensive end-to-end testing suite for all Phase 9 compliance, DX, and hook features. Must spin up real Postgres DBs, real Substrate CLI integrations, and real Git repositories—strictly no mocking. | Unassigned | 💡 Backlog | `(Pending)` |

---

## 🚀 Phase 10: Ecosystem Expansion & Security (Post-V1.0)

**Goal:** Expand Substrate's reach into API gateways and automated security testing.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P10-T01** | 🔴 P1 | **Automated Security Fuzzing (OWASP)** — Upgrade the Phase 6 fuzzer to inject malicious payloads (SQLi, IDOR) based on the schema, acting as an automated pentester. | Jules | 🔒 Blocked (AI Safety Refusal) | `docs/specs/phase-10/p10-t01-security-fuzzing.md` |
| **P10-T03** | 🟢 P3 | **AI Mock Data Generator (QA)** — Scan QA repositories for JSON test fixtures and use the AI engine to auto-update mock data when the upstream API schema changes. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T04** | 🔵 P4 | **AI Spectral Linter (API Governance)** — Enforce plain-English API design rules (e.g. "All endpoints must use camelCase") during the PR diff process to maintain org-wide consistency. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T05** | ⚪ P5 | **Auto-SDK Generator PRs** — Automatically generate TypeScript/Swift/Go clients via OpenAPI Generator when a schema is merged, opening PRs directly in the downstream consumer repos. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T06** | ⚪ P6 | **Traffic-Aware Pruning (Zombies)** — Correlate schema endpoints with live Datadog/OTel metrics to detect unused "zombie" APIs and auto-generate PRs to delete the dead code. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T07** | 🟣 P1 | **MCP Runtime Diffing** — Spin up Model Context Protocol (MCP) servers in a sandbox during CI/CD to dynamically diff `tools/list` and block AI agent breaking changes. | Jules | ✅ Complete | `docs/specs/phase-10/p10-t07-mcp-diffing.md` |
| **P10-T10** | 🚀 P1 | **Consumer-Driven Contract Manifests** — Allow frontend apps to upload `.substrate-consumer.yaml` declaring required fields, directly competing with PactFlow. | Jules | ✅ Complete | `docs/specs/phase-10/p10-t10-consumer-contracts.md` |
| **P10-T13** | 🔴 P1 | **Integrated API Documentation Catalog** — Evolve the registry into an internal Developer Portal by embedding interactive API reference viewers (like Stoplight Elements or ReDoc) directly into the dashboard. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T15** | 🚀 P1 | **Zero-Latency Drift Detection (eBPF)** — Extend runtime drift detection with a zero-latency `cilium/ebpf` kernel probe for high-throughput environments. *UX Goal: Provide a pre-packaged Helm chart (`helm install substrate-ebpf`) that auto-detects pods via Kubernetes labels (e.g., `substrate.io/monitor: "true"`).* | Jules | 🔄 Running (`7308026507766510849`) | `docs/specs/phase-10/p10-t15-ebpf-drift.md` |
| **P10-T18** | 🚀 P1 | **GraphQL Supergraph Federation** — Add native support for Apollo Federation to diff subgraphs and prevent routing breakages. | Jules | ✅ Complete | `docs/specs/phase-10/p10-t18-graphql-federation.md` |
| **P10-T19** | 🚀 P1 | **CRM/Billing Blast Radius** — Integrate with Stripe and Salesforce to map external customer impact on internal API breakages. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T20** | 🔴 P1 | **Phase 10 E2E Validation (No Mocks)** — End-to-end testing of GraphQL Supergraphs, CDC Manifests, eBPF probes, and MCP diffing. Must use real live gateway environments, live Git hooks, and zero mock APIs. | Unassigned | 💡 Backlog | `(Pending)` |

---

## 🗺️ Phase 11: Advanced Graph Visualization (V2.0 UX)

**Goal:** Elevate the Substrate Dependency Graph into a world-class architectural explorer with cascading impact analysis and team-based layouts.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|

### Phase 12: V2.0 Public Launch & Quality Assurance
| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|

---

## 👑 Phase 13: God-Mode & Enterprise Intelligence

**Goal:** Evolve Substrate into a predictive, financial, and auto-healing infrastructure intelligence platform.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|

---

## 🔮 Phase 14: Predictive Intelligence & Viral Growth

**Goal:** Extend existing AI features into proactive predictions and add viral, self-marketing growth loops to Substrate.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P14-T04** | 🟡 P2 | **Contract Score Badge (Viral Growth Mechanism)** — A `shields.io`-style embeddable README badge showing a repo's API contract reliability score (`CONTRACT: A+ \| 98% \| 0 breaks in 90 days`). Score calculated from breaking change frequency, blast radius, and spec-first adoption rate. | Unassigned | 💡 Backlog | `docs/specs/phase-14/p14-t04-contract-score-badge.md` |
| **P14-T05** | 🟡 P2 | **Retroactive Dependency Archaeology (Paid Onboarding Service)** — `substrate archaeology --since 2-years` scans full git history of all connected repos and generates a paid audit report showing every historical breaking change and its estimated incident cost. Priced as a one-time add-on ($500–$2,000/org). | Unassigned | 💡 Backlog | `docs/specs/phase-14/p14-t05-archaeology.md` |
| **P14-T06** | 🟢 P3 | **Substrate Cloud Public Schema Registry (The npm for APIs)** — A hosted public registry where OSS projects and SaaS companies publish versioned API schemas. Teams monitor public APIs (Stripe, GitHub, Twilio) and get alerts on breaking changes. Free: 5 public APIs. Paid: unlimited + private. | Jules | 🔄 Running (`13433426921692773866`) | `docs/specs/phase-14/p14-t06-public-schema-registry.md` |
| **P14-T07** | 🔴 P1 | **Phase 14 E2E Validation (No Mocks)** — Validate all Phase 14 predictive intelligence and viral growth features end-to-end. Includes spinning up live webhooks and live PR interaction tests. | Unassigned | 💡 Backlog | `(Pending)` |

---

## 🌐 Phase 15: Ecosystem Domination & Monetization

**Goal:** Own the API governance ecosystem through community network effects, deep enterprise workflow integrations, AI-native governance, and a partner certification program.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P15-T01** | 🔴 P1 | **Schema Review Assignments ("CODEOWNERS for APIs")** — Auto-request reviews from API owners (not code owners) via a `SCHEMAOWNERS` file when a PR touches a schema. Fills a workflow gap that no existing tool addresses — enterprise API governance teams are distinct from dev teams. | Jules | ✅ Complete | `docs/specs/phase-15/p15-t01-schema-owners.md` |
| **P15-T02** | 🔴 P1 | **Granular GitHub Check Suite** — Replace the single "Substrate" CI check with individually passable/overridable checks: `substrate/security`, `substrate/performance`, `substrate/breaking-changes`, `substrate/pii-detection`. Matches how enterprise CI pipelines actually work. | Jules | ✅ Complete | `docs/specs/phase-15/p15-t02-granular-checks.md` |
| **P15-T03** | 🟡 P2 | **"Dependency SLA" Tracking** — Let consumer teams declare `required_notice_days` in `substrate.yaml`. Substrate warns provider teams when a proposed breaking change will breach a declared SLA before the PR is merged. Enterprise compliance paper trail for inter-team contracts. | Unassigned | 💡 Backlog | `docs/specs/phase-15/p15-t03-dependency-sla.md` |
| **P15-T04** | 🔴 P1 | **"Schema Smell" Detector (AI API Design Linter)** — Proactively detect API design anti-patterns beyond breaking changes: over-fat endpoints, non-descriptive field names, duplicated response objects without `$ref`. Scores APIs 0–100. Shareable/tweetable output drives organic growth. | Unassigned | 💡 Backlog | `docs/specs/phase-15/p15-t04-schema-smell.md` |
| **P15-T05** | 🟡 P2 | **AI Incident Post-Mortem Generator** — `substrate postmortem --incident <date>` correlates the incident window with schema changes, lists every breaking change and blast radius, and estimates incident cost via the FinOps engine (P13-T01). Outputs a ready-to-share Markdown/Notion document. | Unassigned | 💡 Backlog | `docs/specs/phase-15/p15-t05-postmortem-generator.md` |
| **P15-T06** | 🟡 P2 | **Natural Language Governance Rules** — Extend P7-T03 Custom Rules Engine with a plain-English interface. Platform teams type rules like "All payment APIs must require authentication" and Substrate's AI auto-generates the CEL rule with a preview before saving. Lowers barrier for non-engineer governance stakeholders. | Unassigned | 💡 Backlog | `docs/specs/phase-15/p15-t06-nl-governance.md` |
| **P15-T07** | 🤯 P1 | **Substrate for AI Agents (Agent Contract Registry)** — Track which AI agents call which endpoints via MCP. When a breaking change lands, auto-notify the agent owner to update tool definitions. Owns the new category: **"API Governance for the Agentic Era"** — zero competition today. | Jules | ✅ Complete | `docs/specs/phase-15/p15-t07-agent-contract-registry.md` |
| **P15-T08** | 🔴 P1 | **Substrate Marketplace (Community Rules & Plugins)** — A community marketplace for governance rule packs (`substrate-plugin-hipaa`, `substrate-plugin-pci`, `substrate-plugin-owasp`). Published via `substrate plugin publish`. Network effects compound — every contributed rule pack increases value for all orgs. | Unassigned | 💡 Backlog | `docs/specs/phase-15/p15-t08-marketplace.md` |
| **P15-T09** | 🤯 P1 | **"Substrate Certified" Partner Program** — API platform vendors (Kong, AWS API Gateway, Apigee, Cloudflare) pay $2k–$20k/year for certified native integration status. Includes joint marketing, co-sell revenue share (10–15% ACV), and annual Summit sponsorship. Creates deep switching-cost lock-in for enterprise customers. | Unassigned | 💡 Backlog | `docs/specs/phase-15/p15-t09-partner-program.md` |
| **P15-T10** | 🟢 P3 | **Schema Insurance (Enterprise Tier Add-On)** — Premium enterprise add-on: if a breaking change slips through Substrate's monitoring and causes a verified production incident, Substrate pays an SLA credit. Turns Substrate into a risk management instrument, not just a dev tool. | Unassigned | 💡 Backlog | `docs/specs/phase-15/p15-t10-schema-insurance.md` |
| **P15-T11** | 🟡 P2 | **Substrate for Startups (Free Tier with Public Audit Trail)** — Free forever for OSS projects with a public API reliability profile (`substrate.io/profile/myorg/api`). Startups link their Substrate profile in enterprise sales and security questionnaires as proof of API stability — credibility-as-a-service, zero SOC 2 required. | Unassigned | 💡 Backlog | `docs/specs/phase-15/p15-t11-startups-free-tier.md` |
| **P15-T12** | 🚀 P1 | **Enterprise BYOK (KMS & LLM)** — Dual-BYOK architecture. Allows enterprises to encrypt their schemas at rest using AWS KMS/Vault, and route all AI workloads through their own Azure OpenAI/Bedrock VPC endpoints so their proprietary IP never leaves their perimeter. | Jules | 🔄 Running (`5902969312457195193`) | `docs/specs/phase-15/p15-t12-enterprise-byok.md` |
| **P15-T13** | 🔴 P1 | **Phase 15 E2E Validation (No Mocks)** — Live orchestration testing of KMS BYOK, LLM workload routing, and Granular Check Suites against actual GitHub API and AWS/Vault environments. | Unassigned | 💡 Backlog | `(Pending)` |
| **P15-T14** | 🤯 P1 | **Headless Substrate (Full MCP Server Parity)** — 100% of Substrate's GUI/CLI functionality mapped to MCP Tools and Resources. Allows AI agents (Cursor, Claude) to completely control, configure, and manage Substrate without any human intervention. | Jules | 🔄 Running (`17043707948851769530`) | `docs/specs/phase-15/p15-t14-headless-mcp.md` |
