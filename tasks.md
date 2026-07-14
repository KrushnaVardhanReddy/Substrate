# Substrate — Task Tracker

> Last updated: 2026-07-14 (Phase 11 🎨 UI/UX Overhaul in progress. 4 async Jules agents triggered.)
> Tracking all development phases, tasks, and their current status.

---

## Legend

| Symbol | Status |
|---|---|
| ✅ | Complete |
| 🔄 | In Progress |
| ⏳ | Ready to Start (all dependencies met) |
| 🔒 | Blocked (waiting on dependency) |
| 💡 | Planned (not yet started) |

---


> **Note:** Tasks for Phases 0 through 6 have been archived to [completed_tasks.md](./completed_tasks.md)

## 🏗️ Phase 5 Backlog: Stress Testing

**Goal:** Ensure the discovery algorithms and dashboard visualization can scale to enterprise levels (100+ repos).

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|



## 🎨 Phase 6.5: Dynamic UI & Graph Visualization

**Goal:** Transform the Svelte Dashboard from a hardcoded mock into a fully dynamic, interactive dependency map powered by the backend Registry API.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **UI-T01** | 🔴 P1 | **Dynamic Cytoscape Rendering** — Integrate `cytoscape.js` into the Svelte frontend to dynamically render the 100+ dependency nodes and edges returned by the `GET /api/v1/graph` endpoint. | Jules | ✅ Complete | `docs/specs/ui/ui-t01-dynamic-graph.md` |
| **E2E-T01** | 🔴 P1 | **1-Hour Chaos Endurance & UI Polling** — Upgrade `scale_generator.go` to a continuous 1-hour loop and update the Svelte UI to poll and animate the graph changes in real-time. | Jules | ✅ Complete | `docs/specs/e2e/e2e-t01-endurance-mode.md` |
| **UI-T03** | 🟡 P2 | **Graph Filtering & Navigation** — Add a status filter (Show only BREAKING) and protocol filter to the Cytoscape visualization to handle enterprise-scale graphs. | Jules | ✅ Complete | `docs/specs/ui/ui-t03-graph-filtering.md` |
| **UI-T04** | 🟢 P3 | **Enterprise Graph UX Overhaul** — Semantic node coloring, orphan node hiding, and a dedicated Blast Radius Modal for isolating impact analysis. | Antigravity | ✅ Complete | `docs/specs/ui/ui-t04-enterprise-graph-ux.md` |

---

## 🚀 V1.0 Pre-Flight Checklist (Prod Launch)

**Goal:** Finalize the developer experience, onboarding friction, and legal requirements before pushing Substrate to the GitHub Marketplace.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **V1-T07** | 🔴 P1 | **V1.0 System E2E Tests** — Prove the final V1.0 pipeline works as a seamless end-to-end flow using the real local PostgreSQL database. | Jules | ✅ Complete | `docs/specs/v1-preflight/v1-e2e-spec.md` |
| **V1-T08** | 🔴 P1 | **Custom Discovery Rules via YAML** — Expose dependency matching rules via `substrate.yaml` to allow enterprises to define custom regex (e.g., `_ENDPOINT`) for the Kubernetes manifest scanner. | Jules | ✅ Complete | `docs/specs/v1-preflight/v1-t08-custom-discovery-rules.md` |
| **V1-T09** | 🔴 P1 | **True Dogfooding via OpenAPI** — Formalize the internal Substrate API into a valid `openapi.yaml` specification so the engine can monitor and block its own breaking changes. | Antigravity | ✅ Complete | `docs/specs/v1-preflight/v1-t09-dogfooding-openapi.md` |
| **V1-T10** | 🔴 P1 | **GitHub App First-Time Setup** — Ensure the GitHub App detects when a user is uploading a schema for the first time, avoids a baseline fetch error, and returns a friendly "Welcome to Substrate" PR status. | Antigravity | ✅ Complete | `docs/specs/v1-preflight/v1-t10-github-app-first-time-setup.md` |

---

## 🏢 Phase 7: Enterprise Integrations & ITSM (Post-V1.0)

