# Substrate — Task Tracker

> Last updated: 2026-07-12 (Phase 6 E2E merged ✅. Phase 6 complete. Next: V1.0 Pre-flight.)
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
| **P7-T00** | 🔴 P1 | **Zero-Config Org Rollout** — Run engine without `substrate.yaml`, read global `substrate-org.yaml` from `.github` repo, and globally enforce via Dashboard. | Jules | 🔄 In Progress (`17054865686922683816`) | `docs/specs/phase-7/zero-config-org-rollout.md` |
| **P7-T01** | 🔴 P1 | **Enterprise Webhook & Event Egress** — Emit a standardized JSON event whenever a contract is broken to trigger enterprise ITSM workflows (ServiceNow, AWS EventBridge, etc.). | Jules | 🔄 In Progress (`8339582148664302277`) | `docs/specs/phase-7/p7-t01-webhook-egress.md` |
| **P7-T02** | 🟡 P2 | **Targeted Notifications (Slack/Teams)** — Notify specific CODEOWNERS in Slack/Teams when their downstream consumer repo is broken by an upstream change. | Unassigned | 🔒 Blocked (Waiting on T01) | `docs/specs/phase-7/p7-t02-targeted-notifications.md` |
| **P7-T03** | 🟢 P3 | **Custom Rules Engine (CEL/OPA)** — Let enterprises define custom schema rules (e.g., "All APIs must have an X-Correlation-ID header") in `substrate.yaml`. | Jules | ✅ Complete | `docs/specs/phase-7/p7-t03-custom-rules-engine.md` |
| **P7-T04** | 🔵 P4 | **Cross-Repo Auto-Fix PRs** — Use an LLM to automatically generate a draft PR in the downstream consumer repo to fix the breaking dependency. | Jules | ✅ Complete | `docs/specs/phase-7/p7-t04-cross-repo-autofix.md` |
| **P7-T05** | ⚪ P5 | **Runtime Drift Detection (eBPF/Envoy)** — Deploy a sidecar to sample 1% of live API traffic and compare it against the Substrate registry to detect un-documented payloads. | Jules | ✅ Complete | `docs/specs/phase-7/p7-t05-runtime-drift-detection.md` |
| **P7-T06** | 🔴 P1 | **Audit Mode (Shadow Mode)** — Allow enterprises to deploy Substrate without blocking PRs. Substrate comments on PRs and logs cross-repo breaks to the dashboard, providing proof of ROI before switching to blocking mode. | Jules | ✅ Complete | `docs/specs/phase-7/audit-mode-rollout.md` |
| **P7-T07** | 🔴 P1 | **Phase 7 E2E Testing** — Strictly "No Mocks" E2E tests for the enterprise features running against real DB and Go API instances. | Jules | 🔄 In Progress (`3797108567620671748`) | `docs/specs/phase-7/e2e-spec.md` |

---

## 📈 Phase 8: Enterprise Readiness & Scale (The V1.0 Moat)

**Goal:** Provide the critical infrastructure, RBAC, observability, and high-ROI integrations necessary for massive enterprise adoption.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P8-T01** | 🔴 P1 | **Adaptive Job Queue (Graceful Fallback)** — Abstract job processing. If a Redis connection string is present, use a distributed Redis queue (Asynq/KeyDB). If not, gracefully degrade to an in-memory Go channel queue. Prevents webhook timeouts without forcing extra infrastructure. | Unassigned | 💡 Backlog | `(Pending)` |
| **P8-T02** | 🔴 P1 | **Enterprise Authz (Casbin/OpenFGA)** — Implement strict Role-Based Access Control (RBAC) engine for dashboard and API permissions. | Unassigned | 💡 Backlog | `(Pending)` |
| **P8-T03** | 🟡 P2 | **CI/CD Cascading Rollback Gate** — `check-rollback` CLI command to block a provider from rolling back in production if a consumer has already deployed code requiring the newer schema. | Unassigned | 💡 Backlog | `(Pending)` |
| **P8-T04** | 🟢 P3 | **Spotify Backstage Plugin** — Pipe the dependency graph, schema health scores, and API docs directly into Backstage.io developer portals. | Unassigned | 💡 Backlog | `(Pending)` |
| **P8-T05** | 🔵 P4 | **Distributed Tracing (OpenTelemetry)** — Add OpenTelemetry to trace requests across the GitHub Worker, Go API, and diff engine for waterfall debugging. | Unassigned | 💡 Backlog | `(Pending)` |
| **P8-T06** | ⚪ P5 | **Management ROI Dashboard** — A specialized view that calculates the literal engineering hours and monetary value saved by Substrate preventing downstream outages this month. | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |

---

## 🔐 Phase 9: Compliance, IDEs & Developer Experience

**Goal:** Provide compliance auditing, IDE-level developer experience, and governance mapping.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P9-T01** | 🔴 P1 | **Shift-Left IDE Plugins** — VSCode/IntelliJ extensions powered by the MCP server to underline breaking changes as the developer types. | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |
| **P9-T02** | 🟡 P2 | **Continuous AI Sync (`watch`)** — Background daemon that monitors code changes in the IDE and updates the local OpenAPI spec in real-time. | Unassigned | 💡 Backlog | `docs/specs/go-to-market-strategy.md` |
| **P9-T03** | 🟢 P3 | **Compliance Mapping** — Auto-tag schemas with SOC2/GDPR/HIPAA warnings when fields like `ssn` or `medical_history` are detected. | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |
| **P9-T04** | 🔵 P4 | **Quality Gates (SonarQube-style)** — Allow setting different failure thresholds based on service tier (e.g., Tier 1 allows 0 warnings, Beta allows breakages). | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |
| **P9-T05** | ⚪ P5 | **Hexagonal Architecture & `sqlc` Refactor** — Formally isolate engines from HTTP transports, migrate raw `pgx` queries to `sqlc` for type-safe DB layer generation. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T06** | ⚪ P6 | **Configuration Management (Viper)** — Migrate `os.Getenv` calls to Viper for robust `.env`, CLI flag, and YAML configuration loading. | Unassigned | 💡 Backlog | `(Pending)` |

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
