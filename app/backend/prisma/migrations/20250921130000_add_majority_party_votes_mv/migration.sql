-- Create materialized view for party votes on motions (simplified approach)
-- This directly looks at stemmingen (votes) to determine party positions
CREATE MATERIALIZED VIEW majority_party_votes AS
SELECT DISTINCT
    b.zaak_id,
    z.gestart_op,
    f.id as fractie_id,
    s.soort AS vote_type
FROM stemmingen s
JOIN besluiten b ON s.besluit_id = b.id
JOIN zaken z ON b.zaak_id = z.id
JOIN fracties f ON (s.actor_fractie = f.naam_nl OR s.actor_fractie = f.afkorting)
WHERE s.actor_fractie IS NOT NULL
  AND s.soort IN ('Voor', 'Tegen')
  AND z.soort = 'Motie'
  AND f.datum_inactief IS NULL;

-- Create indexes to speed up queries on the materialized view
CREATE INDEX idx_majority_party_votes_zaak_id ON majority_party_votes(zaak_id);
CREATE INDEX idx_majority_party_votes_fractie_id ON majority_party_votes(fractie_id);
CREATE INDEX idx_majority_party_votes_gestart_op ON majority_party_votes(gestart_op);
CREATE INDEX idx_majority_party_votes_vote_type ON majority_party_votes(vote_type);
