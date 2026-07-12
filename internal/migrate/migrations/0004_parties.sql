CREATE TABLE IF NOT EXISTS parties (
  party_key text PRIMARY KEY,
  source_key text NOT NULL REFERENCES data_sources(source_key),
  jurisdiction_key text NOT NULL REFERENCES jurisdictions(jurisdiction_key),
  source_id text NOT NULL,
  number integer,
  short_name text,
  name text,
  name_en text,
  seats integer,
  electoral_votes integer,
  active_from date,
  active_to date,
  source_updated_at timestamptz,
  source_deleted boolean NOT NULL DEFAULT false,
  raw_collection text NOT NULL DEFAULT 'Fractie',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (source_key, source_id)
);

CREATE INDEX IF NOT EXISTS parties_jurisdiction_active_idx
  ON parties (jurisdiction_key, active_to, short_name);

CREATE INDEX IF NOT EXISTS parties_source_updated_idx
  ON parties (source_key, source_updated_at);
