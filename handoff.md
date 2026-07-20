# Substrate — Shift Handoff Document

> **Date:** 2026-07-20
> **Current Focus:** 3-Day Sprint to Crush Backlog (Wave 5 & Wave 6)

## 🚀 Accomplishments from This Session

1. **Wave 5 Tasks Triggered & Merged (Phase 11 - UI/UX)**
   We have successfully submitted, validated, and merged the following tasks from Jules:
   - ✅ **P11-T01: Blast Radius** (PR #166)
   - ✅ **P11-T02: Team Neighborhoods** (PR #165)
2. **Wave 6 Tasks Triggered (Phase 15 - Enterprise)**
   We have dispatched the next batch (Wave 6) to Jules. Keep an eye out for these PRs:
   - 🔄 **P15-T03: Dependency SLA Tracking** (Session: `11792690738245850426`)
   - 🔄 **P15-T04: Schema Smell Detector** (Session: `2377768681235846971`)
   - 🔄 **P15-T05: AI Incident Post-Mortem Generator** (Session: `7617790351664607704`)
   - 🔄 **P15-T06: Natural Language Governance Rules** (Session: `13121575222699399289`)

---

## 🎯 Next Steps for Tomorrow

1. **Review Wave 6 PRs:**
   Wave 6 is currently running in Jules. Monitor GitHub for the incoming PRs and validate their implementations.

2. **Continue 3-Day Sprint:**
   Select the next 4 tasks (Wave 7) from the backlog, generate their specs and prompts, and trigger them.

---

## 🤖 How to Submit Tasks to Jules

Substrate uses a batch submission script (`scripts/jules_submit.py`) to dispatch AI agent coding sessions.

### Step-by-Step Submission:
1. **Ensure your API key is set:** Your `.env.local` must contain `JULES_API_KEY`.
2. **List available tasks:**
   ```bash
   python3 scripts/jules_submit.py --list
   ```
3. **Trigger a specific task:** (e.g. Task `1503` for Dependency SLA)
   ```bash
   python3 scripts/jules_submit.py --task 1503
   ```
4. **Trigger multiple tasks concurrently:**
   ```bash
   python3 scripts/jules_submit.py --task 1503 && \
   python3 scripts/jules_submit.py --task 1504 && \
   python3 scripts/jules_submit.py --task 1505 && \
   python3 scripts/jules_submit.py --task 1506
   ```

### Adding New Tasks to the Script:
When preparing a new wave:
1. Create the `docs/specs/...md` and `prompts/...txt` files.
2. Edit `scripts/jules_submit.py` and add a new entry to the `TASKS` dictionary, linking to the new prompt file.
3. Commit and push the changes, then use the commands above to submit.
