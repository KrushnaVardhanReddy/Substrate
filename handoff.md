# Substrate — Shift Handoff Document

> **Date:** 2026-07-20
> **Current Focus:** 3-Day Sprint to Crush Backlog (Wave 5 & Wave 6)

## 🚀 Accomplishments from This Session

1. **Wave 5 Tasks Triggered & Merged (Phase 11 - UI/UX)**
   We have successfully submitted, validated, and merged the following tasks from Jules:
   - ✅ **P11-T01: Blast Radius** (PR #166)
   - ✅ **P11-T02: Team Neighborhoods** (PR #165)
   - ✅ **P11-T03: Edge Tooltips** (PR #167)
   - ✅ **P11-T04: Volatility Heatmap** (PR #164)

2. **Wave 6 Prepped (Phase 15 - Enterprise)**
   We created the specs and prompts for the next batch (Wave 6) and added them to `scripts/jules_submit.py`:
   - 📝 `P15-T03` Dependency SLA Tracking (Task `1503`)
   - 📝 `P15-T04` Schema Smell Detector (Task `1504`)
   - 📝 `P15-T05` AI Incident Post-Mortem Generator (Task `1505`)
   - 📝 `P15-T06` Natural Language Governance Rules (Task `1506`)

---

## 🎯 Next Steps for Tomorrow

1. **Trigger Wave 6 (Phase 15):**
   Wave 5 is completely merged. Next step is to trigger the Wave 6 tasks using the Jules submission script.

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
