-- Create materialized view for party likeness per motion
-- This depends on the 'majority_party_votes' materialized view
CREATE MATERIALIZED VIEW party_likeness_per_motion AS
SELECT
    mv1.fractie_id as fractie1_id,
    mv2.fractie_id as fractie2_id,
    mv1.zaak_id,
    mv1.gestart_op,
    (mv1.vote_type = mv2.vote_type) as same_vote
FROM majority_party_votes mv1
JOIN majority_party_votes mv2 ON mv1.zaak_id = mv2.zaak_id
WHERE mv1.fractie_id < mv2.fractie_id;

-- Create indexes to speed up queries on the materialized view
CREATE INDEX idx_plpm_gestart_op ON party_likeness_per_motion(gestart_op);
CREATE INDEX idx_plpm_fractie1_id ON party_likeness_per_motion(fractie1_id);
CREATE INDEX idx_plpm_fractie2_id ON party_likeness_per_motion(fractie2_id);
CREATE INDEX idx_plpm_zaak_id ON party_likeness_per_motion(zaak_id);
CREATE INDEX idx_plpm_same_vote ON party_likeness_per_motion(same_vote);