**Goal:** Integrate Substrate deeply into corporate workflows, providing custom governance, automated ticketing, and targeted notifications.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P7-T00** | 🔴 P1 | **Zero-Config Org Rollout** — Run engine without `substrate.yaml`, read global `substrate-org.yaml` from `.github` repo, and globally enforce via Dashboard. | Jules | ✅ Complete | `docs/specs/phase-7/zero-config-org-rollout.md` |
| **P7-T01** | 🔴 P1 | **Enterprise Webhook Egress** — Emit a generic, standardized JSON webhook whenever a contract is broken, allowing enterprises to pipe alerts into their own ITSM systems (ServiceNow/Datadog) rather than building hardcoded integrations. | Jules | ✅ Complete | `docs/specs/phase-7/p7-t01-webhook-egress.md` |
| **P7-T02** | 🟡 P2 | **Targeted Notifications (Slack/Teams)** — *CANCELLED:* Redundant. Enterprises prefer using the Webhook Egress (T01) to pipe alerts into Datadog/PagerDuty rather than rogue Slack apps. | Unassigned | ❌ Cancelled | `(Removed)` |
| **P7-T03** | 🟢 P3 | **Custom Governance Rules (CEL)** — Let enterprises define custom schema diffing rules via easy CEL expressions. | Jules | ✅ Complete | `docs/specs/phase-7/p7-t03-custom-rules-engine.md` |
| **P7-T04** | 🔵 P4 | **Cross-Repo Auto-Fix PRs** — Use an LLM to automatically generate a draft PR in the downstream consumer repo to fix the breaking dependency. | Jules | ✅ Complete | `docs/specs/phase-7/p7-t04-cross-repo-autofix.md` |
| **P7-T05** | ⚪ P5 | **Runtime Drift Detection (Sidecar)** — Deploy a lightweight proxy sidecar to detect un-documented schema payloads. | Jules | ✅ Complete | `docs/specs/phase-7/p7-t05-runtime-drift-detection.md` |
| **P7-T06** | 🔴 P1 | **Audit Mode (Shadow Mode)** — Allow enterprises to deploy Substrate without blocking PRs. Substrate comments on PRs and logs cross-repo breaks to the dashboard, providing proof of ROI before switching to blocking mode. | Jules | ✅ Complete | `docs/specs/phase-7/audit-mode-rollout.md` |
| **P7-T07** | 🔴 P1 | **Phase 7 E2E Testing** — Strictly "No Mocks" E2E tests for the enterprise features running against real DB and Go API instances. | Jules | ✅ Complete | `docs/specs/phase-7/e2e-spec.md` |

---

## 📈 Phase 8: Enterprise Readiness & Scale (The V1.0 Moat)

**Goal:** Provide the critical infrastructure, RBAC, observability, and high-ROI integrations necessary for massive enterprise adoption.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P8-T01** | 🔴 P1 | **PostgreSQL Job Queue (River)** — Offload synchronous diffing, DB inserts, and outbound Webhook Egress delivery (replacing goroutines) to a durable Postgres-backed queue. Prevents timeouts and guarantees event delivery. | Jules | ✅ Complete | `docs/specs/phase-8/p8-t01-postgres-job-queue.md` |
| **P8-T02** | 🔴 P1 | **Enterprise Authz (Casbin/OpenFGA)** — Implement strict Role-Based Access Control (RBAC) engine for dashboard and API permissions. | Jules | ✅ Complete | `docs/specs/phase-8/p8-t02-enterprise-authz.md` |
| **P8-T03** | 🟡 P2 | **CI/CD Cascading Rollback Gate** — `check-rollback` CLI command to block a provider from rolling back in production if a consumer has already deployed code requiring the newer schema. | Jules | ✅ Complete | `docs/specs/phase-8/p8-master-plan.md` |
| **P8-T04** | 🟢 P3 | **Spotify Backstage Plugin** — Pipe the dependency graph, schema health scores, and API docs directly into Backstage.io developer portals. | Jules | ✅ Complete | `docs/specs/phase-8/p8-master-plan.md` |
| **P8-T05** | 🔵 P4 | **Distributed Tracing (OpenTelemetry)** — Add OpenTelemetry to trace requests across the GitHub Worker, Go API, and diff engine for waterfall debugging. | Jules | ✅ Complete | `docs/specs/phase-8/p8-master-plan.md` |
| **P8-T06** | ⚪ P5 | **Management ROI Dashboard** — A specialized view that calculates the literal engineering hours and monetary value saved by Substrate preventing downstream outages this month. | Jules | ✅ Complete | `docs/specs/enterprise-vision.md` |
| **P8-T07** | 🔴 P1 | **Billing & Subscription Engine (Paywall Pause)** — Add `trial_ends_at` to the DB, update the GitHub App to enforce Audit Mode during the 90-day trial, and gracefully pause analysis when the trial expires until Stripe checkout. | Jules | ✅ Complete | `docs/specs/go-to-market-strategy.md` |
| **P8-T08** | 🔴 P1 | **Single Binary VPC Deployment** — Implement `//go:embed` with SvelteKit `adapter-static` to compile the frontend and backend into a single executable for zero-dependency on-prem deployment. | Jules | ✅ Complete | `docs/specs/phase-8/p8-master-plan.md` |
| **P8-T09** | 🔴 P1 | **Docker & Helm Enterprise Delivery** — Create a production-ready `Dockerfile` and Helm chart containing the compiled Go API and Svelte frontend for enterprise Kubernetes clusters. | Jules | ✅ Complete | `docs/specs/phase-8/p8-master-plan.md` |
| **P8-T10** | 🔴 P1 | **Phase 8 E2E Testing** — Test the enterprise readiness components (Job Queue, Authz, Cascading Rollback, Billing) against the real database schema. | Jules | ✅ Complete | `docs/specs/phase-8/p8-t10-e2e-tests.md` |
| **P8-T11** | 🚀 P1 | **Enterprise SSO (SAML/OIDC)** — Integrate Okta, Azure AD, and Google Workspace SSO into the Authz layer so massive enterprises don't have to manage local user accounts. | Unassigned | 💡 Backlog | `(Pending)` |

