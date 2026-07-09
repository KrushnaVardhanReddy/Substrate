# Automated E2E Testing Suite (Phase 5)

## Overview
To guarantee the resilience of the Substrate orchestrator (Diff Engine, Registry API, and Cloudflare Worker) across all supported schema types, we will implement a fully automated end-to-end testing suite written in **Golang**. This suite will interact directly with the live GitHub API to simulate developer workflows, proving that cross-repo breaking changes are correctly identified and blocked in a real-world environment.

## Architecture
- **Language:** Golang
- **Dependencies:** `github.com/google/go-github/v62`
- **Location:** `scripts/e2e/main.go`
- **Execution:** Run locally or in CI against test repositories (`substrate-test-provider` and `substrate-test-consumer`).

## The Automation Flow
For each supported schema adapter (`openapi`, `sql`, `graphql`, `protobuf`, `avro`, `terraform`, `aiml`), the Go script will execute the following state machine concurrently:

1. **Setup & Seeding (Main Branch):**
   - Use the GitHub API to update `substrate.yaml` in both the provider and consumer repositories to the target `schema_type`.
   - Commit and push a valid, baseline schema (e.g., `schema.sql`) to both repositories' `main` branches.
   - *Wait 5 seconds* to allow the Cloudflare Worker to process the push webhook and sync the baselines into the PostgreSQL registry.

2. **Execution of 4 Discrete Scenarios:**
   For each adapter, the suite runs the following permutations sequentially:
   - **Breaking Change:** Removes an element (e.g., column drop, endpoint removal). Expects `❌ BREAKING`.
   - **Safe Extension:** Adds an optional element (e.g., nullable column, new endpoint). Expects `✅ Safe`.
   - **Config Override:** Removes an element but applies an active `substrate.yaml` override. Expects `✅ Substrate — All Clear`.
   - **Warning / Deprecation:** Modifies an element safely but triggers a warning (e.g., changing default, deprecating endpoint). Expects `🟡 WARNING`.

3. **Validation & Polling:**
   - Poll the GitHub API's Issue Comments endpoint for the newly created PR every 3 seconds.
   - Wait for the `substrate-local-test[bot]` to post the cross-repo impact comment.
   - Assert that the comment body contains the expected `rule_id` (e.g., `SQL_COLUMN_DROPPED`) and the `❌ BREAKING` cross-repo impact table.
   - Poll the GitHub API's Status Checks endpoint and assert that `substrate/breaking-changes` is `failure`.

4. **Teardown:**
   - Close the Pull Request.
   - Delete the feature branch.

## Expected Outcomes
- The test suite must pass 100% of the scenarios, proving that all 8 adapters correctly trigger the full webhook → worker → registry → diff engine → github comment pipeline.
- If a check fails, the test suite exits with a non-zero code, identifying the broken schema adapter.
