# Phase 4 Spec: AI Intelligence Layer

> **Status:** 📐 SCAFFOLDED — Ready for implementation  
> **Depends on:** Phase 3 Contract Registry ✅ Complete  
> **Goal:** Substrate evolves from a **breaking-change detector** into an **autonomous schema intelligence agent** — one that explains, remediates, and even auto-fixes breaking changes before they ever reach production.

---

## The Core Shift

| Phase 3 | Phase 4 |
|---|---|
| *"You broke the Billing API."* | *"Here's why, here's the impact, and here's the fix — already PR'd."* |
| Detects the problem | Solves the problem |
| Registry of schemas | Intelligence over schemas |
| Static analysis | Contextual AI reasoning |

Phase 4 is what transforms Substrate from a CI/CD tool into a **developer co-pilot**. It is the moat.

---

## Priority Tiers

| Priority | Tasks | Why |
|---|---|---|
| 🔴 **P1 — Core AI** | P4-T01, P4-T02 | The LLM bridge. Zero AI value without this. |
| 🟡 **P2 — Auto-Fix** | P4-T03, P4-T04 | The flagship feature. Differentiation from every competitor. |
| 🟢 **P3 — Playground** | P4-T05 | Wires real AI to the P3-T16 Playground UI shell. |
| 🔵 **P4 — IDE** | P4-T06 | Shift-Left: Catch breaks before CI. |
| ⚪ **P5 — Enterprise** | P4-T07, P4-T08 | Traffic-aware diffing, PII auditing, compliance mapping. |

---

## P4-T01: AI Reasoning Bridge (MCP Tool Execution)

**Owner:** Antigravity  
**Effort:** ~1 day

### Overview
Wire **any OpenAI-compatible language model** to Substrate's existing MCP server tools (`get_schema_file`, `analyze_breaking_changes`, `get_breaking_change_history`). The Go client uses the OpenAI Chat Completions API format — the same interface exposed by LM Studio, Ollama, Gemini, Claude, and OpenRouter. No vendor lock-in.

### Architecture
```
Developer Input (schema pair)
        ↓
  [Go API: POST /api/v1/ai/analyze]
        ↓
  OpenAI-compatible Client (provider-agnostic)
  ├── LOCAL:  LM Studio → Qwen2.5-Coder-7B  (default, free, no API key)
  ├── CLOUD:  OpenRouter → any model        (pay-per-token, production)
  └── SELF:   Ollama → any model            (CLI-first alternative)
  + Tool: analyze_breaking_changes (MCP)
  + Tool: get_breaking_change_history (MCP)
        ↓
  Streamed AI Response (Server-Sent Events)
        ↓
  Dashboard Playground / IDE Extension
```

### Why Qwen2.5-Coder for Local Dev?
- Runs fully offline on a laptop (7B = ~5GB VRAM, also works CPU-only).
- Fine-tuned on code and structured data — ideal for YAML/JSON/SQL schema diffs.
- Exposed by LM Studio as a local OpenAI-compatible endpoint (`http://localhost:1234/v1`).
- Zero API cost during development and E2E testing.

### Provider Configuration (env vars)

| Environment | `SUBSTRATE_AI_BASE_URL` | `SUBSTRATE_AI_API_KEY` | `SUBSTRATE_AI_MODEL` |
|---|---|---|---|
| **Local (LM Studio)** | `http://localhost:1234/v1` | `lm-studio` (ignored) | `qwen2.5-coder-7b-instruct` |
| **Local (Ollama)** | `http://localhost:11434/v1` | `ollama` (ignored) | `qwen2.5-coder:7b` |
| **Production (OpenRouter)** | `https://openrouter.ai/api/v1` | `sk-or-...` | `qwen/qwen-2.5-coder-32b-instruct` |
| **Enterprise (self-hosted)** | `http://internal-llm/v1` | internal token | any compatible model |