---

## 🔐 Phase 9: Compliance, IDEs & Developer Experience

**Goal:** Provide compliance auditing, IDE-level developer experience, and governance mapping.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P9-T00** | 🚀 P1 | **GitHub Marketplace Launch (WASI)** — Build the `substrate-action` wrapper using `GOOS=wasip1 GOARCH=wasm`. Execute via Node/Wasmtime for sub-second, highly secure, Docker-less CI/CD diffing, and publish to the Marketplace. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T01** | 🔴 P1 | **Shift-Left IDE Plugins** — VSCode/IntelliJ extensions powered by the MCP server to underline breaking changes as the developer types. | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |
| **P9-T02** | 🟡 P2 | **Continuous AI Sync (`watch`)** — Background daemon that monitors code changes in the IDE and updates the local OpenAPI spec in real-time. | Unassigned | 💡 Backlog | `docs/specs/go-to-market-strategy.md` |
| **P9-T03** | 🟢 P3 | **Compliance Mapping** — Auto-tag schemas with SOC2/GDPR/HIPAA warnings when fields like `ssn` or `medical_history` are detected. | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |
| **P9-T04** | 🔵 P4 | **Quality Gates (SonarQube-style)** — Allow setting different failure thresholds based on service tier (e.g., Tier 1 allows 0 warnings, Beta allows breakages). | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |
| **P9-T05** | ⚪ P5 | **Hexagonal Architecture & `sqlc` Refactor** — Formally isolate engines from HTTP transports, migrate raw `pgx` queries to `sqlc` for type-safe DB layer generation. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T06** | ⚪ P6 | **Configuration Management (Viper)** — Migrate `os.Getenv` calls to Viper for robust `.env`, CLI flag, and YAML configuration loading. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T07** | 🔴 P1 | **AI Migration Planner** — Upgrade AI Autofix to generate safe, multi-step migration plans for complex schema/database changes with minimal downtime. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T08** | 🟡 P2 | **Deployment Risk Scoring** — Synthesize breaking change data, infrastructure changes, and downstream blast radius into a holistic "Deployment Risk Score". | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T09** | 🟢 P3 | **AI Impact Analysis Summaries** — Pass cross-repo blast radius checks to the AI handler to generate a plain-English impact summary on PRs. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T10** | 🚀 P1 | **MCP Server WASI Distribution** — Compile the `substrate-mcp` server using `GOOS=wasip1 GOARCH=wasm` to allow secure, sandboxed execution of the MCP server in Claude Desktop or Cursor via Wasmtime/Node.js. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T11** | 🚀 P1 | **Automated Deprecation Campaigns** — Track sunsetting endpoints, auto-open issues in downstream consumer repos, and nag them until 0% usage is reached. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T12** | 🚀 P1 | **Embedded SQLite (LibSQL) Local Caching** — Embed SQLite directly into the CLI and MCP Server to pull background graph updates, enabling sub-millisecond, zero-latency local schema diffs. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T13** | 🟡 P2 | **Embedded Mermaid Blast Radius** — Upgrade the GitHub PR comment bot to render a visual Mermaid.js flowchart of the exact blast radius directly inside the PR, eliminating the need to click away. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T14** | 🟢 P3 | **Ephemeral API Preview URLs** — Generate temporary, shareable Substrate dashboard URLs for PRs so engineers can share proposed schema changes and interactive diffs with frontend teams before merging. | Unassigned | 💡 Backlog | `(Pending)` |

