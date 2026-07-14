# Spec: P8-T01 - PostgreSQL Job Queue (River)

## 1. Overview
As Substrate scales to process webhook events from hundreds of repositories simultaneously, performing synchronous database operations and LLM calls in the API handlers will lead to timeout failures (GitHub Webhooks time out after 10 seconds). 
We need a robust, transactional background job queue. To maintain our "zero-dependency" philosophy for on-premise deployments, we will NOT use Redis or RabbitMQ. Instead, we will use `riverqueue/river`, a high-performance Go job queue backed directly by our existing PostgreSQL database.

## 2. Requirements

### 2.1 Database Integration
- Integrate `github.com/riverqueue/river` and `github.com/riverqueue/river/riverdriver/riverpgxv5`.
- Add a database migration `0008_river_queue.up.sql` using River's CLI/schema tool to generate the necessary `river_job` tables.
- Initialize the River client in `api/internal/server/server.go` using the existing `pgxpool`.

### 2.2 Workers and Jobs
Create the following worker definitions in `api/internal/workers/`:
1. **SyncWebhookWorker**: Handles the repository schema sync processing (currently in `SyncHandler`).
2. **PushWebhookWorker**: Handles GitHub Push events (currently in `webhook.PushHandler`).
3. **CrossRepoCheckWorker**: Offloads the AI LLM diffing and PR generation to a background task.

### 2.3 API Refactoring
- Refactor the current HTTP handlers (`SyncHandler`, `PushHandler`, `CrossRepoCheckHandler`) to quickly insert a job into River and return `202 Accepted`.
- The webhook responder should return immediately, while the workers process the tasks asynchronously.

## 3. Implementation Steps
1. Add `riverqueue` dependencies to `go.mod`.
2. Generate and apply River Postgres migrations.
3. Implement `workers.go` to register all River job structs and workers.
4. Refactor API endpoints to enqueue jobs instead of synchronous execution.
5. Update `cmd/substrate-api/main.go` to start the River worker pool alongside the HTTP server.