### API Endpoint
```
POST /api/v1/ai/analyze
Content-Type: application/json

{
  "org": "my-org",
  "current_schema": "<OpenAPI/GraphQL/SQL text>",
  "proposed_schema": "<OpenAPI/GraphQL/SQL text>",
  "schema_type": "openapi" | "graphql" | "sql"
}

→ Response: text/event-stream (Server-Sent Events)
  data: {"type": "thinking", "content": "Analyzing cross-repo impact..."}
  data: {"type": "finding", "severity": "BREAKING", "message": "..."}
  data: {"type": "fix", "language": "yaml", "code": "..."}
  data: {"type": "done"}
```

### LLM System Prompt
```
You are Substrate AI, an expert in API schema governance.
You have access to tools to check cross-repo contract impacts.
When given a schema diff:
1. Call analyze_breaking_changes to get the raw diff.
2. Call get_breaking_change_history to understand past patterns.
3. Summarize findings in plain English.
4. Suggest the minimal safe remediation (deprecate instead of delete, etc).
5. Generate the exact corrected schema as a code block.
Never hallucinate. Ground all claims in tool outputs.
```

---

## P4-T02: Streaming SSE Response Handler (Go)

**Owner:** Jules  
**Effort:** ~4 hours

### Overview
Implement the `POST /api/v1/ai/analyze` Go handler that streams LLM tokens back to the client using Server-Sent Events (SSE). This decouples the LLM latency from the frontend UX.

### Implementation Notes
- Use `http.Flusher` interface in Go for streaming.
- Read provider config from three env vars: `SUBSTRATE_AI_BASE_URL`, `SUBSTRATE_AI_API_KEY`, `SUBSTRATE_AI_MODEL`.
- **Local dev default:** LM Studio at `http://localhost:1234/v1` with `qwen2.5-coder-7b-instruct` — free, offline, no API key.
- **Production default:** Google Gemini Flash via `https://generativelanguage.googleapis.com/v1beta/openai/` with `GEMINI_API_KEY`.
- Fall back to a deterministic mock response if `SUBSTRATE_AI_BASE_URL` is not set (for CI pipelines with no LLM).
- **Stream Parsing**: The LLM output stream should be piped to the client as `thinking` events. Once the LLM finishes, the handler must parse the accumulated text to extract the `severity` (e.g., matching BREAKING or SAFE) and the exact remediation code (by matching markdown fenced code blocks). The handler must then emit the final `finding` and `fix` events using this extracted data immediately before emitting the `done` event.
- Add the route to `api/internal/server/router.go`.

### Files
- `api/internal/handlers/ai_analyze.go` (CREATE)
- `api/internal/server/router.go` (MODIFY)
- `api/internal/handlers/ai_analyze_test.go` (CREATE)

---

## P4-T03: Safe Schema Patch Generator (Provider-Side)

**Owner:** Jules  
**Effort:** ~1 day

### Overview
Instead of attempting complex consumer-side refactors, Substrate will generate the correct **provider-side schema fix** to avoid breaking changes. If a developer deletes `user_id`, the AI suggests adding `@deprecated` instead.

### Workflow
```
1. Provider pushes PR (e.g., deletes `user_id` from Payments API)
2. Phase 3: Substrate blocks the PR. Breaking change detected.
3. Phase 4 (NEW): Substrate calls P4-T01 AI bridge for auto-fix.
4. AI generates the exact safe schema patch (e.g., restoring `user_id` as deprecated).
5. Substrate outputs this code block in the PR comment for the developer to apply.
```

### API Changes
- New Go handler: `POST /api/v1/ai/autofix`
- Input: `{ provider_org, provider_repo, breaking_change_report }`
- Output: `{ explanation: "...", safe_schema_patch: "..." }`

### Files
- `api/internal/handlers/ai_autofix.go` (CREATE)
- `engine/internal/autofix/generator.go` (CREATE)

---

## P4-T04: GitHub PR Comment Upgrade (AI-Enhanced)