---

## 🚀 Phase 10: Ecosystem Expansion & Security (Post-V1.0)

**Goal:** Expand Substrate's reach into API gateways and automated security testing.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P10-T01** | 🔴 P1 | **Automated Security Fuzzing (OWASP)** — Upgrade the Phase 6 fuzzer to inject malicious payloads (SQLi, IDOR) based on the schema, acting as an automated pentester. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T02** | 🟡 P2 | **API Gateway Auto-Sync** — Automatically push validated OpenAPI schemas to AWS API Gateway, Kong, or Cloudflare API Shield on merge to `main`. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T03** | 🟢 P3 | **AI Mock Data Generator (QA)** — Scan QA repositories for JSON test fixtures and use the AI engine to auto-update mock data when the upstream API schema changes. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T04** | 🔵 P4 | **AI Spectral Linter (API Governance)** — Enforce plain-English API design rules (e.g. "All endpoints must use camelCase") during the PR diff process to maintain org-wide consistency. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T05** | ⚪ P5 | **Auto-SDK Generator PRs** — Automatically generate TypeScript/Swift/Go clients via OpenAPI Generator when a schema is merged, opening PRs directly in the downstream consumer repos. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T06** | ⚪ P6 | **Traffic-Aware Pruning (Zombies)** — Correlate schema endpoints with live Datadog/OTel metrics to detect unused "zombie" APIs and auto-generate PRs to delete the dead code. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T07** | 🟣 P1 | **MCP Runtime Diffing** — Spin up Model Context Protocol (MCP) servers in a sandbox during CI/CD to dynamically diff `tools/list` and block AI agent breaking changes. | Unassigned | 💡 Backlog | `docs/specs/phase-10/p10-t06-mcp-diffing.md` |
| **P10-T08** | 🔵 P4 | **Live Documentation Generator** — Auto-generate live developer portals and Mermaid architecture diagrams from Substrate's live schema/graph metadata. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T09** | 🚀 P1 | **Protobuf & gRPC Schema Registry** — Add native support for parsing, diffing, and visualizing Protobufs to capture backend-to-backend enterprise microservices. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T10** | 🚀 P1 | **Consumer-Driven Contract Manifests** — Allow frontend apps to upload `.substrate-consumer.yaml` declaring required fields, directly competing with PactFlow. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T11** | 🚀 P1 | **Official Terraform Provider** — Build `terraform-provider-substrate` so DevOps teams can manage webhooks, rules, and RBAC policies entirely via Infrastructure-as-Code. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T12** | 🚀 P1 | **Tree-sitter Deterministic Impact Analysis** — Parse downstream consumer repositories using Tree-sitter AST to deterministically pinpoint exactly *which lines of code* are broken by an upstream API change. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T13** | 🔴 P1 | **Integrated API Documentation Catalog** — Evolve the registry into an internal Developer Portal by embedding interactive API reference viewers (like Stoplight Elements or ReDoc) directly into the dashboard. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T14** | 🚀 P1 | **Turing-Complete Governance Rules (WASM)** — Support WebAssembly (Wazero) plugins for custom enterprise governance rules written in Rust/Go/TS. *UX Goal: Provide a simple `substrate plugin create` CLI and integrate via `wasm_plugin: "./policy.wasm"` in `substrate.yaml`.* | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T15** | 🚀 P1 | **Zero-Latency Drift Detection (eBPF)** — Extend runtime drift detection with a zero-latency `cilium/ebpf` kernel probe for high-throughput environments. *UX Goal: Provide a pre-packaged Helm chart (`helm install substrate-ebpf`) that auto-detects pods via Kubernetes labels (e.g., `substrate.io/monitor: "true"`).* | Unassigned | 💡 Backlog | `(Pending)` |
---

## 🗺️ Phase 11: Advanced Graph Visualization (V2.0 UX)

