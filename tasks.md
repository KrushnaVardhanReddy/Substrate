# Substrate — Task Tracker

> Last updated: 2026-07-07 (repo private, distribution model changed to Docker Hub public image, P1-MVP blog post deferred post-Phase 2)
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
| P1c-T01 | GraphQL SDL parser using `vektah/gqlparser` (MIT) | Jules | 🔒 |
| P1c-T02 | GraphQL breaking change rules spec (~15–20 rules) | Antigravity | 🔒 |
| P1c-T03 | GraphQL rule engine + tests | Jules | 🔒 |

### Phase 1d — Protobuf & gRPC 🔒 BLOCKED on Phase 2 🏆

> **Library decision:** **`bufbuild/buf`** (Apache 2.0, written in Go). `buf breaking` is the `oasdiff` of Protobuf — 100+ wire-compatibility rules, field number checks, service/method detection. Drop into `go.mod`, write a `buf` checker adapter (identical pattern to `oasdiff` P1-T06). Parser + rules both covered. This is the easiest Wave 2 phase.

| Task ID | Name | Owner | Status |
|---|---|---|---|
| P1d-T01 | `buf` checker adapter spec (maps `buf breaking` output → `DiffReport`) | Antigravity | 🔒 |
| P1d-T02 | `buf` checker adapter implementation + tests | Jules | 🔒 |

### Phase 1e — AsyncAPI & Apache Avro 🔒 BLOCKED on Phase 2

> **Library decision (AsyncAPI):** `asyncapi/parser-go` (Apache 2.0) for parsing. Custom diff rules (channel removed, operation changed, message schema breaking).
> **Library decision (Avro):** Do **not** build a custom Avro compatibility checker. Call the Confluent/Apicurio Schema Registry **compatibility check API** — offloads complex Avro evolution rules (union promotion, defaults, field ordering) to a battle-tested engine. Substrate wraps the JSON response into `DiffReport`.

| Task ID | Name | Owner | Status |
|---|---|---|---|
| P1e-T01 | AsyncAPI parser using `asyncapi/parser-go` | Jules | 🔒 |
| P1e-T02 | Avro adapter using Schema Registry compatibility API | Jules | 🔒 |
| P1e-T03 | AsyncAPI + Avro breaking change rules spec | Antigravity | 🔒 |
| P1e-T04 | Rule engine + tests | Jules | 🔒 |

### Phase 1f — AI/ML Model Contracts ⭐ 🔒 BLOCKED on Phase 2

> **Library decision:** No new dependencies needed. `gopkg.in/yaml.v3` (model contracts) and `santhosh-tekuri/jsonschema` (dataset schema contracts) are **already in `go.mod`**. Substrate owns the `substrate.yaml` model contract spec — we define the format, the diff is field-by-field rule comparison. Drop 3 parser tasks from the original plan.

| Task ID | Name | Owner | Status |
|---|---|---|---|
| P1f-T01 | `substrate.yaml` ML model contract spec (inputs, outputs, version constraints) | Antigravity | 🔒 |
| P1f-T02 | ML contract diff engine using `yaml.v3` + `jsonschema` (no new deps) | Jules | 🔒 |
| P1f-T03 | AI/ML rule engine + tests | Jules | 🔒 |

### Phase 1g — Enterprise Metadata (Salesforce & SOAP) 🔒 BLOCKED on Phase 2

> **Library decision (WSDL/SOAP):** Use Go stdlib `encoding/xml` to parse WSDL/XSD into a struct tree. Rule set is small (operation removed, message type changed, required element added) — fully buildable on stdlib. No external dependency.
> **Library decision (Salesforce):** Salesforce metadata files are XML snapshots (`*.object-meta.xml`, `*.field-meta.xml`). Parse with Go stdlib `encoding/xml`. Users retrieve snapshots via the Salesforce CLI (`sf project retrieve start`) as a **separate user step** — Substrate only receives and diffs two XML snapshot directories. No subprocess. **Single-binary promise maintained.**
> **Note:** This is the highest-effort Wave 2 phase — ships last. Both parsers use stdlib only.

