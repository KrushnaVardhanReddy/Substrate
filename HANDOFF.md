# Substrate Project Handoff

**Date:** July 8, 2026 (End of Session)
**Current Phase:** Phase 3 (Contract Registry) & Phase 1e (AsyncAPI/Avro) 🔄 IN PROGRESS — 2 Jules sessions running.

---

## What Jules Is Working On Right Now

Both sessions were submitted to the `feature/dev` branch. Review PRs tomorrow morning.

| Jules Session | Task | What It Builds |
|---|---|---|
| `3806940771454033087` | P1e-T01 | AsyncAPI 2.x/3.x Breaking Change Adapter |
| `11212679817250327809` | P1e-T02 | Apache Avro Schema Registry Compatibility Adapter |

---

## What We Completed Today

1. **Merged PRs**:
   - P1d-T01 (Protobuf Adapter) merged ✅
   - P3-T01 (Go API Server Scaffold) merged ✅
   - P3-T02 (substrate.yaml Consumer Parser) merged ✅
   - P3-T02b (Contract Registry Sync push-to-main webhook) merged ✅
2. **Specs Synced**: Ensured all actual implementation details from the merged PRs (Go structs, Store interfaces, env vars, `installation_id` fields) are documented in the respective specs.
3. **Phase 1e Launched**: Wrote full spec for AsyncAPI and Avro adapters, and submitted T01 and T02 prompts to Jules.

---

## Next Steps (After Jules PRs Are Merged)

**Tomorrow (priority order):**

> 🚨 **CRITICAL RULE FOR ALL FUTURE TASKS:** 🚨
> **First, specs MUST be checked and updated. ONLY THEN can prompts be created or submitted for ANY task.** Do not write a prompt unless the spec-first rule is fully satisfied.

1. Review and merge Jules PRs for P1e-T01 and P1e-T02.
2. **P3-T02c (Cross-Repo Compatibility Check on Provider PR)**
   - Unblocked now that P3-T02b is merged.
   - *Check the spec first* (`docs/specs/phase-3/contract-registry.md`). It should be complete, but verify.
   - Write prompt and submit to Jules.
3. **P3-T02d (Cross-repo E2E fixture tests)**
   - Write prompt after P3-T02c is submitted.

---

## Key File Locations

| File | Purpose |
|---|---|
| `docs/specs/phase-3/contract-registry.md` | Phase 3 Registry Spec (READ-ONLY) |
| `docs/specs/phase-1/asyncapi-avro-adapter.md` | Phase 1e AsyncAPI/Avro Spec (READ-ONLY) |
| `tasks.md` | Full roadmap with priorities and Jules Session IDs |
| `scripts/jules_submit.py` | Run `python3 scripts/jules_submit.py --list` to see available prompts |
