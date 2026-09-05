# Substrate — Task Tracker

> Last updated: 2026-08-01 (Phase 19 specs, prompts, and Jules tasks added. ✅)
> Tracking all development phases, tasks, and their current status.

---

## Legend

| Symbol | Status |
|---|---|
| 🔄 | In Progress |
| 🤖 | Submitted to Jules (PR pending) |
| ⏳ | Ready to Start (all dependencies met) |
| 🔒 | Blocked (waiting on dependency) |
| 💡 | Planned (not yet started) |

---


> **Note:** Tasks for Phases 0 through 6 have been archived to [completed_tasks.md](./completed_tasks.md)

> **Note:** Phases 3, 9, 10, 11, 12, and 13 are fully complete. See [completed_tasks.md](./completed_tasks.md) for their full task lists.

## 🔌 Phase 19: Universal Data Ingestion (Airbyte Integration)

**Goal:** Accelerate Substrate's data ingestion capabilities by leveraging the open-source Airbyte ecosystem to automatically pipe CRM, database, and third-party API data directly into Substrate's "Schema Insurance" and "CRM Blast Radius" pipelines without writing custom connectors.

> **Strategic framing:** Substrate acts as an Airbyte *destination* (consumer), not an orchestration host. Enterprise upsell: denominate blast radius in real MRR from any of Airbyte's 300+ connectors.
> **Base branch:** `fix/dependency-graph-filtering-5385298189871839623`

| Task ID | Tier | Name & Description | Owner | Jules ID | Status | Spec Link |
|---|---|---|---|---|---|---|
| **P19-T01** | 🚀 P1 | **Airbyte Ingestion Adapter (sqlc)** — Webhook destination endpoint, CRUD source configs, bulk staging insert, KMS config encryption, hexagonal `AirbyteStore` port. **⚠️ Wave 1 — submit first, blocks T02 & T03.** | Jules | `1901` | 🤖 Submitted | `docs/specs/phase-19/p19-t01-airbyte-adapter.md` |
| **P19-T02** | 🟡 P2 | **Dynamic Schema Insurance for Ingested Streams** — Async stream validator hooked into ingest handler. Violations → `insurance_claims`. **Wave 2 — parallel with T03 & T04, after T01 merges.** | Jules | `1902` | 🔒 Blocked (T01) | `docs/specs/phase-19/p19-t02-schema-insurance-streams.md` |
| **P19-T03** | 🔴 P1 | **MCP Tooling for Airbyte Configs** — 4 MCP tools (`configure_airbyte_source`, `trigger_airbyte_sync`, `list_airbyte_sources`, `get_airbyte_violations`). Config fields auto-redacted. **Wave 2 — parallel with T02 & T04.** | Jules | `1903` | 🔒 Blocked (T01) | `docs/specs/phase-19/p19-t03-mcp-airbyte-tools.md` |
| **P19-T04** | 🟢 P3 | **Airbyte Sync Status Dashboard** — Svelte 5 `/org/{org}/settings/integrations` page, SSE violation feed, source card components. **Wave 2 — pure frontend, zero backend conflict.** | Jules | `1904` | 🔒 Blocked (T01) | `docs/specs/phase-19/p19-t04-sync-status-dashboard.md` |
| **P19-T99** | 🔴 P1 | **Phase 19 E2E Validation (No Mocks)** — Source lifecycle, ingestion + violation detection, MCP tool + redaction assertions. **Wave 3 — after T01+T02+T03 merged.** | Jules | `1999` | 🔒 Blocked (T01+T02+T03) | `docs/specs/phase-19/p19-t99-e2e.md` |

---

## 📡 Phase 20: OTel/APM Ingest (Traffic-Aware Blast Radius)

**Goal:** Accept OTLP/HTTP payloads from any OTel-compatible APM (Datadog, Grafana, Prometheus) and enrich blast radius with real traffic weight — turning theoretical risk into quantitative req/min scores.

> **Strategic framing:** Makes every existing Substrate metric quantitative. No new feature needed — blast radius becomes "12k req/min at P99 87ms" instead of just "consumer exists".
> **Base branch:** `fix/dependency-graph-filtering-5385298189871839623`

