# Substrate Handoff

## Current Status (Mid-Day)
* **Merged Feature Branches:** We successfully reviewed, validated, and merged high-priority feature branches into `feature/dev`:
    * P12-T16: Advanced E2E UI Implementation (TDD)
    * P10-T16: Webhook Auto-Discovery Integration
    * P9-T10: MCP Server WASI Distribution
    * P13-T03: AI Chaos Engineering Auto-Tests
    * P9-T12: Embedded SQLite (LibSQL) Local Caching (CLI)
    * P12-T09: AI Support Copilot (UI)
    * P10-T12: Tree-sitter Deterministic Impact Analysis (Engine)
    * P13-T01: FinOps Cost Prediction (API)
    * P10-T02: API Gateway Auto-Sync (Engine)
    * P13-T02: DB Performance Breakages Analytics (Engine/API)
    * P10-T09: Native Protobuf and gRPC parsing (Engine)
    * *Note: The core API, engine, and E2E UI tests (after fixing the Can-Deploy stub) are all passing.*
* **New Specs Created:** Created new markdown specs and updated `openapi.yaml` to strictly align the API schema with the frontend implementation for the Impact and Telemetry APIs.

* **New Tasks Dispatched:** We just fired off another massive wave of Enterprise expansion tasks to Jules via `jules_submit.py`:
    * **P9-T01:** Shift-Left IDE Plugins — *(✅ Done & Merged)*
    * **P13-T04:** Enterprise E2E Validation — *(Running)*
    * **P9-T00:** GitHub Marketplace Launch (WASI) — *(✅ Done & Merged)*
    * **P14-T03:** Living API Changelog — *(✅ Done & Merged)*
    * **P14-T02:** "Time to Break" Predictive Scoring — *(✅ Done & Merged)*

## Next Steps
1. **Monitor Jules PRs:** Track the progress of P9-T01.
2. **Review TDD Stubs:** The UI for "Can-Deploy / Can-Rollback" was just implemented to pass the TDD E2E tests. Make sure the backend logic mapping (`docs/specs/phase-9/p9-t15-impact-api.md`) actually aligns with the final architecture.
3. **Plan Phase 10 Finale:** We only have a few tasks left to complete Phase 10 (e.g. eBPF drift detection). Start drafting the specs for them!
4. **Trigger P10-T14:** Now that the massive merges are done, we can trigger `python3 scripts/jules_submit.py --file prompts/phase-10/t14_custom_governance.txt` safely!

* **Note:** The `tasks.md` tracker has been cleaned up. Completed tasks from phases 9-13 have been moved to `completed_tasks.md` for better readability.

## Quick Start Reminders
* **Run Tests:** `make test-all`
* **Start Forgejo locally:** Run `make forgejo` (Starts the container on port 3000).
* **Start Backend Stack:** Run `make start-bg`.
* **View Graph:** Go to `http://localhost:5173/org/<forgejo-username>/graph`.

## 🧪 Comprehensive Demo Testing Guide
* **Axis 1 (Breakages):** Test schema breaking changes (e.g., removing a field) vs safe additions.
* **Axis 2 (Governance):** Test pushing a breaking change with an `overrides` block in `substrate.yaml` (`rule_id: "*"`) to verify the "Acknowledged" (Amber) path.
