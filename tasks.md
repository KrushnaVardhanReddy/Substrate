# Substrate — Task Tracker

> Last updated: 2026-07-12 (Phase 6 features merged ✅. Phase 6 E2E Tests in progress. Active Jules session running.)
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


## 🚀 V1.0 Pre-Flight Checklist (Prod Launch)

**Goal:** Finalize the developer experience, onboarding friction, and legal requirements before pushing Substrate to the GitHub Marketplace.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **V1-T01** | 🔴 P1 | **GitHub App Auto-Discovery** — Zero-touch onboarding. Auto-scan repos for OpenAPI and open PRs with `substrate.yaml`. | Antigravity | 💡 Backlog | `docs/specs/v1-preflight/v1-spec.md` |
| **V1-T02** | 🟡 P2 | **CLI AI Architect (`init --design`)** — Conversational LLM interface to scaffold an API contract before writing code. | Antigravity | 💡 Backlog | `docs/specs/v1-preflight/v1-spec.md` |
| **V1-T03** | 🟢 P3 | **Interactive Diff Viewer UI** — Vercel-style preview URL inside PR comments showing a visual side-by-side schema diff. | Antigravity | 💡 Backlog | `docs/specs/v1-preflight/v1-spec.md` |
| **V1-T04** | 🔵 P4 | **`substrate check-deploy`** — CI/CD deployment safety gate to ensure safe deploy ordering (P2-T06 Deferred). | Antigravity | 💡 Backlog | `docs/specs/v1-preflight/v1-spec.md` |
| **V1-T05** | ⚪ P5 | **Local `validate` CLI** — Local validation command for developers to test schema changes against the registry offline. | Antigravity | 💡 Backlog | `docs/specs/v1-preflight/v1-spec.md` |
| **V1-T06** | ⚪ P5 | **Legal & Licensing Audit** — Audit `go.mod` for GPL licenses, add UI Credits/Notices. | Antigravity | 💡 Backlog | `docs/specs/v1-preflight/v1-spec.md` |
| **V1-T07** | 🟡 P6 | **V1.0 System E2E Tests** — Final E2E orchestration tests mimicking production flow across all modules. | Antigravity | 💡 Backlog | `docs/specs/v1-preflight/v1-spec.md` |

---

## 🏢 Phase 7: Enterprise Integrations & ITSM (Post-V1.0)

**Goal:** Integrate Substrate deeply into corporate workflows, providing custom governance, automated ticketing, and targeted notifications.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
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

## How to Submit Jules Tasks

```bash
# From repo root
python3 scripts/jules_submit.py --list
python3 scripts/jules_submit.py --task 1
python3 scripts/jules_submit.py --task 1 --branch feat/diff-engine
```