| Task ID | Tier | Name & Description | Owner | Jules ID | Status | Spec Link |
|---|---|---|---|---|---|---|
| **P20-T01** | 🚀 P1 | **OTel OTLP Receiver** — OTLP/HTTP endpoint accepting metrics + traces. Aggregates into per-route traffic windows. Enriches `GET /impact` with `req_per_min`, `p99_ms`, `risk` fields. **Wave 1 — foundational, blocks T02/T03.** | Jules | `2001` | ⏳ Ready | `docs/specs/phase-20/p20-t01-otel-receiver.md` |
| **P20-T02** | 🟢 P3 | **Traffic Observability Dashboard** — Svelte 5 `/org/{org}/observability` page. OTel source provisioning, route traffic charts, live SSE feed. Graph node glow by traffic intensity. **Wave 2 — pure frontend.** | Jules | `2002` | 🔒 Blocked (T01) | `docs/specs/phase-20/p20-t02-traffic-dashboard.md` |
| **P20-T03** | 🔴 P1 | **MCP Tools for OTel Traffic** — 3 MCP tools: `get_route_traffic`, `get_top_traffic_routes`, `get_traffic_enriched_blast_radius`. **Wave 2 — parallel with T02.** | Jules | `2003` | 🔒 Blocked (T01) | `docs/specs/phase-20/p20-t03-mcp-otel-tools.md` |
| **P20-T99** | 🔴 P1 | **Phase 20 E2E Validation (No Mocks)** — OTLP JSON ingestion, traffic-enriched blast radius assertion, MCP parity, auth rejection. **Wave 3.** | Jules | `2099` | 🔒 Blocked (T01+T03) | `docs/specs/phase-20/p20-t99-e2e.md` |

---

## 🎵 Phase 21: Kafka / Confluent Schema Registry Governance

**Goal:** Receive schema change webhooks from Confluent Schema Registry and Apicurio, auto-detect breaking changes using the existing diff engine, and govern the entire event-driven Kafka stack — a market segment no competitor covers today.

> **Strategic framing:** Substrate already parses Avro + Protobuf (Phase 1d/1e). This adds live webhook triggering — zero new diff logic needed, just a receiver + event pipeline.
> **Base branch:** `fix/dependency-graph-filtering-5385298189871839623`

| Task ID | Tier | Name & Description | Owner | Jules ID | Status | Spec Link |
|---|---|---|---|---|---|---|
| **P21-T01** | 🚀 P1 | **Kafka Schema Registry Webhook Receiver** — Accept Confluent + Apicurio (CloudEvents) payloads. Diff against previous version using existing Avro/Protobuf adapters. Breaking changes → `events` table + SSE broadcast. **Wave 1 — foundational.** | Jules | `2101` | ⏳ Ready | `docs/specs/phase-21/p21-t01-kafka-schema-registry.md` |
| **P21-T02** | 🟢 P3 | **Kafka Schema Governance Dashboard** — Svelte 5 `/org/{org}/kafka` page. Source management, subject table with version history drawer, live breaking-change SSE feed. **Wave 2 — pure frontend.** | Jules | `2102` | 🔒 Blocked (T01) | `docs/specs/phase-21/p21-t02-kafka-dashboard.md` |
| **P21-T03** | 🔴 P1 | **MCP Tools for Kafka Governance** — 3 MCP tools: `list_kafka_subjects`, `get_kafka_schema_diff`, `get_kafka_breaking_events`. **Wave 2 — parallel with T02.** | Jules | `2103` | 🔒 Blocked (T01) | `docs/specs/phase-21/p21-t03-mcp-kafka-tools.md` |
| **P21-T99** | 🔴 P1 | **Phase 21 E2E Validation (No Mocks)** — Source lifecycle, Avro schema ingestion, breaking change detection via real diff engine, MCP parity. **Wave 3.** | Jules | `2199` | 🔒 Blocked (T01+T03) | `docs/specs/phase-21/p21-t99-e2e.md` |

---

## 🧪 E2E Coverage Tracker

> **Purpose:** Track which phases have E2E test coverage. As a solo developer, E2E is the primary safety net against production regressions.
>
> 📊 **Current coverage: ~78% of backend API surface area.**

