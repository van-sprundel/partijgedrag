-- Partial index backing the party_positions / classified CTEs used by
-- party-likeness, party-focus and the voting compass. Scanning the full votes
-- table is dominated by rows we filter out (mistakes, deleted rows, non
-- Voor/Tegen vote types); this index keeps only what those aggregations need
-- and lets the planner satisfy them with an index-only scan.
CREATE INDEX IF NOT EXISTS votes_position_idx
  ON votes (motion_key, party_source_id, vote_type)
  WHERE source_deleted = false
    AND mistake = false
    AND vote_type IN ('Voor', 'Tegen')
    AND party_source_id IS NOT NULL;

ANALYZE votes;
