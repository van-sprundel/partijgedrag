-- Matches the `ORDER BY proposed_at DESC NULLS LAST, motion_key` used by the
-- motion lists and the voting compass. motions_jurisdiction_proposed_idx sorts
-- NULLs first, so the planner cannot use it for that order and instead sorts
-- every motion in the jurisdiction before applying LIMIT. With this index a
-- LIMIT query stops scanning once it has enough rows, which matters when the
-- per-row filters are expensive (the compass checks each motion's votes).
CREATE INDEX IF NOT EXISTS motions_jurisdiction_recent_idx
  ON motions (jurisdiction_key, proposed_at DESC NULLS LAST, motion_key);

ANALYZE motions;
