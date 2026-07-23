# P10-T14: Custom Governance Rules (CRUD + CEL + JS)

## 1. Context
Substrate currently identifies breaking changes automatically. However, enterprise platform teams need to enforce custom API design policies (e.g., "All endpoints must have an X-Correlation-ID header" or "No pagination limits over 100"). Rather than forcing them to use a separate linter or compile WASM plugins, we embed a pure-Go JavaScript engine (`github.com/dop251/goja`) directly into the Substrate CLI, and provide a full CRUD API for governance rules backed by Postgres.

## 2. Requirements
- Add the `github.com/dop251/goja` dependency to `engine/go.mod`.
- Modify `engine/internal/config/config.go` to support a new `javascript` field in the `CustomRule` struct.
- In `engine/internal/checker/checker.go` or a new `engine/internal/rules/js_engine.go`, initialize a Goja runtime.
- For every custom rule defined in `substrate.yaml` with a `javascript` block, execute the script. The script should expose a `validate(schema)` function.
- We must pass the parsed JSON schema map to the JS environment.
- If the JS function returns a string, it is treated as a validation failure (the string is the error message). If it returns `true`, it passes.
- **MCP Server Integration:** Add a new MCP tool `test_js_rule` to `engine/cmd/substrate-mcp/main.go`. This tool should accept a `javascript` string and a `schema` (JSON string or object) to allow AI agents to instantly test custom rules while generating `substrate.yaml` configs.

## 3. Example `substrate.yaml` UX
```yaml
custom_rules:
  - id: "require-correlation-id"
    description: "Every endpoint must trace requests"
    javascript: |
      function validate(schema) {
        for (const path in schema.paths) {
          for (const method in schema.paths[path]) {
            const params = schema.paths[path][method].parameters || [];
            const hasHeader = params.some(p => p.in === 'header' && p.name === 'X-Correlation-ID');
            if (!hasHeader) {
              return "Missing X-Correlation-ID header on " + method.toUpperCase() + " " + path;
            }
          }
        }
        return true;
      }
```

## 4. Implementation Details
- Ensure the JS engine is heavily sandboxed (no file system access, no network access). Goja is inherently sandboxed as it has no native event loop or I/O bindings by default.
- Integrate the returned errors into the standard `report.DiffReport` as `Warnings` or `BreakingChanges` based on the rule severity.

## 5. Implementation Status: ✅ Implemented

### Database Schema
Migration `0046_create_governance_rules.up.sql` creates the `governance_rules` table:
```sql
CREATE TABLE IF NOT EXISTS governance_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    rule_text TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### API Endpoints (`api/internal/server/router.go`)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `GET` | `/api/v1/org/{org}/rules` | `AuthzMiddleware` (admin JWT or service token) | List all governance rules for an org |
| `POST` | `/api/v1/org/{org}/rules` | `AuthzMiddleware` | Create a new governance rule |
| `DELETE` | `/api/v1/org/{org}/rules/{ruleID}` | `AuthzMiddleware` | Delete a rule by ID |
| `POST` | `/api/governance/generate-cel` | `JWTValidMiddleware` | Translate natural language rule to CEL expression via AI |

### Handler (`api/internal/handlers/governance_rules.go`)
- `chi.URLParam(r, "org")` is used to extract the org name from the URL path (maps to `{org}` in the router).
- Org lookup uses `store.GetOrgIDByName(ctx, orgName)` which queries `organizations.github_org_name`.

### CEL Generation (`api/internal/handlers/governance.go`)
- Accepts a natural language `prompt` and returns a validated CEL expression.
- Falls back to a deterministic mock CEL expression when no AI API key is configured (useful for local testing).
- Validates the generated CEL using `github.com/google/cel-go`.

### E2E Validation
- Covered by `TestPhase10SystemE2E/Scenario_3:_Governance_rules_CRUD_and_API_Governance_Comment` in `scripts/e2e/phase10_e2e_test.go`.
- Test creates a rule via `POST /api/v1/org/{org}/rules` and asserts `201 Created`.
- PR comment assertion is skipped when local Forgejo is not running.
