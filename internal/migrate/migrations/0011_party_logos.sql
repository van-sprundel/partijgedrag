ALTER TABLE parties
  ADD COLUMN IF NOT EXISTS logo_data bytea,
  ADD COLUMN IF NOT EXISTS logo_content_type text,
  ADD COLUMN IF NOT EXISTS logo_synced_at timestamptz;

-- The party list pages only need to know whether a logo exists; the bytes are
-- fetched separately per party by the image route.
CREATE INDEX IF NOT EXISTS parties_logo_idx
  ON parties (jurisdiction_key, source_id)
  WHERE logo_data IS NOT NULL;
