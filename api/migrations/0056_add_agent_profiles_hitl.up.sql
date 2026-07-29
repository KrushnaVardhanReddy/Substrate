CREATE TABLE agent_profiles (
  id            SERIAL PRIMARY KEY,
  org           TEXT NOT NULL,
  name          TEXT NOT NULL,
  allowed_tools TEXT[] NOT NULL,
  hitl_enabled  BOOLEAN NOT NULL DEFAULT false,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE hitl_queue (
  id            SERIAL PRIMARY KEY,
  org           TEXT NOT NULL,
  profile_id    INT REFERENCES agent_profiles(id) ON DELETE CASCADE,
  tool_name     TEXT NOT NULL,
  arguments     JSONB NOT NULL,
  status        TEXT NOT NULL DEFAULT 'pending', -- pending | approved | rejected
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  resolved_at   TIMESTAMPTZ
);
