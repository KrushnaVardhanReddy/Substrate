# Substrate — Shift Handoff Document

> **Date:** 2026-07-23
> **Current Focus:** E2E Coverage Audit + sqlc Migration

## ✅ Completed This Session

1. **Wave 5 Merged (Phase 11 — UI/UX)**
   - ✅ P11-T01: Blast Radius (PR #166)
   - ✅ P11-T02: Team Neighborhoods (PR #165)
   - ✅ P11-T03: Edge Tooltips (PR #167)
   - ✅ P11-T04: Volatility Heatmap (PR #164)

2. **Wave 6 Merged (Phase 15 — Enterprise)**
   - ✅ P15-T03: Dependency SLA Tracking (PR #170)
   - ✅ P15-T04: Schema Smell Detector (PR #171)
   - ✅ P15-T05: AI Incident Post-Mortem Generator (PR #168)
   - ✅ P15-T06: Natural Language Governance Rules (PR #169)

3. **Wave 7 Merged (Viral Growth + Enterprise)**
   - ✅ P14-T04: Contract Score Badge (PR #173)
   - ✅ P14-T05: Retroactive Dependency Archaeology (PR #175)
   - ✅ P15-T08: Substrate Marketplace (PR #174)
   - ✅ P15-T10: Schema Insurance (PR #172)

4. **Wave 8 Merged (Phase 9 — All New Files)**
   - ✅ P9-T02: AI Sync Watch (PR #178)
   - ✅ P9-T11: Deprecation Campaigns (PR #177)
   - ✅ P9-T14: Preview URLs (PR #179)
   - ✅ P9-T17: Phase 9 E2E (PR #176)

5. **Wave 9 Merged (Compliance & Zombies & Free Tier)**
   - ✅ P9-T03: Compliance Mapping (PR #182)
   - ✅ P10-T03: AI Mock Data Gen (PR #180)
   - ✅ P10-T06: Zombie Pruning (PR #181)
   - ✅ P15-T11: Startups Free Tier (PR #183)

6. **Wave 10 Merged**
   - ✅ P9-T08: Deployment Risk Score (PR #184)
   - ✅ P10-T04: AI Spectral Linter (PR #185)
   - ✅ P10-T13: API Docs Catalog (PR #186)
   - ✅ P15-T09: Partner Program (PR #187)

7. **sqlc Pipeline Restored (2026-07-23)**
   - ✅ Fixed `sqlc.yaml` — schema path corrected to `api/migrations`
   - ✅ Fixed all `.sql` query files (table name drifts, ambiguous columns, duplicate query names)
   - ✅ sqlc now generates to `api/internal/db/sqlcgen/` (separate package to avoid struct conflicts)
   - ✅ `pgstore.go` — `UpsertEndpointTraffic`, `GetZeroTrafficEndpoints`, `GetROIMetrics` now use sqlc-generated code
   - ✅ `api/` builds clean: `go build ./...` passes

8. **E2E Coverage Audit (2026-07-23)**
   - ✅ Full audit of all 15 phases — identified ~78% backend API coverage
   - ✅ Confirmed: Phases 6–10, 12–15 all green
   - ⚠️ Gap 1: **Phase 11 backend** (Graph, Impact, SSE, Diff) — zero coverage
   - ⚠️ Gap 2: **Phase 3 Contract Registry** (Graph/Can-Deploy/Repos) — only indirectly tested
   - ✅ Specs written: `docs/specs/phase-11/p11-t18-e2e-validation.md`, `docs/specs/phase-3/p3-t12-e2e-validation.md`
   - ✅ Prompts written: `prompts/phase-11/t18_e2e_validation.txt`, `prompts/phase-3-registry/t12_e2e_validation.txt`
   - ✅ Both tasks registered in `scripts/jules_submit.py` (IDs: 1118, 312)
   - ✅ Both submitted to Jules (sessions below)

---

## 🌊 Full Wave Queue (All Pre-Planned)

> Tasks within each wave are **safe to run in parallel** — they touch different files and won't cause merge conflicts.
> Tasks that modify `github-app/src/formatter.ts` are limited to **one per wave** (marked ⚠️).

### Wave 11 — RUNNING (Phase 9+10 — 1 formatter.ts touch)
| Task ID | Jules ID |
|---|---|
| P15-T13 Phase 15 E2E | `2826767874429659759` |
| P14-T07 Phase 14 E2E | `1407` — New `scripts/e2e/` file ✓ |
| P15-T13 Phase 15 E2E | `1513` — New `scripts/e2e/` file ✓ |

### Wave 12 — RUNNING (Phase 9+10 — 1 formatter.ts touch)
| Task ID | Jules ID |
|---|---|
| P9-T13 Mermaid Blast Radius | `9797585847433362378` |
| P9-T04 Quality Gates | `17462911419978641490` |
| P10-T20 Phase 10 E2E | `8954309360691061006` |
| P9-T06 Viper Config | `5081940009452583220` |

### Wave 13 — RUNNING (High-Risk Refactors — Trigger Last)
> ⚠️ These two tasks are the highest conflict risk. Trigger them together after all other waves are merged.

| Task ID | Jules ID |
|---|---|
| P10-T19 CRM/Billing Blast Radius | `1883281273610373932` |
| P9-T05 Hexagonal Architecture | `17783096233236153624` |

### Wave 14 — (Phase 15 — On-Premise)
| Task ID | Script Key | Conflict Notes |
|---|---|---|
| P15-T15 Air-Gapped License | `1515` | New `engine/licensing/` pkg ✓ |
```bash
python3 scripts/jules_submit.py --task 1515
```

### Wave 15 — E2E Coverage Gaps (NEW — Submitted 2026-07-23)
> These close the two backend E2E gaps found in the coverage audit.
> Both touch only `scripts/e2e/` — zero conflict risk. Safe to run in parallel.

| Task ID | Script Key | Jules Session | Files Created |
|---|---|---|---|
| P11-T18 Phase 11 Backend E2E | `1118` | `11577987586214929019` | `scripts/e2e/phase11_e2e_test.go` |
| P3-T12 Phase 3 Registry E2E | `312` | `5507342138131598347` | `scripts/e2e/phase3_e2e_test.go` |

```bash
# Already submitted — monitor at https://jules.google.com/
# To resubmit if needed:
python3 scripts/jules_submit.py --task 1118
python3 scripts/jules_submit.py --task 312
```

### Wave 16 — MCP Full Parity (NEW — Submitted 2026-07-23)
> Complete API MCP server parity with 28 new tools and SSE transport.

| Task ID | Script Key | Jules Session | Files Created/Edited |
|---|---|---|---|
| P-MCP-01 Full MCP Parity | `901` | `2715584438838468656` | `api/internal/mcp/*`, `scripts/e2e/phase_mcp_e2e_test.go` |

```bash
# Already submitted — monitor at https://jules.google.com/
# To resubmit if needed:
python3 scripts/jules_submit.py --task 901
```

---

## 🤖 How to Submit Tasks to Jules

1. Ensure `JULES_API_KEY` is in `.env.local`.
2. List all tasks: `python3 scripts/jules_submit.py --list`
3. Trigger using the `bash` blocks above for each wave.
4. After Jules creates PRs, merge them sequentially (not all at once) and run `npm run check` + `go test ./...` after each merge.

---

## 🔁 After Wave 15 PRs Merge

1. Update `tasks.md`: flip P11-T18 and P3-T12 from `⏳ Ready` → `✅ PR Merged`
2. Update E2E Coverage Tracker table in `tasks.md`: Phase 3 and Phase 11 rows → `✅ Green`
3. Run full E2E suite to confirm 100% pass:
   ```bash
   export GITHUB_TOKEN=mock_token
   cd scripts/e2e && go test -v -p 1 -run "." ./...
   ```
4. Coverage target: **~90%+** after both PRs merge (up from 78%).
