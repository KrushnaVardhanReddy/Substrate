# Phase 16 - Task 01: Markdown Guide Ingestion (Diátaxis)

## 1. Goal
Pull `docs/**/*.md` files from provider repositories (via the Substrate registry) alongside their OpenAPI specs, and render them as human-readable "Guides" in the Substrate Developer Portal, stitched together with the auto-generated API schema reference.

## 2. Requirements
- During the schema sync webhook (`POST /api/v1/sync`), also clone and extract any Markdown files found under the repo's `docs/` directory.
- Store ingested Markdown content in a new `repo_guides` PostgreSQL table with columns: `org`, `repo`, `file_path`, `title`, `content`, `updated_at`.
- Expose a new endpoint `GET /api/v1/docs/{org}/{repo}` that returns the list of guide documents for a repo.
- Expose `GET /api/v1/docs/{org}/{repo}/{slug}` to return a single rendered guide's content.
- In the Svelte catalog page (`dashboard/src/routes/(app)/org/[org]/catalog/[repo]/+page.svelte`), add a "Guides" tab that lists and renders the ingested Markdown documents alongside the existing API schema reference.

## 3. Database Schema
```sql
CREATE TABLE repo_guides (
  id          SERIAL PRIMARY KEY,
  org         TEXT NOT NULL,
  repo        TEXT NOT NULL,
  file_path   TEXT NOT NULL,
  title       TEXT NOT NULL,
  content     TEXT NOT NULL,
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(org, repo, file_path)
);
```

## 4. Constraints
- Markdown rendering in the UI must use a safe sanitiser (e.g., `marked` + `DOMPurify`) to prevent XSS.
- Max ingested file size: 500KB per document.
- Gracefully handle repos with no `docs/` directory — return empty array, no errors.
- Must include Vitest unit tests for the new Svelte Guides tab component.
