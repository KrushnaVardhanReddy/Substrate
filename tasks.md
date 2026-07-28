# Substrate — Task Tracker

> Last updated: 2026-07-28 (Phase 12 & 13 complete ✅. All automated tasks done. Manual QA remaining.)
> Tracking all development phases, tasks, and their current status.

---

## Legend

| Symbol | Status |
|---|---|
| 🔄 | In Progress |
| 🤖 | Submitted to Jules (PR pending) |
| ⏳ | Ready to Start (all dependencies met) |
| 🔒 | Blocked (waiting on dependency) |
| 💡 | Planned (not yet started) |

---


> **Note:** Tasks for Phases 0 through 6 have been archived to [completed_tasks.md](./completed_tasks.md)

> **Note:** Phases 3, 9, 10, 11, 12, and 13 are fully complete. See [completed_tasks.md](./completed_tasks.md) for their full task lists.

### Cross-Cutting Refactors (All Complete ✅)
| Task | Priority | Title | Assignee | Status | Spec |
|---|---|---|---|---|---|
| **CC-T01** | 🛡️ P1 | **PASETO Security Migration** | Jules | ✅ PR Merged | `docs/wiki/concepts/architecture.md` |
| **CC-T02** | 🚀 P1 | **E2E Overhaul: Full-Stack PGlite Harness** | Jules | ✅ Complete | `prompts/e2e/pglite_infrastructure.txt` |
| **CC-T03** | 🚀 P1 | **Live VCS E2E Integration (Forgejo)** | Jules | ✅ Complete | `docs/specs/cross-cutting/cc-t03-live-vcs.md` |
| **CC-T04** | 🟢 P3 | **UI Org Context & API Keys Refactor** | Jules | ✅ PR Merged | `docs/specs/ui/org_context_and_apikeys.md` |
| **CC-T05** | 🟢 P3 | **E2E Validation: UI Org Context** | Jules | ✅ Complete | `docs/specs/e2e/ui_org_context_e2e.md` |
| **CC-T06** | 🟢 P2 | **API Keys Backend Integration** | Jules | ✅ PR Merged | `docs/specs/ui/api_keys_backend.md` |
| **CC-T07** | 🟢 P3 | **E2E Validation: API Keys Backend** | Jules | ⚪ Redundant (in T06) | `docs/specs/e2e/api_keys_e2e.md` |
| **CC-T08** | 🚀 P1 | **Spec Sync Governance & IDE Reminders** | Jules | ⏳ Next Up | `docs/specs/cross_cutting/cc-t08-spec-sync-governance.md` |

---

## 🗺️ Phase 12: V2.0 Public Launch & Quality Assurance ✅ COMPLETE

