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
| **P5-T06** | 🟡 P6 | **Phase 5 E2E Tests** — Cross-repo dependency discovery tests validating Env, Package, Terraform, and Event scanners against mock repositories. | Antigravity | ✅ Merged | `(Pending)` |
| **P5-T07** | 🟡 P7 | **100-Repo Scale & Chaos Simulation** — Built a highly concurrent APM simulator that fires 100 repositories with jitter, tests poison pill schemas (50k lines), mutates the graph, and outputs a telemetry reporting matrix. | Jules | ✅ Merged | `docs/specs/phase-5/p5-t07-scale-simulation.md` |
| **P5-T08** | 🔴 P1 | **Chi Router Migration & Panic Recovery** — Migrate standard `http.ServeMux` to `go-chi/chi/v5` and implement `middleware.Recoverer` to prevent poison pill panics during stress testing. | Jules | ✅ Merged | `docs/specs/phase-5/p5-t08-chi-router.md` |

---

## 🧪 Phase 6: QA & Automation Layer (The SDET Co-Pilot)

**Goal:** Automatically generate, update, and cover QA test infrastructure using the explicit schema contracts.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P6-T01** | 🔴 P1 | **Auto-Updating Postman** — Sync schema changes to Postman collections via API or auto-generated folders. | Jules | ✅ Merged | `prompts/phase-6-qa/t01_postman_sync.txt` |
| **P6-T02** | 🟡 P2 | **Shadow API Coverage** — Map OTel/Datadog traces to OpenAPI to find untested fields. Update Dashboard UI. | Jules | ✅ Merged | `prompts/phase-6-qa/t02_shadow_coverage.txt` |
| **P6-T03** | 🟢 P3 | **Auto-Generating Test Code** — Fuzz API constraints to generate executable Playwright/Go tests. | Jules | ✅ Merged | `prompts/phase-6-qa/t03_test_code_generation.txt` |
| **P6-T04** | 🔵 P4 | **Mock Server Time Machine** — CLI command to spin up local mock servers for historical API versions. | Jules | ✅ Merged | `prompts/phase-6-qa/t04_mock_server.txt` |
| **P6-T05** | 🟡 P6 | **Phase 6 E2E Tests** — Cross-functional E2E suite validating Mock Server, Test Generation, Shadow Coverage, and Postman sync. | Antigravity | ✅ Merged | `docs/specs/phase-6/p6-t05-e2e.md` |

---

## 🚀 V1.0 Pre-Flight Checklist (Prod Launch)

**Goal:** Finalize the developer experience, onboarding friction, and legal requirements before pushing Substrate to the GitHub Marketplace.

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **V1-T01** | 🔴 P1 | **GitHub App Auto-Discovery** — Zero-touch onboarding. Auto-scan repos for OpenAPI and open PRs with `substrate.yaml`. | Jules | ✅ Merged | `docs/specs/v1-preflight/v1-spec.md` |
| **V1-T02** | 🟡 P2 | **CLI AI Architect (`init --design`)** — Conversational LLM interface to scaffold an API contract before writing code. | Jules | ✅ Merged | `docs/specs/v1-preflight/v1-spec.md` |
| **V1-T03** | 🟢 P3 | **Interactive Diff Viewer UI** — Vercel-style preview URL inside PR comments showing a visual side-by-side schema diff. | Jules | ✅ Merged | `docs/specs/v1-preflight/v1-spec.md` |
| **V1-T04** | 🔵 P4 | **`substrate check-deploy`** — CI/CD deployment safety gate to ensure safe deploy ordering (P2-T06 Deferred). | Jules | ✅ Merged | `docs/specs/v1-preflight/v1-spec.md` |
| **V1-T05** | ⚪ P5 | **Local `validate` CLI** — Local validation command for developers to test schema changes against the registry offline. | Jules | ✅ Merged | `docs/specs/v1-preflight/v1-spec.md` |
| **V1-T06** | ⚪ P5 | **Legal & Licensing Audit** — Audit `go.mod` for GPL licenses, add UI Credits/Notices. | Jules | ✅ Merged | `docs/specs/v1-preflight/v1-spec.md` |
| **V1-T07** | 🟡 P6 | **V1.0 System E2E Tests** — Final E2E orchestration tests mimicking production flow across all modules. | Jules | ✅ Merged | `docs/specs/v1-preflight/v1-e2e-spec.md` |

---

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



## Recently Completed Tasks (Moved from tasks.md)

### Legend

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| ✅ | Complete |

### 🔐 Phase 9: Compliance, IDEs & Developer Experience

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P9-T10** | 🚀 P1 | **MCP Server WASI Distribution** — Compile the `substrate-mcp` server using `GOOS=wasip1 GOARCH=wasm` to allow secure, sandboxed execution of the MCP server in Claude Desktop or Cursor via Wasmtime/Node.js. | Jules | ✅ Complete | `docs/specs/phase-9/p9-t10-mcp-wasi.md` |
| **P9-T15** | 🚀 P1 | **Public Impact API & MCP Server** — Expose a REST API (`/api/v1/impact`) and an MCP tool (`get_blast_radius`) allowing CI/CD pipelines to block merges based on downstream risk scores, and enabling AI agents (Cursor/Claude) to autonomously fix downstream breaking changes. | Jules | ✅ Complete | `docs/specs/phase-9/p9-t15-impact-api.md` |

