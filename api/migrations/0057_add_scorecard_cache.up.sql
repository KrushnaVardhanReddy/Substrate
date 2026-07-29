CREATE TABLE scorecard_cache (
  org         TEXT NOT NULL,
  repo        TEXT NOT NULL,
  grade       TEXT NOT NULL,
  total       INT NOT NULL,
  breakdown   JSONB NOT NULL,
  computed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (org, repo)
);

CREATE TABLE repo_compliance (
  org            TEXT NOT NULL,
  repo           TEXT NOT NULL,
  soc2_compliant BOOLEAN NOT NULL DEFAULT false,
  has_pii        BOOLEAN NOT NULL DEFAULT false,
  PRIMARY KEY (org, repo)
);
