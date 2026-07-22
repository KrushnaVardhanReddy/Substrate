# P15-T02: Granular GitHub Check Suite

## Objective
Replace the single monolithic "Substrate" GitHub CI check with highly granular, individually passable checks. Enterprise pipelines require separation of concerns, allowing different teams (Security, Platform, Data) to bypass or enforce specific policies without blocking the entire build.

## Architecture
1. **Check Suite Breakdown:**
   Substrate's GitHub App must spawn multiple independent Checks on a PR instead of one:
   - `substrate / breaking-changes` (Fails if APIs drop fields)
   - `substrate / security-rules` (Fails if authentication headers are removed)
   - `substrate / pii-compliance` (Fails if an `ssn` field is added without encryption tags)
   - `substrate / schema-linting` (Fails on missing descriptions or poor casing)

2. **Webhook Orchestration (`api/internal/github/checks.go`):**
   - On the `pull_request` event, call the GitHub API to create all 4 checks in `queued` state.
   - As the Diff Engine processes the schema, stream the results back to the API.
   - The API updates the specific check based on the `rule_id` mapping. (e.g. `FIELD_REMOVED` updates `breaking-changes`, while `AUTH_REMOVED` updates `security-rules`).

3. **Bypass Configurations:**
   - Update `substrate.yaml` to allow granular enforcement:
     ```yaml
     checks:
       breaking-changes: blocking
       schema-linting: advisory   # Will show as neutral/warning, won't block PR
     ```

## Deliverables
- Refactored `api/internal/github/checks.go` to support creating and updating multiple distinct checks per PR.
- Mapping logic in the API to bucket diff anomalies into the correct check category.
- E2E tests mocking GitHub's API to ensure 4 distinct POST requests are made for the check suite.
