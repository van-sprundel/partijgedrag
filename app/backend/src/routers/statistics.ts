import { implement } from "@orpc/server";
import {
	apiContract,
	type PartyCategoryLikeness,
	type PartyFocusCategory,
	type PartyLikeness,
} from "../contracts/index.js";
import { sql, sqlOneOrNull } from "../lib/db.js";
import type { PartyMapped } from "../lib/db-types.js";
import { mapPartyToContract } from "../utils/mappers.js";

const os = implement(apiContract);

export const statisticsRouter = {
	getPartyLikenessMatrix: os.statistics.getPartyLikenessMatrix.handler(
		async ({ input }) => {
			const { dateFrom, dateTo } = input || {};

			// Simplified approach: directly compare votes between parties
			// Based on the PHP implementation which looks at all stemmingen
			const results: PartyLikeness[] = dateFrom && dateTo
				? await sql`
				WITH PartyVotes AS (
					SELECT DISTINCT
						b.zaak_id,
						f.id as fractie_id,
						f.afkorting as fractie_name,
						s.soort as vote_type
					FROM stemmingen s
					JOIN besluiten b ON s.besluit_id = b.id
					JOIN zaken z ON b.zaak_id = z.id
					JOIN fracties f ON s.fractie_id = f.id
					WHERE s.fractie_id IS NOT NULL
						AND s.soort IN ('Voor', 'Tegen')
						AND z.soort = 'Motie'
						AND b.stemming_datum BETWEEN ${dateFrom} AND ${dateTo}
				),
				PartyComparisons AS (
					SELECT
						p1.fractie_id as party1_id,
						p1.fractie_name as party1_name,
						p2.fractie_id as party2_id,
						p2.fractie_name as party2_name,
						COUNT(*) as common_motions,
						SUM(CASE WHEN p1.vote_type = p2.vote_type THEN 1 ELSE 0 END) as same_votes
					FROM PartyVotes p1
					JOIN PartyVotes p2 ON p1.zaak_id = p2.zaak_id AND p1.fractie_id < p2.fractie_id
					GROUP BY p1.fractie_id, p1.fractie_name, p2.fractie_id, p2.fractie_name
				)
				SELECT
					party1_id AS "party1Id",
					party1_name AS "party1Name",
					party2_id AS "party2Id",
					party2_name AS "party2Name",
					common_motions AS "commonMotions",
					same_votes AS "sameVotes",
					CASE
						WHEN common_motions > 0
						THEN (same_votes::float / common_motions::float) * 100
						ELSE 0
					END AS "likenessPercentage"
				FROM PartyComparisons
				ORDER BY party1_name, "likenessPercentage" DESC;
			`
				: await sql`
				WITH PartyVotes AS (
					SELECT DISTINCT
						b.zaak_id,
						f.id as fractie_id,
						f.afkorting as fractie_name,
						s.soort as vote_type
					FROM stemmingen s
					JOIN besluiten b ON s.besluit_id = b.id
					JOIN zaken z ON b.zaak_id = z.id
					JOIN fracties f ON s.fractie_id = f.id
					WHERE s.fractie_id IS NOT NULL
						AND s.soort IN ('Voor', 'Tegen')
						AND z.soort = 'Motie'
				),
				PartyComparisons AS (
					SELECT
						p1.fractie_id as party1_id,
						p1.fractie_name as party1_name,
						p2.fractie_id as party2_id,
						p2.fractie_name as party2_name,
						COUNT(*) as common_motions,
						SUM(CASE WHEN p1.vote_type = p2.vote_type THEN 1 ELSE 0 END) as same_votes
					FROM PartyVotes p1
					JOIN PartyVotes p2 ON p1.zaak_id = p2.zaak_id AND p1.fractie_id < p2.fractie_id
					GROUP BY p1.fractie_id, p1.fractie_name, p2.fractie_id, p2.fractie_name
				)
				SELECT
					party1_id AS "party1Id",
					party1_name AS "party1Name",
					party2_id AS "party2Id",
					party2_name AS "party2Name",
					common_motions AS "commonMotions",
					same_votes AS "sameVotes",
					CASE
						WHEN common_motions > 0
						THEN (same_votes::float / common_motions::float) * 100
						ELSE 0
					END AS "likenessPercentage"
				FROM PartyComparisons
				ORDER BY party1_name, "likenessPercentage" DESC;
			`;

			// Also get the reverse relationships (party2 to party1)
			const reverseResults: PartyLikeness[] = results.map((r) => ({
				party1Id: r.party2Id,
				party1Name: r.party2Name,
				party2Id: r.party1Id,
				party2Name: r.party1Name,
				commonMotions: r.commonMotions,
				sameVotes: r.sameVotes,
				likenessPercentage: r.likenessPercentage,
			}));

			const allResults = [...results, ...reverseResults];

			return allResults.map((r) => ({
				...r,
				commonMotions: Number(r.commonMotions),
				sameVotes: Number(r.sameVotes),
				likenessPercentage: r.likenessPercentage
					? Number(r.likenessPercentage)
					: 0,
			}));
		},
	),

	getPartyFocus: os.statistics.getPartyFocus.handler(async ({ input }) => {
		const { partyId, dateFrom, dateTo } = input;

		const party = await sqlOneOrNull<PartyMapped>`
			SELECT
				id,
				nummer as number,
				afkorting as "shortName",
				naam_nl as "nameNl",
				naam_en as "nameEn",
				aantal_zetels as seats,
				aantal_stemmen as "votesCount",
				datum_actief as "activeFrom",
				datum_inactief as "activeTo",
				content_type as "contentType",
				content_length as "contentLength",
				gewijzigd_op as "updatedAt",
				api_gewijzigd_op as "apiUpdatedAt",
				logo_data as "logoData",
				verwijderd as removed
			FROM fracties
			WHERE id = ${partyId}
		`;

		if (!party) {
			return null;
		}

		const results: PartyFocusCategory[] = dateFrom && dateTo
			? await sql`
			SELECT
				mc.id AS "categoryId",
				mc.name AS "categoryName",
				mc.type AS "categoryType",
				COUNT(DISTINCT z.id) AS "motionCount"
			FROM
				zaak_actors za
			JOIN
				zaken z ON za.zaak_id = z.id
			JOIN
				fracties f ON za.actor_fractie = f.naam_nl OR za.actor_fractie = f.afkorting
			JOIN
				zaak_categories zc ON z.id = zc.zaak_id
			JOIN
				motion_categories mc ON zc.category_id = mc.id
			JOIN
				besluiten b ON z.id = b.zaak_id
			WHERE
				f.id = ${partyId}
				AND za.relatie = 'Indiener'
				AND z.soort = 'Motie'
				AND b.stemming_datum BETWEEN ${dateFrom} AND ${dateTo}
			GROUP BY
				mc.id, mc.name, mc.type
			ORDER BY
				"motionCount" DESC;
		`
			: await sql`
			SELECT
				mc.id AS "categoryId",
				mc.name AS "categoryName",
				mc.type AS "categoryType",
				COUNT(DISTINCT z.id) AS "motionCount"
			FROM
				zaak_actors za
			JOIN
				zaken z ON za.zaak_id = z.id
			JOIN
				fracties f ON za.actor_fractie = f.naam_nl OR za.actor_fractie = f.afkorting
			JOIN
				zaak_categories zc ON z.id = zc.zaak_id
			JOIN
				motion_categories mc ON zc.category_id = mc.id
			JOIN
				besluiten b ON z.id = b.zaak_id
			WHERE
				f.id = ${partyId}
				AND za.relatie = 'Indiener'
				AND z.soort = 'Motie'
			GROUP BY
				mc.id, mc.name, mc.type
			ORDER BY
				"motionCount" DESC;
		`;

		return {
			party: mapPartyToContract(party),
			categories: results.map((r) => ({
				...r,
				motionCount: Number(r.motionCount),
			})),
		};
	}),

	getPartyCategoryLikeness: os.statistics.getPartyCategoryLikeness.handler(
		async ({ input }) => {
			const { partyId, dateFrom, dateTo } = input;

			// Simplified approach: directly analyze votes by category
			const results: PartyCategoryLikeness[] = dateFrom && dateTo
				? await sql`
				WITH PartyVotesByCategory AS (
					SELECT DISTINCT
						b.zaak_id,
						f.id as fractie_id,
						s.soort as vote_type,
						zc.category_id
					FROM stemmingen s
					JOIN besluiten b ON s.besluit_id = b.id
					JOIN zaken z ON b.zaak_id = z.id
					JOIN fracties f ON s.fractie_id = f.id
					JOIN zaak_categories zc ON z.id = zc.zaak_id
					WHERE s.fractie_id IS NOT NULL
						AND s.soort IN ('Voor', 'Tegen')
						AND z.soort = 'Motie'
						AND b.stemming_datum BETWEEN ${dateFrom} AND ${dateTo}
				),
				CategoryComparisons AS (
					SELECT
						pv1.category_id,
						pv2.fractie_id as other_party_id,
						COUNT(*) as total_votes,
						SUM(CASE WHEN pv1.vote_type = pv2.vote_type THEN 1 ELSE 0 END) as same_votes
					FROM PartyVotesByCategory pv1
					JOIN PartyVotesByCategory pv2 ON pv1.zaak_id = pv2.zaak_id
						AND pv1.category_id = pv2.category_id
						AND pv1.fractie_id != pv2.fractie_id
					WHERE pv1.fractie_id = ${partyId}
					GROUP BY pv1.category_id, pv2.fractie_id
				)
				SELECT
					mc.id AS "categoryId",
					mc.name AS "categoryName",
					f.id AS "party2Id",
					f.afkorting AS "party2Name",
					CASE
						WHEN cc.total_votes > 0
						THEN (cc.same_votes::float / cc.total_votes::float) * 100
						ELSE 0
					END AS "likenessPercentage"
				FROM CategoryComparisons cc
				JOIN motion_categories mc ON cc.category_id = mc.id
				JOIN fracties f ON cc.other_party_id = f.id
				ORDER BY mc.name, f.afkorting;
			`
				: await sql`
				WITH PartyVotesByCategory AS (
					SELECT DISTINCT
						b.zaak_id,
						f.id as fractie_id,
						s.soort as vote_type,
						zc.category_id
					FROM stemmingen s
					JOIN besluiten b ON s.besluit_id = b.id
					JOIN zaken z ON b.zaak_id = z.id
					JOIN fracties f ON s.fractie_id = f.id
					JOIN zaak_categories zc ON z.id = zc.zaak_id
					WHERE s.fractie_id IS NOT NULL
						AND s.soort IN ('Voor', 'Tegen')
						AND z.soort = 'Motie'
				),
				CategoryComparisons AS (
					SELECT
						pv1.category_id,
						pv2.fractie_id as other_party_id,
						COUNT(*) as total_votes,
						SUM(CASE WHEN pv1.vote_type = pv2.vote_type THEN 1 ELSE 0 END) as same_votes
					FROM PartyVotesByCategory pv1
					JOIN PartyVotesByCategory pv2 ON pv1.zaak_id = pv2.zaak_id
						AND pv1.category_id = pv2.category_id
						AND pv1.fractie_id != pv2.fractie_id
					WHERE pv1.fractie_id = ${partyId}
					GROUP BY pv1.category_id, pv2.fractie_id
				)
				SELECT
					mc.id AS "categoryId",
					mc.name AS "categoryName",
					f.id AS "party2Id",
					f.afkorting AS "party2Name",
					CASE
						WHEN cc.total_votes > 0
						THEN (cc.same_votes::float / cc.total_votes::float) * 100
						ELSE 0
					END AS "likenessPercentage"
				FROM CategoryComparisons cc
				JOIN motion_categories mc ON cc.category_id = mc.id
				JOIN fracties f ON cc.other_party_id = f.id
				ORDER BY mc.name, f.afkorting;
			`;

			return results.map((r) => ({
				...r,
				likenessPercentage: r.likenessPercentage
					? Number(r.likenessPercentage)
					: 0,
			}));
		},
	),
};
