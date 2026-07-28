CREATE TABLE schema_prune_cache (
  org           TEXT NOT NULL,
  repo          TEXT NOT NULL,
  intent_hash   TEXT NOT NULL,
  pruned_schema JSONB NOT NULL,
  computed_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (org, repo, intent_hash)
);
