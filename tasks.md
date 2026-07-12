# Substrate — Task Tracker

> Last updated: 2026-07-10 (Phase 2 complete ✅ — Phase 3 Contract Registry in progress. **P3-T09 MCP Server merged ✅.** P3-T12 Astro Docs submitted to Jules 2026-07-10. Active Jules session running.)
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
| Diff engine architecture plan | Antigravity | ✅ | `docs/specs/phase-1/diff-engine-plan.md` |
| **DiffReport JSON schema spec** | Antigravity | ✅ | `docs/specs/phase-1/diff-report-schema.md` |
| **Breaking change rules spec** | Antigravity | ✅ | `docs/specs/phase-1/breaking-change-rules.md` |

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
| **P1-T09** | **Pending acknowledgments in `substrate.yaml`** — `status: pending` + optional `until:` expiry date for acknowledged breaking changes | Jules | ⏳ | spec: `docs/specs/override-config.md` (needs update) |

> **P1-T09 rationale (inspired by Pact's "Pending Pacts"):** Today `substrate.yaml` lets teams override/ignore a breaking change. That's a binary on/off. We need a richer model: `status: pending` means "we know this breaks, we're fixing it in the next sprint." Substrate shows "⏳ Acknowledged" instead of blocking the merge. An optional `until: 2026-08-01` date auto-expires the acknowledgment, forcing re-review. This is the enterprise compliance feature that makes audit logs meaningful.

**Spec Gate:** P1-T07 spec approved ✅ — Jules prompt written ✅ — submit via `python3 scripts/jules_submit.py --task 7`

---

### Phase 1 MVP — GitHub Action (Distribution Wedge) 🏆 MOSTLY COMPLETE

> **v0.1.0 tagged and published to GitHub Marketplace on 2026-07-06.**
> **Distribution model changed:** Repo is now private. Action distributes via Docker Hub public image (`kpakkiragari/substrate-engine`). Source is never exposed.

| Task ID | Name | Owner | Status |
|---|---|---|---|
| P1-MVP-1 | Create `action.yml` + `Dockerfile` + `entrypoint.sh` | Antigravity | ✅ |
| P1-MVP-2 | Dogfood the action via `.github/workflows/test-action.yml` | Antigravity | ✅ |
| P1-MVP-3 | Tag `v0.1.0` + publish to GitHub Marketplace | User | ✅ |
| P1-MVP-5 | Create public repo `substrate-engine` in Docker Hub | User | ✅ |
| P1-MVP-6 | Add `DOCKERHUB_USERNAME` + `DOCKERHUB_TOKEN` secrets to GitHub repo | User | ✅ |
| P1-MVP-7 | Update `action.yml` Docker Hub image + verify release CI workflow fires on next tag | Antigravity | ✅ |
| **P1-MVP-8** | **Create Public Wrapper Repo** — Create `substrate-action` public repo with just `action.yml` and `README.md` to publish to GitHub Marketplace while keeping source private | User | 💡 |

> **P1-MVP-5 and P1-MVP-6 require manual steps in the browser (Docker Hub + GitHub Settings). See instructions below.**
>
> **Docker Hub setup:** Go to [hub.docker.com](https://hub.docker.com) → Create public repo `substrate-engine` → Generate an access token → Add `DOCKERHUB_USERNAME=kpakkiragari` and `DOCKERHUB_TOKEN=<token>` as GitHub Actions secrets in this repo's Settings → Security → Secrets.
>
> **First image push:** After secrets are added, run `git tag v0.1.1 && git push origin v0.1.1` (or any new tag) to trigger the release workflow and push the first Docker Hub image. Then update `action.yml` to pin to that tag.

---

### Phase 1b — SQL Migrations 🔄 IN PROGRESS

> **Dependency:** Phase 1a ✅ COMPLETE. Phase 1b is now unblocked.
> **Spec Gate:** `docs/specs/breaking-change-rules-sql.md` ✅ APPROVED — 26 rules (14 BREAKING, 6 WARNING, 6 SAFE).
> **Phase 1b done gate (minimum to unblock Phase 2):** T01 + T02 + T03 complete ✅. T04, T05, T06 are post-Phase 2 backlog — they do **not** block Phase 2 from starting.

| Task ID | Name | Owner | Status |
|---|---|---|---|
| P1b-T01 | SQL breaking change rules spec | Antigravity | ✅ |
| P1b-T02 | SQL DDL parser + diff engine + rule engine (PostgreSQL) | Jules | ✅ |
| P1b-T03 | SQL rule engine + tests (expand coverage, edge cases) | Jules | ✅ |
| P1b-T04 | **dbt `schema.yml` adapter** — parse dbt model contracts as SQL schema input | Jules | 🔒 *deferred: post-Phase 2* |
| P1b-T05 | **dbt sources/exposures ingestion** — auto-seed dependency graph from dbt projects | Jules | 🔒 *deferred: post-Phase 2* |
| P1b-T06 | **`stripe/pg-schema-diff` spike** — evaluate replacing hand-rolled `DiffSchemas()` with Stripe's Apache 2.0 library | Antigravity | 🔒 *deferred: post-Phase 2* |

### Phase 1c — GraphQL SDL 🔒 BLOCKED on Phase 2

> **Library decision:** `vektah/gqlparser` (MIT) for SDL parsing. No Go library with pre-built breaking change rules exists — Substrate owns ~15–20 rules (field removed, type changed, argument made required, directive removed). Do **not** shell out to Node.js `graphql-hive` — breaks the single-binary promise.

| Task ID | Name | Owner | Status |
|---|---|---|---|
| P1c-T01 | GraphQL SDL Diff Adapter (`graphql`) | Jules | ✅ | `prompts/phase-1c-graphql/t01_parser.txt` |

### Phase 1d — Protobuf & gRPC ✅ COMPLETE

> **Library decision:** **`bufbuild/buf`** (Apache 2.0, written in Go). `buf breaking` is the `oasdiff` of Protobuf — 100+ wire-compatibility rules, field number checks, service/method detection. Invoked via `exec.Command` (binary, not Go lib import). Adapter pattern identical to P1-T06 oasdiff. **Spec approved. Jules submitted 2026-07-08.**
> **Spec:** `docs/specs/phase-1/protobuf-checker-adapter.md` ✅ APPROVED

| Task ID | Name | Owner | Status | Jules Prompt |
|---|---|---|---|---|
| P1d-T01 | `buf` checker adapter spec + implementation + tests | Jules | ✅ | `prompts/phase-1d-protobuf/t01_buf_adapter.txt` |

### Phase 1e — AsyncAPI & Apache Avro ✅ COMPLETE

> **Library decision (AsyncAPI):** `asyncapi/parser-go` (Apache 2.0) for parsing. Custom diff rules (channel removed, operation changed, message schema breaking).
> **Library decision (Avro):** Do **not** build a custom Avro compatibility checker. Call the Confluent/Apicurio Schema Registry **compatibility check API** — offloads complex Avro evolution rules (union promotion, defaults, field ordering) to a battle-tested engine. Substrate wraps the JSON response into `DiffReport`.
> **Spec:** `docs/specs/phase-1/asyncapi-avro-adapter.md` ✅ APPROVED

| Task ID | Name | Owner | Status |
|---|---|---|---|
| P1e-T01 | AsyncAPI parser using `asyncapi/parser-go` | Jules | ✅ |
| P1e-T02 | Avro adapter using Schema Registry compatibility API | Jules | ✅ |
| P1e-T03 | AsyncAPI + Avro breaking change rules spec | Antigravity | ✅ |
| P1e-T04 | Rule engine + tests | Jules | 🔒 after T01+T02 |

### Phase 1f — AI/ML Model Contracts ⭐ ✅ COMPLETE

> **Library decision:** No new dependencies needed. `gopkg.in/yaml.v3` and `github.com/santhosh-tekuri/jsonschema/v6` are **already in `go.mod`**. Substrate owns the `substrate.yaml` ml_model contract spec.
> **Spec:** `docs/specs/phase-1/aiml-adapter.md` ✅ APPROVED

| Task ID | Name | Owner | Status | Jules Prompt |
|---|---|---|---|---|
| P1f-T01 | AI/ML model contract diff adapter implementation | Jules | ✅ | `prompts/phase-1f-aiml/t01_aiml_adapter.txt` |


### Phase 1g — Enterprise Metadata (Salesforce & SOAP) ✅ COMPLETE

> **Library decision:** Use Go stdlib `encoding/xml` to parse both WSDL/XSD and Salesforce XML metadata. Both are purely XML diffing operations based on specific tags (`<CustomObject>`, `<definitions>`). **Single-binary promise maintained.**
> **Spec:** `docs/specs/phase-1/enterprise-adapter.md` ✅ APPROVED

| Task ID | Name | Owner | Status | Jules Prompt |
|---|---|---|---|---|
| **P1g-T01** | **Enterprise metadata diff adapter (Salesforce, SOAP)** | Jules | ✅ | `prompts/phase-1g-enterprise/t01_enterprise_adapter.txt` |

### Phase 1h — Infrastructure as Code (Terraform) ✅ COMPLETE

> **Library decision (Terraform):** Use Go stdlib `encoding/json` to parse `terraform show -json tfplan` output. No HCL parsing. 
> **Spec:** `docs/specs/phase-1/iac-adapter.md` ✅ APPROVED

| Task ID | Name | Owner | Status | Jules Prompt |
|---|---|---|---|---|
| P1h-T01 | Terraform Plan JSON adapter (`terraform-plan`) | Jules | ✅ | `prompts/phase-1h-iac/t01_terraform_adapter.txt` |

---

> **Wave 2 execution order (Phase 2 complete ✅ — Wave 2 unblocked):** Follow effort order, not numerical order:
> `1d ✅ (Protobuf)` → `1e ✅ (AsyncAPI/Avro)` → `1h ✅ (Terraform)` → `1c ✅ (GraphQL)` → `1f ✅ (AI/ML)` → `1g ✅ (Enterprise)`

---

## Phase 2 — GitHub App ✅ COMPLETE

> **Dependency:** Phase 1 diff engine binary complete ✅
> **Spec Gate:** `docs/specs/github-app.md` ✅ APPROVED
> **Status:** All implementation tasks successfully deployed and validated end-to-end.

| Task ID | Name | Owner | Status | Jules Prompt |
|---|---|---|---|---|
| P2-T01 | GitHub App scaffold + Cloudflare Worker webhook receiver | Jules | ✅ | `prompts/phase-2-github-app/t01_github_app_scaffold.txt` |
| **P2-T02a** | **`substrate serve` HTTP mode** — adds `POST /diff` to Go binary (container service) | Jules | ✅ | `prompts/phase-2-github-app/t02a_binary_serve_mode.txt` |
| P2-T03 | PR comment formatter (standalone TypeScript module) | Jules | ✅ | `prompts/phase-2-github-app/t03_pr_comment_formatter.txt` |
| **P2-T02b** | **Wire Worker → Container → GitHub APIs** — real diff results posted to PRs | Jules | ✅ | `prompts/phase-2-github-app/t02b_wire_worker_container.txt` |
| P2-T04 | Required status check (merge blocker) | Jules | ✅ | `prompts/phase-2-github-app/t04_required_status_check.txt` |
| P2-T05 | Deployment (Cloudflare Workers + Container) | Jules | ✅ | `prompts/phase-2-github-app/t05_deployment.txt` |
| **P2-T06** | **`substrate check-deploy`** — registry query: safe to deploy to production? | Antigravity | 🔒 *deferred: needs Phase 3 registry* | — |
| **P2-T07** | **Webhook re-verification** — consumer spec drift → auto re-check provider | Jules | 🔒 *deferred: needs Phase 3 registry* | — |
| **P2-T08** | **Blog post + HackerNews launch** — official Substrate GitHub App launch | Antigravity | 💡 *after Phase 2 ships* | — |

---

## Phase 3 — Dashboard, API & MCP Server ✅ COMPLETED

> **Dependency:** Phase 2 GitHub App working end-to-end ✅

### The Core Phase 3 Feature: Cross-Repo Contract Registry

> **Problem:** A SQL column is renamed in `backend-api` (PR #45). `frontend` and `mobile-app` both consume that field — but they have no open PRs. How does Substrate know they break?
>
> **Solution:** Substrate maintains a **contract registry** — a stored snapshot of each consumer's current spec pulled from their `main` branch. When a provider's PR lands, Substrate checks the proposed change against every registered consumer's snapshot — no cross-repo PRs needed.
>
> **Design:**
> ```yaml
> # substrate.yaml in frontend repo
> consumers:
>   - name: "users-api"
>     type: openapi
>     source: "github.com/myorg/backend-api"
>     path: "openapi.yaml"
>     branch: "main"
> ```
> When `backend-api` opens a PR with a breaking OpenAPI change, Substrate fetches `frontend`'s stored `schema.graphql` from the registry and checks compatibility. It comments directly on the provider's PR: *"⚠️ Consumer `frontend` will break — field `user_name` no longer exists."*
>
> **How we can test this today (before Phase 3 infra):** Simulate multiple repos as subfolders in `engine/cmd/substrate/testdata/cross-repo/`. E2E tests run Substrate against `(provider_new_spec, consumer_stored_spec)` and assert cross-boundary breaks are detected. No real cross-repo infra needed for the engine tests.

> **Spec:** `docs/specs/phase-3/contract-registry.md`  
> **Execution order:** P3-T01 → P3-T02 → P3-T02b → **P3-T02c+P3-T02d (parallel)** → Dashboard → MCP
> **Jules sessions:** P3-T01 ✅ | P3-T02 ✅ | P3-T02b ✅ | P3-T02c ⏳ ready to spec

| Task ID | Priority | Name | Owner | Status |
|---|---|---|---|---|
| P3-T01 | 🔴 P1 | PostgreSQL schema + Go API server | Jules | ✅ |
| P3-T02 | 🔴 P1 | `substrate.yaml` consumer declaration parser | Jules | ✅ |
| **P3-T02b** | 🔴 P1 | **Contract registry sync** — snapshot consumer specs on push to `main` | Jules | ✅ | `prompts/phase-3-registry/t02b_contract_sync.txt` |
| **P3-T02c** | 🔴 P1 | **Cross-repo check on PR** — validate provider PR against all consumer snapshots | Jules | ✅ | `prompts/phase-3-registry/t02c_cross_repo_check.txt` |
| **P3-T06** | 🟡 P2 | **GitHub OAuth + org management** | Jules | ✅ Merged `feature/dev` 2026-07-10 (`1fc0f16`) | `prompts/phase-3-registry/t06_auth_org.txt` |
| **P3-T02d** | 🟡 P2 | **Cross-repo E2E fixture tests** — `testdata/cross-repo/` | Jules | ✅ | `prompts/phase-3-registry/t02d_cross_repo_e2e.txt` |
| **P3-T03** | 🟢 P3 | **SvelteKit dashboard project setup + design system** | Jules | ✅ | `prompts/phase-3-registry/t03_dashboard.txt` |
| **P3-T04** | 🟢 P3 | **Connected repos list + schema browser** | Jules | ✅ Merged `feature/dev` 2026-07-10 (`c2fa3ce`) | `prompts/phase-3-registry/t04_dashboard_repos.txt` |
| **P3-T05** | 🟢 P3 | **Dependency graph visualization** | Antigravity | ✅ | `docs/specs/phase-3/dashboard.md` |
| **P3-T11** | 🟢 P3 | **Compatibility matrix dashboard** — provider × consumer version grid | Antigravity | ✅ | `docs/specs/phase-3/dashboard.md` |
| **P3-T15** | 🟡 P2 | **Dashboard E2E Tests (Playwright)** — test nav, graph, and matrix with API mocks | Jules/Antigravity | ✅ Merged `feature/dev` 2026-07-10 | `prompts/phase-3-registry/t15_dashboard_e2e.txt` |
| **P3-T16** | 🟢 P3 | **AI Schema Validator Playground (UI)** — Interactive split-pane diff & AI auto-remediation | Jules | ✅ Merged `feature/dev` 2026-07-10 | `prompts/phase-3-registry/t16_playground.txt` |
| **P3-T08** | 🔵 P4 | **MCP server spec** (`docs/specs/phase-3/mcp-server.md`) | Antigravity | ✅ |
| **P3-T09** | 🔵 P4 | **MCP server implementation** (Go, JSON-RPC 2.0) | Jules | ✅ Merged `feature/dev` 2026-07-09 | `prompts/phase-3-registry/t09_mcp_implementation.txt` |
| **P3-T09b** | 🔵 P4 | **Wire MCP to live Registry API** — `get_schema_file` (new endpoint), `analyze_repository` (local fs walk), `get_substrate_docs` (embedded docs). `get_breaking_change_history` deferred to Phase 4 (needs new DB table). | Jules | ✅ Merged `feature/dev` 2026-07-10 (`7b3c15e`) | `prompts/phase-3-registry/t09b_wire_mcp_registry.txt` |
| **P3-T10** | 🔵 P4 | **MCP deployment + IDE integration docs** | Jules | ✅ Merged `feature/dev` 2026-07-10 | `prompts/phase-3-registry/t10_mcp_docs.txt` |
| **P3-T12** | 🔵 P4 | **Comprehensive Substrate Platform Docs** — Scaffold Astro Starlight Diátaxis site | Jules | 🔄 Submitted (Session 6304847416252553404) | `prompts/phase-3-registry/t12_astro_starlight_diataxis.txt` |
| **P3-T13** | 🟢 P3 | **Execution Modes (Enterprise Rollout)** — Add `--mode=legacy\|strict` to CLI for shadow mode / dry-run deployments | Jules | ✅ Merged `feature/dev` 2026-07-09 | `prompts/phase-3-registry/t13_enterprise_modes.txt` |
| **P3-T14** | 🟢 P3 | **Interactive Diff Viewer URL** — Worker adds a "View Dashboard" link to GitHub comments, routing to the Svelte dashboard | Jules | ✅ Merged `feature/dev` 2026-07-10 (`7ab9f8a`) | `prompts/phase-3-registry/t14_interactive_diff_url.txt` |
| **P3-T07** | ⚪ P5 | **Free tier limits + production deployment** | Jules | ✅ Merged `feature/dev` 2026-07-10 | `prompts/phase-3-registry/t07_rate_limits.txt` |

> **P3-T11 rationale (inspired by PactFlow's compatibility matrix):** PactFlow's most requested enterprise dashboard feature. For a platform team managing 20+ microservices, this is the central control panel: "which version of `users-api` is compatible with which version of `frontend` and `mobile-app`?" Every cell in the grid is a green tick or red cross. This is what makes Substrate indispensable for large orgs and is a core enterprise upsell feature.


---

## Phase 4 — AI Intelligence Layer 🧠 VISION

> **Dependency:** Phase 3 contract registry + dependency graph complete.
>
> **The Core Insight:** Phase 3 collects the data. Phase 4 uses it. The moat is not the model — the moat is the data. Substrate's accumulated schema evolution graph (versions, dependencies, breaking change history, override patterns) gives AI systems context that no competitor can replicate.
>
> **Architecture decision:** We do NOT train a general-purpose LLM (that's a $100M+ project). We use a **hybrid approach**:
> - **MCP + Foundation Models** (Claude/GPT/Gemini) for reasoning-heavy tasks — these models reason over Substrate's data via MCP tools. Ship in weeks, not months.
> - **Small Specialized Models** (classical ML / fine-tuned classifiers) for narrow, well-defined tasks where a small model clearly outperforms a prompted LLM (anomaly detection, prediction).
> - **Never** fine-tune a general LLM — it won't beat GPT-4o even with schema data, and it'll be obsoleted by the next model release anyway.
>
> **Timeline reality:** Phase 4 splits into two sub-phases with very different timelines. Phase 4a is fast. Phase 4b needs data accumulation.

### Phase 4a — MCP + Foundation Models 🚀 (fast — weeks after Phase 3 ships)

> These features are essentially: **MCP server live + good system prompt + UI wrapper = done.** No ML training, no data accumulation needed. The foundation model does the reasoning. Your value is in the data tools you expose via MCP.

| Task ID | Name | AI Approach | Owner | Status |
|---|---|---|---|---|
| **P4a-T01** | **AI Impact Analyst** — "What changed in the payments API last month that could affect checkout?" | MCP tools → Claude/GPT reasons over schema graph | Antigravity | ✅ Solved by P3-T09 (MCP Server) |
| **P4a-T02** | **Zero-Touch Auto-Discovery PR** — On installation, GitHub App scans legacy repos and automatically opens PR adding `substrate.yaml` | MCP/LLM scans repo and generates config | Jules | 💡 |
| **P4a-T03** | **AI API Architect (Spec-First)** — `substrate design` CLI command generates best-practice OpenAPI from text | LLM generates perfect schema | Antigravity | 💡 |
| **P4a-T04** | **Auto-Migration Generator** — breaking change detected → AI generates migration code diff for each affected consumer | MCP tools → foundation model generates code | Jules | 💡 |
| **P4a-T05** | **Smart Deprecation Planner** — repeatedly acknowledged break → AI suggests 30-day deprecation plan | MCP tools → foundation model reasons over history | Antigravity | 💡 |
| **P4a-T06** | **AI-powered PR review assistant** — "This PR touches 5 schemas. Full cross-repo impact + suggested reviewers" | MCP tools → foundation model, triggered on PR open | Antigravity | 💡 |

### Phase 4 Implementation Tasks 🏗️ (Ready to execute after P3-T15/P3-T16)

> Concrete engineering tasks. Full spec: `docs/specs/phase-4/ai-intelligence.md`

| Task ID | Priority | Description | Owner | Status |
|---|---|---|---|---|
| **P4-T01** | 🔴 P1 | **AI Reasoning Bridge** — Wire LLM (Gemini/Claude) to MCP tools via Go API | Antigravity | ✅ Done |
| **P4-T02** | 🔴 P1 | **Streaming SSE Handler** — `POST /api/v1/ai/analyze` with real-time token streaming | Jules | ✅ Done | `prompts/phase-4/t02_streaming_sse_ai_handler.txt` |
| **P4-T03** | 🟡 P2 | **Safe Schema Patch Generator** — AI suggests exact YAML/SQL provider fix | Jules | ✅ Done | `prompts/phase-4/t03_safe_schema_patch_generator.txt` |
| **P4-T04** | 🟡 P2 | **GitHub PR Comment Upgrade** — Add AI explanation + fix snippet to PRs | Jules | ✅ Done | `prompts/phase-4/t04_github_pr_comment_upgrade.txt` |
| **P4-T05** | 🟢 P3 | **Wire Real AI to Playground** — Replace mock in P3-T16 with live SSE stream | Antigravity | ✅ Done |
| **P4-T06** | 🔵 P4 | **Shift-Left VSCode Extension** — Inline breaking-change warnings in the editor | Jules | ✅ Done | `prompts/phase-4/t06_vscode_extension.txt` |
| **P4-T07** | ⚪ P5 | **Traffic-Aware Diffing** — Prometheus/OTel integration to suppress unused-field warnings | Jules | ✅ Done | `prompts/phase-4/t07_traffic_aware_diffing.txt` |
| **P4-T08** | ⚪ P5 | **PII & Compliance Auditing** — Auto-tag `[HIPAA]`, `[PCI]`, `[PII]` fields on diff | Jules | ✅ Done | `prompts/phase-4/t08_compliance_auditing.txt` |
| **P4-T09** | 🔵 P4 | **Breaking Change History (MCP)** — Postgres table and API for MCP tool | Jules | ⏳ Ready | `prompts/phase-4/t09_breaking_change_history.txt` |

### Phase 4b — Specialized ML (slow — needs 6-12 months of real user data)

> These features need enough accumulated data to train/tune on. **Do not start these until Phase 3 has been live for at least 6 months** and you have meaningful signal. Classical ML (no LLM) for the numerical tasks. Small classifier for prediction.

| Task ID | Name | AI Approach | Owner | Status |
|---|---|---|---|---|
| **P4b-T01** | **Schema Health Score** — per-team grade on breaking change frequency vs org average | Statistics on historical registry data. No ML needed — just counting. Start here. | Jules | 💡 |
| **P4b-T02** | **Anomaly Detection** — alert when schema change velocity spikes (early warning before bad release) | Classical ML: time-series anomaly detection (isolation forest / z-score baseline). NOT an LLM task. | Jules | 💡 |
| **P4b-T03** | **Predictive Breaking Change Detection** — "this diff pattern has historically broken consumers in 87% of similar cases" | Small binary classifier trained on Substrate's accumulated diff history. Fine-tune only if >10k labelled examples. | Antigravity | 💡 |

> **Phase 4b note:** P4b-T01 (Schema Health Score) is actually statistics, not ML — it can ship as soon as Phase 3 has any data at all. P4b-T02 and P4b-T03 are the real "needs data" items. Start collecting the training signal from day one of Phase 3.

---

## Backlog / Future Phases 💡

These are not yet scheduled but are on the product roadmap:

**Phase 7 — Expanded Data Ecosystem** 🗄️
*   **Expanded SQL Dialects:** Add parsers for MySQL/MariaDB, SQL Server (T-SQL), Oracle, and SQLite. (Leverages the existing generic SQL rules engine).
*   **NoSQL Schema Diffing (MongoDB / DynamoDB):** Since NoSQL is "schema-less" at the DB layer, Substrate will parse application-level ODMs (Mongoose schemas, Prisma, Python Pydantic models) or generic JSON Schema exports to detect when a document structure changes and breaks downstream consumers.

**Phase 5 — Enterprise Automation & Governance** 🏢
See the full spec: `docs/specs/enterprise-vision.md`
- [ ] **Cross-Repo Auto-Fix PR Generator** — Substrate opens fix PRs in consumer repos automatically (Deferred from Phase 4)
- [ ] **Dynamic Dependency Discovery** — Full spec: `docs/specs/phase-5/dependency-discovery.md`
  - **Tier 1a — Env Var Scanner:** `.env.example`, `docker-compose.yml`, K8s manifests, GitHub Actions `env:` blocks, `fly.toml`, `Dockerfile ENV` → extract `*_API_URL` / `*_ENDPOINT` patterns, resolve against URL→Repo registry
  - **Tier 1b — Package Manifest Scanner:** `package.json` (`@myorg/*`), `go.mod` (internal modules), `requirements.txt`, `pom.xml` → SDK import = contract dependency
  - **Tier 1c — OpenAPI Generator Config Scanner:** `openapitools.json`, `.openapi-generator-config.yaml`, `Makefile` generator targets → `inputSpec` URL is a direct, deterministic schema reference
  - **Tier 1d — Docker Compose / K8s / Helm:** `depends_on` blocks + env var values with K8s DNS patterns (`*.svc.cluster.local`)
  - **Tier 2a — Terraform Extended:** env var injections from resource references + `output` URL scraping
  - **Tier 2b — Message Queue / Event-Driven:** AsyncAPI `$ref` cross-repo URLs + Kafka topic consumer/producer mapping
  - **Tier 3 — Runtime Confirmation:** OTel/Datadog trace ingestion + eBPF/Service Mesh (enterprise)
  - **URL→Repo Registry:** GitHub Deployments API auto-population + `substrate.yaml deployed_urls` declaration + Terraform output scraping
  - **Confidence Scoring:** Multi-signal scoring system (0–100 pts) — High (80+), Medium (50–79), Low (<50) — shown in dashboard dependency graph
- [ ] Custom Rules Engine (CEL / OPA Rego)
- [ ] Traffic-Aware Diffing (Datadog/OTel integration for `WARNING (Unused)`)
- [ ] ITSM Integration (Jira & ServiceNow Auto-Ticketing)
- [ ] Auto-SDK PR Generation
- [ ] Local Time-Travel Mock Servers
- [ ] Runtime Drift Detection (eBPF / Envoy)
- [ ] Security & PII Auditing
- [ ] Compatibility Gates (SonarQube-style quality gates)
- [ ] Shift-Left IDE Plugins (SonarLint equivalent via MCP)
- [ ] Cross-Repo Auto-Fix PRs (Snyk-style auto-remediation)
- [ ] Cross-Repo Data Flow Taint Analysis (Checkmarx style)
- [ ] Compliance Mapping (SOC2, HIPAA, GDPR tags)
- [ ] **LLM Direct Spec Inference (For LEGACY Projects):** If no spec exists, use an LLM (Approach 1) to read raw routing code (`routes.ts`) and dynamically generate the OpenAPI spec in the background without user intervention.

**Override & Skip Config** ⚠️ Must ship before Phase 2 GitHub App
- [x] Write `docs/specs/override-config.md` spec (P1-T00)
- [x] Parse `substrate.yaml` overrides in the diff engine (P1-T03)
- [ ] Show "✅ Acknowledged" vs "❌ Blocking" in PR comments
- [ ] Override expiry enforcement (auto-re-trigger on expired overrides)
- [ ] Override audit trail in the dashboard (who approved what, when)

**User Feedback Signal** 📊 Required before Wave 2 starts
- [ ] Pin a GitHub Issue: "What schema format do you need next? 👍 the comment" — captures demand signal to order Wave 2
- [ ] Add analytics to the GitHub Action (opt-in usage ping) to see which schema types are most requested
- [ ] Set up a Discord or GitHub Discussions channel for user feedback

**Library Selection Checklist** — Required gate before every Jules implementation task
- [ ] For every new schema format: check if an `oasdiff`-equivalent library exists (Go, Apache 2.0 / MIT) BEFORE writing the Jules prompt
- [ ] Document the library decision in the task row — not after implementation, before it
- [ ] Verify license (Apache 2.0 or MIT only — no GPL, no commercial)
- [ ] Confirm no subprocess / Node.js dependency — single-binary promise must hold for all formats

**MCP Server** ⭐ Phase 3 — Turns Substrate into an IDE-native knowledge layer
- [x] Write `docs/specs/mcp-server.md` spec (P3-T08) ✅
- [x] Implement Go MCP server with 7 tools: `get_dependency_graph`, `check_compatibility`, `get_breaking_change_history`, `get_substrate_docs`, `analyze_repository`, `execute_cli_command`, `get_schema_file` ✅ (P3-T09 merged 2026-07-09)
- [x] **Test `check_compatibility`:** Verified live integration with the Diff Engine via local execution.
- [ ] **Wire Mock Endpoints to Postgres:** Currently, tools like `get_dependency_graph` and `get_breaking_change_history` return mock data. They need to be wired to the live PostgreSQL Registry API and validated later.
- [ ] Deploy MCP server endpoint (P3-T10 — ⏳ next)
- [ ] Write IDE integration guide (Antigravity, Cursor, Claude, Copilot)
- [ ] Expose MCP server as self-hostable for Enterprise tier
- [ ] **AI Pre-flight Validation Workflow:** Expand MCP tools to allow the AI to automatically create, validate configurations/schemas locally (using the local Diff Engine CLI), and auto-remediate breaking changes *before* committing code.

**Legal / Pre-launch Checklist** ⚖️
- [x] Add `NOTICES` file at repo root with Apache 2.0 attribution for oasdiff
- [ ] Add "Credits" section to dashboard UI footer
- [ ] Review all Go module dependencies in `go.mod` for license compatibility before v1.0 release


**Developer Experience (DX) & Tooling**
- [ ] Surface exact Engine error logs (e.g. invalid schema_type, parsing errors) inside GitHub PR comments for faster debugging.
- [ ] Implement a local CLI command (`substrate validate` or `substrate diff`) to allow developers to test schema compatibility and configurations locally before committing.
- [ ] **Bite-Sized Video Tutorials** — Record 60-second YouTube shorts (Installation, config, AI Auto-fix) and embed them directly in the Astro Starlight docs.

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

## 🧭 Phase 5: Automated Dependency Discovery (Enterprise)

**Goal:** Eliminate all manual `substrate.yaml` consumer declarations by automatically building the full cross-repo dependency graph from static and runtime signals.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P5-T01** | 🔴 P1 | **Env Var & URL Registry** — Scan `.env.example`, k8s, docker-compose. Populate URL→Repo registry via GitHub Deployments. | Jules | ✅ Merged | `prompts/phase-5-discovery/t01_env_scanner.txt` |
| **P5-T02** | 🟡 P2 | **Package & Generator Scanners** — Parse `package.json`, `go.mod`, and `openapi-generator-config.yaml` for internal SDKs. | Jules | ✅ Merged | `prompts/phase-5-discovery/t02_package_scanner.txt` |
| **P5-T03** | 🟢 P3 | **Terraform & UI** — Analyze Terraform env injection/outputs. Build Confidence Scoring UI in the Svelte Dashboard. | Antigravity | ✅ Merged | `prompts/phase-5-discovery/t03_terraform_and_ui.txt` |
| **P5-T04** | 🔵 P4 | **Event-Driven Discovery** — Map Kafka topics and AsyncAPI `$ref` cross-references for message queues. | Jules | ✅ Merged | `prompts/phase-5-discovery/t04_event_discovery.txt` |
| **P5-T05** | ⚪ P5 | **Runtime Confirmation** — Ingest OTel/Datadog traces and eBPF network logs to confirm static graph edges. | Antigravity | ✅ Merged | `prompts/phase-5-discovery/t05_runtime_confirmation.txt` |
| **P5-T06** | 🟡 P6 | **Phase 5 E2E Tests** — Cross-repo dependency discovery tests validating Env, Package, Terraform, and Event scanners against mock repositories. | Antigravity | ⏳ Ready to Start | `(Pending)` |

---

## 🧪 Phase 6: QA & Automation Layer (The SDET Co-Pilot)

**Goal:** Automatically generate, update, and cover QA test infrastructure using the explicit schema contracts.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P6-T01** | 🔴 P1 | **Auto-Updating Postman** — Sync schema changes to Postman collections via API or auto-generated folders. | Jules | ⏳ Jules PR Pending | `prompts/phase-6-qa/t01_postman_sync.txt` |
| **P6-T02** | 🟡 P2 | **Shadow API Coverage** — Map OTel/Datadog traces to OpenAPI to find untested fields. Update Dashboard UI. | Jules | ⏳ Jules PR Pending | `prompts/phase-6-qa/t02_shadow_coverage.txt` |
| **P6-T03** | 🟢 P3 | **Auto-Generating Test Code** — Fuzz API constraints to generate executable Playwright/Go tests. | Jules | ⏳ Jules PR Pending | `prompts/phase-6-qa/t03_test_code_generation.txt` |
| **P6-T04** | 🔵 P4 | **Mock Server Time Machine** — CLI command to spin up local mock servers for historical API versions. | Jules | ⏳ Jules PR Pending | `prompts/phase-6-qa/t04_mock_server.txt` |

---

## How to Submit Jules Tasks

```bash
# From repo root
python3 scripts/jules_submit.py --list
python3 scripts/jules_submit.py --task 1
python3 scripts/jules_submit.py --task 1 --branch feat/diff-engine
```
