# P9-T05: Hexagonal Architecture & `sqlc` Refactor

## Overview
Formally isolate the engine from HTTP transports and migrate raw `pgx` queries to `sqlc` for a type-safe DB layer.

## Requirements
1. **Per-Table Store Interfaces**: Split the current monolithic `db.Store` interface into per-table interfaces (e.g., `RepoStore`, `ChangeStore`, `OrgStore`). This is **mandatory** to avoid merge conflicts during parallel development.
2. **sqlc Migration**: Write `sqlc` query files under `api/internal/db/queries/` and replace all raw `pgx.Query` calls.
3. **Hexagonal Ports**: Define clean interfaces in `api/internal/ports/` for each domain service.
4. **Transport Isolation**: Ensure HTTP handlers only call port interfaces, never directly touching the DB package.
