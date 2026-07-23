1. **Implement 28 new tools in `api/internal/mcp/tools.go`**:
   - Implement Group A: `diff_schemas`, `sync_schema`, `check_deploy`, `check_rollback`, `get_dependency_graph`, `get_impact`, `get_diff_report`, `cross_repo_check`, `trigger_webhook`.
   - Implement Group B: `create_governance_rule`, `list_governance_rules`, `delete_governance_rule`, `generate_cel_rule`.
   - Implement Group C: `register_webhook`, `list_repos`, `get_schema`, `get_breaking_history`, `get_changes_feed`, `get_zombies`, `ingest_otel_metrics`.
   - Implement Group D: `predict_egress_cost`, `get_roi_metrics`.
   - Implement Group E: `file_insurance_claim`, `get_insurance_policy`, `list_insurance_claims`.
   - Implement Group F: `publish_plugin`, `list_plugins`, `ai_analyze_schema`, `get_public_profile`, `get_changelog`.
   For handlers requiring HTTP round-trips to existing handlers, use `httptest.NewRecorder()`. If store provides logic, use store directly.

2. **Fix resources stub and add 8 resources in `api/internal/mcp/resources.go`**:
   - Change `substrate://governance-rules` to `substrate://governance-rules/{org}` and implement it to use `store.ListGovernanceRules`.
   - Add `substrate://graph/{org}`.
   - Add `substrate://repos/{org}`.
   - Add `substrate://history/{org}/{repo}`.
   - Add `substrate://diff/{id}`.
   - Add `substrate://zombies/{org}`.
   - Add `substrate://roi/{org}`.
   - Add `substrate://insurance/{org}/policy`.
   - Add `substrate://profile/{org}`.
   - Add 5 new prompts in `RegisterPrompts`: `substrate_diff_workflow`, `substrate_governance_setup`, `substrate_incident_response`, `substrate_onboard_org`, `substrate_schema_smell`.

3. **Add SSE transport methods to `api/internal/mcp/server.go`**:
   - Add `ServeSSE(w http.ResponseWriter, r *http.Request)`.
   - Add `ServeHTTPMessage(w http.ResponseWriter, r *http.Request)` that reads POST JSON body, calls `HandleMessage(body)`, and writes the result. (Using a sync.Map for SSE connections if needed, but per the prompt, just return the response directly in HTTP response body).

4. **Register SSE routes in `api/internal/server/router.go`**:
   - I need to export `mcpServer` from somewhere. Oh, the prompt says "Find where `mcp.RegisterTools` and `mcp.RegisterResources` are called. Add after those calls:". But looking at `api/cmd/server/main.go`, `mcpServer` is instantiated inside `if headlessMCP { ... }`.
   - Wait, if the SSE transport is for standard API server, `mcpServer` needs to be instantiated globally and passed to router? No, the prompt specifically says "Find where mcp.RegisterTools and mcp.RegisterResources are called. Add after those calls: `r.Method("GET", "/mcp/sse", ...)`". However, these are currently NOT called in `server.NewRouter`, they are called in `api/cmd/server/main.go` under `if headlessMCP`. Let me check `router.go` closely or maybe we need to create `mcpServer := mcp.NewServer()` in `router.go`? I'll look into `router.go` to see how to inject it, or initialize it there. Let me re-read the spec. Oh, the instructions say "Find where mcp.RegisterTools and mcp.RegisterResources are called - you will add mcp.ServeSSE() registration here too." Wait, it's NOT called in `router.go`. It says "Find where mcp.RegisterTools and mcp.RegisterResources are called. Add after those calls..." Let's clarify if we should add it in `router.go` or `main.go`. I'll create `mcpServer` inside `NewRouter`.

5. **Create `scripts/e2e/phase_mcp_e2e_test.go`**:
   - Implement `TestPhaseMCPSystemE2E` with 10 scenarios.

6. **Complete pre commit steps**
   - Make sure proper testing, verifications, reviews and reflections are done.

7. **Submit the change.**
   - Commit and submit.