| Phase | Test File | Coverage | Priority | Status |
|-------|-----------|----------|----------|--------|
| V1 (Phases 1–5 core) | `scripts/e2e/v1_e2e_test.go` | ✅ Webhook, Diff, Deploy Gate, AI Scaffold | — | ✅ Green |
| Phase 3 (Contract Registry) | `scripts/e2e/phase3_e2e_test.go` | ✅ Full | — | ✅ Green |
| Phase 6 (QA Feedback) | `scripts/e2e/phase6_e2e_test.go` | ✅ Full | — | ✅ Green |
| Phase 7 (Enterprise) | `scripts/e2e/phase7_e2e_test.go` | ✅ Audit Mode, CEL Rules, Drift, AI Autofix | — | ✅ Green |
| Phase 8 (Readiness) | `scripts/e2e/phase8_e2e_test.go` | ✅ Job Queue, RBAC(1 route), Rollback, Billing | ⚠️ RBAC breadth | ✅ Green (partial) |
| Phase 9 (DX) | `scripts/e2e/phase9_e2e_test.go` | ✅ Compliance, Quality Gates, Watch Daemon | — | ✅ Green |
| Phase 10 (Ecosystem) | `scripts/e2e/phase10_e2e_test.go` | ✅ OTel, Zombies, Governance, SDK | Sc3 SKIP (Forgejo) | ✅ Green |
| Phase 10+13 (Enterprise) | `scripts/e2e/phase10_13_e2e_test.go` | ✅ Protobuf, FinOps, TreeSitter, Gateway | — | ✅ Green |
| Phase 11 (Graph/SSE) | `scripts/e2e/phase11_e2e_test.go` | ✅ Graph, Impact, SSE, Diff retrieval | — | ✅ Green |
| Phase 12 (SSE/WASM) | `scripts/e2e/phase12_sse_test.go` + `wasm_boundary_test.go` | ✅ SSE Broker, WASM Boundary, Fuzz | — | ✅ Green |
| Phase 13 (Enterprise) | `scripts/e2e/phase10_13_e2e_test.go` | ✅ FinOps, Protobuf, TreeSitter | — | ✅ Green |
| Phase 14 (Predictive) | `scripts/e2e/phase14_e2e_test.go` | ✅ Badge, Archaeology, Negotiation | — | ✅ Green |
| Phase 15 (Ecosystem) | `scripts/e2e/phase15_e2e_test.go` | ✅ NL Governance, Marketplace, Insurance | — | ✅ Green |
| History/Sync/RBAC (Cross-cutting) | `phase_history...`, `phase_sync...`, `phase8...` | ✅ History, Sync, Schema, RBAC breadth | 🔴 P1 | ✅ Green |
| UI — Diff Viewer / Impact / Governance | `diff-viewer`, `impact-page`, `governance` | ✅ `/diff/[id]`, `/impact`, `/governance` | 🔴 P1 | ✅ Green |

---

## 🚨 E2E Gap Tasks (Full Audit — 2026-07-23)

> Full route-by-route audit of all 39 backend routes + 22 frontend pages. Gaps ranked by production risk.

### 🔴 Priority 1 — Backend API Gaps

| Task ID | Name | Routes Covered | Spec | Status |
|---------|------|----------------|------|--------|
| **P1-T09** | Phase 1f/1g E2E Validation | AI/ML and Salesforce schema adapters | `docs/specs/e2e/phase1f-1g-adapters.md` | ✅ Complete |
| **P11-T18** | Phase 11 Backend E2E | `GET /api/v1/graph/{org}`, `GET /api/v1/impact/{org}/{repo}`, `GET /api/v1/events`, `GET /api/v1/diff/{id}` | `docs/specs/phase-11/p11-t18-e2e-validation.md` | ✅ PR Merged |
| **P3-T12** | Phase 3 Registry E2E | `GET /api/v1/graph/{org}`, `GET /api/v1/registry/can-deploy`, `GET /api/v1/repos/{org}` | `docs/specs/phase-3/p3-t12-e2e-validation.md` | ✅ Complete |
| **P-HIST-01** | History & Changes API | `POST /api/v1/history`, `GET /api/v1/history/{org}/{repo}`, `GET /api/v1/changes` | `docs/specs/cross-cutting/p-hist-01-e2e-history-api.md` | ✅ Complete |
| **P-SYNC-01** | Contract Sync Pipeline | `POST /api/v1/sync`, `GET /api/v1/schema/{owner}/{repo}`, `GET /api/v1/spec/{org}/{repo}` | `docs/specs/cross-cutting/p-sync-01-e2e-sync-pipeline.md` | ✅ Complete |
| **P-RBAC-01** | RBAC Breadth | All `authzMW` write routes — rules, insurance, zombies/pr, partners | `docs/specs/cross-cutting/p-rbac-01-e2e-rbac-breadth.md` | ✅ Complete |
| **P-MCP-01** | Full MCP Server Parity | 28 Tools, 8 Resources, 5 Prompts + SSE Transport | `docs/specs/cross-cutting/p-mcp-01-full-mcp-parity.md` | ✅ Complete |
| **P-MCP-02** | MCP HTTP SSE Transport | Implement the missing HTTP Server-Sent Events Transport for MCP. | `docs/specs/mcp/mcp-http-transport.md` | ✅ Complete |

### 🟡 Priority 2 — Backend API Gaps

| Task ID | Name | Routes Covered | Spec | Status |
|---------|------|----------------|------|--------|
| **P-ROI-01** | ROI + FinOps E2E | `GET /api/v1/telemetry/roi/{org}`, `POST /api/v1/finops/predict` | `docs/specs/cross-cutting/p-roi-01-e2e-finops.md` | ✅ Complete |
| **P-INS-01** | Insurance Lifecycle | `GET /api/v1/org/{org}/insurance/policy`, `GET /api/v1/org/{org}/insurance/claims` | `docs/specs/cross-cutting/p-ins-01-e2e-insurance.md` | ✅ Complete |

