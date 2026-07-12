ALTER TABLE ingestion_runs
ADD COLUMN IF NOT EXISTS cursor_saved boolean NOT NULL DEFAULT false,
ADD COLUMN IF NOT EXISTS stop_reason text;
