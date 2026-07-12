ALTER TABLE motions ADD COLUMN IF NOT EXISTS bullet_points jsonb;
ALTER TABLE motions ADD COLUMN IF NOT EXISTS document_url text;
ALTER TABLE motions ADD COLUMN IF NOT EXISTS document_synced_at timestamptz;

CREATE INDEX IF NOT EXISTS motions_document_backfill_idx
  ON motions (source_key, source_deleted, document_synced_at, proposed_at DESC);