**Goal:** Elevate the Substrate Dependency Graph into a world-class architectural explorer with cascading impact analysis and team-based layouts.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P11-T01** | 🟡 P2 | **Cascading Blast Radius (Nth-Degree)** — Add an "Impact Depth" slider to the focus mode to animate and reveal 2nd and 3rd-degree downstream consumers to track rippling failures. | Unassigned | 💡 Backlog | `(Pending)` |
| **P11-T02** | 🟡 P2 | **Team Neighborhoods (Compound Nodes)** — Group repository nodes physically inside Cytoscape compound boundary boxes based on their `CODEOWNERS` or organizational team structure. | Unassigned | 💡 Backlog | `(Pending)` |
| **P11-T03** | 🟢 P3 | **Interactive Edge Tooltips** — Hovering over a dependency edge reveals a tooltip showing exactly which endpoints/contracts are being consumed (e.g., `GET /api/v1/customers`). | Unassigned | 💡 Backlog | `(Pending)` |
| **P11-T04** | 🔵 P4 | **Historical Volatility Heatmap** — Add a toggle to color-code the graph by historical breaking changes (Red = frequent breakers, Blue = stable core services). | Unassigned | 💡 Backlog | `(Pending)` |
| **P11-T05** | ⚪ P5 | **Bird's Eye Mini-Map** — Introduce a Cytoscape navigator widget in the bottom-left corner for maintaining context when zoomed into a localized blast radius on 100+ repo graphs. | Unassigned | 💡 Backlog | `(Pending)` |
| **P11-T07** | 🚀 P1 | **Visual API Design Studio** — Drag-and-drop OpenAPI designer built directly into the Substrate UI to empower PMs and Architects to design before coding. | Unassigned | 💡 Backlog | `(Pending)` |
| **P11-T08** | 🤯 P1 | **Substrate WASM Engine (In-Browser Diffing)** — Compile the Go Diff Engine to WebAssembly (`GOOS=js GOARCH=wasm`) so users can test schema breakages instantly in the Svelte dashboard with zero backend latency. | Unassigned | 💡 Backlog | `(Pending)` |
| **P11-T09** | 🚀 P1 | **Server-Sent Events (SSE) Real-Time UI** — Stream real-time cross-repo diff results from the River queue directly to the dashboard via SSE and Go channels, eliminating UI polling. | Unassigned | 💡 Backlog | `(Pending)` |
| **P11-T10** | 🔴 P1 | **Svelte Flow Migration** — Migrate the entire graph visualization from Cytoscape.js to `Svelte Flow` for native reactivity, gorgeous HTML custom nodes, and built-in minimap support. | Stitch | 🔄 In Progress | `docs/specs/phase-11/p11-t10-svelte-flow.md` |
| **P11-T11** | 🔴 P1 | **Global Command Palette (Cmd+K)** — Implement a Raycast-style command palette for instant global search across 100+ microservices, allowing users to jump directly to specific nodes. | Stitch | 🔄 In Progress | `docs/specs/phase-11/p11-t11-command-palette.md` |
| **P11-T12** | 🟡 P2 | **Rich Side-by-Side Diff Viewer** — Build an interactive, syntax-highlighted side-by-side YAML diff viewer in the dashboard to review exact line deletions, mirroring the GitHub PR experience. | Stitch | 💡 Backlog | `(Pending)` |
| **P11-T13** | 🟢 P3 | **Time-Travel Graph Replay** — Add a timeline scrubber to the bottom of the graph to view the architectural dependencies of the enterprise exactly as they existed on any historical date. | Stitch | 💡 Backlog | `(Pending)` |
| **P11-T14** | 🔴 P1 | **Premium Aesthetics System** — Overhaul the UI with a strict focus on Enterprise SaaS aesthetics: Deep Dark Mode, Glassmorphism modals, curated HSL color palettes, and edge-flow micro-animations. | Stitch | 🔄 In Progress | `docs/specs/phase-11/p11-t14-premium-aesthetics.md` |
| **P11-T15** | 🔴 P1 | **Zero-to-One Onboarding Wizard** — Create a frictionless, animated onboarding wizard (Connect GitHub → Scan Repos → Build Graph) to guarantee a flawless 5-minute enterprise onboarding experience. | Stitch | 🔄 In Progress | `docs/specs/phase-11/p11-t15-onboarding-wizard.md` |
