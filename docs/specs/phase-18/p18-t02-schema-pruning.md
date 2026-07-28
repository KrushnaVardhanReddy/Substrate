# Phase 18 - Task 02: Context-Aware Schema Pruning

## 1. Goal
Implement a "Pruned Schema API" that accepts natural language user intent and returns only the relevant subset of the OpenAPI spec, dramatically reducing token usage when feeding schemas to LLMs.

## 2. Requirements
- Add `POST /api/v1/schema/prune` with body `{ org, repo, intent: "string" }`.
- Use the existing AI handler (`api/internal/ai/`) to embed the intent string and compare it semantically against each endpoint's `summary` and `description` fields.
- Return a valid OpenAPI JSON object containing only the top-N (configurable, default 5) most semantically relevant endpoints, with all unrelated paths stripped.
- Cache pruned results in a `schema_prune_cache` table keyed on `(org, repo, intent_hash)` with a 30-minute TTL.
- Expose a `max_endpoints` query param to let callers control the returned set size.

## 3. Pruning Algorithm
```
1. Fetch full OpenAPI spec for org/repo from registry.
2. For each path+method, compute cosine similarity between
   intent embedding and endpoint summary embedding.
3. Sort descending by similarity score.
4. Return top-N paths as a stripped OpenAPI object.
```

## 4. Database Schema
```sql
CREATE TABLE schema_prune_cache (
  org           TEXT NOT NULL,
  repo          TEXT NOT NULL,
  intent_hash   TEXT NOT NULL,
  pruned_schema JSONB NOT NULL,
  computed_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (org, repo, intent_hash)
);
```
