# Substrate — Task Tracker

> Last updated: 2026-07-15 (Phase 11 🎨 Core UI/UX Overhaul tasks T10, T11, T14, T15 merged ✅)
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
| **P9-T15** | 🚀 P1 | **Public Impact API & MCP Server** — Expose a REST API (`/api/v1/impact`) and an MCP tool (`get_blast_radius`) allowing CI/CD pipelines to block merges based on downstream risk scores, and enabling AI agents (Cursor/Claude) to autonomously fix downstream breaking changes. | Jules | ✅ Complete | `docs/specs/phase-9/p9-t15-impact-api.md` |

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
| **P11-T01** | 🟡 P2 | **Cascading Blast Radius (Nth-Degree)** — Add an "Impact Depth" slider to the focus mode to animate and reveal 2nd and 3rd-degree downstream consumers to track rippling failures. | Jules | ✅ Complete | `docs/specs/phase-11/p11-t01-blast-radius.md` |
| **P11-T02** | 🟡 P2 | **Team Neighborhoods (Compound Nodes)** — Group repository nodes physically inside Cytoscape compound boundary boxes based on their `CODEOWNERS` or organizational team structure. | Jules | ✅ Complete | `docs/specs/phase-11/p11-t02-team-neighborhoods.md` |
| **P11-T03** | 🟢 P3 | **Interactive Edge Tooltips** — Hovering over a dependency edge reveals a tooltip showing exactly which endpoints/contracts are being consumed (e.g., `GET /api/v1/customers`). | Jules | ✅ Complete | `docs/specs/phase-11/p11-t03-edge-tooltips.md` |
| **P11-T04** | 🔵 P4 | **Historical Volatility Heatmap** — Add a toggle to color-code the graph by historical breaking changes (Red = frequent breakers, Blue = stable core services). | Jules | ✅ Complete | `docs/specs/phase-11/p11-t04-volatility-heatmap.md` |
| **P11-T05** | ⚪ P5 | **Bird's Eye Mini-Map** — Introduce a Cytoscape navigator widget in the bottom-left corner for maintaining context when zoomed into a localized blast radius on 100+ repo graphs. | Stitch | ✅ Complete | `(Handled by P11-T10)` |
| **P11-T07** | 🚀 P1 | **Visual API Design Studio** — Drag-and-drop OpenAPI designer built directly into the Substrate UI to empower PMs and Architects to design before coding. | Jules | ✅ Complete | `docs/specs/phase-11/p11-t07-visual-api-studio.md` |
| **P11-T08** | 🤯 P1 | **Substrate WASM Engine (In-Browser Diffing)** — Compile the Go Diff Engine to WebAssembly (`GOOS=js GOARCH=wasm`) so users can test schema breakages instantly in the Svelte dashboard with zero backend latency. | Stitch | ✅ Complete | `docs/specs/phase-11/p11-t08-wasm-engine.md` |
| **P11-T09** | 🚀 P1 | **Server-Sent Events (SSE) Real-Time UI** — Stream real-time cross-repo diff results from the River queue directly to the dashboard via SSE and Go channels, eliminating UI polling. | Stitch | ✅ Complete | `docs/specs/phase-11/p11-t09-sse-ui.md` |
| **P11-T10** | 🔴 P1 | **Svelte Flow Migration** — Migrate the entire graph visualization from Cytoscape.js to `Svelte Flow` for native reactivity, gorgeous HTML custom nodes, and built-in minimap support. | Stitch | ✅ Complete | `docs/specs/phase-11/p11-t10-svelte-flow.md` |
| **P11-T11** | 🔴 P1 | **Global Command Palette (Cmd+K)** — Implement a Raycast-style command palette for instant global search across 100+ microservices, allowing users to jump directly to specific nodes. | Stitch | ✅ Complete | `docs/specs/phase-11/p11-t11-command-palette.md` |
| **P11-T12** | 🟡 P2 | **Rich Side-by-Side Diff Viewer (& Sign Out)** — Build an interactive, syntax-highlighted side-by-side YAML diff viewer in the dashboard to review exact line deletions. Also includes adding the global Sign Out button. | Jules | ✅ Complete | `docs/specs/phase-11/p11-t12-diff-viewer.md` |
| **P11-T13** | 🟢 P3 | **Time-Travel Graph Replay** — Add a timeline scrubber to the bottom of the graph to view the architectural dependencies of the enterprise exactly as they existed on any historical date. | Jules | ✅ Complete | `docs/specs/phase-11/p11-t13-time-travel.md` |
| **P11-T14** | 🔴 P1 | **Premium Aesthetics System** — Overhaul the UI with a strict focus on Enterprise SaaS aesthetics: Deep Dark Mode, Glassmorphism modals, curated HSL color palettes, and edge-flow micro-animations. | Stitch | ✅ Complete | `docs/specs/phase-11/p11-t14-premium-aesthetics.md` |
| **P11-T15** | 🔴 P1 | **Zero-to-One Onboarding Wizard** — Create a frictionless, animated onboarding wizard (Connect GitHub → Scan Repos → Build Graph) to guarantee a flawless 5-minute enterprise onboarding experience. | Stitch | ✅ Complete | `docs/specs/phase-11/p11-t15-onboarding-wizard.md` |
| **P11-T16** | 🔴 P1 | **Taxonomy & Metadata Tagging** — Extend `substrate.yaml` with a `metadata` block (type, team, databases). Update the database with a JSONB column, pass it via the API, and render visually distinct SVG icons (frontend, database, mobile) in the Svelte Flow graph. | Jules | ✅ Complete | `docs/specs/phase-11/p11-t16-taxonomy-metadata.md` |
| **P11-T17** | 🟢 P3 | **Graph Image Export** — Allow users to export their current Svelte Flow dependency graph as a high-resolution PNG for use in RFCs and compliance audits, using `html-to-image`. | Jules | ✅ Complete | `docs/specs/phase-11/p11-t17-graph-export.md` |

