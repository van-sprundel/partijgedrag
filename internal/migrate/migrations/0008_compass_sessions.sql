CREATE TABLE IF NOT EXISTS compass_sessions (
  session_key text PRIMARY KEY,
  jurisdiction_key text NOT NULL REFERENCES jurisdictions(jurisdiction_key),
  answers jsonb NOT NULL,
  min_overlap integer NOT NULL DEFAULT 5,
  created_at timestamptz NOT NULL DEFAULT now()
);
