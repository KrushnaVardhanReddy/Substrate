# Substrate — Task Tracker

> Last updated: 2026-07-07 (P1-T07 ✅ merged, P1b-T01 SQL spec ✅ approved, P1b-T02 SQL engine ✅ merged)
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

## Phase 0 — Specs & Project Setup ✅ COMPLETE

| Task | Owner | Status | Output |
|---|---|---|---|
| Product vision & README | Antigravity | ✅ | `README.md` |
| Business model & monetization | Antigravity | ✅ | `README.md` (Business Model section) |
| Spec-first contribution rules | Antigravity | ✅ | `CONTRIBUTING.md`, `.cursorrules` |
| Project folder structure | Antigravity | ✅ | `docs/`, `prompts/`, `scripts/`, etc. |
| Jules/Stitch automation scripts | Antigravity | ✅ | `scripts/jules_submit.py`, `scripts/stitch_submit.py` |
| Delegation workflow doc | Antigravity | ✅ | `docs/stitch_jules_workflow.md` |
| Environment setup | Antigravity | ✅ | `.env.local`, `.env.example`, `.gitignore` |
| Diff engine architecture plan | Antigravity | ✅ | `docs/specs/diff-engine-plan.md` |
| **DiffReport JSON schema spec** | Antigravity | ✅ | `docs/specs/diff-report-schema.md` |
| **Breaking change rules spec** | Antigravity | ✅ | `docs/specs/breaking-change-rules.md` |

---

## Phase 1 — Breaking Change Prevention (Incremental Sub-phases)

> **Dependency:** Phase 0 specs complete ✅
> **Architecture:** OpenAPI diffing is powered by **oasdiff** (Go library, Apache 2.0). We wrap it, not rebuild it. Custom parsers handle SQL/GraphQL/ML schemas.

### Phase 1a — OpenAPI 3.x ✅ COMPLETE (v0.1.0 shipped 🚀)

> 🎯 **oasdiff decision:** Using `github.com/oasdiff/oasdiff` as a Go library dependency.
> Gives us 160+ rules, allOf flattening, stability levels, spec validation — for free.
> Phase 1a shrunk from 9 tasks → 4 tasks. Shipped in ~2 weeks.

