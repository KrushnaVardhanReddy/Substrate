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
| **UI-T01** | 🔴 P1 | **Dynamic Cytoscape Rendering** — Integrate `cytoscape.js` into the Svelte frontend to dynamically render the 100+ dependency nodes and edges returned by the `GET /api/v1/graph` endpoint. | Unassigned | 💡 Backlog | `docs/specs/ui/ui-t01-dynamic-graph.md` |

---

## 🚀 V1.0 Pre-Flight Checklist (Prod Launch)

**Goal:** Finalize the developer experience, onboarding friction, and legal requirements before pushing Substrate to the GitHub Marketplace.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|

---

## 🏢 Phase 7: Enterprise Integrations & ITSM (Post-V1.0)

**Goal:** Integrate Substrate deeply into corporate workflows, providing custom governance, automated ticketing, and targeted notifications.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P7-T00** | 🔴 P1 | **Zero-Config Org Rollout** — Run engine without `substrate.yaml`, read global `substrate-org.yaml` from `.github` repo, and globally enforce via Dashboard. | Unassigned | 💡 Backlog | `docs/specs/phase-7/zero-config-org-rollout.md` |
| **P7-T01** | 🔴 P1 | **ServiceNow/Jira Dynamic CAB** — Auto-create ITSM tickets for breaking changes and assign specific downstream Tech Leads as approvers. | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |
| **P7-T02** | 🟡 P2 | **Targeted Notifications (Slack/Teams)** — Notify specific CODEOWNERS in Slack/Teams when their downstream consumer repo is broken by an upstream change. | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |
| **P7-T03** | 🟢 P3 | **Custom Rules Engine (CEL/OPA)** — Let enterprises define custom schema rules (e.g., "All APIs must have an X-Correlation-ID header") in `substrate.yaml`. | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |
| **P7-T04** | 🔵 P4 | **Cross-Repo Auto-Fix PRs** — Use an LLM to automatically generate a draft PR in the downstream consumer repo to fix the breaking dependency. | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |
| **P7-T05** | ⚪ P5 | **Runtime Drift Detection (eBPF/Envoy)** — Deploy a sidecar to sample 1% of live API traffic and compare it against the Substrate registry to detect un-documented payloads. | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |

---

## 📈 Phase 8: Compliance, IDEs & Analytics (The Enterprise Moat)

**Goal:** Provide compliance auditing, IDE-level developer experience, and management-level reporting to justify enterprise adoption.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P8-T01** | 🔴 P1 | **Shift-Left IDE Plugins** — VSCode/IntelliJ extensions powered by the MCP server to underline breaking changes as the developer types. | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |
| **P8-T02** | 🟡 P2 | **Continuous AI Sync (`watch`)** — Background daemon that monitors code changes in the IDE and updates the local OpenAPI spec in real-time. | Unassigned | 💡 Backlog | `docs/specs/go-to-market-strategy.md` |
| **P8-T03** | 🟢 P3 | **Compliance Mapping** — Auto-tag schemas with SOC2/GDPR/HIPAA warnings when fields like `ssn` or `medical_history` are detected. | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |
| **P8-T04** | 🔵 P4 | **Quality Gates (SonarQube-style)** — Allow setting different failure thresholds based on service tier (e.g., Tier 1 allows 0 warnings, Beta allows breakages). | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |
| **P8-T05** | ⚪ P5 | **Management ROI Dashboard** — A dashboard view that calculates the literal hours and money saved by preventing outages this month. | Unassigned | 💡 Backlog | `docs/specs/enterprise-vision.md` |

---

## 🔐 Phase 9: Ecosystem Expansion & Security (Post-V1.0)

**Goal:** Expand Substrate's reach into developer portals, API gateways, and automated security testing.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P9-T01** | 🔴 P1 | **Automated Security Fuzzing (OWASP)** — Upgrade the Phase 6 fuzzer to inject malicious payloads (SQLi, IDOR) based on the schema, acting as an automated pentester. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T02** | 🟡 P2 | **Spotify Backstage Plugin** — Pipe the dependency graph, schema health scores, and API docs directly into Backstage.io developer portals. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T03** | 🟢 P3 | **API Gateway Auto-Sync** — Automatically push validated OpenAPI schemas to AWS API Gateway, Kong, or Cloudflare API Shield on merge to `main`. | Unassigned | 💡 Backlog | `(Pending)` |
| **P9-T04** | 🔵 P4 | **Hexagonal Architecture Refactor** — Formally isolate the core Discovery/Diff engines from HTTP transports and SQL storage to allow pluggable graph databases (Neo4j) and robust in-memory unit testing. | Unassigned | 💡 Backlog | `(Pending)` |

---

## 🚀 Phase 10: The Ultimate Enterprise SDLC (Post-V1.0 Vision)

**Goal:** Extend Substrate beyond schema safety into auto-remediation, code generation, and runtime API governance.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P10-T01** | 🔴 P1 | **CI/CD Cascading Rollback Gate** — `substrate check-rollback` CLI command to block a provider from rolling back in production if a consumer has already deployed code requiring the newer schema. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T02** | 🟡 P2 | **AI Mock Data Generator (QA)** — Scan QA repositories for JSON test fixtures and use the AI engine to auto-update mock data when the upstream API schema changes. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T03** | 🟢 P3 | **AI Spectral Linter (API Governance)** — Enforce plain-English API design rules (e.g. "All endpoints must use camelCase") during the PR diff process to maintain org-wide consistency. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T04** | 🔵 P4 | **Auto-SDK Generator PRs** — Automatically generate TypeScript/Swift/Go clients via OpenAPI Generator when a schema is merged, opening PRs directly in the downstream consumer repositories. | Unassigned | 💡 Backlog | `(Pending)` |
| **P10-T05** | ⚪ P5 | **Traffic-Aware Pruning (Zombies)** — Correlate schema endpoints with live Datadog/OTel metrics to detect unused "zombie" APIs and auto-generate PRs to delete the dead code. | Unassigned | 💡 Backlog | `(Pending)` |

---

## How to Submit Jules Tasks

```bash
# From repo root
python3 scripts/jules_submit.py --list
python3 scripts/jules_submit.py --task 1
python3 scripts/jules_submit.py --task 1 --branch feat/diff-engine
```
