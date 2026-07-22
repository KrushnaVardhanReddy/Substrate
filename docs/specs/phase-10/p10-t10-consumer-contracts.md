# P10-T10: Consumer-Driven Contract Manifests

## Objective
Enable a "Consumer-Driven Contracts" (CDC) architecture, allowing frontend or downstream applications to declare exactly which API fields they consume. Substrate will use these manifests to prune "breaking" changes from the diff report if no consumer actually uses the dropped field, achieving zero false positives.

## Architecture
1. **The Manifest (`.substrate-consumer.yaml`):**
   Consumers commit a file declaring their dependencies.
   ```yaml
   schema_version: "1.0"
   provider: "github.com/myorg/backend-api"
   consumes:
     - path: "GET /api/v1/users"
       fields:
         - "id"
         - "email"
         # Note: "last_name" is NOT consumed.
   ```

2. **Registry Ingestion API:**
   - When the consumer repo pushes to `main`, the Substrate GitHub App pushes this manifest to the Substrate API (`POST /api/v1/consumers/manifest`).
   - The API stores this manifest mapping in the PostgreSQL database.

3. **Impact Pruning Logic:**
   - During a provider PR, the Diff Engine detects that `last_name` was removed from `GET /api/v1/users`.
   - The engine queries the Registry API for consumer impact.
   - The API checks the stored manifests. Since no consumer declared `last_name`, the API explicitly flags this change as `SAFE_NO_CONSUMERS` instead of `BREAKING`.

## Deliverables
- Parser for `.substrate-consumer.yaml`.
- Postgres schema updates (`consumer_manifests` table).
- Updated Impact Engine logic to filter breaking changes against known consumer manifests.
- Unit and E2E tests proving a breaking change is ignored if unused.
