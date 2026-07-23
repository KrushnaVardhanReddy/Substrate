# P-MCP-01: Full MCP Parity — Tools, Resources, Prompts + SSE Transport

## 1. Overview

Substrate currently has two MCP servers:

| Server | Binary | Transport | Tools | Purpose |
|--------|--------|-----------|-------|---------|
| **Engine MCP** | `substrate-mcp` | stdio | 7 tools (diff, schema file, history, docs, analyze, compat, CLI) | Local IDE use (Cursor, Claude Desktop) |
| **API MCP** | `api/internal/mcp` embedded in API server | stdio only (via `--headless-mcp`) | 5 tools only | Headless registry ops in VPC |

**Decision: Keep them separate.** Engine MCP = local IDE assistant. API MCP = full headless registry control in VPC.

**Goal for this task:** Fill all gaps in the **API MCP server** so that 100% of Substrate registry operations can be performed headlessly, and add **SSE transport** to the API MCP so Cursor/Claude Desktop can also connect to a live running Substrate instance (not just the CLI binary).

---

## 2. Decisions Made

### Transport
Add **both transports** to `api/internal/mcp/server.go`:
- **stdio** (keep existing) — for `--headless-mcp` VPC deployments
- **SSE/HTTP** (new) — expose MCP over `GET /mcp/sse` + `POST /mcp/message` on the existing API server, so Cursor/Claude Desktop can connect with:
  ```json
  { "mcpServers": { "substrate": { "url": "http://substrate.mycompany.com/mcp/sse" } } }
  ```

### Auth on MCP
Use the **existing service token** (`REGISTRY_API_TOKEN`). MCP clients include it as:
```
Authorization: Bearer local-dev-token
```
The SSE endpoint validates this token using the existing `ServiceTokenMiddleware`. This is consistent with all other service-to-service calls — no new auth mechanism needed.

### Governance-rules resource stub
Wire `substrate://governance-rules` to real DB call: `store.ListGovernanceRules(ctx, org)`.

---

## 3. Missing Tools to Add (`api/internal/mcp/tools.go`)

### Group A — Core Registry Workflow (P1)
All of these already have underlying handlers/store methods — just need MCP wrappers.

| Tool Name | Calls | Input | Output |
|-----------|-------|-------|--------|
| `diff_schemas` | `handlers.SaveDiffHandler` (or direct store call) | `{org, provider_repo, schema_type, base_content, head_content}` | `{id, breaking_count, breaking_changes[]}` |
| `sync_schema` | `handlers.SyncHandler` (HTTP POST to self, or direct store) | `{org, provider_repo, consumer_repo, schema_type, commit_sha, raw_content}` | `{status, contract_id}` |
| `check_deploy` | `handlers.CanDeployHandler` (HTTP GET to self, or direct store) | `{repo, commit_sha}` | `{can_deploy: bool, blocked_by: []}` |
| `check_rollback` | `handlers.CanRollbackHandler` | `{repo, target_sha}` | `{can_rollback: bool, blocked_by: []}` |
| `get_dependency_graph` | `handlers.GraphHandler` / `store.GetDependencyGraph` | `{org}` | `[{consumer, provider, status}]` |
| `get_impact` | `handlers.ImpactHandler` / `store.GetImpact` | `{org, repo}` | `[{repo, status}]` |
| `get_diff_report` | `store.GetDiffReportByID` | `{id}` | `{report_data, org, created_at}` |
| `cross_repo_check` | `handlers.CrossRepoCheckHandler` | `{provider_repo, head_schema_content, schema_type}` | `{consumers: [{repo, status, breaking_changes}]}` |
| `trigger_webhook` | `webhook.PushHandler` | `{installation_id, org, repo, commit_sha, branch}` | `{status}` |

### Group B — Governance (P1)

| Tool Name | Calls | Input | Output |
|-----------|-------|-------|--------|
| `create_governance_rule` | `governanceHandler.CreateRule` | `{org, name, description, cel_expression, severity}` | `{id, status}` |
| `list_governance_rules` | `governanceHandler.ListRules` | `{org}` | `[{id, name, cel_expression, severity}]` |
| `delete_governance_rule` | `governanceHandler.DeleteRule` | `{org, rule_id}` | `{status}` |
| `generate_cel_rule` | `handlers.GenerateCELHandler` | `{plain_english_rule}` | `{cel_expression, explanation}` |

### Group C — Org Management (P2)

| Tool Name | Calls | Input | Output |
|-----------|-------|-------|--------|
| `register_webhook` | `handlers.RegisterWebhookHandler` | `{org, url, secret}` | `{status, id}` |
| `list_repos` | `handlers.ReposHandler` | `{org}` | `[{full_name, schema_type, status}]` |
| `get_schema` | `handlers.SchemaHandler` | `{owner, repo}` | `{raw_content, schema_type, commit_sha}` |
| `get_breaking_history` | `handlers.HistoryGetHandler` | `{org, repo, limit?}` | `[{commit_sha, breaking_changes, created_at}]` |
| `get_changes_feed` | `handlers.ChangesHandler` | `{}` | `[{org, repo, breaking_count, created_at}]` |
| `get_zombies` | `handlers.GetZombiesHandler` | `{org}` | `[{endpoint, last_seen, traffic_count}]` |
| `ingest_otel_metrics` | `handlers.OTelMetricsHandler` | `{org, repo, endpoint, request_count}` | `{status}` |

### Group D — FinOps & Telemetry (P2)

| Tool Name | Calls | Input | Output |
|-----------|-------|-------|--------|
| `predict_egress_cost` | `handlers.HandlePredictCost` | `{org, repo, breaking_changes[]}` | `{estimated_cost_usd, risk_tier}` |
| `get_roi_metrics` | `handlers.ROIHandler` | `{org}` | `{prevented_outages, hours_saved, dollars_saved}` |