### 🚀 Phase 10: Ecosystem Expansion & Security (Post-V1.0)

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P10-T08** | 🔵 P4 | **Live Documentation Generator** — Auto-generate live developer portals and Mermaid architecture diagrams from Substrate's live schema/graph metadata. | Jules | ✅ Done | `docs/specs/phase-10/p10-t08-live-docs.md` |
| **P10-T11** | 🚀 P1 | **Official Terraform Provider** — Build `terraform-provider-substrate` so DevOps teams can manage webhooks, rules, and RBAC policies entirely via Infrastructure-as-Code. | Jules | ✅ Done | `docs/specs/phase-10/p10-t11-terraform-provider.md` |
| **P10-T16** | 🔴 P1 | **Webhook Auto-Discovery Integration** — Wire the Phase 5 Discovery Engine (`api/internal/discovery`) into the Cloudflare Webhook pipeline to automatically scan repos for dependencies, eliminating the need for manual `substrate.yaml` consumer mapping. | Jules | ✅ Complete | `docs/specs/phase-10/p10-t16-webhook-auto-discovery.md` |
| **P10-T17** | ⭐ P1 | **Implicit Infrastructure Discovery (IID)** — Scan package manifests, Docker Compose, and SaaS SDKs to auto-generate infrastructure dependency graphs (Databases, Queues, Stripe, AWS) with Zero-Config. | Jules | ✅ Done | `docs/specs/phase-10/p10-t17-implicit-infrastructure-discovery.md` |

### 🗺️ Phase 11: Advanced Graph Visualization (V2.0 UX)

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
| **P12-T10** | 🔴 P1 | **VCS-Agnostic Webhook & API Adapter** — Refactor the Cloudflare worker to implement an Adapter pattern, supporting GitHub, GitLab, and Gitea/Forgejo payloads. | Jules | ✅ Done | `docs/specs/phase-12/p12-t10-vcs-agnostic-adapter.md` |
| **P12-T11** | 🟡 P2 | **Multi-VCS Onboarding UI** — Update the dashboard onboarding flow to present three connect options: GitHub (Cloud), GitLab (Cloud), and Custom Server (URL & Token) for self-hosted instances. | Jules | ✅ Done | `docs/specs/phase-12/p12-t11-multi-vcs-onboarding-ui.md` |
| **P12-T12** | 🚀 P1 | **System Matrix E2E Test** — Playwright suite that runs `git push` against local Forgejo to test the Red (Breaking), Green (Safe), and Yellow (Override) paths natively end-to-end. | Jules | ✅ Done | `docs/specs/phase-12/p12-t12-system-matrix-e2e.md` |
| **P12-T13** | ⭐ P1 | **Zero-Config Developer Portal (Catalog UI)** — Build a "Backstage Killer" Service Catalog in the dashboard to auto-render API docs using Stoplight Elements without manual yaml config. | Jules | ✅ Done | `docs/specs/phase-12/p12-t13-zero-config-catalog.md` |
| **P12-T14** | 🔴 P1 | **System Matrix Overrides (Yellow Path)** — Extend the E2E Matrix to SQL and AsyncAPI, and implement the Acknowledged status in the Go backend and Svelte UI. | Jules | ✅ Done | `docs/specs/phase-12/p12-t14-yellow-path-matrix.md` |
| **P12-T15** | 🚀 P1 | **Advanced Feature Suites E2E** — Automate E2E testing for WASM Engine, MCP Impact API, AI Autofix, SSE Real-Time, Blast Radius, and Enterprise Dashboard UI. | Jules | ✅ Complete | `docs/specs/phase-12/p12-t15-advanced-e2e.md` |
| **P12-T16** | 🔴 P1 | **Advanced E2E UI Implementation (TDD)** — Implement the SvelteKit frontend UI (Enterprise Routes, Impact API UI, Transitive Blast Radius, Telemetry) to satisfy the failing TDD E2E tests from P12-T15. | Jules | ✅ Complete | `docs/specs/phase-12/p12-t16-advanced-e2e-ui.md` |

### 👑 Phase 13: God-Mode & Enterprise Intelligence

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P13-T03** | 🤯 P1 | **AI Chaos Engineering Auto-Tests** — Automatically write Playwright/Jest tests that prove an API breakage, run them in a sandbox, and post the failing test logs to the PR. | Jules | ✅ Complete | `docs/specs/phase-13/p13-t03-ai-chaos-tests.md` |

### 📥 Archived from tasks.md

