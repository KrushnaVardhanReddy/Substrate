# Substrate Project Handoff

## 📌 Current State
- **Branch:** `feature/dev`
- **Latest Fixes:** 
  - **Database Collision Fix:** Resolved the PostgreSQL unique constraint collision where discovered providers (without a `github_repo_id`) were overwriting each other. Implemented a `crc32` hashing pseudo-ID in `api/internal/webhook/handler.go`. 100-repo scale simulation now successfully populates the DB with 100% accuracy.
  - **Graph UX Spec Update:** Formalized the "Two-Step Exploration" UI flow in `ui-t03-graph-filtering.md` (Strict Isolation + Click to Explore "Blast Radius" + Neighbor Checkbox).

## 🤖 Active Jules Sessions (2)
1. **`V1-T07 — V1.0 System E2E Tests`**
   - **Status:** Pending PR
   - **Note:** Jules encountered a Docker-in-Docker layer extraction permission issue (whiteout files) within their sandbox. We instructed them to simply verify compilation and push the code so we can run the live Postgres E2E tests locally on our end.
2. **`UI-T03 — Graph Filtering & Navigation`** 
   - **Session ID:** `2522520351791071517`
   - **Status:** In Progress
   - **Note:** Jules is implementing the newly specced Two-Step UX Flow (Highlight matching nodes only, click to explore subgraph, with a neighbor override toggle) in Svelte/Cytoscape.

## 🚀 Next Steps for Next Session
1. **Review Jules PRs:**
   - Pull down the `V1-T07` branch and run the System E2E tests against our local `make postgres` instance to validate their scripts.
   - Pull down the `UI-T03` branch, run the dashboard (`npm run dev`), and manually verify the new two-step graph filtering UX behaves elegantly.
2. **V1.0 Pre-Flight:**
   - Once these two PRs are merged, continue burning down the remaining `V1.0 Pre-Flight Checklist` (e.g., GitHub App Auto-Discovery, CLI AI Architect, Legal Audit) located in `tasks.md`.
3. **Database TRUNCATE Reminder:**
   - Before running new E2E tests, remember to clear stale graph data using `PGPASSWORD=postgres psql -h localhost -U postgres -d substrate -c "TRUNCATE TABLE dependencies CASCADE; TRUNCATE TABLE contracts CASCADE; TRUNCATE TABLE repositories CASCADE;"`.
