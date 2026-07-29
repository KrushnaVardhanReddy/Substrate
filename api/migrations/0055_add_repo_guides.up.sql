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
