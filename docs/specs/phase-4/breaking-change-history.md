# Phase 4 Spec: Breaking Change History

> **Status:** 📐 SCAFFOLDED  
> **Goal:** Track all historical breaking changes to enable AI reasoning over past schema evolution and power the `get_breaking_change_history` MCP tool (which was deferred from Phase 3).

## 1. Database Schema (SQL Migration)

Create a new migration `api/migrations/0002_breaking_change_history.up.sql` and `0002_breaking_change_history.down.sql`.

**`0002_breaking_change_history.up.sql`**:
```sql
CREATE TABLE breaking_change_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    repo_id UUID REFERENCES repositories(id) ON DELETE CASCADE,
    org_name TEXT NOT NULL,
    repo_name TEXT NOT NULL,
    git_sha TEXT NOT NULL,
    timestamp TIMESTAMPTZ DEFAULT NOW(),
    breaking_changes JSONB NOT NULL
);
```

**`0002_breaking_change_history.down.sql`**:
```sql
DROP TABLE IF EXISTS breaking_change_history;
```

## 2. Store Interface Updates

In `api/internal/db/store.go`, add the following structs and methods:
```go
type BreakingChangeRecord struct {
    ID                 uuid.UUID
    RepoID             uuid.UUID
    CommitSHA          string
    SchemaType         string
    BreakingRulesCount int
    DiffReport         json.RawMessage
    CreatedAt          time.Time
}

type Store interface {
    // ... existing methods ...
    RecordBreakingChange(ctx context.Context, repoID uuid.UUID, commitSHA, schemaType string, breakingCount int, diffReport []byte) error
    GetBreakingChangeHistory(ctx context.Context, orgName, repoName string, limit int) ([]BreakingChangeRecord, error)
}
```

## 3. API Endpoints

Add two new endpoints in `api/internal/server/router.go`:

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/api/v1/history` | Record a new breaking change. Called by the Cloudflare Worker when a PR with breaking changes is merged. |
| `GET` | `/api/v1/history/{org}/{repo}?limit=10` | Retrieve breaking change history for a repo. |

## 4. MCP Tool Implementation

Update the MCP Server (`api/internal/mcp/server.go`) to replace the mock data for `get_breaking_change_history`. 
It should make a GET request to the new Registry API endpoint: `GET /api/v1/history/{org}/{repo}?limit={limit}` using the `REGISTRY_API_URL` environment variable.
