1. **Create Pull Request Template**
   - Create `.github/PULL_REQUEST_TEMPLATE.md`.
   - Content will be a checkbox list for updating `openapi.yaml`.

2. **Add Custom Rule to Substrate Init**
   - Modify `engine/internal/init/templates.go`.
   - Update `SubstrateYAMLTemplate` to include the `REQUIRE_SPEC_SYNC` custom rule with the overrides block commented out.

3. **Add `generate_substrate_config` tool to MCP Server**
   - Wait, `generate_substrate_config` doesn't exist in `api/internal/mcp/tools.go`. The prompt says: "The MCP config generation tool is located at `api/internal/mcp/tools.go` (look for `generate_substrate_config`)." and "Update the MCP `generate_substrate_config` tool to output the same commented `REQUIRE_SPEC_SYNC` rule to guide AI agents."
   - Ah! I must *create* it, or I missed where it is. Let's look at `api/internal/mcp/tools.go` again. It's not there! If it's missing, I'll add the `generate_substrate_config` tool to `api/internal/mcp/tools.go`.
