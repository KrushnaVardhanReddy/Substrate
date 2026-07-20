# Substrate — Shift Handoff Document

> **Date:** 2026-07-20
> **Current Focus:** 3-Day Sprint — Crushing All 26 Remaining Backlog Tasks

## ✅ Completed This Session

1. **Wave 5 Merged (Phase 11 — UI/UX)**
   - ✅ P11-T01: Blast Radius (PR #166)
   - ✅ P11-T02: Team Neighborhoods (PR #165)
   - ✅ P11-T03: Edge Tooltips (PR #167)
   - ✅ P11-T04: Volatility Heatmap (PR #164)

2. **Wave 6 Running (Phase 15 — Enterprise)**
   - 🔄 P15-T03: Dependency SLA Tracking (Session: `11792690738245850426`)
   - 🔄 P15-T04: Schema Smell Detector (Session: `2377768681235846971`)
   - 🔄 P15-T05: AI Incident Post-Mortem Generator (Session: `7617790351664607704`)
   - 🔄 P15-T06: Natural Language Governance Rules (Session: `13121575222699399289`)

3. **All 26 Remaining Tasks Have Specs + Prompts**
   Every remaining backlog task now has a `docs/specs/` file and a `prompts/` file registered in `scripts/jules_submit.py`.

---

## 🌊 Full Wave Queue (All Pre-Planned)

> Tasks within each wave are **safe to run in parallel** — they touch different files and won't cause merge conflicts.
> Tasks that modify `github-app/src/formatter.ts` are limited to **one per wave** (marked ⚠️).

### Wave 6 — RUNNING (Phase 15 Enterprise)
| Task ID | Jules ID |
|---|---|
| P15-T03 Dependency SLA | `11792690738245850426` |
| P15-T04 Schema Smell | `2377768681235846971` |
| P15-T05 Post-Mortem Gen | `7617790351664607704` |
| P15-T06 NL Governance | `13121575222699399289` |

### Wave 7 — READY TO TRIGGER (Viral Growth + Enterprise)
| Task ID | Script Key | Conflict Notes |
|---|---|---|
| P15-T08 Marketplace | `1508` | All new files ✓ |
| P15-T10 Schema Insurance | `1510` | All new files ✓ |
| P14-T04 Contract Badge | `1404` | All new files ✓ |
| P14-T05 Archaeology CLI | `1405` | All new files ✓ |
```bash
python3 scripts/jules_submit.py --task 1508 && \
python3 scripts/jules_submit.py --task 1510 && \
python3 scripts/jules_submit.py --task 1404 && \
python3 scripts/jules_submit.py --task 1405
```

### Wave 8 — (Phase 9 — All New Files)
| Task ID | Script Key | Conflict Notes |
|---|---|---|
| P9-T02 AI Sync Watch | `902` | New `engine/watcher/` pkg ✓ |
| P9-T11 Deprecation Campaigns | `911` | New `engine/deprecation/` pkg ✓ |
| P9-T14 Preview URLs | `914` | New handler + public route ✓ |
| P9-T17 Phase 9 E2E | `917` | New `scripts/e2e/` file ✓ |
```bash
python3 scripts/jules_submit.py --task 902 && \
python3 scripts/jules_submit.py --task 911 && \
python3 scripts/jules_submit.py --task 914 && \
python3 scripts/jules_submit.py --task 917
```

### Wave 9 — (Phase 9+10 — 1 formatter.ts touch)
| Task ID | Script Key | Conflict Notes |
|---|---|---|
| P9-T03 Compliance Mapping | `903` | ⚠️ modifies `formatter.ts` |
| P10-T03 AI Mock Data Gen | `1003` | New `engine/mockgen/` pkg ✓ |
| P10-T06 Zombie Pruning | `1006` | New `engine/zombie/` + route ✓ |
| P15-T11 Startups Free Tier | `1511` | New public profile route ✓ |
```bash
python3 scripts/jules_submit.py --task 903 && \
python3 scripts/jules_submit.py --task 1003 && \
python3 scripts/jules_submit.py --task 1006 && \
python3 scripts/jules_submit.py --task 1511
```

### Wave 10 — (Phase 9+10 — 1 formatter.ts touch)
| Task ID | Script Key | Conflict Notes |
|---|---|---|
| P9-T08 Deployment Risk Score | `908` | ⚠️ modifies `formatter.ts` |
| P10-T04 AI Spectral Linter | `1004` | New `engine/spectral/` + routes ✓ |
| P10-T13 API Docs Catalog | `1013` | New handler + new Svelte route ✓ |
| P15-T09 Partner Program | `1509` | New handler + admin route ✓ |
```bash
python3 scripts/jules_submit.py --task 908 && \
python3 scripts/jules_submit.py --task 1004 && \
python3 scripts/jules_submit.py --task 1013 && \
python3 scripts/jules_submit.py --task 1509
```

### Wave 11 — (Phase 9+10 — 1 formatter.ts touch)
| Task ID | Script Key | Conflict Notes |
|---|---|---|
| P9-T09 AI Impact Analysis | `909` | ⚠️ modifies `formatter.ts` + `ai_analyze.go` |
| P10-T05 Auto-SDK Generator | `1005` | New `engine/sdkgen/` ✓ |
| P14-T07 Phase 14 E2E | `1407` | New `scripts/e2e/` file ✓ |
| P15-T13 Phase 15 E2E | `1513` | New `scripts/e2e/` file ✓ |
```bash
python3 scripts/jules_submit.py --task 909 && \
python3 scripts/jules_submit.py --task 1005 && \
python3 scripts/jules_submit.py --task 1407 && \
python3 scripts/jules_submit.py --task 1513
```

### Wave 12 — (Phase 9+10 — 1 formatter.ts touch)
| Task ID | Script Key | Conflict Notes |
|---|---|---|
| P9-T13 Mermaid Blast Radius | `913` | ⚠️ modifies `formatter.ts` |
| P9-T04 Quality Gates | `904` | New `engine/gates/` + `checks.ts` ✓ |
| P10-T20 Phase 10 E2E | `1020` | New `scripts/e2e/` file ✓ |
| P9-T06 Viper Config | `906` | Modifies all `os.Getenv` Go files |
```bash
python3 scripts/jules_submit.py --task 913 && \
python3 scripts/jules_submit.py --task 904 && \
python3 scripts/jules_submit.py --task 1020 && \
python3 scripts/jules_submit.py --task 906
```

### Wave 13 — (High-Risk Refactors — Trigger Last)
> ⚠️ These two tasks are the highest conflict risk. Trigger them together after all other waves are merged.

| Task ID | Script Key | Conflict Notes |
|---|---|---|
| P10-T19 CRM/Billing Blast Radius | `1019` | ⚠️ modifies `formatter.ts` |
| P9-T05 Hexagonal Architecture | `905` | Refactors many Go handler files |
```bash
python3 scripts/jules_submit.py --task 1019 && \
python3 scripts/jules_submit.py --task 905
```

---

## 🤖 How to Submit Tasks to Jules

1. Ensure `JULES_API_KEY` is in `.env.local`.
2. List all tasks: `python3 scripts/jules_submit.py --list`
3. Trigger using the `bash` blocks above for each wave.
4. After Jules creates PRs, merge them sequentially (not all at once) and run `npm run check` + `go test ./...` after each merge.
