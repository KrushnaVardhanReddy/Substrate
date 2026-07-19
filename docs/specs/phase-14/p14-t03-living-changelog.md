# P14-T03: Living API Changelog

## 1. Objective
Auto-generate a beautiful, public-facing versioned changelog page (similar to Stripe's API Changelog) from every tracked schema change. This serves as a public source of truth for downstream consumers and helps organizations build trust by documenting their API's evolution.

## 2. Architecture & Database Changes
Currently, the `diff_reports` table only stores the raw JSON `report_data` and timestamps. To implement a changelog per repository, we must store the organization and repository directly in the database.

1. **Database Migration:**
   - Create a new SQL migration in `api/migrations/` (e.g., `0011_add_org_repo_to_diff_reports`).
   - Add `org_name VARCHAR(255)` and `repo_name VARCHAR(255)` to the `diff_reports` table.
2. **Store Update:**
   - Update `SaveDiffReport` in `api/internal/db/store.go` and `api/internal/db/queries.go` to accept `orgName string, repoName string` and `INSERT` them.
3. **Handler Update:**
   - Update `SaveDiffHandler` in `api/internal/handlers/diff.go` to pass `req.Org` and `req.ProviderRepo` to the store method.

## 3. API Implementation
Implement `GET /api/v1/changelog/{org}/{repo}` in `api/internal/handlers/changelog.go`.
- Query: `SELECT report_data, created_at FROM diff_reports WHERE org_name = $1 AND repo_name = $2 ORDER BY created_at DESC LIMIT 20`.
- Response: Transform the raw diffs into a chronological summary JSON array containing dates, endpoints added, and endpoints removed.

## 4. UI Implementation (Stitch)
Once the API is merged, the UI will consume the new endpoint to build a shareable route at `substrate.io/myorg/payments-api/changelog`.
