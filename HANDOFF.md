# Substrate Project Handoff

**Date:** July 8, 2026 (End of Session)
**Current Phase:** Phase 3 (Contract Registry) & Phase 1d (Protobuf) 🔄 IN PROGRESS — 3 Jules sessions running.

---

## What Jules Is Working On Right Now

All 3 sessions were submitted to the `feature/dev` branch. Review PRs tomorrow morning.

| Jules Session | Task | What It Builds |
|---|---|---|
| `3925306055466858734` | P3-T01 | Go API Server Scaffold + PostgreSQL DB Schema for the Registry |
| `4407958795511618619` | P3-T02 | `substrate.yaml` Consumer Parser (reads `consumers:` block in config) |
| `15152674839790590105` | P1d-T01 | Protobuf & gRPC Diff Engine Adapter using `buf breaking` |

**These 3 tasks are fully parallel — no cross-dependencies. Once P3-T01 and P3-T02 are merged, the next tasks are P3-T02b (Sync) and P3-T02c (Cross-Repo Check).**

---

## PR Review Checklist (for Tomorrow)

For each Jules PR, verify:
- [ ] `go build ./...` + `go vet ./...` pass.
- [ ] All new tests are green (`go test ./... -v`).
- [ ] No files outside the FILES LIST were touched.
- [ ] Commit message starts with `jules: ` prefix.
- [ ] **P3-T01:** Ensure handlers use dependency injection (DB pool), not global state.
- [ ] **P1d-T01:** Ensure `buf` is invoked via `exec.Command` and not imported as a Go module.

---

## What We Completed Today

1. **Phase 3 Launched:** Contract Registry spec approved. Task priorities set (P1 to P5).
2. **Phase 1d Launched:** Protobuf `buf` adapter spec approved. 
3. **Automated Submission Fixed:** Updated `jules_submit.py` to correctly use `feature/dev` as the default target branch.
4. **Go-To-Market Strategy Documented:** Created `docs/specs/go-to-market-strategy.md` outlining:
   - "Zero-Touch Onboarding" (GitHub App auto-PR)
   - "AI Spec Generation" (CLI `substrate init --ai` generating `openapi.yaml`)
   - "AI API Architect" (`substrate design` for new projects / Spec-First approach)
   - SaaS vs Enterprise (BYO AI) Security Models.

---

## Next Steps (After Jules PRs Are Merged)

**Tomorrow (priority order):**
1. Review and merge Jules PRs for P3-T01, P3-T02, and P1d-T01.
2. Draft and submit `P3-T02b` Jules prompt: The Sync logic (saving schemas to Postgres on push to main).
3. Draft and submit `P3-T02c` Jules prompt: The Cross-Repo logic (querying DB and running diff on PR).

**Future Vision Tasks (Backlog):**
- Prototyping the `substrate design` Conversational AI CLI.
- SvelteKit Dashboard UI build.

---

## Key File Locations

| File | Purpose |
|---|---|
| `docs/specs/phase-3/contract-registry.md` | Phase 3 Registry Spec (READ-ONLY) |
| `docs/specs/phase-1/protobuf-checker-adapter.md` | Phase 1d Protobuf Spec (READ-ONLY) |
| `docs/specs/go-to-market-strategy.md` | Business Strategy & AI Onboarding Vision |
| `tasks.md` | Full roadmap with priorities and Jules Session IDs |
| `scripts/jules_submit.py` | Run `python3 scripts/jules_submit.py --list` to see available prompts |
