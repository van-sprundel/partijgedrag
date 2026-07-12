CREATE TABLE IF NOT EXISTS cabinet_periods (
  period_key text PRIMARY KEY,
  jurisdiction_key text NOT NULL REFERENCES jurisdictions(jurisdiction_key),
  name text NOT NULL,
  started_on date NOT NULL,
  ended_on date,
  parties text[] NOT NULL DEFAULT '{}'::text[],
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK (ended_on IS NULL OR ended_on > started_on)
);

CREATE INDEX IF NOT EXISTS cabinet_periods_jurisdiction_dates_idx
  ON cabinet_periods (jurisdiction_key, started_on DESC);

INSERT INTO cabinet_periods (period_key, jurisdiction_key, name, started_on, ended_on, parties)
VALUES
  ('rutte-iii', 'nl-tweede-kamer', 'Rutte III', '2017-10-26', '2022-01-10', ARRAY['VVD', 'CDA', 'D66', 'CU']),
  ('rutte-iv', 'nl-tweede-kamer', 'Rutte IV', '2022-01-10', '2024-07-02', ARRAY['VVD', 'D66', 'CDA', 'CU']),
  ('schoof-i', 'nl-tweede-kamer', 'Schoof I', '2024-07-02', '2026-02-23', ARRAY['PVV', 'VVD', 'NSC', 'BBB']),
  ('jetten-i', 'nl-tweede-kamer', 'Jetten I', '2026-02-23', NULL, ARRAY['D66', 'VVD', 'CDA'])
ON CONFLICT (period_key) DO UPDATE
SET name = EXCLUDED.name,
    started_on = EXCLUDED.started_on,
    ended_on = EXCLUDED.ended_on,
    parties = EXCLUDED.parties,
    updated_at = now();