**Owner:** Jules  
**Effort:** ~2 hours

### Overview
Upgrade the existing PR comment formatter (built in Phase 2) to include the Phase 4 AI analysis. The PR comment should now include:
- The standard breaking change table (Phase 2).
- A new **AI Impact Analysis** section with the plain-English explanation.
- A **"Suggested Safe Remediation"** section with a copy-pasteable schema patch (P4-T03).

### PR Comment Template (Additions)
```markdown
---
### 🤖 AI Impact Analysis
> Removing `user_id` will break the `BillingService` consumer in 2 downstream repos.
> **Root Cause:** `BillingService v2.3` uses `user_id` for invoice correlation.

### 🔧 Suggested Safe Remediation
To unblock this PR safely, apply the following schema change:
```yaml
# Restore the field but mark it deprecated
properties:
  user_id:
    type: string
    deprecated: true
```

---

## P4-T05: Wire Real AI to Playground UI

**Owner:** Antigravity  
**Effort:** ~2 hours

### Overview
Connect the P3-T16 Playground UI shell (which uses mock/simulated responses) to the real `POST /api/v1/ai/analyze` SSE endpoint built in P4-T01/P4-T02.

### Changes
- Replace the simulated `setTimeout` mock in `playground/+page.svelte` with a real `fetch` to the SSE endpoint.
- Use the `EventSource` API (or `fetch` with `ReadableStream`) to consume the streamed tokens and update the UI in real-time as tokens arrive.
- **Add Schema Type Dropdown:** Introduce a `<select>` bound to `schemaType` covering all supported formats (OpenAPI, GraphQL, SQL, Protobuf, AsyncAPI, Terraform).
- **Interactive Templates:** When the schema type changes, auto-fill the `currentSchema` and `proposedSchema` text areas with realistic breaking-change examples (e.g., deleting an OpenAPI endpoint, dropping a SQL column, or removing a Protobuf field) to demonstrate the AI's multi-format capabilities.
- The Playground becomes a fully functional product demo.

---

## P4-T06: Shift-Left IDE Extension (VSCode)

**Owner:** Jules  
**Effort:** ~2 days

### Overview
A VSCode extension that uses the Substrate MCP server to catch breaking changes **before the developer even commits**. This is the "SonarLint for API schemas" moment.

### Behaviour
- On save of `openapi.yaml`, `schema.graphql`, or `schema.sql`, the extension:
  1. Compares the file against the last committed version.
  2. Calls the local `substrate` binary (or the MCP server) for a diff.
  3. If breaking changes are detected, underlines the changed section in **red** in the editor.
  4. Shows an inline error tooltip: *"⚠️ Deleting `user_id` will break `BillingService`. Click to view consumers."*

### Key Files
- `vscode-extension/` (CREATE — new top-level directory)
- `vscode-extension/src/extension.ts`
- `vscode-extension/src/substrate-client.ts`
- `vscode-extension/package.json`
- `vscode-extension/README.md`

---

## P4-T07: Traffic-Aware Diffing (Zero False Positives)

**Owner:** Jules  
**Effort:** ~1 day

### Overview
Integrate with OpenTelemetry/Prometheus to eliminate false positives. If a breaking change targets an endpoint/field with **0 traffic in the last 30 days**, auto-downgrade severity from `BREAKING` to `WARNING (Unused)`.

### API Contract
```go
// New interface for traffic providers
type TrafficProvider interface {
    GetFieldUsage(org, repo, fieldPath string, days int) (int64, error)
}

// Implementations
type OTelTrafficProvider struct { ... }  // OpenTelemetry Collector
type PrometheusTrafficProvider struct { ... } // Prometheus query
type MockProvider struct { ... } // E2E testing (always returns 0)
type NoOpTrafficProvider struct { ... }  // Default: no data = assume used
```

### Config (`substrate.yaml`)
```yaml
traffic:
  provider: prometheus
  endpoint: "http://prometheus.internal:9090"
  lookback_days: 30
  downgrade_threshold: 0  # 0 calls = downgrade to WARNING