| Task ID | Tier | Name & Description | Owner | Status | Spec Link |
|---|---|---|---|---|---|
| **P9-T00** | 🚀 P1 | **GitHub Marketplace Launch (WASI)** — Build the `substrate-action` wrapper using `GOOS=wasip1 GOARCH=wasm`. Execute via Node/Wasmtime for sub-second, highly secure, Docker-less CI/CD diffing, and publish to the Marketplace. | Jules | ✅ Done | `docs/specs/phase-9/p9-t00-wasi-action.md` |
| **P9-T01** | 🔴 P1 | **Shift-Left IDE Plugins** — VSCode/IntelliJ extensions powered by the MCP server to underline breaking changes as the developer types. | Jules | ✅ Done | `docs/specs/phase-9/p9-t01-ide-plugins.md` |
| **P10-T02** | 🟡 P2 | **API Gateway Auto-Sync** — Automatically push validated OpenAPI schemas to AWS API Gateway, Kong, or Cloudflare API Shield on merge to `main`. | Jules | ✅ Done | `docs/specs/phase-10/p10-t02-api-gateway.md` |
| **P10-T09** | 🚀 P1 | **Protobuf & gRPC Schema Registry** — Add native support for parsing, diffing, and visualizing Protobufs to capture backend-to-backend enterprise microservices. | Jules | ✅ Done | `docs/specs/phase-10/p10-t09-protobuf-grpc.md` |
| **P10-T12** | 🚀 P1 | **Tree-sitter Deterministic Impact Analysis** — Parse downstream consumer repositories using Tree-sitter AST to deterministically pinpoint exactly *which lines of code* are broken by an upstream API change. | Jules | ✅ Done | `docs/specs/phase-10/p10-t12-treesitter-analysis.md` |
| **P12-T09** | 🟢 P3 | **AI Support Copilot** — A floating AI chat widget in the dashboard that uses our existing Phase 4 Intelligence Layer to answer questions, generate `substrate.yaml` configs, and troubleshoot user graphs in real-time. | Jules | ✅ Done | `docs/specs/phase-12/p12-t09-ai-copilot.md` |
| **P13-T01** | 🤯 P1 | **FinOps Cost Prediction** — Connect schema diffs to Datadog traffic to calculate the exact USD egress cost increase of payload size changes. | Jules | ✅ Done | `docs/specs/phase-13/p13-t01-finops.md` |
| **P13-T02** | 🤯 P1 | **DB Performance Breakages** — Dry-run Prisma/PlanetScale migrations to predict table-locks and performance outages before they merge. | Jules | ✅ Done | `docs/specs/phase-13/p13-t02-db-performance.md` |
| **P13-T04** | 🔴 P1 | **Enterprise E2E Validation** — Create a `phase10_13_e2e_test.go` suite to programmatically validate Protobuf diffing, Embedded JS Governance Rules, FinOps egress calculations, and Tree-sitter AST impact analysis. | Jules | ✅ Done | `docs/specs/phase-13/p13-t04-e2e-validation.md` |
| **P14-T02** | 🚀 P1 | **"Time to Break" Predictive Scoring** — Upgrade P4b anomaly detection with a concrete user-facing output: a color-coded risk score per graph node indicating the probability of a breaking change in the next N sprints, plus a weekly digest to Platform teams. | Jules & Stitch | ✅ Done | `docs/specs/phase-14/p14-t02-predictive-scoring.md` |
| **P14-T03** | 🔴 P1 | **Living API Changelog (Auto-Generated Public Page)** — Auto-generate a beautiful, public-facing versioned changelog page (like Stripe's API Changelog) from every tracked schema change. Shareable at `substrate.io/myorg/payments-api/changelog`. Embeddable via iframe/JS widget. | Jules | ✅ Done | `docs/specs/phase-14/p14-t03-living-changelog.md` |
| **P9-T12** | 🚀 P1 | **Embedded SQLite (LibSQL) Local Caching** — Embed SQLite directly into the CLI and MCP Server to pull background graph updates, enabling sub-millisecond, zero-latency local schema diffs. | Jules | ✅ Done | `docs/specs/phase-9/p9-t12-embedded-sqlite.md` |
| **P10-T14** | 🚀 P1 | **Embedded JS Governance Rules (Goja)** — Support writing custom enterprise governance rules directly in `substrate.yaml` using JavaScript. Uses the `goja` pure-Go JS engine for secure, Turing-complete execution without compilation overhead. | Jules | ✅ Done | `docs/specs/phase-10/p10-t14-custom-governance.md` |
| **P14-T01** | 🚀 P1 | **AI Contract Negotiation** — Upgrade P7-T04 Auto-Fix PRs with an async team negotiation workflow. When a breaking change is detected, Substrate posts a structured GitHub comment tagging all affected consumer leads, tracks their approval/rejection reactions, and only turns the provider PR green when all consumers have acknowledged. | Jules | ✅ Done | `docs/specs/phase-14/p14-t01-contract-negotiation.md` |
| **P12-T08** | 🔴 P1 | **V2.0 Production Cutover** — Final pipeline updates to bundle Svelte static assets and WASM binary into the Go single-binary deployment. | Jules | ✅ Done | `docs/specs/phase-12/p12-t08-production-cutover.md` |
