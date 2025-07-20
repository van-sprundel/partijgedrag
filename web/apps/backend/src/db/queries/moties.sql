/* @name get */
SELECT z.id, z.nummer, z.onderwerp, z.titel, z.updated_at, z.created_at, z.status, z.gestart_op, z.vergaderjaar, v.besluit_tekst, v.party_votes FROM "zaken" as z
INNER JOIN "besluiten"  as b ON z.id = b.zaak_id
INNER JOIN "voting_results" as v ON b.id = v.besluit_id
LIMIT :limit! OFFSET :offset!;