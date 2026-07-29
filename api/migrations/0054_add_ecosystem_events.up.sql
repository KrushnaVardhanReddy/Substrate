CREATE TABLE IF NOT EXISTS ecosystem_events (
  id          SERIAL PRIMARY KEY,
  org         TEXT NOT NULL,
  repo        TEXT NOT NULL,
  event_type  TEXT NOT NULL,
  description TEXT,
  event_time  TIMESTAMPTZ NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ecosystem_events_org ON ecosystem_events(org);