| Task ID | Name | Owner | Status |
|---|---|---|---|
| P1g-T01 | SOAP/WSDL parser using Go stdlib `encoding/xml` | Jules | 🔒 |
| P1g-T02 | Salesforce metadata XML snapshot parser using Go stdlib `encoding/xml` | Jules | 🔒 |
| P1g-T03 | Enterprise breaking change rules spec | Antigravity | 🔒 |
| P1g-T04 | Enterprise rule engine + tests | Jules | 🔒 |

---

> **Wave 2 execution order (after Phase 2):** Follow effort order, not numerical order. Start with the easiest win to build momentum:
> `1d (Protobuf/buf — 2 tasks, near-zero custom code)` → `1c (GraphQL — custom rules on solved parser)` → `1f (AI/ML — no new deps)` → `1e (AsyncAPI/Avro — two formats, higher complexity)` → `1g (Enterprise — highest effort, ships last)`
> Order within this list may shift based on real user demand signals after Phase 2 launches.

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

## Phase 3 — Dashboard, API & MCP Server 🔄 IN PROGRESS

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

| Task ID | Name | Owner | Status |
|---|---|---|---|
| P3-T01 | PostgreSQL schema + Go API server | Jules | 🔄 |
| P3-T02 | `substrate.yaml` multi-contract parser + dependency registration | Jules | 🔄 |
| **P3-T02b** | **Contract registry** — store + sync consumer spec snapshots from registered repos | Jules | 🔄 |
| **P3-T02c** | **Cross-repo compatibility check** — on provider PR, validate change against all consumer snapshots | Antigravity | 🔄 |
| **P3-T02d** | **Cross-repo E2E fixture tests** — `testdata/cross-repo/` simulating multi-repo scenario | Jules | 🔄 |
| P3-T03 | SvelteKit project setup + design system | Antigravity | 💡 |
| P3-T04 | Connected repos list + schema browser | Antigravity + Jules | 💡 |
| P3-T05 | Dependency graph visualization | Antigravity | 💡 |
| P3-T06 | GitHub OAuth + org management | Jules | 💡 |
| P3-T07 | Free tier limits + production deployment | Jules | 💡 |
| **P3-T08** | **MCP server spec** (`docs/specs/mcp-server.md`) | Antigravity | 💡 |
| **P3-T09** | **MCP server implementation** (Go, exposes graph + schema tools) | Jules | 💡 |
| **P3-T10** | **MCP server deployment + IDE integration docs** | Antigravity | 💡 |
| **P3-T11** | **Compatibility matrix dashboard** — provider version × consumer version grid showing green/red compatibility status for all registered cross-repo contracts | Antigravity | 💡 |

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
| **P4a-T01** | **AI Impact Analyst** — "What changed in the payments API last month that could affect checkout?" | MCP tools → Claude/GPT reasons over schema graph | Antigravity | 💡 |
| **P4a-T02** | **Auto-Migration Generator** — breaking change detected → AI generates migration code diff for each affected consumer | MCP tools → foundation model generates code | Jules | 💡 |
| **P4a-T03** | **Smart Deprecation Planner** — repeatedly acknowledged break → AI suggests 30-day deprecation plan + consumer notifications | MCP tools → foundation model reasons over acknowledgment history | Antigravity | 💡 |
| **P4a-T04** | **AI-powered PR review assistant** — "This PR touches 5 schemas. Full cross-repo impact + suggested reviewers" | MCP tools → foundation model, triggered on PR open | Antigravity | 💡 |

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
- [ ] Write `docs/specs/mcp-server.md` spec (P3-T08)
- [ ] Implement Go MCP server with tools: `get_schema`, `check_impact`, `list_consumers`, `get_change_history`, `validate_change`
- [ ] Deploy MCP server endpoint
- [ ] Write IDE integration guide (Antigravity, Cursor, Claude, Copilot)
- [ ] Expose MCP server as self-hostable for Enterprise tier

**Legal / Pre-launch Checklist** ⚖️
- [x] Add `NOTICES` file at repo root with Apache 2.0 attribution for oasdiff
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
