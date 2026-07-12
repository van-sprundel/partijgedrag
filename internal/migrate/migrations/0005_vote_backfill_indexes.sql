CREATE INDEX IF NOT EXISTS motions_vote_backfill_idx
  ON motions (source_key, source_deleted, votes_synced_at, proposed_at DESC);
