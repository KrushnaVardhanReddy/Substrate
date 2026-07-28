1. **Add `RiskScoreHandler` to Router**: Ensure `RiskScoreHandler` is registered in `api/internal/server/router.go` under `/api/v1/risk/{org}/{repo}/{pr}` so the API test can hit it.
2. **Create `scripts/e2e/phase9_api_test.go`**: Use the provided PGlite database connection to seed data. Hit the live API (`http://localhost:8090/api/v1/risk/...`) and verify the risk scoring and compliance behavior.
3. **Create `dashboard/tests/e2e/phase9.spec.ts`**: Implement the Playwright UI test navigating to `http://localhost:5173/org/mcp-org/` and asserting the expected repository and UI elements render correctly.
4. **Pre Commit Steps**: Ensure test passes and dependencies are checked.