### Phase 12: V2.0 Public Launch & Quality Assurance
| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P12-T01** | 🔴 P1 | **Zero-to-One Onboarding E2E** — Automate the user journey from GitHub token entry to Graph rendering using Playwright. | Stitch | ✅ Complete | `docs/specs/phase-12/p12-t01-onboarding-e2e.md` |
| **P12-T02** | 🔴 P1 | **Svelte Flow Interaction E2E** — Playwright tests for Canvas interactions: clicking nodes (blast radius), toggling heatmaps, hovering edges. | Stitch | ✅ Complete | `docs/specs/phase-12/p12-t02-svelteflow-interactions.md` |
| **P12-T03** | 🟡 P2 | **Visual Studio Resilience Test** — Verify bidirectional YAML binding and ensure invalid YAML does not crash Svelte state. | Stitch | ✅ Complete | `docs/specs/phase-12/p12-t03-playground-e2e.md` |
| **P12-T04** | 🔴 P1 | **SSE Connection Resilience Test** — Go tests to simulate 100+ concurrent clients and abrupt disconnects to prevent memory leaks in the broadcast registry. | Jules | ✅ Complete | `docs/specs/phase-12/p12-t04-sse-resilience.md` |
| **P12-T05** | 🟡 P2 | **WASM Engine Boundary Tests** — Ensure passing massive/malformed payloads to the WebAssembly diff engine returns safe JS errors instead of panics. | Jules | ✅ Complete | `docs/specs/phase-12/p12-t05-wasm-boundary.md` |
| **P12-T06** | 🟢 P3 | **1,000-Node UI Stress Test** — Generate a massive mock graph to ensure Dagre layouts under 2s and canvas renders at 60fps. | Jules | ✅ Complete | `docs/specs/phase-12/p12-t06-stress-test.md` |
| **P12-T07** | 🔵 P4 | **Telemetry & Crash Reporting** — Integrate PostHog/Sentry to trace live production errors in the Svelte Flow canvas. | Stitch | ✅ Complete | `docs/specs/phase-12/p12-t07-telemetry.md` |
| **P12-T08** | 🔴 P1 | **V2.0 Production Cutover** — Final pipeline updates to bundle Svelte static assets and WASM binary into the Go single-binary deployment. | Jules | 🔄 Running (`7567015248302150297`) | `docs/specs/phase-12/p12-t08-production-cutover.md` |
| **P12-T09** | 🟢 P3 | **AI Support Copilot** — A floating AI chat widget in the dashboard that uses our existing Phase 4 Intelligence Layer to answer questions, generate `substrate.yaml` configs, and troubleshoot user graphs in real-time. | Unassigned | 💡 Backlog | `docs/specs/phase-12/p12-t09-ai-copilot.md` |
