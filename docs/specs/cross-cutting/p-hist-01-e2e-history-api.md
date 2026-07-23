# P-HIST-01: Breaking Change History & Changes API E2E Validation

## 1. Overview
The history pipeline (`POST /api/v1/history` → `GET /api/v1/history/{org}/{repo}` → `GET /api/v1/changes`) is the backbone of the `substrate archaeology` feature and the Diff Viewer timeline. It has zero direct E2E assertions despite being a live, production route.

## 2. APIs Under Test

| Route | Auth | Description |
|-------|------|-------------|
| `POST /api/v1/history` | Service Token | Store a breaking change event for a repo |
| `GET /api/v1/history/{org}/{repo}` | Service Token or JWT | List breaking change history for a specific repo |
| `GET /api/v1/changes` | Service Token | Recent breaking changes across the org |

## 3. Test Scenarios

### Scenario 1: Store and Retrieve Breaking Change History
1. `setupPhase_Hist_Database` — clean `breaking_change_history`, `repositories`, `organizations`.
2. Seed org + two repos (provider, consumer) via pgxpool.
3. `POST /api/v1/history` with payload: `{org, repo, commit_sha, breaking_changes: [{description}]}` → assert **201**.
4. `GET /api/v1/history/hist-org/backend` → assert **200**.
5. Unmarshal as `[]map[string]interface{}` — assert `len >= 1`.
6. Assert `result[0]["commit_sha"] == "sha-hist-test"`.
7. Assert `result[0]["breaking_changes"]` is non-empty array.

### Scenario 2: Changes Feed Returns Recent Events
1. Reuse DB state from Scenario 1 (2 history entries).
2. `GET /api/v1/changes` with `Authorization: Bearer local-dev-token` → assert **200**.
3. Assert response contains at least the repo name `"hist-org/backend"` in the result.

### Scenario 3: History Returns Empty for Repo With No History
1. Seed a clean repo with no breaking changes.
2. `GET /api/v1/history/hist-org/clean-repo` → assert **200**.
3. Assert response body is an empty JSON array `[]`.

## 4. Constraints
- File: `scripts/e2e/phase_history_e2e_test.go`
- All prefix vars/functions: `pHist`
- No mocks. Real Postgres + real API.

## 5. Implementation Status
⏳ **NOT YET IMPLEMENTED** — created 2026-07-23
