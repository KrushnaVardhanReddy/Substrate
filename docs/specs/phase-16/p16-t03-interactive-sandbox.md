# Phase 16 - Task 03: Interactive API Sandbox ("Try It Out")

## 1. Goal
Build a secure API Proxy inside the Substrate Go backend that allows developers to issue live requests to a registered API directly from the Substrate documentation UI using an ephemeral test token — without leaving the browser.

## 2. Requirements
- Add endpoint `POST /api/v1/sandbox/token` — generates a short-lived (15min TTL) ephemeral PASETO token scoped to a specific `org/repo`.
- Add endpoint `POST /api/v1/sandbox/request` — receives `{ method, path, headers, body }` from the UI and proxies the request to the target API's registered `base_url`, injecting the ephemeral token as an `Authorization` header. Returns the upstream response (status, headers, body).
- Store the `base_url` for each registered repo in a new `repo_config` column in the `repos` table (nullable, set via `substrate.yaml`).
- In the Svelte catalog page, add a "Try It" button on each endpoint card that opens a request builder panel (method, path params, body editor) and shows the live response inline.

## 3. Security Constraints
- The sandbox proxy endpoint (`/api/v1/sandbox/request`) MUST be authenticated with a valid org-scoped PASETO token.
- Outbound requests from the proxy MUST timeout after 10 seconds.
- Block requests to private RFC-1918 IP ranges (SSRF prevention).
- Log every proxied request to a `sandbox_audit_log` table: `org`, `repo`, `method`, `path`, `status_code`, `user_id`, `timestamp`.

## 4. Database Schema
```sql
-- Add column to existing repos table
ALTER TABLE repos ADD COLUMN base_url TEXT;

CREATE TABLE sandbox_audit_log (
  id          SERIAL PRIMARY KEY,
  org         TEXT NOT NULL,
  repo        TEXT NOT NULL,
  method      TEXT NOT NULL,
  path        TEXT NOT NULL,
  status_code INT,
  user_id     TEXT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```