### Group E — Insurance (P2)

| Tool Name | Calls | Input | Output |
|-----------|-------|-------|--------|
| `file_insurance_claim` | `handlers.InsuranceFileClaimHandler` | `{org, incident_date, repo, description}` | `{claim_id, status}` |
| `get_insurance_policy` | `handlers.InsuranceGetPolicyHandler` | `{org}` | `{tier, coverage_usd, active}` |
| `list_insurance_claims` | `handlers.InsuranceGetClaimsHandler` | `{org}` | `[{id, status, created_at}]` |

### Group F — Marketplace & Public (P3)

| Tool Name | Calls | Input | Output |
|-----------|-------|-------|--------|
| `publish_plugin` | `marketplace.PublishHandler` | `{name, description, cel_rules[], author}` | `{id, status}` |
| `list_plugins` | `marketplace.ListPluginsHandler` | `{}` | `[{name, description, author}]` |
| `ai_analyze_schema` | `services.AIAnalyzeHandler` | `{org, schema_type, current_schema, proposed_schema}` | `{findings[]}` |
| `get_public_profile` | `publicProfileHandler.GetPublicProfile` | `{org}` | `{breaking_changes, uptime, score}` |
| `get_changelog` | `handlers.ChangelogHandler` | `{org, repo}` | `[{version, changes, date}]` |

---

## 4. Missing Resources to Add (`api/internal/mcp/resources.go`)

| URI | Maps To | Description |
|-----|---------|-------------|
| `substrate://graph/{org}` | `store.GetDependencyGraph(org)` | Full dependency graph — agent reads before deciding to diff |
| `substrate://repos/{org}` | `store.GetReposByOrg(org)` | All registered repos for an org |
| `substrate://history/{org}/{repo}` | `store.GetBreakingChangeHistory(org, repo, 10)` | Breaking change log |
| `substrate://diff/{id}` | `store.GetDiffReportByID(id)` | Stored diff report |
| `substrate://zombies/{org}` | `store.GetZeroTrafficEndpoints(org)` | Zero-traffic endpoints |
| `substrate://roi/{org}` | `store.GetROIMetrics(org)` | ROI dashboard data |
| `substrate://insurance/{org}/policy` | `store.GetInsurancePolicy(org)` | Insurance coverage info |
| `substrate://profile/{org}` | `store.GetPublicProfile(org)` | Public reliability profile |

**Fix stub:** `substrate://governance-rules` currently returns `{"rules":[]}` — wire to real `store.ListGovernanceRules(ctx, org)`.  Since this resource has no `{org}` param, either (a) require the org in the URI (`substrate://governance-rules/{org}`) or (b) infer from the token. Use option (a).

---

## 5. Missing Prompts to Add (`api/internal/mcp/resources.go` `RegisterPrompts`)

| Prompt Name | Description |
|-------------|-------------|
| `substrate_diff_workflow` | Step-by-step: how to diff two schemas and interpret the output |
| `substrate_governance_setup` | How to create CEL governance rules for an org from scratch |
| `substrate_incident_response` | How to handle a production breaking-change incident via MCP: check impact → file postmortem → notify |
| `substrate_onboard_org` | Full org onboarding: register repos → sync schema → set up webhook → create first rule |
| `substrate_schema_smell` | How to detect and fix API design anti-patterns using Substrate |

---

## 6. SSE Transport (`api/internal/mcp/server.go` + `api/internal/server/router.go`)

Add two new routes to the router (behind `serviceTokenMW`):
```
GET  /mcp/sse      — SSE stream, sends JSON-RPC responses as SSE events
POST /mcp/message  — Accepts JSON-RPC requests, dispatches, responds via SSE
```

Wire these into the existing `Server.Handle(req)` dispatch method. No new dependencies — use `net/http` `http.ResponseWriter` flushing.

**Cursor/Claude Desktop config after this:**
```json
{
  "mcpServers": {
    "substrate": {
      "url": "http://your-substrate.internal/mcp/sse",
      "headers": { "Authorization": "Bearer your-service-token" }
    }
  }
}
```

---

## 7. E2E Test (`scripts/e2e/phase_mcp_e2e_test.go`) — NEW

Every new tool must be tested. File: `scripts/e2e/phase_mcp_e2e_test.go`

### Scenarios

| Scenario | Tool | What it asserts |
|----------|------|----------------|
| 1 | MCP initialize handshake | `tools/list` returns ≥ 20 tools after all additions |
| 2 | `diff_schemas` | Call via JSON-RPC → assert response has `breaking_count` field |
| 3 | `get_dependency_graph` | Seed DB → call tool → assert edges returned |
| 4 | `check_deploy` | Seed breaking change → call tool → assert `can_deploy: false` |
| 5 | `create_governance_rule` + `list_governance_rules` | Create CEL rule → list → assert it appears |
| 6 | `delete_governance_rule` | Delete rule created in Sc5 → list → assert gone |
| 7 | `sync_schema` + `get_schema` | Sync schema → read back via resource → assert content matches |
| 8 | SSE transport connectivity | HTTP GET `/mcp/sse` → assert `Content-Type: text/event-stream` → `POST /mcp/message` with `tools/list` → assert SSE event received |
| 9 | `file_insurance_claim` + `list_insurance_claims` | File claim → list → assert pending |
| 10 | `get_roi_metrics` | Seed telemetry → call tool → assert numeric fields |

---

## 8. Implementation Status
⏳ **NOT YET IMPLEMENTED** — created 2026-07-23