```

---

## P4-T08: PII & Compliance Auditing

**Owner:** Antigravity  
**Effort:** ~1 day  

### Overview
Substrate scans schema diffs for fields that match known PII, financial, or health data patterns and auto-tags them with compliance labels.

### PII Detection Rules
```go
var piiPatterns = map[string]string{
    `(?i)(ssn|social.?security)`:    "PII:SSN",
    `(?i)(password|passwd|secret)`:  "SECURITY:CREDENTIAL",
    `(?i)(card.?number|pan|cvv)`:    "PCI:PAYMENT",
    `(?i)(medical|diagnosis|hipaa)`: "HIPAA:PHI",
    `(?i)(email|phone|address)`:     "PII:CONTACT",
}
```

### Behaviour
- The compliance auditor scans BOTH the `change.Path` and `change.Description` strings for the regex patterns. This ensures detection even for nested fields where adapters (like oasdiff) might report the deep field name in the description rather than the path.
- If a match is found in a PR, Substrate:
  1. Adds a `[PII]` or `[HIPAA]` tag to the field in the registry.
  2. Posts a compliance alert in the PR comment.
  3. Notifies the security team channel (Slack/Teams webhook).

---

## Milestone Summary

| Task | Feature | Est. Effort | Status |
|---|---|---|---|
| **P4-T01** | AI Reasoning Bridge (MCP + LLM) | 1 day | ✅ Done |
| **P4-T02** | Streaming SSE Handler (Go) | 4 hrs | ✅ Done |
| **P4-T03** | Safe Schema Patch Generator | 1 day | ✅ Done |
| **P4-T04** | GitHub PR Comment Upgrade | 2 hrs | ✅ Done |
| **P4-T05** | Wire Real AI to Playground | 2 hrs | ✅ Done |
| **P4-T06** | Shift-Left VSCode Extension | 2 days | ✅ Done |
| **P4-T07** | Traffic-Aware Diffing | 1 day | ✅ Done |
| **P4-T08** | PII & Compliance Auditing | 1 day | ✅ Done |

**Total Estimated Phase 4 Duration:** ~1–2 weeks (with Jules parallelization)

---

## Env Variables Added in Phase 4

```bash
# ── LOCAL DEV: LM Studio + Qwen2.5-Coder (free, offline, no API key needed) ──
SUBSTRATE_AI_BASE_URL=http://localhost:1234/v1
SUBSTRATE_AI_API_KEY=lm-studio          # ignored by LM Studio, just needs a value
SUBSTRATE_AI_MODEL=qwen2.5-coder-7b-instruct

# ── PRODUCTION: Google Gemini Flash ───────────────────────────────────────────
# SUBSTRATE_AI_BASE_URL=https://generativelanguage.googleapis.com/v1beta/openai/
# SUBSTRATE_AI_API_KEY=your-gemini-api-key   # from console.cloud.google.com
# SUBSTRATE_AI_MODEL=gemini-2.0-flash

# ── ALTERNATIVE PROD: OpenRouter (multi-model, pay-per-token) ─────────────────
# SUBSTRATE_AI_BASE_URL=https://openrouter.ai/api/v1
# SUBSTRATE_AI_API_KEY=sk-or-your-key
# SUBSTRATE_AI_MODEL=qwen/qwen-2.5-coder-32b-instruct

# ── Traffic Integration (optional, Phase 4-T07) ───────────────────────────────
SUBSTRATE_PROMETHEUS_ENDPOINT=http://prometheus.internal:9090

# ── PII Alerting (optional, Phase 4-T08) ──────────────────────────────────────
SUBSTRATE_SLACK_WEBHOOK=https://hooks.slack.com/...
SUBSTRATE_SECURITY_CHANNEL=#security-alerts
```
