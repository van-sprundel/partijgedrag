import { implement } from "@orpc/server";
import {
	apiContract,
	type PartyCategoryLikeness,
	type PartyFocusCategory,
	type PartyLikeness,
} from "../contracts/index.js";
import { sql, sqlOneOrNull } from "../services/db/sql-tag.js";
import { mapPartyToContract } from "../utils/mappers.js";

const os = implement(apiContract);

export const statisticsRouter = {
	getPartyLikenessMatrix: os.statistics.getPartyLikenessMatrix.handler(
		async ({ input }) => {
			const { dateFrom, dateTo } = input || {};

			// Simplified approach: directly compare votes between parties
			// Based on the PHP implementation which looks at all stemmingen
			const results: PartyLikeness[] = await sql<PartyLikeness>`
				WITH PartyVotes AS (
					SELECT DISTINCT
						d.case_id,
						p.id as party_id,
						p.short_name as party_name,
						v.type as vote_type
					FROM votes v
					JOIN decisions d ON v.decision_id = d.id
					JOIN cases c ON d.case_id = c.id
					JOIN parties p ON v.party_id = p.id
					WHERE v.party_id IS NOT NULL
						AND v.type IN ('Voor', 'Tegen')
						AND c.type = 'Motie'
						AND p.active_to IS NULL
						${dateFrom && dateTo ? sql`AND c.started_at BETWEEN ${dateFrom} AND ${dateTo}` : sql``}
				),
				PartyComparisons AS (
					SELECT
						p1.party_id as party1_id,
						p1.party_name as party1_name,
						p2.party_id as party2_id,
						p2.party_name as party2_name,
						COUNT(*) as common_motions,
						SUM(CASE WHEN p1.vote_type = p2.vote_type THEN 1 ELSE 0 END) as same_votes
					FROM PartyVotes p1
					JOIN PartyVotes p2 ON p1.case_id = p2.case_id AND p1.party_id < p2.party_id
					GROUP BY p1.party_id, p1.party_name, p2.party_id, p2.party_name
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
				ORDER BY party1_name, "likenessPercentage" DESC
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

		const party = await sqlOneOrNull<{
			id: string;
			name_nl: string | null;
			short_name: string | null;
			seats: bigint | null;
			active_from: Date | null;
			active_to: Date | null;
			content_type: string | null;
			logo_data: Buffer | null;
			updated_at: Date | null;
			api_updated_at: Date | null;
		}>`
			SELECT id, name_nl, short_name, seats, active_from, active_to, content_type, logo_data, updated_at, api_updated_at
			FROM parties
			WHERE id = ${partyId}
		`;

		if (!party) {
			return null;
		}

		const results: PartyFocusCategory[] = await sql<PartyFocusCategory>`
			SELECT
				mc.id AS "categoryId",
				mc.name AS "categoryName",
				mc.type AS "categoryType",
				COUNT(c.id) AS "motionCount"
			FROM
				case_actors ca
			JOIN
				cases c ON ca.case_id = c.id
			JOIN
				parties p ON ca.actor_party = p.name_nl OR ca.actor_party = p.short_name
			JOIN
				case_categories cc ON c.id = cc.case_id
			JOIN
				motion_categories mc ON cc.category_id = mc.id
			WHERE
				p.id = ${partyId}
				AND ca.relation = 'Indiener'
				AND c.type = 'Motie'
				${dateFrom && dateTo ? sql`AND c.started_at BETWEEN ${dateFrom} AND ${dateTo}` : sql``}
			GROUP BY
				mc.id, mc.name, mc.type
			ORDER BY
				"motionCount" DESC
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
			const results: PartyCategoryLikeness[] = await sql<PartyCategoryLikeness>`
				WITH PartyVotesByCategory AS (
					SELECT DISTINCT
						d.case_id,
						p.id as party_id,
						v.type as vote_type,
						cc.category_id
					FROM votes v
					JOIN decisions d ON v.decision_id = d.id
					JOIN cases c ON d.case_id = c.id
					JOIN parties p ON v.party_id = p.id
					JOIN case_categories cc ON c.id = cc.case_id
					WHERE v.party_id IS NOT NULL
						AND v.type IN ('Voor', 'Tegen')
						AND c.type = 'Motie'
						AND p.active_to IS NULL
						${dateFrom && dateTo ? sql`AND c.started_at BETWEEN ${dateFrom} AND ${dateTo}` : sql``}
				),
				CategoryComparisons AS (
					SELECT
						pv1.category_id,
						pv2.party_id as other_party_id,
						COUNT(*) as total_votes,
						SUM(CASE WHEN pv1.vote_type = pv2.vote_type THEN 1 ELSE 0 END) as same_votes
					FROM PartyVotesByCategory pv1
					JOIN PartyVotesByCategory pv2 ON pv1.case_id = pv2.case_id
						AND pv1.category_id = pv2.category_id
						AND pv1.party_id != pv2.party_id
					WHERE pv1.party_id = ${partyId}
					GROUP BY pv1.category_id, pv2.party_id
				)
				SELECT
					mc.id AS "categoryId",
					mc.name AS "categoryName",
					p.id AS "party2Id",
					p.short_name AS "party2Name",
					CASE
						WHEN cc.total_votes > 0
						THEN (cc.same_votes::float / cc.total_votes::float) * 100
						ELSE 0
					END AS "likenessPercentage"
				FROM CategoryComparisons cc
				JOIN motion_categories mc ON cc.category_id = mc.id
				JOIN parties p ON cc.other_party_id = p.id
				WHERE p.active_to IS NULL
				ORDER BY mc.name, p.short_name
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
