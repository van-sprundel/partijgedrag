ALTER TABLE motions
ADD COLUMN IF NOT EXISTS votes_synced_at timestamptz;

CREATE TABLE IF NOT EXISTS decisions (
  decision_key text PRIMARY KEY,
  source_key text NOT NULL REFERENCES data_sources(source_key),
  motion_key text NOT NULL REFERENCES motions(motion_key) ON DELETE CASCADE,
  source_id text NOT NULL,
  agenda_point_source_id text,
  voting_type text,
  decision_type text,
  decision_text text,
  comment text,
  status text,
  decision_order integer,
  source_updated_at timestamptz,
  source_deleted boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (source_key, source_id)
);

CREATE INDEX IF NOT EXISTS decisions_motion_idx
  ON decisions (motion_key);

CREATE INDEX IF NOT EXISTS decisions_source_updated_idx
  ON decisions (source_key, source_updated_at);

CREATE TABLE IF NOT EXISTS votes (
  vote_key text PRIMARY KEY,
  source_key text NOT NULL REFERENCES data_sources(source_key),
  motion_key text NOT NULL REFERENCES motions(motion_key) ON DELETE CASCADE,
  decision_key text NOT NULL REFERENCES decisions(decision_key) ON DELETE CASCADE,
  source_id text NOT NULL,
  vote_type text,
  party_source_id text,
  party_name text,
  actor_name text,
  party_size integer,
  mistake boolean NOT NULL DEFAULT false,
  person_source_id text,
  source_updated_at timestamptz,
  source_deleted boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (source_key, source_id)
);

CREATE INDEX IF NOT EXISTS votes_motion_idx
  ON votes (motion_key);

CREATE INDEX IF NOT EXISTS votes_decision_idx
  ON votes (decision_key);

CREATE INDEX IF NOT EXISTS votes_party_idx
  ON votes (party_source_id);
