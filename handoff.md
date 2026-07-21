# Substrate — Shift Handoff Document

> **Date:** 2026-07-20
> **Current Focus:** 3-Day Sprint — Crushing All 26 Remaining Backlog Tasks

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

6. **All Remaining Tasks Have Specs + Prompts**
   Every remaining backlog task now has a `docs/specs/` file and a `prompts/` file registered in `scripts/jules_submit.py`.

---

## 🌊 Full Wave Queue (All Pre-Planned)

> Tasks within each wave are **safe to run in parallel** — they touch different files and won't cause merge conflicts.
> Tasks that modify `github-app/src/formatter.ts` are limited to **one per wave** (marked ⚠️).






### Wave 11 — RUNNING (Phase 9+10 — 1 formatter.ts touch)
| Task ID | Jules ID |
|---|---|
| P15-T13 Phase 15 E2E | `2826767874429659759` |
| P15-T13 Phase 15 E2E | `2826767874429659759` |
| P14-T07 Phase 14 E2E | `1407` | New `scripts/e2e/` file ✓ |
| P15-T13 Phase 15 E2E | `1513` | New `scripts/e2e/` file ✓ |


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

### Wave 14 — (Phase 15 — On-Premise)
| Task ID | Script Key | Conflict Notes |
|---|---|---|
| P15-T15 Air-Gapped License | `1515` | New `engine/licensing/` pkg ✓ |
```bash
python3 scripts/jules_submit.py --task 1515
```

---

## 🤖 How to Submit Tasks to Jules

1. Ensure `JULES_API_KEY` is in `.env.local`.
2. List all tasks: `python3 scripts/jules_submit.py --list`
3. Trigger using the `bash` blocks above for each wave.
4. After Jules creates PRs, merge them sequentially (not all at once) and run `npm run check` + `go test ./...` after each merge.