> Full task list in [completed_tasks.md](./completed_tasks.md#phase-12)

| Task ID | Tier | Summary | Status |
|---|---|---|---|
| **P12-T01** | 🔴 P1 | Zero-to-One Onboarding E2E | ✅ Complete |
| **P12-T02** | 🔴 P1 | Svelte Flow Interaction E2E | ✅ Complete |
| **P12-T03** | 🟡 P2 | AI Playground & Diff Viewer E2E | ✅ Complete |
| **P12-T04** | 🔴 P1 | SSE Connection Resilience Test | ✅ Complete |
| **P12-T05** | 🟡 P2 | WASM Engine Boundary Tests | ✅ Complete |
| **P12-T06** | 🟢 P3 | 1,000-Node UI Stress Test | ✅ Complete |
| **P12-T07** | 🔵 P4 | Telemetry & Crash Reporting (PostHog) | ✅ Complete |
| **P12-T08** | 🔴 P1 | V2.0 Production Cutover | ✅ Done |
| **P12-T09** | 🟢 P3 | AI Support Copilot | ✅ Done |
| **P12-T10** | 🔴 P1 | VCS-Agnostic Webhook & API Adapter | ✅ Done |
| **P12-T11** | 🟡 P2 | Multi-VCS Onboarding UI | ✅ Done |
| **P12-T12** | 🚀 P1 | System Matrix E2E (Red/Green/Yellow) | ✅ Done |
| **P12-T13** | ⭐ P1 | Zero-Config Developer Portal (Catalog UI) | ✅ Done |
| **P12-T14** | 🔴 P1 | System Matrix Overrides (Yellow Path) | ✅ Done |
| **P12-T15** | 🚀 P1 | Advanced Feature Suites E2E | ✅ Complete |
| **P12-T16** | 🔴 P1 | Advanced E2E UI Implementation (TDD) | ✅ Complete |

---

## 👑 Phase 13: God-Mode & Enterprise Intelligence ✅ COMPLETE

> Full task list in [completed_tasks.md](./completed_tasks.md#phase-13)

| Task ID | Tier | Summary | Status |
|---|---|---|---|
| **P13-T01** | 🤯 P1 | FinOps Cost Prediction (Egress Calculator) | ✅ Done |
| **P13-T02** | 🤯 P1 | DB Performance Breakages (CLI Analytics) | ✅ Done |
| **P13-T03** | 🤯 P1 | AI Chaos Engineering Auto-Tests | ✅ Complete |
| **P13-T04** | 🔴 P1 | Enterprise E2E Validation | ✅ Done |

---

## 🔮 Phase 14: Predictive Intelligence & Viral Growth

**Goal:** Extend existing AI features into proactive predictions and add viral, self-marketing growth loops to Substrate.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P14-T04** | 🟡 P2 | **Contract Score Badge (Viral Growth Mechanism)** — A `shields.io`-style embeddable README badge showing a repo's API contract reliability score (`CONTRACT: A+ \| 98% \| 0 breaks in 90 days`). Score calculated from breaking change frequency, blast radius, and spec-first adoption rate. | Jules | ✅ PR Merged | `docs/specs/phase-14/p14-t04-contract-score-badge.md` |
| **P14-T05** | 🟡 P2 | **Retroactive Dependency Archaeology (Paid Onboarding Service)** — `substrate archaeology --since 2-years` scans full git history of all connected repos and generates a paid audit report showing every historical breaking change and its estimated incident cost. Priced as a one-time add-on ($500–$2,000/org). | Jules | ✅ PR Merged | `docs/specs/phase-14/p14-t05-archaeology.md` |
| **P14-T06** | 🟢 P3 | **Substrate Cloud Public Schema Registry (The npm for APIs)** — A hosted public registry where OSS projects and SaaS companies publish versioned API schemas. Teams monitor public APIs (Stripe, GitHub, Twilio) and get alerts on breaking changes. Free: 5 public APIs. Paid: unlimited + private. | Jules | ✅ PR Merged | `docs/specs/phase-14/p14-t06-public-schema-registry.md` |
| **P14-T07** | 🔴 P1 | **Phase 14 E2E Validation (No Mocks)** — Validate all new engine components (Archaeology, Contracts) end-to-end against real repositories. | Jules | ✅ PR Merged | `docs/specs/phase-14/p14-t07-e2e-validation.md` |

---

## 🌐 Phase 15: Ecosystem Domination & Monetization

**Goal:** Own the API governance ecosystem through community network effects, deep enterprise workflow integrations, AI-native governance, and a partner certification program.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P15-T01** | 🔴 P1 | **Schema Review Assignments ("CODEOWNERS for APIs")** — Auto-request reviews from API owners (not code owners) via a `SCHEMAOWNERS` file when a PR touches a schema. Fills a workflow gap that no existing tool addresses — enterprise API governance teams are distinct from dev teams. | Jules | ✅ Complete | `docs/specs/phase-15/p15-t01-schema-owners.md` |
| **P15-T02** | 🔴 P1 | **Granular GitHub Check Suite** — Replace the single "Substrate" CI check with individually passable/overridable checks: `substrate/security`, `substrate/performance`, `substrate/breaking-changes`, `substrate/pii-detection`. Matches how enterprise CI pipelines actually work. | Jules | ✅ Complete | `docs/specs/phase-15/p15-t02-granular-checks.md` |
| **P15-T03** | 🟡 P2 | **"Dependency SLA" Tracking** — Let consumer teams declare `required_notice_days` in `substrate.yaml`. Substrate warns provider teams when a proposed breaking change will breach a declared SLA before the PR is merged. Enterprise compliance paper trail for inter-team contracts. | Jules | ✅ PR Merged | `docs/specs/phase-15/p15-t03-dependency-sla.md` |
| **P15-T04** | 🔴 P1 | **"Schema Smell" Detector (AI API Design Linter)** — Proactively detect API design anti-patterns beyond breaking changes: over-fat endpoints, non-descriptive field names, duplicated response objects without `$ref`. Scores APIs 0–100. Shareable/tweetable output drives organic growth. | Jules | ✅ PR Merged | `docs/specs/phase-15/p15-t04-schema-smell.md` |
| **P15-T05** | 🟡 P2 | **AI Incident Post-Mortem Generator** — `substrate postmortem --incident <date>` correlates the incident window with schema changes, lists every breaking change and blast radius, and estimates incident cost via the FinOps engine (P13-T01). Outputs a ready-to-share Markdown/Notion document. | Jules | ✅ PR Merged | `docs/specs/phase-15/p15-t05-postmortem-generator.md` |
| **P15-T06** | 🟡 P2 | **Natural Language Governance Rules** — Extend P7-T03 Custom Rules Engine with a plain-English interface. Platform teams type rules like "All payment APIs must require authentication" and Substrate's AI auto-generates the CEL rule with a preview before saving. Lowers barrier for non-engineer governance stakeholders. | Jules | ✅ PR Merged | `docs/specs/phase-15/p15-t06-nl-governance.md` |
| **P15-T07** | 🤯 P1 | **Substrate for AI Agents (Agent Contract Registry)** — Track which AI agents call which endpoints via MCP. When a breaking change lands, auto-notify the agent owner to update tool definitions. Owns the new category: **"API Governance for the Agentic Era"** — zero competition today. | Jules | ✅ Complete | `docs/specs/phase-15/p15-t07-agent-contract-registry.md` |
| **P15-T08** | 🔴 P1 | **Substrate Marketplace (Community Rules & Plugins)** — A community marketplace for governance rule packs (`substrate-plugin-hipaa`, `substrate-plugin-pci`, `substrate-plugin-owasp`). Published via `substrate plugin publish`. Network effects compound — every contributed rule pack increases value for all orgs. | Jules | ✅ PR Merged | `docs/specs/phase-15/p15-t08-marketplace.md` |
| **P15-T09** | 🤯 P1 | **"Substrate Certified" Partner Program** — API platform vendors (Kong, AWS API Gateway, Apigee, Cloudflare) pay $2k–$20k/year for certified native integration status. Includes joint marketing, co-sell revenue share (10–15% ACV), and annual Summit sponsorship. Creates deep switching-cost lock-in for enterprise customers. | Jules | ✅ PR Merged | `docs/specs/phase-15/p15-t09-partner-program.md` |
| **P15-T10** | 🟢 P3 | **Schema Insurance (Enterprise Tier Add-On)** — Premium enterprise add-on: if a breaking change slips through Substrate's monitoring and causes a verified production incident, Substrate pays an SLA credit. Turns Substrate into a risk management instrument, not just a dev tool. | Jules | ✅ PR Merged | `docs/specs/phase-15/p15-t10-schema-insurance.md` |
| **P15-T11** | 🟡 P2 | **Substrate for Startups (Free Tier with Public Audit Trail)** — Free forever for OSS projects with a public API reliability profile (`substrate.io/profile/myorg/api`). Startups link their Substrate profile in enterprise sales and security questionnaires as proof of API stability — credibility-as-a-service, zero SOC 2 required. | Jules | ✅ PR Merged | `docs/specs/phase-15/p15-t11-startups-free-tier.md` |
| **P15-T12** | 🚀 P1 | **Enterprise BYOK (KMS & LLM)** — Dual-BYOK architecture. Allows enterprises to encrypt their schemas at rest using AWS KMS/Vault, and route all AI workloads through their own Azure OpenAI/Bedrock VPC endpoints so their proprietary IP never leaves their perimeter. | Jules | ✅ PR Merged | `docs/specs/phase-15/p15-t12-enterprise-byok.md` |
| **P15-T13** | 🔴 P1 | **Phase 15 E2E Validation (No Mocks)** — Live orchestration testing of KMS BYOK, LLM workload routing, and Granular Check Suites against actual GitHub API and AWS/Vault environments. | Jules | ✅ PR Merged | `(Pending)` |
| **P15-T14** | 🤯 P1 | **Headless Substrate (Full MCP Server Parity)** — 100% of Substrate's GUI/CLI functionality mapped to MCP Tools and Resources. Allows AI agents (Cursor, Claude) to completely control, configure, and manage Substrate without any human intervention. | Jules | ✅ PR Merged | `docs/specs/phase-15/p15-t14-headless-mcp.md` |
| **P15-T15** | 🟢 P2 | **Air-Gapped License Validator (On-Premise)** — Cryptographically signed `.lic` JWT validation middleware for On-Premise VPC deployments. Automatically degrades the engine to Free Tier if the license expires or is tampered with. | Unassigned | ✅ PR Merged | `docs/specs/phase-15/p15-t15-license-validator.md` |

---

## 📚 Phase 16: The Absolute SSOT (API Documentation)

**Goal:** Transform Substrate from a schema registry into a complete Developer Portal, eliminating the need for external tools like ReadMe or Backstage by merging technical schemas with human-written guides and interactive tools.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P16-T01** | 🚀 P1 | **Markdown Guide Ingestion (Diátaxis)** — Pull `docs/**/*.md` files from provider repos alongside OpenAPI specs. Stitch human-written tutorials, how-tos, and references together with the auto-generated API schema in the UI. | Jules | ⏳ In Progress | `(Pending)` |
| **P16-T02** | 🟡 P2 | **Business Metadata Block** — Expand `substrate.yaml` to accept `metadata` (PM, Slack channel, PagerDuty link, SLA Tier) and display it prominently at the top of the API's documentation page. | Unassigned | 💡 Planned | `(Pending)` |
| **P16-T03** | 🔴 P1 | **Interactive API Sandbox ("Try It Out")** — Build an API proxy into the Go backend that allows developers to generate ephemeral test tokens and issue live requests to the API directly from the Substrate documentation UI. | Jules | ✅ PR Merged | `(Pending)` |
| **P16-T04** | 🟢 P3 | **Auto-Generated SDK Code Snippets** — Dynamically generate copy-pasteable request snippets (Curl, Python, Node, Go) for every endpoint in the documentation based on the OpenAPI schema. | Unassigned | 💡 Planned | `(Pending)` |
| **P16-T05** | 🔵 P4 | **Configurable Graph Node Branding (YAML + MCP)** — Allow teams to define `metadata.node_color` in `substrate.yaml`. Plumb this custom hex color down to the Svelte Flow Dependency Graph UI *and* expose it via the MCP Server resources so AI agents know the visual branding of the nodes they are analyzing. | Unassigned | 💡 Planned | `(Pending)` |
| **P16-T99** | 🔴 P1 | **Phase 16 E2E Validation (No Mocks)** — Validate the full documentation portal: Markdown ingestion from real repos, Interactive Sandbox proxy forwarding real requests via ephemeral tokens, SDK snippet generation, and Business Metadata rendering. Must include Go API tests and Playwright UI tests. | Jules | 💡 Planned | `docs/specs/phase-16/p16-t99-e2e.md` |

---

## 📈 Phase 17: DevOps & Management Intelligence

**Goal:** Surface financial, operational, and structural data to Engineering Managers and DevOps teams, providing zero-touch automation for API Gateways and high-level incident context.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P17-T01** | 🚀 P1 | **API Maturity Scorecards (Manager Dashboard)** — Provide a high-level UI grading APIs (A-F) based on documentation completeness, SOC2/PII compliance, volatility, and shadow test coverage. | Jules | ⏳ Next Up | `(Pending)` |
| **P17-T02** | 🟡 P2 | **Legacy API FinOps Translation** — Expand the zombie API detection to calculate and display the exact estimated dollar amount saved by sunsetting legacy endpoints. | Unassigned | 💡 Planned | `(Pending)` |
| **P17-T03** | 🔴 P1 | **GitOps API Gateway Sync (CRD Generation)** — Automatically generate Kubernetes CRDs (or Terraform state) for Kong/AWS API Gateway based on Substrate schemas, pushing them when PRs are merged. | Jules | ✅ PR Merged | `(Pending)` |
| **P17-T04** | 🔴 P1 | **PagerDuty Blast Radius Injection** — Integrate with Datadog/PagerDuty. On incident creation, Substrate injects the visual Mermaid Dependency Graph into the incident description to instantly show downstream blast radius. | Jules | ⏳ Next Up | `(Pending)` |
| **P17-T05** | 🚀 P1 | **Auto-Rollback via ArgoCD/Flux** — Wire Substrate's `Can-Rollback` and eBPF Drift engine to ArgoCD/Flux webhooks, automatically reverting a deployment if it causes severe schema violations in production. | Jules | ⏳ Next Up | `(Pending)` |
| **P17-T06** | 🚀 P1 | **Deployment & Incident Graph Overlay** — Ingest deployment/incident webhooks (GitHub Actions, Datadog) into an `events` timeline. Visualize these events directly on the Svelte Dependency Graph so developers can instantly correlate downstream failures with upstream changes. | Jules | ⏳ In Progress | `docs/specs/phase-17/p17-t06-deployment-overlay.md` |
| **P17-T99** | 🔴 P1 | **Phase 17 E2E Validation (No Mocks)** — Validate all DevOps integrations: API Maturity Scorecard scoring logic, PagerDuty webhook injection, ArgoCD rollback trigger, and GitOps CRD generation. Use real ArgoCD/Forgejo test environments — no mocks. | Jules | 💡 Planned | `docs/specs/phase-17/p17-t99-e2e.md` |

---

## 🤖 Phase 18: AI Agent Governance & FDAIE Tooling

**Goal:** Provide Forward Deployed AI Engineers (FDAIEs) with the necessary tooling to safely deploy, monitor, and optimize AI agents communicating with enterprise APIs via the Model Context Protocol (MCP).

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P18-T01** | 🔴 P1 | **Granular MCP Tool Governance** — Allow FDAIEs to define "Agent Profiles" in Substrate (e.g., support vs. devops). Dynamically generate restricted MCP toolsets based on the profile, enforcing Human-in-the-Loop requirements for destructive `POST/DELETE` operations. | Jules | ⏳ In Progress | `(Pending)` |
| **P18-T02** | 🚀 P1 | **Context-Aware Schema Pruning** — Implement a "Pruned Schema API" that takes natural language intent, determines the necessary subset of the OpenAPI spec, and returns only the relevant endpoints to minimize LLM token usage and latency. | Jules | ✅ PR Merged | `(Pending)` |
| **P18-T03** | 🟡 P2 | **Agent Execution Audit Trails** — Build an AI Forensics Dashboard that logs every MCP tool request, the exact schema injected into the prompt context, and the final payload executed by the AI, providing a verifiable audit log for hallucinations. | Unassigned | 💡 Planned | `(Pending)` |
| **P18-T04** | 🟢 P3 | **Dependency-Aware RAG Bundling** — Automatically bundle upstream and downstream service schemas into an agent's context window when modifying a target microservice, mathematically guaranteeing cross-repo contract safety. | Unassigned | 💡 Planned | `(Pending)` |
| **P18-T99** | 🔴 P1 | **Phase 18 E2E Validation (No Mocks)** — Validate AI governance features end-to-end: Agent Profile restrictions, Schema Pruning accuracy against real NL prompts, Audit Trail completeness logging all MCP tool calls, and RAG Bundle context injection. Must use real LLM completions — no simulated responses. | Jules | 💡 Planned | `docs/specs/phase-18/p18-t99-e2e.md` |

---

## 🧪 E2E Coverage Tracker

> **Purpose:** Track which phases have E2E test coverage. As a solo developer, E2E is the primary safety net against production regressions.
>
> 📊 **Current coverage: ~78% of backend API surface area.**

| Phase | Test File | Coverage | Priority | Status |
|-------|-----------|----------|----------|--------|
| V1 (Phases 1–5 core) | `scripts/e2e/v1_e2e_test.go` | ✅ Webhook, Diff, Deploy Gate, AI Scaffold | — | ✅ Green |
| Phase 3 (Contract Registry) | `scripts/e2e/phase3_e2e_test.go` | ✅ Full | — | ✅ Green |
| Phase 6 (QA Feedback) | `scripts/e2e/phase6_e2e_test.go` | ✅ Full | — | ✅ Green |
| Phase 7 (Enterprise) | `scripts/e2e/phase7_e2e_test.go` | ✅ Audit Mode, CEL Rules, Drift, AI Autofix | — | ✅ Green |
| Phase 8 (Readiness) | `scripts/e2e/phase8_e2e_test.go` | ✅ Job Queue, RBAC(1 route), Rollback, Billing | ⚠️ RBAC breadth | ✅ Green (partial) |
| Phase 9 (DX) | `scripts/e2e/phase9_e2e_test.go` | ✅ Compliance, Quality Gates, Watch Daemon | — | ✅ Green |
| Phase 10 (Ecosystem) | `scripts/e2e/phase10_e2e_test.go` | ✅ OTel, Zombies, Governance, SDK | Sc3 SKIP (Forgejo) | ✅ Green |
| Phase 10+13 (Enterprise) | `scripts/e2e/phase10_13_e2e_test.go` | ✅ Protobuf, FinOps, TreeSitter, Gateway | — | ✅ Green |
| Phase 11 (Graph/SSE) | `scripts/e2e/phase11_e2e_test.go` | ✅ Graph, Impact, SSE, Diff retrieval | — | ✅ Green |
| Phase 12 (SSE/WASM) | `scripts/e2e/phase12_sse_test.go` + `wasm_boundary_test.go` | ✅ SSE Broker, WASM Boundary, Fuzz | — | ✅ Green |
| Phase 13 (Enterprise) | `scripts/e2e/phase10_13_e2e_test.go` | ✅ FinOps, Protobuf, TreeSitter | — | ✅ Green |
| Phase 14 (Predictive) | `scripts/e2e/phase14_e2e_test.go` | ✅ Badge, Archaeology, Negotiation | — | ✅ Green |
| Phase 15 (Ecosystem) | `scripts/e2e/phase15_e2e_test.go` | ✅ NL Governance, Marketplace, Insurance | — | ✅ Green |
| History/Sync/RBAC (Cross-cutting) | `phase_history...`, `phase_sync...`, `phase8...` | ✅ History, Sync, Schema, RBAC breadth | 🔴 P1 | ✅ Green |
| UI — Diff Viewer / Impact / Governance | `diff-viewer`, `impact-page`, `governance` | ✅ `/diff/[id]`, `/impact`, `/governance` | 🔴 P1 | ✅ Green |

---

## 🚨 E2E Gap Tasks (Full Audit — 2026-07-23)

> Full route-by-route audit of all 39 backend routes + 22 frontend pages. Gaps ranked by production risk.

### 🔴 Priority 1 — Backend API Gaps

| Task ID | Name | Routes Covered | Spec | Status |
|---------|------|----------------|------|--------|
| **P1-T09** | Phase 1f/1g E2E Validation | AI/ML and Salesforce schema adapters | `docs/specs/e2e/phase1f-1g-adapters.md` | ✅ Complete |
| **P11-T18** | Phase 11 Backend E2E | `GET /api/v1/graph/{org}`, `GET /api/v1/impact/{org}/{repo}`, `GET /api/v1/events`, `GET /api/v1/diff/{id}` | `docs/specs/phase-11/p11-t18-e2e-validation.md` | ✅ PR Merged |
| **P3-T12** | Phase 3 Registry E2E | `GET /api/v1/graph/{org}`, `GET /api/v1/registry/can-deploy`, `GET /api/v1/repos/{org}` | `docs/specs/phase-3/p3-t12-e2e-validation.md` | ✅ Complete |
| **P-HIST-01** | History & Changes API | `POST /api/v1/history`, `GET /api/v1/history/{org}/{repo}`, `GET /api/v1/changes` | `docs/specs/cross-cutting/p-hist-01-e2e-history-api.md` | ✅ Complete |
| **P-SYNC-01** | Contract Sync Pipeline | `POST /api/v1/sync`, `GET /api/v1/schema/{owner}/{repo}`, `GET /api/v1/spec/{org}/{repo}` | `docs/specs/cross-cutting/p-sync-01-e2e-sync-pipeline.md` | ✅ Complete |
| **P-RBAC-01** | RBAC Breadth | All `authzMW` write routes — rules, insurance, zombies/pr, partners | `docs/specs/cross-cutting/p-rbac-01-e2e-rbac-breadth.md` | ✅ Complete |
| **P-MCP-01** | Full MCP Server Parity | 28 Tools, 8 Resources, 5 Prompts + SSE Transport | `docs/specs/cross-cutting/p-mcp-01-full-mcp-parity.md` | ✅ Complete |
| **P-MCP-02** | MCP HTTP SSE Transport | Implement the missing HTTP Server-Sent Events Transport for MCP. | `docs/specs/mcp/mcp-http-transport.md` | ✅ Complete |

### 🟡 Priority 2 — Backend API Gaps

| Task ID | Name | Routes Covered | Spec | Status |
|---------|------|----------------|------|--------|
| **P-ROI-01** | ROI + FinOps E2E | `GET /api/v1/telemetry/roi/{org}`, `POST /api/v1/finops/predict` | `docs/specs/cross-cutting/p-roi-01-e2e-finops.md` | ✅ Complete |
| **P-INS-01** | Insurance Lifecycle | `GET /api/v1/org/{org}/insurance/policy`, `GET /api/v1/org/{org}/insurance/claims` | `docs/specs/cross-cutting/p-ins-01-e2e-insurance.md` | ✅ Complete |

### 🔴 Priority 1 — Frontend UI Gaps (Playwright)

| Task ID | Page | Risk | Spec | Status |
|---------|------|------|------|--------|
| **P-UI-DIFF** | `/diff/[id]` — Diff Viewer | Core product page — blank on regression | `docs/specs/cross-cutting/p-ui-01-e2e-missing-pages.md` | ✅ Complete |
| **P-UI-IMPACT** | `/org/{org}/repo/[repo]/impact` — Impact Page | Blast radius feature | `docs/specs/cross-cutting/p-ui-02-e2e-impact.md` | ✅ Complete |
| **P-UI-GOV** | `/org/{org}/governance` — Governance Rules | Rule CRUD UI | `docs/specs/cross-cutting/p-ui-03-e2e-gov.md` | ✅ Complete |

### 🟡 Priority 2 — Frontend UI Gaps (Playwright)

| Task ID | Page | Status |
|---------|------|--------|
| **P-UI-ZOMBIE** | `/org/{org}/zombies` | ✅ Complete |
| **P-UI-INS** | `/org/{org}/settings/insurance` | ✅ Complete |
| **P-UI-PREVIEW** | `(public)/preview/[token]` | ✅ Complete |

### 🟢 Priority 3 — CLI Command Gaps

| Task ID | Commands | Status |
|---------|----------|--------|
| **P-CLI-POST** | `substrate postmortem`, `substrate schema-smell`, `substrate plugin publish` | ✅ Complete |
| **P-PART-01** | Partners CRUD (`/api/v1/org/{org}/partners` full lifecycle) | ✅ Complete |

### 🟣 Priority 4 — Frontend Unit Testing Gaps (Vitest)

| Task ID | Component/Page | Status |
|---------|----------------|--------|
| **P-UNIT-01** | Backfill Svelte unit tests for QA Dashboard and API Keys pages (19 missing files) | 💡 Planned |

---

### E2E Quick Commands
```bash
# Run all E2E tests (requires API + Postgres running)
export GITHUB_TOKEN=mock_token && cd scripts/e2e && go test -v -p 1 -run "." ./...

# Run a specific phase
cd scripts/e2e && go test -v -p 1 -run TestPhase11SystemE2E ./...
cd scripts/e2e && go test -v -p 1 -run TestPhase3ContractRegistry ./...

# Run Playwright UI tests
cd dashboard && npx playwright test

# Start API server
cd api && DATABASE_URL="postgresql://postgres:postgres@localhost:5432/substrate?sslmode=disable" \
  REGISTRY_API_TOKEN="local-dev-token" INTERNAL_SERVICE_TOKEN="local-dev-token" \
  JWT_SECRET="local-jwt-secret" GITHUB_CLIENT_ID="mock-client-id" \
  GITHUB_CLIENT_SECRET="mock-client-secret" DASHBOARD_URL="http://localhost:5173" \
  ENVIRONMENT="development" go run ./cmd/server/main.go
```

---

## 🧪 Final Manual QA Checklist

Once all automated E2E tests are complete, these four manual workflows must be verified to clear Substrate for production:

- `[ ]` **1. GitHub App (Real-World Webhooks)**
  - Install the Substrate GitHub App on a live test repository.
  - Push a breaking OpenAPI change.
  - Verify the Substrate Bot successfully posts a Markdown comment on the PR containing the correct blast radius formatting.
- `[ ]` **2. VS Code Extension & UI Feel**
  - Verify the Cytoscape dependency graph renders correctly *inside* the VS Code extension webview panel.
  - Open the SvelteKit Dashboard in a browser and load the 1,000-node scale test to ensure the UI remains snappy and responsive.
- `[x]` **3. MCP Server in Claude/Cursor**
  - Add the `substrate-mcp` server to your `claude_desktop_config.json`.
  - Ask Claude an impact analysis question (e.g., *"What happens if I delete the email field from the Payments API?"*).
  - Verify Claude correctly invokes the `get_blast_radius` tool and hallucinates nothing.
- `[x]` **4. Zero-Config Scaffold (`substrate init`)**
  - Run `substrate init` in a fresh, empty directory.
  - Verify the generated `substrate.yaml` and `.github/workflows/substrate.yml` are perfectly formatted and intuitive.
