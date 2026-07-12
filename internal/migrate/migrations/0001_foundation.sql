CREATE TABLE IF NOT EXISTS schema_migrations (
  version text PRIMARY KEY,
  applied_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS jurisdictions (
  jurisdiction_key text PRIMARY KEY,
  kind text NOT NULL CHECK (kind IN ('country', 'municipality')),
  name text NOT NULL,
  country_code text NOT NULL DEFAULT 'NL',
  municipality_code text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (country_code, municipality_code)
);

CREATE TABLE IF NOT EXISTS data_sources (
  source_key text PRIMARY KEY,
  jurisdiction_key text NOT NULL REFERENCES jurisdictions(jurisdiction_key),
  provider text NOT NULL,
  name text NOT NULL,
  base_url text NOT NULL,
  config jsonb NOT NULL DEFAULT '{}'::jsonb,
  enabled boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS source_cursors (
  source_key text NOT NULL REFERENCES data_sources(source_key),
  pipeline text NOT NULL,
  cursor jsonb NOT NULL DEFAULT '{}'::jsonb,
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (source_key, pipeline)
);

CREATE TABLE IF NOT EXISTS ingestion_runs (
  id bigserial PRIMARY KEY,
  source_key text NOT NULL REFERENCES data_sources(source_key),
  pipeline text NOT NULL,
  status text NOT NULL CHECK (status IN ('running', 'succeeded', 'failed')),
  cursor_before jsonb NOT NULL DEFAULT '{}'::jsonb,
  cursor_after jsonb NOT NULL DEFAULT '{}'::jsonb,
  records_seen integer NOT NULL DEFAULT 0,
  records_changed integer NOT NULL DEFAULT 0,
  error_message text,
  started_at timestamptz NOT NULL DEFAULT now(),
  finished_at timestamptz
);

CREATE TABLE IF NOT EXISTS raw_records (
  source_key text NOT NULL REFERENCES data_sources(source_key),
  collection text NOT NULL,
  source_id text NOT NULL,
  source_updated_at timestamptz,
  source_deleted boolean NOT NULL DEFAULT false,
  payload jsonb NOT NULL,
  payload_hash text NOT NULL,
  first_seen_at timestamptz NOT NULL DEFAULT now(),
  last_seen_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (source_key, collection, source_id)
);

CREATE INDEX IF NOT EXISTS raw_records_collection_updated_idx
  ON raw_records (source_key, collection, source_updated_at);

CREATE TABLE IF NOT EXISTS motions (
  motion_key text PRIMARY KEY,
  source_key text NOT NULL REFERENCES data_sources(source_key),
  jurisdiction_key text NOT NULL REFERENCES jurisdictions(jurisdiction_key),
  source_id text NOT NULL,
  number text,
  title text,
  subject text,
  status text,
  kind text,
  parliamentary_year text,
  proposed_at timestamptz,
  source_updated_at timestamptz,
  source_deleted boolean NOT NULL DEFAULT false,
  raw_collection text NOT NULL DEFAULT 'Zaak',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (source_key, source_id)
);

CREATE INDEX IF NOT EXISTS motions_source_updated_idx
  ON motions (source_key, source_updated_at);

CREATE INDEX IF NOT EXISTS motions_jurisdiction_proposed_idx
  ON motions (jurisdiction_key, proposed_at DESC);

INSERT INTO jurisdictions (jurisdiction_key, kind, name, country_code)
VALUES ('nl-tweede-kamer', 'country', 'Tweede Kamer', 'NL')
ON CONFLICT (jurisdiction_key) DO UPDATE
SET name = EXCLUDED.name,
    updated_at = now();

INSERT INTO data_sources (source_key, jurisdiction_key, provider, name, base_url, config)
VALUES (
  'tweedekamer-odata-v2',
  'nl-tweede-kamer',
  'tweedekamer-odata',
  'Tweede Kamer OData 2.0',
  'https://gegevensmagazijn.tweedekamer.nl/OData/v4/2.0',
  '{}'::jsonb
)
ON CONFLICT (source_key) DO UPDATE
SET base_url = EXCLUDED.base_url,
    updated_at = now();
