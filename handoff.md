# Substrate Handoff

## Current Status (End of Day)
* **Phase 12 Testing Automated:** We have successfully achieved full automated E2E test coverage for the Substrate platform. The "Real-World" pipeline (Git Push -> Webhook -> Go API -> WASM Diff Engine -> UI Sync) is fully tested across all 7 schema types via Playwright.
* **TDD E2E Suites Merged (P12-T15):** Jules successfully wrote and merged the Advanced Feature Suites E2E (`feature-dev-14546955130784840283`). This includes 16 tests, 6 of which are failing *by design* because they assert against UI elements we haven't built yet (TDD approach).
* **Phase 10 Roadmap Launched:** We formally defined 17 massive enterprise features for Phase 10, generated all specs, and dispatched Jules to build three of them in parallel (T11 Terraform Provider, T08 Live Docs, T17 Implicit Infra). 
* **AI Safety Block:** T01 (Security Fuzzing) was refused by Jules due to AI safety guardrails (preventing the generation of offensive SQLi/XSS payloads) and has been marked as Blocked.
* **Phase 13 "God-Mode" Added:** We laid the conceptual groundwork for Phase 13, adding FinOps Cost Prediction, DB Performance Breakages, and AI Chaos Engineering to the future roadmap.

## Next Steps for Tomorrow
1. **Frontend Polish (P12-T16):** Your first task should be to jump into the SvelteKit codebase (`dashboard/`) and build the missing UI elements (Enterprise Routes `/org/admin/telemetry`, Impact API Text `"Can-Deploy"`, and Transitive Blast Radius highlights). Run `npx playwright test` to track your progress until all 16 tests turn green.
2. **Review Jules's Phase 10 PRs:** Check on the three background sessions Jules is running. Review and merge the Terraform Provider, Live Docs API, and Implicit Infrastructure Discovery PRs.
3. **Pivot Blocked Tasks:** Decide what to do with the blocked Security Fuzzing task (P10-T01)—either build it manually, pivot to a CLI wrapper (OWASP ZAP), or ignore it and assign Jules to the Tree-sitter AST analysis (P10-T12).

## Quick Start Reminders
* **Run Tests:** `cd dashboard && npx playwright test` (to see the 6 failing TDD tests).
* **Start Forgejo locally:** Run `make forgejo` (Starts the container on port 3000).
* **Start Backend Stack:** Run `make start-bg`.
* **View Graph:** Go to `http://localhost:5173/org/<forgejo-username>/graph`.

## 🏛️ Architectural Decisions Log
* **Cloud Database Selection:** We have officially selected **Neon (Serverless Postgres)** as our managed cloud database provider for the SaaS tier. 
  * *Why:* It scales to zero, meaning the Free Tier (0.5GB storage) will cost us literally **$0/month** until we land our first paying enterprise client. Once we scale, Neon's "Database Branching" feature will perfectly align with Substrate's PR-based schema preview model.

## 🧪 Comprehensive Demo Testing Guide

### Axis 1: Breaking vs. Non-Breaking Changes
For each repository, test both a schema breaking change and a safe (non-breaking) addition to verify the diff engine logic.
* **Microservices (Protobuf):** *Breaking:* Remove `CartItem.product_id` | *Safe:* Add `string notes = 3`
* **GitHub GraphQL:** *Breaking:* Remove `User.email` | *Safe:* Add `User.age: Int`
* **Stripe (OpenAPI):** *Breaking:* Remove `/v1/charges` | *Safe:* Add `/v2/beta/charges`
* **Jaffle Shop (SQL):** *Breaking:* Drop `customer_lifetime_value` | *Safe:* Add `age INTEGER`
* **Slack (AsyncAPI):** *Breaking:* Remove `channel_id` | *Safe:* Add `thread_ts`

### Axis 2: Governance & Overrides (The "Yellow Path")
* **Intentional Breakage:** Push a breaking change, but include an `overrides` block in `substrate.yaml` with `rule_id: "*"`.
* **Verification:** Ensure that the CI check *passes* (with a warning) instead of failing, and the UI marks the change as "Acknowledged" (Amber/Yellow Status).

### 🤖 Playwright E2E Automation Note
> **Important Architectural Note for the E2E Tests:** 
> The Playwright script dynamically generates lightweight "stub" schema files (e.g., a 10-line `openapi.yaml`) in a `/tmp` folder instead of copying the real `demo-repos/` from disk. 
> *Why did we do this?* Because massive repos like `stripe/openapi` are 1.2GB and are explicitly `.gitignore`d. Generating stubs guarantees the test is lightning-fast and 100% reproducible on any machine.
