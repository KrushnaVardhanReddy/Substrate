# Substrate Project Handoff

**Date:** July 8, 2026 (End of Session)
**Current Phase:** Phase 2 (GitHub App) 🔄 IN PROGRESS — 3 Jules sessions running.

---

## What Jules Is Working On Right Now

All 3 sessions submitted to `feature/dev` branch on 2026-07-08. Review PRs tomorrow morning.

| Jules Session | Task | What It Builds |
|---|---|---|
| `332461586460335586` | P2-T01 | Cloudflare Worker scaffold + GitHub App webhook receiver + HMAC validation |
| `12378420499826865949` | P2-T02a | `substrate serve --port 8080` HTTP mode on the Go binary |
| `9046194793485505468` | P2-T03 | PR comment formatter TypeScript module (DiffReport → markdown) |

**These 3 tasks are fully parallel — no cross-dependencies. Once all 3 PRs are merged, the next task is P2-T02b (wire them together).**

---

## PR Review Checklist (for Tomorrow)

For each Jules PR, verify:
- [ ] `tsc --noEmit` passes with zero errors (TypeScript tasks)
- [ ] `go build ./...` + `go vet ./...` pass (Go tasks)
- [ ] All new tests are green
- [ ] No files outside the FILES LIST were touched
- [ ] Commit message starts with `jules: ` prefix
- [ ] No files in `docs/specs/` were modified

---

## What We Completed Today (July 7–8 Session)

1. **Docker Hub distribution fixed and live** — v0.1.6 ships via automated release CI. Dockerfile fixed (gcc + git + CGO). Auth scopes corrected.
2. **Phase 2 architecture designed** — all 5 open questions resolved (container service, personal account, per-repo+org scope, configurable block/warn, setup guide comment).
3. **Phase 2 spec written** — `docs/specs/github-app.md` approved. Full architecture including HMAC flow, Container Service HTTP contract, PR comment templates, commit status logic.
4. **3 Phase 2 Jules prompts written** — all following t06 quality standard: MANDATORY RULES at top, CONTEXT section, DELIVERABLES per file, FILES LIST with READ-ONLY gates.
5. **Roadmap extended** — Phase 4 added (4a: MCP + foundation models, 4b: specialized ML). 4 Pact-inspired features added (P1-T09, P2-T06, P2-T07, P3-T11). Phase 4 hybrid architecture documented.
6. **HN launch task** — moved from Phase 1 backlog to P2-T08 (end of Phase 2). Right timing.

---

## Next Steps (After Jules PRs Are Reviewed)

**Tomorrow (priority order):**
1. Review and merge Jules PRs for P2-T01, P2-T02a, P2-T03
2. Write and submit `P2-T02b` Jules prompt — wires Worker → Container → GitHub APIs
3. Once T02b is done, write `P2-T04` Jules prompt (status check integration)

**After Phase 2 ships:**
4. Register the GitHub App at `github.com/settings/apps/new` under `kpakkiragari` (manual step)
5. Set Cloudflare Workers secrets: `GITHUB_APP_ID`, `GITHUB_APP_PRIVATE_KEY`, `GITHUB_WEBHOOK_SECRET`
6. Deploy the Container Service (Cloudflare Containers or Fly.io)
7. HN launch post (P2-T08)

---

## Architecture Decisions Locked

| Decision | Choice | Rationale |
|---|---|---|
| Binary execution model | Option C: Container Service (HTTP) | CGO support needed for SQL parser; 2–5s latency acceptable |
| GitHub App owner | `kpakkiragari` personal | Can migrate to `substratehq` org later |
| Installation scope | Both per-repo and org | GitHub handles natively, no code change |
| Breaking change default | `block` (fail PR) | Configurable via `on_breaking_change: warn` |
| No `substrate.yaml` | Post setup guide comment | Converts installs to active users via `substrate init` |
| Container hosting | Cloudflare Containers (preferred) / Fly.io (backup) | Same Docker image already published |
| Phase 4 AI model | Hybrid: MCP + foundation models for reasoning, classical ML for anomaly/prediction | Never train a general LLM |

---

## Key File Locations

| File | Purpose |
|---|---|
| `docs/specs/github-app.md` | Phase 2 authoritative spec (READ-ONLY for Jules) |
| `prompts/phase-2-github-app/t01_github_app_scaffold.txt` | Jules prompt for P2-T01 |
| `prompts/phase-2-github-app/t02a_binary_serve_mode.txt` | Jules prompt for P2-T02a |
| `prompts/phase-2-github-app/t03_pr_comment_formatter.txt` | Jules prompt for P2-T03 |
| `tasks.md` | Full roadmap — Phase 1 through Phase 4 |
