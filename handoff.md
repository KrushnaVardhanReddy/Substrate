# Substrate Handoff & Session Plan

## Current State

- **System E2E Status:** 🟩 100% Passing (Phases 16, 17, 18 all green).
- **Active Branch:** `fix/dependency-graph-filtering-5385298189871839623`
- **Just Pushed:** Phase 19 specs, prompts, tasks.md update, and jules_submit.py Phase 19 task entries.
- **Next Phase:** Phase 19 — Universal Data Ingestion (Airbyte Integration).

---

## Phase 19 Jules Submission Plan

> Jules will branch off `fix/dependency-graph-filtering-5385298189871839623` and raise individual PRs per task.

### ⚠️ Dependency Order — DO NOT submit all at once

Phase 19 tasks have strict dependencies. Submit in waves:

---

### 🌊 Wave 1 — Submit Now (blocks everything else)

| Task | Jules ID | Command | Why first |
|---|---|---|---|
| **P19-T01** — Airbyte Ingestion Adapter | `1901` | `python3 scripts/jules_submit.py --task 1901 --branch current` | Creates `AirbyteStore` port interface, DB schema, and ingest handler. T02 and T03 both import this port. |

**Wait for T01 PR to be merged before Wave 2.**

---

### 🌊 Wave 2 — Submit in Parallel (after T01 merges)

These three tasks touch **completely different files** — zero merge conflict risk.

| Task | Jules ID | Command | Files touched |
|---|---|---|---|
| **P19-T02** — Stream Insurance | `1902` | `python3 scripts/jules_submit.py --task 1902 --branch current` | `services/stream_validator.go`, modifies `airbyte_handler.go` (adds async dispatch) |
| **P19-T03** — MCP Tools | `1903` | `python3 scripts/jules_submit.py --task 1903 --branch current` | `mcp/server.go` (adds 4 tools), `mcp/airbyte_tools_test.go` |
| **P19-T04** — Sync Dashboard | `1904` | `python3 scripts/jules_submit.py --task 1904 --branch current` | Pure frontend — `dashboard/src/routes/`, `dashboard/src/lib/components/` |

> ⚠️ **T02 and T03 both modify `airbyte_handler.go` / `mcp/server.go`** — they won't conflict because they touch different sections, but review PRs carefully before merging. Merge T02 first, then T03, to be safe.

---

### 🌊 Wave 3 — Submit after T01 + T02 + T03 merge

| Task | Jules ID | Command | Why last |
|---|---|---|---|
| **P19-T99** — E2E Validation | `1999` | `python3 scripts/jules_submit.py --task 1999 --branch current` | Tests the full stack — needs real handlers, validator, and MCP tools in place. |

---

## Active Jules Sessions

| Task | Session ID | Status |
|---|---|---|
| **P19-T01** — Airbyte Ingestion Adapter | `12733402407044387061` | 🤖 Submitted |

---

## How to Execute E2E Tests

### Full Suite (Backend + Frontend)
```bash
export GITHUB_TOKEN=dummy; make e2e
```
*(Output logged to `scripts/e2e/e2e_test.logs`)*

### Phase 19 Only (after T99 merges)
```bash
cd scripts/e2e && go test -v -p 1 -run TestPhase19 ./...
```

### Specific Playwright Test
```bash
cd dashboard
npx playwright test tests/e2e/org-context.spec.ts --project=chromium --headed
```
