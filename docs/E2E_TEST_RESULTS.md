# E2E Integration Testing Results (Phase 5)

This document serves as the official sign-off log for the Phase 5 End-to-End Cross-Repo Integration test. We successfully executed all four required scenarios between `substrate-test-provider` and `substrate-test-consumer`, verifying the orchestration between the GitHub App (Cloudflare Worker), the Diff Engine, and the Registry API.

## Tested Scenarios

### Scenario 1: Uncoordinated Breaking Change
* **Action:** Removed the `GET /users/{id}` endpoint from `openapi.yaml`.
* **Expected Result:** Substrate engine flags `ENDPOINT_REMOVED` as `🔴 BREAKING`. The cross-repo registry reports the downstream consumer `substrate-test-consumer` as broken. The GitHub status check fails.
* **Actual Result:** **PASS**. The pipeline blocked the PR successfully. The PR comment properly rendered both the local repository error and the cross-repo impact table.

### Scenario 2: Safe Extension
* **Action:** Added a net-new endpoint `GET /users/details/safe-endpoint`.
* **Expected Result:** Substrate engine detects an addition, flags it as `✅ Safe`, and the cross-repo registry reports 0 broken consumers. The GitHub status check passes.
* **Actual Result:** **PASS**. The pipeline correctly identified the safe extension and allowed the PR to proceed.

### Scenario 3: Config Override (Acknowledged Breaking Change)
* **Action:** Removed `GET /users/{id}` but added an explicit override block in `substrate.yaml` for the `ENDPOINT_REMOVED` rule, citing coordination with consumers.
* **Bug Encountered:** The override worked for the *single-repo* check, but the *cross-repo* check still failed.
* **Root Cause & Fix:** 
  1. Found that `index.ts` (Cloudflare Worker) was not passing the `substrate.yaml` file content to the Registry API. Updated `CrossRepoCheckRequest` to include `config_content`.
  2. Found that `config.go` (Diff Engine) was attempting to use `os.Stat` to verify the existence of `openapi.yaml` on the disk, which fails when the engine runs in a stateless HTTP container. Removed the disk-check logic for cloud execution.
* **Expected Result (Post-Fix):** Engine parses the config, matches the rule ID, and downgrades the breaking change to a warning (`✅ Acknowledged`). Cross-repo impact is mitigated.
* **Actual Result:** **PASS**. The GitHub check turned green and the PR was unblocked.

### Scenario 4: Warnings Only
* **Action:** Restored the `GET /users/{id}` endpoint (no longer breaking), but added `deprecated: true` to the schema. Removed the `substrate.yaml` override block.
* **Expected Result:** Substrate detects `ENDPOINT_DEPRECATED`. Flags it as a `🟡 WARNING` rather than an error. The GitHub check passes.
* **Actual Result:** **PASS**.

### Scenario 5: SQL Adapter Matrix Testing
* **Action:** Ran the full suite of SQL automated tests (`make e2e-sql-safe`, `make e2e-sql-breaking`, `make e2e-sql-warning`, `make e2e-sql-override`).
* **Bugs Encountered:**
  1. Found that `COLUMN_REMOVED` was missing from `KnownRules` in `engine/internal/config/config.go`, which caused overrides to be rejected and fail as breaking changes.
  2. Found that `endpoint-added` and `api-path-added` mappings were missing from `engine/internal/diff/openapi.go`, causing safe endpoint additions to default to Warnings and failing the `All Clear` assertion.
* **Root Cause & Fix:** 
  1. Populated `KnownRules` with all 27 SQL-specific rule IDs.
  2. Mapped the missing OpenAPI rules to their correct `SAFE` spec values.
  3. Aligned the `docs/E2E_TEST_PLAN.md` with the engine (`SQL_COLUMN_DROPPED` -> `COLUMN_REMOVED`).
  4. Updated the E2E script to strictly assert the exact `Rule ID` in the GitHub PR bot comment to ensure spec-first compliance.
* **Actual Result:** **PASS**. All 4 SQL test vectors passed successfully and accurately verified the underlying rule IDs.

## Conclusion
The Phase 5 cross-repo impact tracking pipeline is fully operational for both OpenAPI and SQL schemas. The Cloudflare Worker correctly marshals spec snapshots and configuration state to both the stateless Diff Engine and the stateful PostgreSQL Registry API, surfacing real-time dependency impact directly into developer workflows.
