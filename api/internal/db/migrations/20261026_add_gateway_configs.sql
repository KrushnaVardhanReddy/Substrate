CREATE TABLE gateway_configs (
  id          SERIAL PRIMARY KEY,
  org         TEXT NOT NULL,
  repo        TEXT NOT NULL,
  gateway_type TEXT NOT NULL,
  crd_yaml    TEXT NOT NULL,
  pr_url      TEXT,
  generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
