1. **Config struct changes**
   - Update `api/internal/config/config.go` -> `Metadata` struct to add `Owner`, `SlackChannel`, `PagerDuty`, `PM`, `SLATier`, `NodeColor`.
2. **Database Migration**
   - Create `api/migrations/20240101000000_add_repo_metadata.up.sql` (and down.sql) as an empty file or dummy query (`SELECT 1;`), because the `repositories` table already has a `metadata` JSONB column which stores schemaless JSON.
3. **Webhook Updates**
   - Check `api/internal/services/push.go` (or `api/internal/webhook/push.go`). The logic in `push.go` currently passes `cfg.Metadata` mapped to `json.RawMessage` to `store.UpsertRepo`. Since we are updating the `config.Metadata` struct, the JSON marshalling will naturally include the new fields, so `push.go` might not even need an update. However, we'll verify it carefully to ensure `metadata` is successfully marshaled and persisted.
4. **Dashboard Frontend Updates**
   - In `dashboard/src/routes/(app)/org/[org]/catalog/[repo]/+page.svelte`: add a UI section (e.g., a "Business Context" card) to display `metadata.owner`, `metadata.slack_channel`, `metadata.pagerduty`, `metadata.pm`, `metadata.sla_tier`.
   - In `dashboard/src/lib/components/ServiceNode.svelte`: Use `metadata.node_color` for `backgroundColor` or `border-color`. If present, override the default heatmap coloring.
5. **AI Tools Update**
   - In `api/internal/mcp/tools.go`, check `get_dependency_graph` or `analyze_repository`. Since `get_dependency_graph` calls `GetDependencyGraph` which fetches the graph, `consumer_metadata` and `provider_metadata` are fetched from DB. This is automatically exposed to AI tools if the API response is unmarshaled as JSON.
6. **Pre-commit step**
   - Ensure proper testing, verification, review, and reflection are done (will call `pre_commit_instructions`).
