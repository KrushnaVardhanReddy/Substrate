# Spec: P7-T06 - Audit Mode (Shadow Mode) Rollout

## 1. Overview
Enterprise adoption of CI/CD blocking tools suffers from the "friction problem". If a new tool immediately blocks developer PRs across 50 repositories on day one, developers will push back, and the tool will be uninstalled.

**Audit Mode (Shadow Mode)** is the solution. It allows organizations to deploy Substrate across their entire architecture without blocking a single merge. Substrate will silently analyze, post informational comments, and build a report of prevented outages on the management dashboard, providing undeniable ROI before the organization officially enforces strict checking.

## 2. Requirements

### 2.1 Configuration
Add a `mode` field to `substrate.yaml` and a global `--mode=audit` flag to the CLI.

```yaml
# substrate.yaml
service: billing-api
schema_type: openapi
spec_path: openapi.yaml
mode: audit # Can be audit, legacy, default, strict
```

### 2.2 CLI Execution Behavior
When `mode == audit`:
1. The diff engine performs standard analysis.
2. It detects `BREAKING` changes.
3. It prints the standard CLI output.
4. **Crucial:** It overrides the exit code. Instead of exiting with `2` (Fail), it logs:
   `[AUDIT MODE] Breaking changes detected, but exiting with 0 to allow merge.`
   This MUST be printed to `os.Stderr` to prevent corrupting `--format json` output.
   And exits with `0`.

### 2.3 PR Comment Formatter
When posting a GitHub PR comment in Audit Mode, the header must clearly state that the PR is NOT blocked:

```markdown
### ⚠️ Substrate Analysis (Audit Mode)
Breaking changes were detected, but **this PR is not blocked**. 
Substrate is currently running in Shadow Mode to evaluate schema changes.

**Detected Breaks:**
- Removed field `customer.email` (Affects `billing-api`)
```

### 2.4 Dashboard ROI Tracking
The API server must accept a boolean `is_audit_mode: true` in the webhook payload. 
These events should be tracked in the database as "Preventable Outages".
The Management ROI Dashboard will feature a specific widget:
**"If Substrate was in Blocking Mode this week, you would have prevented X cross-repo outages."**

## 3. Implementation Steps
1. Update `engine/internal/config/config.go` to support the `mode` enum.
2. Modify `engine/cmd/substrate/main.go` exit code logic to intercept `BREAKING` status and return `0` if `mode == audit`.
3. Update `github-app/src/formatter.ts` to include the "(Audit Mode)" disclaimer if the payload flag is set.
4. Update `api/internal/db/schema.sql` to add an `is_audit_mode` column to the `webhook_events` table for dashboard aggregation.
