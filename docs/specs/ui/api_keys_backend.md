# Substrate UI Org Context & API Keys Refactor (Part 2: Backend)

## Objective
Implement the full backend database schema, Go API endpoints, and SvelteKit integration for Organization-scoped API Keys to replace the current placeholder UI.

## Backend Requirements

### 1. Database Schema
- Table: `api_keys`
- Columns:
  - `id` (UUID, primary key)
  - `org_id` (UUID, foreign key to `organizations` ON DELETE CASCADE)
  - `name` (VARCHAR)
  - `prefix` (VARCHAR, e.g., first 4 chars of token + '...')
  - `hash` (VARCHAR, bcrypt hash of the token)
  - `created_at` (TIMESTAMP)
  - `last_used_at` (TIMESTAMP, optional/nullable)

### 2. API Endpoints
- `POST /api/v1/org/{org}/apikeys`
  - Body: `{ "name": "My Key" }`
  - Generates a secure token (e.g., using `crypto/rand` to generate a random 32-byte hex string or PASETO v4.local).
  - Hashes token using bcrypt.
  - Stores hash, prefix, name, org_id in DB.
  - Returns raw token EXACTLY ONCE to the client along with the key metadata.
- `GET /api/v1/org/{org}/apikeys`
  - Returns a list of API keys for the organization (without the raw token, only prefix and metadata).
- `DELETE /api/v1/org/{org}/apikeys/{id}`
  - Revokes (deletes) the specified API key.

### 3. Frontend Integration
- Update `dashboard/src/routes/(app)/org/[org]/apikeys/+page.svelte` (and create `+page.ts`) to fetch and display the list of API keys.
- Implement the "Generate New Key" button to call the POST endpoint and display the raw key in a modal or one-time alert.
- Implement revocation/deletion from the UI.

### 4. Testing
- Unit tests for API endpoints (`api_keys_test.go`).
- Playwright E2E test for the UI flow (`dashboard/tests/e2e/api_keys.spec.ts`).
