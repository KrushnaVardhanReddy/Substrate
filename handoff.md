# Substrate — Shift Handoff Document

> **Date:** 2026-07-25
> **Current Focus:** Fixing Dependency Graph Frontend Rendering (Svelte 5 + Cytoscape)

## ✅ Completed This Session

1. **Dependency Graph Rendering Engine Migration**
   - ✅ **Issue:** The dependency graph on the frontend (`/org/[org]/graph`) was completely blank. This was caused by `SvelteFlow` v1.6.2 throwing `lifecycle_outside_component` context tracker errors in Svelte 5, especially when combined with async data loading in SvelteKit's declarative routing.
   - ✅ **Fix:** We ripped out the incompatible `SvelteFlow` library and reverted the rendering engine back to **Cytoscape** and **cytoscape-dagre**, which the project originally used 10 days ago.
   - ✅ **Integration:** Integrated Cytoscape directly with the new Server-Sent Events (SSE) logic natively in Svelte 5 using the `$effect` rune.
   - ✅ **Layout:** Fixed CSS layout overlap bugs by setting the `cyContainer` flex wrapper to `flex: 1` rather than absolute positioning, allowing it to seamlessly fit below the header.
   - ✅ **Styling:** Updated the node padding and sizing styles (`width: 'auto'`, `height: 'auto'`) to conform to modern Cytoscape standards (preventing text bleed and deprecation warnings).
   - ✅ **Clean up:** Removed the manual `getLayoutedElements` DAGRE math function (since Cytoscape computes layout natively), cleaned up all `local-dev-token` hardcoding, and purged all temporary API proxy endpoints/debug files created during the debugging session.

### 1. Vendor PGlite & Wrapper Script
- We will download the PGlite WASM binaries into the repository.
- Jules will write a lightweight Node.js wrapper (`scripts/e2e/pglite_server.js`) that spins up PGlite on a local TCP port and seeds it.

### 2. The Full-Stack Tear-Up Pipeline
- The runner (`scripts/e2e/run_full_e2e.sh`) will spin up the PGlite Database, the Go Backend API, the MCP Server, and the SvelteKit Frontend in the background.
- It will execute the Go API tests, the MCP tests, and Playwright UI tests in a single, offline CI pass.
- **Status:** Prompt created and submitted to Jules (`CC-T02`).

### 3. E2E Phase 1 (Contract Registry)
- Re-wrote the Contract Registry prompt to use the new full-stack harness (Go API + Playwright UI).
- **Status:** Prompt created and submitted to Jules (`P3-T12`).

---

## 🗺️ E2E Full-Stack Roadmap (The 15 Prompts)

We have successfully submitted the first four phases. Phase 3 (Discovery Scanners) has been successfully merged! The remaining phases have been fully generated and wired up in `scripts/jules_submit.py` for batch submission (3 sessions at a time):

1. ✅ **Phase 0:** PGlite Infrastructure (CC-T02) -> *Submitted* (`--task 0`)
2. ✅ **Phase 1:** Contract Registry (Phase 3) -> *Submitted* (`--task 1`)
3. ✅ **Phase 2:** AI Diff Engine & Analyzers (Phase 4) -> *Submitted* (`--task 2`)
4. ✅ **Phase 3:** Discovery Scanners (Phase 5) -> *Merged!* (`--task 3`)
5. 📝 **Phase 4:** QA & Shadow API (Phase 6) -> *Generated* (`--task 60`)
6. 📝 **Phase 5:** Enterprise Rollout & Drift (Phase 7) -> *Generated* (`--task 70`)
7. 📝 **Phase 6:** Readiness, Authz & Jobs (Phase 8) -> *Generated* (`--task 80`)
8. 📝 **Phase 7:** Compliance & Risk Scoring (Phase 9) -> *Generated* (`--task 90`)
9. 📝 **Phase 8:** Ecosystem & Zombie Pruning (Phase 10) -> *Generated* (`--task 100`)
10. 📝 **Phase 9:** Graph UI & Visual Studio (Phase 11) -> *Generated* (`--task 110`)
11. 📝 **Phase 10:** SSE Boundaries & Scaling (Phase 12) -> *Generated* (`--task 120`)
12. 📝 **Phase 11:** God Mode & FinOps (Phase 13) -> *Generated* (`--task 130`)
13. 📝 **Phase 12:** Predictive Intelligence (Phase 14) -> *Generated* (`--task 140`)
14. 📝 **Phase 13:** Monetization & Insurance (Phase 15) -> *Generated* (`--task 150`)
15. 📝 **Phase 14:** MCP Parity & Cross-Cutting -> *Generated* (`--task 990`)
16. 📝 **Application Phase 1:** Core Diff Engine -> *Generated* (`--task 10`)
17. 📝 **Application Phase 2:** GitHub App -> *Generated* (`--task 20`)

---

## 🔁 Next Steps for the Evening Shift

1. ✅ **Verify Offline PGlite Execution:** Pulled the PRs and fixed River Queue & Svelte Playwright timeout issues. `scripts/e2e/run_full_e2e.sh` is now 100% green!
2. **Execute 3-Session Parallel E2E Batching:** We can now trigger the remaining phases (Phases 6-15, plus MCP parity) in batches of 3 using `jules_submit.py`.
