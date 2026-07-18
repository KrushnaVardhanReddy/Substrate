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
| **P9-T00** | 🚀 P1 | **GitHub Marketplace Launch (WASI)** — Build the `substrate-action` wrapper using `GOOS=wasip1 GOARCH=wasm`. Execute via Node/Wasmtime for sub-second, highly secure, Docker-less CI/CD diffing, and publish to the Marketplace. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T01** | 🔴 P1 | **Shift-Left IDE Plugins** — VSCode/IntelliJ extensions powered by the MCP server to underline breaking changes as the developer types. | Jules | ⏳ Ready | `docs/specs/phase-9/p9-t01-ide-plugins.md` |
| **P9-T02** | 🟡 P2 | **Continuous AI Sync (`watch`)** — Background daemon that monitors code changes in the IDE and updates the local OpenAPI spec in real-time. | Unassigned | 💡 Backlog | `docs/specs/go-to-market-strategy.md` |
| **P9-T03** | 🟢 P3 | **Compliance Mapping** — Auto-tag schemas with SOC2/GDPR/HIPAA warnings when fields like `ssn` or `medical_history` are detected. | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |
| **P9-T04** | 🔵 P4 | **Quality Gates (SonarQube-style)** — Allow setting different failure thresholds based on service tier (e.g., Tier 1 allows 0 warnings, Beta allows breakages). | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |
| **P9-T05** | ⚪ P5 | **Hexagonal Architecture & `sqlc` Refactor** — Formally isolate engines from HTTP transports, migrate raw `pgx` queries to `sqlc` for type-safe DB layer generation. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T06** | ⚪ P6 | **Configuration Management (Viper)** — Migrate `os.Getenv` calls to Viper for robust `.env`, CLI flag, and YAML configuration loading. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T07** | 🔴 P1 | **AI Migration Planner** — Upgrade AI Autofix to generate safe, multi-step migration plans for complex schema/database changes with minimal downtime. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T08** | 🟡 P2 | **Deployment Risk Scoring** — Synthesize breaking change data, infrastructure changes, and downstream blast radius into a holistic "Deployment Risk Score". | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T09** | 🟢 P3 | **AI Impact Analysis Summaries** — Pass cross-repo blast radius checks to the AI handler to generate a plain-English impact summary on PRs. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T11** | 🚀 P1 | **Automated Deprecation Campaigns** — Track sunsetting endpoints, auto-open issues in downstream consumer repos, and nag them until 0% usage is reached. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T12** | 🚀 P1 | **Embedded SQLite (LibSQL) Local Caching** — Embed SQLite directly into the CLI and MCP Server to pull background graph updates, enabling sub-millisecond, zero-latency local schema diffs. | Jules | 🔄 Running (`10646242211838366953`) | `docs/specs/phase-9/p9-t12-embedded-sqlite.md` |
| **P9-T13** | 🟡 P2 | **Embedded Mermaid Blast Radius** — Upgrade the GitHub PR comment bot to render a visual Mermaid.js flowchart of the exact blast radius directly inside the PR, eliminating the need to click away. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T14** | 🟢 P3 | **Ephemeral API Preview URLs** — Generate temporary, shareable Substrate dashboard URLs for PRs so engineers can share proposed schema changes and interactive diffs with frontend teams before merging. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T16** | 🚀 P1 | **WASM Git Pre-Commit Hooks** — Blazing fast local Git hooks that run `substrate diff` in 0.02s before code ever leaves the developer's laptop. | Unassigned | 💡 Backlog | `(Pending)` |

---

## 🚀 Phase 10: Ecosystem Expansion & Security (Post-V1.0)

