# Substrate API Keys Backend Integration

## Objective
Implement full backend and frontend integration for Organization-scoped API Keys to replace the current placeholder UI.

## Problem Statement
The "API Keys" page (`/org/[org]/apikeys`) currently features scaffolding and a non-functional "Generate New Key" button. To complete this functionality, we must allow users to generate, view, and revoke API keys for their organization.

## Solution Specification

### 1. Database Schema
Add a new PostgreSQL table `api_keys`:
- `id`: UUID (Primary Key)
- `org_name`: VARCHAR (References `organizations.name`)
- `name`: VARCHAR (User-defined name for the key)
- `key_hash`: VARCHAR (Bcrypt/Argon2 hash of the actual key)
- `prefix`: VARCHAR (First 4-8 chars of the key for display purposes)
- `created_at`: TIMESTAMP
- `expires_at`: TIMESTAMP (Nullable)

### 2. Backend API Endpoints (Go)
Implement the following routes under `/api/v1/org/{org}/apikeys`:
- `GET /`: List all API keys for the organization (returns ID, Name, Prefix, CreatedAt).
- `POST /`: Generate a new API key. Return the *raw* key exactly once in the response, and store the `key_hash` and `prefix` in the DB.
- `DELETE /{id}`: Revoke/delete an API key.

### 3. Frontend Integration (SvelteKit)
Update `dashboard/src/routes/(app)/org/[org]/apikeys/+page.svelte`:
- Implement a data load function (`+page.server.ts` or standard fetch) to list the keys.
- Wire the "Generate New Key" button to the `POST` endpoint.
- Display the generated raw key in a copyable modal component (warning the user it will only be shown once).
- Add a "Revoke" button to each row in the table, wired to the `DELETE` endpoint.

## Security Constraints
- **Never** store raw API keys in plaintext in the database.
- Use PASETO `v4.local` or secure token generation for the raw keys themselves.
- Ensure the API endpoints are protected by the existing `authzMW` middleware to guarantee the user belongs to the requested `{org}`.