| Task ID | Name | Owner | Status | Jules Prompt |
|---|---|---|---|---|
| P1-T00 | **`substrate.yaml` override config spec** | Antigravity | ✅ | `docs/specs/override-config.md` |
| P1-T01 | Go module scaffold + DiffReport structs + oasdiff dependency | Jules | ✅ | `prompts/phase-1-diff-engine/t01_go_scaffold.txt` |
| P1-T02 | oasdiff adapter (wraps oasdiff output → `DiffReport` format) | Jules | ✅ | `prompts/phase-1-diff-engine/t02_oasdiff_adapter.txt` |
| P1-T03 | Override config parser + CLI (`substrate.yaml`, `--format json/text/changelog`) | Jules | ✅ | [Jules Session (T03)](https://jules.google.com/u/1/session/8542866566283395371) |
| P1-T04 | Unit tests + E2E tests (80% coverage, real OpenAPI fixture files) | Jules | ✅ | `prompts/phase-1-diff-engine/t04_tests.txt` |
| **P1-T05** | **Add `NOTICES` file** — Apache 2.0 attribution for oasdiff ⚖️ | Antigravity | ✅ | `NOTICES` |
| **P1-T06** | **oasdiff `checker` adapter** — replace shallow diff with semantic 37-rule engine | Jules | ✅ | `prompts/phase-1-diff-engine/t06_checker_adapter.txt` (spec: `docs/specs/openapi-checker-adapter.md`) |
| **P1-T07** | **`substrate init` command** — scaffolds `substrate.yaml` + GitHub Actions workflow for new users | Jules | ✅ | `prompts/phase-1-diff-engine/t07_init_command.txt` (spec: `docs/specs/init-command.md`) |
| **P1-T08** | **OASDiff Rules Refinement** — map granular required-property-removed rules to `FIELD_REMOVED` | Antigravity | ✅ | `docs/specs/openapi-checker-adapter.md` |

**Spec Gate:** P1-T07 spec approved ✅ — Jules prompt written ✅ — submit via `python3 scripts/jules_submit.py --task 7`

---

### Phase 1 MVP — GitHub Action (Distribution Wedge) 🏆 ✅ COMPLETE

> **v0.1.0 tagged and published to GitHub Marketplace on 2026-07-06.**
> **Dogfooding:** `.github/workflows/test-action.yml` validated the action in-repo before launch.

| Task ID | Name | Owner | Status |
|---|---|---|---|
| P1-MVP-1 | Create `action.yml` + `Dockerfile` + `entrypoint.sh` | Antigravity | ✅ |
| P1-MVP-2 | Dogfood the action via `.github/workflows/test-action.yml` | Antigravity | ✅ |
| P1-MVP-3 | Tag `v0.1.0` + publish to GitHub Marketplace | User | ✅ |
| P1-MVP-4 | Write "Optic Alternative" blog post + HackerNews launch | Antigravity | ⏳ |

---

### Phase 1b — SQL Migrations 🔄 IN PROGRESS

> **Dependency:** Phase 1a ✅ COMPLETE. Phase 1b is now unblocked.
> **Spec Gate:** `docs/specs/breaking-change-rules-sql.md` ✅ APPROVED — 26 rules (14 BREAKING, 6 WARNING, 6 SAFE).

| Task ID | Name | Owner | Status |
|---|---|---|---|
| P1b-T01 | SQL breaking change rules spec | Antigravity | ✅ |
| P1b-T02 | SQL DDL parser + diff engine + rule engine (PostgreSQL) | Jules | ✅ |
| P1b-T03 | SQL rule engine + tests (expand coverage, edge cases) | Jules | ⏳ |

### Phase 1c — GraphQL SDL 🔒 BLOCKED on 1a

| Task ID | Name | Owner | Status |
|---|---|---|---|
| P1c-T01 | GraphQL schema parser | Jules | 🔒 |
| P1c-T02 | GraphQL breaking change rules spec | Antigravity | 🔒 |
| P1c-T03 | GraphQL rule engine + tests | Jules | 🔒 |

### Phase 1d — Protobuf & gRPC 🔒 BLOCKED on 1a

| Task ID | Name | Owner | Status |
|---|---|---|---|
| P1d-T01 | Protobuf `.proto` / gRPC parser | Jules | 🔒 |
| P1d-T02 | Protobuf breaking change rules spec | Antigravity | 🔒 |
| P1d-T03 | Protobuf rule engine + tests | Jules | 🔒 |

### Phase 1e — AsyncAPI & Apache Avro 🔒 BLOCKED on 1a

| Task ID | Name | Owner | Status |
|---|---|---|---|
| P1e-T01 | AsyncAPI & Avro schema parsers | Jules | 🔒 |
| P1e-T02 | Event-driven breaking change rules spec | Antigravity | 🔒 |
| P1e-T03 | Event-driven rule engine + tests | Jules | 🔒 |

### Phase 1f — AI/ML Model Contracts ⭐ 🔒 BLOCKED on 1a

| Task ID | Name | Owner | Status |
|---|---|---|---|
| P1f-T01 | `substrate.yaml` model contract spec | Antigravity | 🔒 |
| P1f-T02 | Model contract parser | Jules | 🔒 |
| P1f-T03 | Dataset schema parser (CSV/Parquet) | Jules | 🔒 |
| P1f-T04 | LLM structured output schema parser | Jules | 🔒 |
| P1f-T05 | AI/ML rule engine + tests | Jules | 🔒 |

### Phase 1g — Enterprise Metadata (Salesforce & SOAP) 🔒 BLOCKED on 1a

| Task ID | Name | Owner | Status |
|---|---|---|---|
| P1g-T01 | Salesforce Custom Object/XML parser | Jules | 🔒 |
| P1g-T02 | SOAP WSDL parser | Jules | 🔒 |
| P1g-T03 | Enterprise breaking change rules spec | Antigravity | 🔒 |
| P1g-T04 | Enterprise rule engine + tests | Jules | 🔒 |

---

## Phase 2 — GitHub App ⏳ BLOCKED

> **Dependency:** Phase 1 diff engine binary complete

| Task ID | Name | Owner | Status | Jules Prompt |
|---|---|---|---|---|
| P2-T01 | GitHub App scaffold + webhook receiver | Jules | 🔒 | `prompts/phase-2-github-app/t01_github_app_scaffold.txt` |
| P2-T02 | PR diff trigger → calls Go engine binary | Jules | 🔒 | — |
| P2-T03 | PR comment formatter (impact report) | Jules | 🔒 | — |
| P2-T04 | Required status check (merge blocker) | Antigravity | 🔒 | — |
| P2-T05 | Deployment (Cloudflare Workers) | Jules | 🔒 | — |

---

## Phase 3 — Dashboard, API & MCP Server 💡 PLANNED

> **Dependency:** Phase 2 GitHub App working end-to-end

| Task ID | Name | Owner | Status |
|---|---|---|---|
| P3-T01 | PostgreSQL schema + Go API server | Jules | 💡 |
| P3-T02 | `substrate.yaml` parser + dependency registration | Jules | 💡 |
| P3-T03 | SvelteKit project setup + design system | Antigravity | 💡 |
| P3-T04 | Connected repos list + schema browser | Antigravity + Jules | 💡 |
| P3-T05 | Dependency graph visualization | Antigravity | 💡 |
| P3-T06 | GitHub OAuth + org management | Jules | 💡 |
| P3-T07 | Free tier limits + production deployment | Jules | 💡 |
| **P3-T08** | **MCP server spec** (`docs/specs/mcp-server.md`) | Antigravity | 💡 |
| **P3-T09** | **MCP server implementation** (Go, exposes graph + schema tools) | Jules | 💡 |
| **P3-T10** | **MCP server deployment + IDE integration docs** | Antigravity | 💡 |

---

## Backlog / Future Phases 💡

These are not yet scheduled but are on the product roadmap:

**Override & Skip Config** ⚠️ Must ship before Phase 2 GitHub App
- [x] Write `docs/specs/override-config.md` spec (P1-T00)
- [x] Parse `substrate.yaml` overrides in the diff engine (P1-T03)
- [ ] Show "✅ Acknowledged" vs "❌ Blocking" in PR comments
- [ ] Override expiry enforcement (auto-re-trigger on expired overrides)
- [ ] Override audit trail in the dashboard (who approved what, when)

**MCP Server** ⭐ Phase 3 — Turns Substrate into an IDE-native knowledge layer
- [ ] Write `docs/specs/mcp-server.md` spec (P3-T08)
- [ ] Implement Go MCP server with tools: `get_schema`, `check_impact`, `list_consumers`, `get_change_history`, `validate_change`
- [ ] Deploy MCP server endpoint
- [ ] Write IDE integration guide (Antigravity, Cursor, Claude, Copilot)
- [ ] Expose MCP server as self-hostable for Enterprise tier

**Legal / Pre-launch Checklist** ⚖️
- [ ] Add `NOTICES` file at repo root with Apache 2.0 attribution for oasdiff (`https://github.com/oasdiff/oasdiff`)
- [ ] Add "Credits" section to dashboard UI footer
- [ ] Review all Go module dependencies in `go.mod` for license compatibility before v1.0 release



**Integrations**
- [ ] Slack integration for PR notifications
- [ ] Jira integration
- [ ] PagerDuty integration
- [ ] GitLab & Bitbucket support

**Enterprise**
- [ ] Self-hosted / VPC deployment option
- [ ] SSO / SAML (Okta)
- [ ] Custom rules engine
- [ ] Compliance reporting & audit logs (SOC2)

---

## How to Submit Jules Tasks

```bash
# From repo root
python3 scripts/jules_submit.py --list
python3 scripts/jules_submit.py --task 1
python3 scripts/jules_submit.py --task 1 --branch feat/diff-engine
```