### 🔴 Priority 1 — Frontend UI Gaps (Playwright)

| Task ID | Page | Risk | Spec | Status |
|---------|------|------|------|--------|
| **P-UI-DIFF** | `/diff/[id]` — Diff Viewer | Core product page — blank on regression | `docs/specs/cross-cutting/p-ui-01-e2e-missing-pages.md` | ✅ Complete |
| **P-UI-IMPACT** | `/org/{org}/repo/[repo]/impact` — Impact Page | Blast radius feature | `docs/specs/cross-cutting/p-ui-02-e2e-impact.md` | ✅ Complete |
| **P-UI-GOV** | `/org/{org}/governance` — Governance Rules | Rule CRUD UI | `docs/specs/cross-cutting/p-ui-03-e2e-gov.md` | ✅ Complete |

### 🟡 Priority 2 — Frontend UI Gaps (Playwright)

| Task ID | Page | Status |
|---------|------|--------|
| **P-UI-ZOMBIE** | `/org/{org}/zombies` | ✅ Complete |
| **P-UI-INS** | `/org/{org}/settings/insurance` | ✅ Complete |
| **P-UI-PREVIEW** | `(public)/preview/[token]` | ✅ Complete |

### 🟢 Priority 3 — CLI Command Gaps

| Task ID | Commands | Status |
|---------|----------|--------|
| **P-CLI-POST** | `substrate postmortem`, `substrate schema-smell`, `substrate plugin publish` | ✅ Complete |
| **P-PART-01** | Partners CRUD (`/api/v1/org/{org}/partners` full lifecycle) | ✅ Complete |

### 🟣 Priority 4 — Frontend Unit Testing Gaps (Vitest)

| Task ID | Component/Page | Status |
|---------|----------------|--------|
| **P-UNIT-01** | Backfill Svelte unit tests for QA Dashboard and API Keys pages (19 missing files) | ✅ PR Merged |

---

### E2E Quick Commands
```bash
# Run all E2E tests (requires API + Postgres running)
export GITHUB_TOKEN=mock_token && cd scripts/e2e && go test -v -p 1 -run "." ./...

# Run a specific phase
cd scripts/e2e && go test -v -p 1 -run TestPhase11SystemE2E ./...
cd scripts/e2e && go test -v -p 1 -run TestPhase3ContractRegistry ./...

# Run Playwright UI tests
cd dashboard && npx playwright test

# Start API server
cd api && DATABASE_URL="postgresql://postgres:postgres@localhost:5432/substrate?sslmode=disable" \
  REGISTRY_API_TOKEN="local-dev-token" INTERNAL_SERVICE_TOKEN="local-dev-token" \
  JWT_SECRET="local-jwt-secret" GITHUB_CLIENT_ID="mock-client-id" \
  GITHUB_CLIENT_SECRET="mock-client-secret" DASHBOARD_URL="http://localhost:5173" \
  ENVIRONMENT="development" go run ./cmd/server/main.go
```

---

## 🧪 Final Manual QA Checklist

Once all automated E2E tests are complete, these four manual workflows must be verified to clear Substrate for production:

- `[ ]` **1. GitHub App (Real-World Webhooks)**
  - Install the Substrate GitHub App on a live test repository.
  - Push a breaking OpenAPI change.
  - Verify the Substrate Bot successfully posts a Markdown comment on the PR containing the correct blast radius formatting.
- `[ ]` **2. VS Code Extension & UI Feel**
  - Verify the Cytoscape dependency graph renders correctly *inside* the VS Code extension webview panel.
  - Open the SvelteKit Dashboard in a browser and load the 1,000-node scale test to ensure the UI remains snappy and responsive.
- `[x]` **3. MCP Server in Claude/Cursor**
  - Add the `substrate-mcp` server to your `claude_desktop_config.json`.
  - Ask Claude an impact analysis question (e.g., *"What happens if I delete the email field from the Payments API?"*).
  - Verify Claude correctly invokes the `get_blast_radius` tool and hallucinates nothing.
- `[x]` **4. Zero-Config Scaffold (`substrate init`)**
  - Run `substrate init` in a fresh, empty directory.
  - Verify the generated `substrate.yaml` and `.github/workflows/substrate.yml` are perfectly formatted and intuitive.

---

## 🚀 Future Enhancements (Post-Demo)
- `[ ]` **Support for Arbitrary/Custom Metadata Tags**
  - Update `Metadata` struct in `api/internal/config/config.go` to support custom inline maps (e.g., `yaml:",inline"`).
  - Allow enterprise users to pass proprietary fields (like `support_group`, `team_dl`, or arbitrary `tags`) without being dropped by the YAML unmarshaler.
  - Ensure the MCP tools and UI properly render these custom keys.
