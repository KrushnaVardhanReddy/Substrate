# Substrate End-to-End (E2E) Live Testing Plan

To guarantee Substrate works in a real-world environment without any mocked data, we will perform a live, end-to-end test. This plan outlines the exact steps to spin up the local infrastructure and simulate real GitHub PRs across all 8 supported schema types.

## 1. Local Infrastructure Setup
To test without mocking, we must run all three core components of Substrate locally.

1. **PostgreSQL Registry Database:**
   - Spin up a local PostgreSQL container (`docker run --name substrate-db -p 5432:5432 -e POSTGRES_PASSWORD=postgres -d postgres`).
   - Run the database migrations.
2. **Go Registry API & MCP Server:**
   - Start the Go API server: `cd api && go run cmd/server/main.go`. This connects to the local PostgreSQL database.
3. **SvelteKit Dashboard:**
   - Start the frontend: `cd dashboard && npm run dev`. This will query the local Go API to visualize our test contracts.

## 2. Webhook Tunneling (The Cloudflare Worker)
We need GitHub to send real webhooks to our local Cloudflare Worker.
1. Run `ngrok http 8787` (or `smee.io`) to get a public URL for your local machine.
2. Start the GitHub App worker locally: `cd github-app && npx wrangler dev`.
3. Go to the Substrate GitHub App settings on github.com and update the **Webhook URL** to the ngrok/smee address.

## 3. "Test Labs" Setup (Real GitHub Repos)
Create two blank repositories on GitHub:
- `substrate-test-provider` (Simulating the backend team)
- `substrate-test-consumer` (Simulating the frontend/mobile team)

Install the Substrate GitHub App on both repositories.

## 4. The Live E2E Matrix (Testing All 8 Adapters)
With the infrastructure running, we will test the entire diff engine by pushing different file types to the provider repo, and then opening a PR that breaks them.

| Adapter | Test File | Safe Change (Merged) | Breaking Change PR (Blocked) | Expected Rule |
| :--- | :--- | :--- | :--- | :--- |
| **OpenAPI** | `openapi.yaml` | Add a new `/health` endpoint | Delete an existing `/users` endpoint | `ENDPOINT_REMOVED` |
| **GraphQL** | `schema.graphql` | Add a new `type Post` | Remove a field from `type User` | `GRAPHQL_FIELD_REMOVED` |
| **Protobuf** | `user.proto` | Add a new optional field | Change a field's type from `int32` to `string` | `PROTO_FIELD_TYPE_CHANGED` |
| **Avro** | `user.avsc` | Add a field with a default | Remove a required field | `AVRO_FIELD_REMOVED` |
| **SQL (PG)** | `schema.sql` | `CREATE TABLE logs...` | `ALTER TABLE users DROP COLUMN id;` | `SQL_COLUMN_DROPPED` |
| **Terraform** | `main.tf` | Add an S3 bucket | Change the provider version / delete an output | `TF_OUTPUT_REMOVED` |
| **AI/ML** | `model.yaml` | Add a new tag | Change the required input tensor shape | `AIML_INPUT_SHAPE_CHANGED` |
| **Enterprise**| `Account.object` | Add a custom field | Change field type from `Text` to `Number` | `SFDC_FIELD_TYPE_CHANGED` |

## 5. Testing the MCP Server (AI IDE Integration)
Once the repositories are populated with history, we will test the MCP server integration.
1. Edit `claude_desktop_config.json` to point to the local `substrate-mcp` binary.
2. Open Claude Desktop and prompt it: *"What happens if I merge a PR that deletes the `/users` endpoint in `substrate-test-provider`?"*
3. Claude will autonomously use the `check_compatibility` MCP tool to run a local simulation against the database and report back the exact blast radius.
