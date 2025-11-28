import { useClient } from "../services/db/client.js";

/**
 * Refreshes materialized views in the correct order
 * Call this after data changes to update the precomputed statistics
 */
export async function refreshMaterializedViews(): Promise<void> {
	try {
		using client = await useClient();

		// Refresh majority_party_votes first (no dependencies)
		await client.query("REFRESH MATERIALIZED VIEW majority_party_votes;");

		// Then refresh party_likeness_per_motion (depends on majority_party_votes)
		await client.query("REFRESH MATERIALIZED VIEW party_likeness_per_motion;");

		console.log("Materialized views refreshed successfully");
	} catch (error) {
		console.error("Error refreshing materialized views:", error);
		throw error;
	}
}

/**
 * Refreshes materialized views concurrently (use with caution)
 * Only use this if you're sure the views don't depend on each other
 */
export async function refreshMaterializedViewsConcurrently(): Promise<void> {
	try {
		using client = await useClient();

		await Promise.all([
			client.query("REFRESH MATERIALIZED VIEW CONCURRENTLY majority_party_votes;"),
			client.query("REFRESH MATERIALIZED VIEW CONCURRENTLY party_likeness_per_motion;"),
		]);

		console.log("Materialized views refreshed concurrently");
	} catch (error) {
		console.error("Error refreshing materialized views concurrently:", error);
		throw error;
	}
}

/**
 * Drops and recreates materialized views (for schema changes)
 */
export async function recreateMaterializedViews(): Promise<void> {
	try {
		using client = await useClient();

		// Drop views in reverse dependency order
		await client.query("DROP MATERIALIZED VIEW IF EXISTS party_likeness_per_motion;");
		await client.query("DROP MATERIALIZED VIEW IF EXISTS majority_party_votes;");

		// Recreate majority_party_votes
		await client.query(`
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
		`);

		// Recreate indexes for majority_party_votes
		await client.query("CREATE INDEX idx_majority_party_votes_zaak_id ON majority_party_votes(zaak_id);");
		await client.query("CREATE INDEX idx_majority_party_votes_fractie_id ON majority_party_votes(fractie_id);");
		await client.query("CREATE INDEX idx_majority_party_votes_gestart_op ON majority_party_votes(gestart_op);");
		await client.query("CREATE INDEX idx_majority_party_votes_vote_type ON majority_party_votes(vote_type);");

		// Recreate party_likeness_per_motion
		await client.query(`
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
		`);

		// Recreate indexes for party_likeness_per_motion
		await client.query("CREATE INDEX idx_plpm_gestart_op ON party_likeness_per_motion(gestart_op);");
		await client.query("CREATE INDEX idx_plpm_fractie1_id ON party_likeness_per_motion(fractie1_id);");
		await client.query("CREATE INDEX idx_plpm_fractie2_id ON party_likeness_per_motion(fractie2_id);");
		await client.query("CREATE INDEX idx_plpm_zaak_id ON party_likeness_per_motion(zaak_id);");
		await client.query("CREATE INDEX idx_plpm_same_vote ON party_likeness_per_motion(same_vote);");

		console.log("Materialized views recreated successfully");
	} catch (error) {
		console.error("Error recreating materialized views:", error);
		throw error;
	}
}
