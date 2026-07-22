# P9-T14: Ephemeral API Preview URLs

## Overview
Generate temporary, shareable Substrate dashboard URLs for PRs so engineers can share proposed schema changes and interactive diffs with frontend teams before merging.

## Requirements
1. **Preview Record**: On PR open, create a `preview_sessions` DB record with a unique `token` (UUID) and a `pr_number`.
2. **API Route**: Add `GET /preview/:token` in the API that serves the diff data for a given PR.
3. **Dashboard Route**: Add `GET /preview/:token` in the SvelteKit dashboard that renders the schema diff in read-only mode.
4. **Expiry**: Preview URLs expire 7 days after the PR is merged or closed.
