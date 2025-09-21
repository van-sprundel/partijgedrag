-- Create materialized view for pre-calculating majority party votes on motions
CREATE MATERIALIZED VIEW majority_party_votes AS
WITH PartyMotionVotes AS (
    SELECT
        b.zaak_id,
        z.gestart_op,
        f.id as fractie_id,
        s.soort AS vote_type,
        COUNT(*) AS vote_count
    FROM stemmingen s
    JOIN besluiten b ON s.besluit_id = b.id
    JOIN zaken z ON b.zaak_id = z.id
    JOIN fracties f ON s.actor_fractie = f.naam_nl OR s.actor_fractie = f.afkorting
    WHERE s.actor_fractie IS NOT NULL
      AND s.soort IN ('Voor', 'Tegen')
      AND z.soort = 'Motie'
    GROUP BY b.zaak_id, z.gestart_op, f.id, s.soort
)
SELECT
    zaak_id,
    gestart_op,
    fractie_id,
    vote_type
FROM (
    SELECT
        zaak_id,
        gestart_op,
        fractie_id,
        vote_type,
        ROW_NUMBER() OVER(PARTITION BY zaak_id, fractie_id ORDER BY vote_count DESC) as rn
    FROM PartyMotionVotes
) ranked_votes
WHERE rn = 1;

-- Create indexes to speed up queries on the materialized view
CREATE INDEX idx_majority_party_votes_zaak_id ON majority_party_votes(zaak_id);
CREATE INDEX idx_majority_party_votes_fractie_id ON majority_party_votes(fractie_id);
CREATE INDEX idx_majority_party_votes_gestart_op ON majority_party_votes(gestart_op);
