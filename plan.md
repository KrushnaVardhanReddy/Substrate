1. **Create `engine/pkg/ai/migration.go`**: Implement the LLM integration logic using `github.com/sashabaranov/go-openai`. Read API key from `os.Getenv("OPENAI_API_KEY")` (Wait, memory/existing code says `SUBSTRATE_AI_API_KEY`, but prompt says `OPENAI_API_KEY` - I'll support both, prioritizing `OPENAI_API_KEY` as requested in the task).
   - Write a `GenerateMigrationPlan(diffReport json.RawMessage) (string, error)` function.
   - Craft a system prompt instructing the AI to output an "Expand-and-Contract" deployment strategy (e.g. 1. Add new field, 2. Dual write, 3. Backfill, 4. Drop old field) based on breaking changes in the diff report.
   - Force the LLM to output a strictly formatted Markdown checklist.
   - Include a fallback mechanism if the API call fails or times out.

2. **Create tests for `engine/pkg/ai/migration.go`**: Write `migration_test.go` with table-driven unit tests, achieving at least 80% coverage. Include mock LLM client responses.

3. **Update `api/internal/handlers/diff.go`**:
   - In `SaveDiffHandler`, after unmarshaling `DiffReport` and identifying high-severity breaking changes (e.g., `diffReport.Summary.BreakingCount > 0`), trigger the migration planner logic.
   - If a migration plan is successfully generated, it needs to be injected into the GitHub PR as a comment. The spec says "The generated plan is injected into the GitHub PR as an actionable checklist for the developer."
   - We might need to use `api/internal/github/client.go` to comment on the PR. Let's trace how comments are created or create a `CreateIssueComment` function if missing. The task states "Formatting logic to output the response nicely in the existing GitHub PR comment bot." So I will add `CreateIssueComment` to the GitHub client.

4. **Update `api/internal/github/client.go` and `api/internal/github/comments.go`**:
   - Add `CreateIssueComment(ctx context.Context, owner, repo string, issueNumber int, body string) error` to the GitHub `Client` interface and its implementations.
   - Add a formatting helper `GenerateMigrationComment(plan string) string` in `comments.go`.

5. **Wire it up in `api/internal/handlers/diff.go`**:
   - Inject the `github.Client` into `SaveDiffHandler` (or instantiate it) and invoke `CreateIssueComment` using the generated migration plan, for the correct `Org`, `ProviderRepo`, and `PRNumber` from the `SaveDiffRequest`.
   - Update `api/internal/server/router.go` to pass the `github.Client` to `SaveDiffHandler`.

6. **Pre-commit checks**: Ensure all tests run, linter passes, and Go build is successful. E2E check.