**Goal:** Expand Substrate's reach into API gateways and automated security testing.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P10-T01** | 🔴 P1 | **Automated Security Fuzzing (OWASP)** — Upgrade the Phase 6 fuzzer to inject malicious payloads (SQLi, IDOR) based on the schema, acting as an automated pentester. | Jules | 🔒 Blocked (AI Safety Refusal) | `docs/specs/phase-10/p10-t01-security-fuzzing.md` |
| **P10-T02** | 🟡 P2 | **API Gateway Auto-Sync** — Automatically push validated OpenAPI schemas to AWS API Gateway, Kong, or Cloudflare API Shield on merge to `main`. | Jules | ⏳ Ready | `docs/specs/phase-10/p10-t02-api-gateway.md` |
| **P10-T03** | 🟢 P3 | **AI Mock Data Generator (QA)** — Scan QA repositories for JSON test fixtures and use the AI engine to auto-update mock data when the upstream API schema changes. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T04** | 🔵 P4 | **AI Spectral Linter (API Governance)** — Enforce plain-English API design rules (e.g. "All endpoints must use camelCase") during the PR diff process to maintain org-wide consistency. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T05** | ⚪ P5 | **Auto-SDK Generator PRs** — Automatically generate TypeScript/Swift/Go clients via OpenAPI Generator when a schema is merged, opening PRs directly in the downstream consumer repos. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T06** | ⚪ P6 | **Traffic-Aware Pruning (Zombies)** — Correlate schema endpoints with live Datadog/OTel metrics to detect unused "zombie" APIs and auto-generate PRs to delete the dead code. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T07** | 🟣 P1 | **MCP Runtime Diffing** — Spin up Model Context Protocol (MCP) servers in a sandbox during CI/CD to dynamically diff `tools/list` and block AI agent breaking changes. | Unassigned | 💡 Backlog | `docs/specs/phase-10/p10-t06-mcp-diffing.md` |
| **P10-T09** | 🚀 P1 | **Protobuf & gRPC Schema Registry** — Add native support for parsing, diffing, and visualizing Protobufs to capture backend-to-backend enterprise microservices. | Jules | ⏳ Ready | `docs/specs/phase-10/p10-t09-protobuf-grpc.md` |
| **P10-T10** | 🚀 P1 | **Consumer-Driven Contract Manifests** — Allow frontend apps to upload `.substrate-consumer.yaml` declaring required fields, directly competing with PactFlow. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T12** | 🚀 P1 | **Tree-sitter Deterministic Impact Analysis** — Parse downstream consumer repositories using Tree-sitter AST to deterministically pinpoint exactly *which lines of code* are broken by an upstream API change. | Jules | 🔄 Running (`7482648289839328684`) | `docs/specs/phase-10/p10-t12-treesitter-analysis.md` |
| **P10-T13** | 🔴 P1 | **Integrated API Documentation Catalog** — Evolve the registry into an internal Developer Portal by embedding interactive API reference viewers (like Stoplight Elements or ReDoc) directly into the dashboard. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T14** | 🚀 P1 | **Turing-Complete Governance Rules (WASM)** — Support WebAssembly (Wazero) plugins for custom enterprise governance rules written in Rust/Go/TS. *UX Goal: Provide a simple `substrate plugin create` CLI and integrate via `wasm_plugin: "./policy.wasm"` in `substrate.yaml`.* | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T15** | 🚀 P1 | **Zero-Latency Drift Detection (eBPF)** — Extend runtime drift detection with a zero-latency `cilium/ebpf` kernel probe for high-throughput environments. *UX Goal: Provide a pre-packaged Helm chart (`helm install substrate-ebpf`) that auto-detects pods via Kubernetes labels (e.g., `substrate.io/monitor: "true"`).* | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T18** | 🚀 P1 | **GraphQL Supergraph Federation** — Add native support for Apollo Federation to diff subgraphs and prevent routing breakages. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T19** | 🚀 P1 | **CRM/Billing Blast Radius** — Integrate with Stripe and Salesforce to map external customer impact on internal API breakages. | Unassigned | 💡 Backlog | `(Pending)` |

---

## 🗺️ Phase 11: Advanced Graph Visualization (V2.0 UX)

**Goal:** Elevate the Substrate Dependency Graph into a world-class architectural explorer with cascading impact analysis and team-based layouts.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|

### Phase 12: V2.0 Public Launch & Quality Assurance
| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P12-T08** | 🔴 P1 | **V2.0 Production Cutover** — Final pipeline updates to bundle Svelte static assets and WASM binary into the Go single-binary deployment. | Jules | 🔄 Running (`7567015248302150297`) | `docs/specs/phase-12/p12-t08-production-cutover.md` |
| **P12-T09** | 🟢 P3 | **AI Support Copilot** — A floating AI chat widget in the dashboard that uses our existing Phase 4 Intelligence Layer to answer questions, generate `substrate.yaml` configs, and troubleshoot user graphs in real-time. | Jules | 🔄 Running (`14882584206088153250`) | `docs/specs/phase-12/p12-t09-ai-copilot.md` |

---

## 👑 Phase 13: God-Mode & Enterprise Intelligence

**Goal:** Evolve Substrate into a predictive, financial, and auto-healing infrastructure intelligence platform.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P13-T01** | 🤯 P1 | **FinOps Cost Prediction** — Connect schema diffs to Datadog traffic to calculate the exact USD egress cost increase of payload size changes. | Jules | 🔄 Running (`11581841415853721360`) | `docs/specs/phase-13/p13-t01-finops.md` |
| **P13-T02** | 🤯 P1 | **DB Performance Breakages** — Dry-run Prisma/PlanetScale migrations to predict table-locks and performance outages before they merge. | Jules | ⏳ Ready | `docs/specs/phase-13/p13-t02-db-performance.md` |
