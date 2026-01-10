import { implement } from "@orpc/server";
import { apiContract } from "../contracts/index.js";
import { sql } from "../services/db/sql-tag.js";

const os = implement(apiContract);

export const analysisRouter = {
	getCoalitionAlignment: os.analysis.getCoalitionAlignment.handler(
		async ({ input }) => {
			const { dateFrom, dateTo } = input || {};

			const results = dateFrom && dateTo
				? await sql<{ 'fractie1Id': string | null; 'fractie2Id': string | null; 'fractie1Name': string | null; 'fractie2Name': string | null; alignmentPct: string; sameVotes: string | null; totalVotes: string }>`
					SELECT
						plpm.fractie1_id AS "fractie1Id",
						plpm.fractie2_id AS "fractie2Id",
						f1.afkorting AS "fractie1Name",
						f2.afkorting AS "fractie2Name",
						ROUND(AVG(CASE WHEN same_vote THEN 1.0 ELSE 0.0 END) * 100, 2) AS "alignmentPct",
						SUM(CASE WHEN same_vote THEN 1 ELSE 0 END) AS "sameVotes",
						COUNT(*) AS "totalVotes"
					FROM party_likeness_per_motion plpm
					JOIN fracties f1 ON plpm.fractie1_id = f1.id
					JOIN fracties f2 ON plpm.fractie2_id = f2.id
					WHERE plpm.gestart_op BETWEEN ${dateFrom} AND ${dateTo}
					GROUP BY plpm.fractie1_id, plpm.fractie2_id, f1.afkorting, f2.afkorting
					HAVING COUNT(*) >= 10
					ORDER BY "alignmentPct" DESC
				`
				: await sql<{ 'fractie1Id': string | null; 'fractie2Id': string | null; 'fractie1Name': string | null; 'fractie2Name': string | null; alignmentPct: string; sameVotes: string | null; totalVotes: string }>`
					SELECT
						plpm.fractie1_id AS "fractie1Id",
						plpm.fractie2_id AS "fractie2Id",
						f1.afkorting AS "fractie1Name",
						f2.afkorting AS "fractie2Name",
						ROUND(AVG(CASE WHEN same_vote THEN 1.0 ELSE 0.0 END) * 100, 2) AS "alignmentPct",
						SUM(CASE WHEN same_vote THEN 1 ELSE 0 END) AS "sameVotes",
						COUNT(*) AS "totalVotes"
					FROM party_likeness_per_motion plpm
					JOIN fracties f1 ON plpm.fractie1_id = f1.id
					JOIN fracties f2 ON plpm.fractie2_id = f2.id
					GROUP BY plpm.fractie1_id, plpm.fractie2_id, f1.afkorting, f2.afkorting
					HAVING COUNT(*) >= 10
					ORDER BY "alignmentPct" DESC
				`;

			return results
				.filter((r) => r.fractie1Id && r.fractie2Id && r.fractie1Name && r.fractie2Name)
				.map((r) => ({
					fractie1Id: r.fractie1Id!,
					fractie2Id: r.fractie2Id!,
					fractie1Name: r.fractie1Name!,
					fractie2Name: r.fractie2Name!,
					alignmentPct: Number(r.alignmentPct),
					sameVotes: Number(r.sameVotes),
					totalVotes: Number(r.totalVotes),
				}));
		},
	),

	getMPDeviations: os.analysis.getMPDeviations.handler(async ({ input }) => {
		const { dateFrom, dateTo, limit = 50 } = input || {};

		const results = dateFrom && dateTo
			? await sql<{ persoonId: string | null; fractieId: string | null; persoonNaam: string | null; fractieNaam: string | null; deviationPct: string; deviationCount: string | null; totalVotes: string }>`
				WITH party_majority AS (
					SELECT
						b.id as besluit_id,
						s.fractie_id,
						s.soort as majority_vote
					FROM stemmingen s
					JOIN besluiten b ON s.besluit_id = b.id
					JOIN zaken z ON b.zaak_id = z.id
					WHERE s.fractie_id IS NOT NULL
					  AND s.soort IN ('Voor', 'Tegen')
					  AND z.soort = 'Motie'
					  AND z.gestart_op BETWEEN ${dateFrom} AND ${dateTo}
				),
				individual_votes AS (
					SELECT
						s.persoon_id,
						s.fractie_id,
						b.id as besluit_id,
						s.soort as vote
					FROM stemmingen s
					JOIN besluiten b ON s.besluit_id = b.id
					JOIN zaken z ON b.zaak_id = z.id
					WHERE s.persoon_id IS NOT NULL
					  AND s.fractie_id IS NOT NULL
					  AND s.soort IN ('Voor', 'Tegen')
					  AND z.soort = 'Motie'
					  AND z.gestart_op BETWEEN ${dateFrom} AND ${dateTo}
				)
				SELECT
					iv.persoon_id AS "persoonId",
					iv.fractie_id AS "fractieId",
					COALESCE(p.roepnaam, p.voornamen) || ' ' || COALESCE(p.tussenvoegsel || ' ', '') || p.achternaam AS "persoonNaam",
					f.afkorting AS "fractieNaam",
					ROUND(AVG(CASE WHEN iv.vote != pm.majority_vote THEN 1.0 ELSE 0.0 END) * 100, 2) AS "deviationPct",
					SUM(CASE WHEN iv.vote != pm.majority_vote THEN 1 ELSE 0 END) AS "deviationCount",
					COUNT(*) AS "totalVotes"
				FROM individual_votes iv
				JOIN party_majority pm ON iv.besluit_id = pm.besluit_id AND iv.fractie_id = pm.fractie_id
				JOIN personen p ON iv.persoon_id = p.id
				JOIN fracties f ON iv.fractie_id = f.id
				GROUP BY iv.persoon_id, iv.fractie_id, p.roepnaam, p.voornamen, p.tussenvoegsel, p.achternaam, f.afkorting
				HAVING COUNT(*) >= 20
				ORDER BY "deviationPct" DESC
				LIMIT ${limit}
			`
			: await sql<{ persoonId: string | null; fractieId: string | null; persoonNaam: string | null; fractieNaam: string | null; deviationPct: string; deviationCount: string | null; totalVotes: string }>`
				WITH party_majority AS (
					SELECT
						b.id as besluit_id,
						s.fractie_id,
						s.soort as majority_vote
					FROM stemmingen s
					JOIN besluiten b ON s.besluit_id = b.id
					JOIN zaken z ON b.zaak_id = z.id
					WHERE s.fractie_id IS NOT NULL
					  AND s.soort IN ('Voor', 'Tegen')
					  AND z.soort = 'Motie'
				),
				individual_votes AS (
					SELECT
						s.persoon_id,
						s.fractie_id,
						b.id as besluit_id,
						s.soort as vote
					FROM stemmingen s
					JOIN besluiten b ON s.besluit_id = b.id
					JOIN zaken z ON b.zaak_id = z.id
					WHERE s.persoon_id IS NOT NULL
					  AND s.fractie_id IS NOT NULL
					  AND s.soort IN ('Voor', 'Tegen')
					  AND z.soort = 'Motie'
				)
				SELECT
					iv.persoon_id AS "persoonId",
					iv.fractie_id AS "fractieId",
					COALESCE(p.roepnaam, p.voornamen) || ' ' || COALESCE(p.tussenvoegsel || ' ', '') || p.achternaam AS "persoonNaam",
					f.afkorting AS "fractieNaam",
					ROUND(AVG(CASE WHEN iv.vote != pm.majority_vote THEN 1.0 ELSE 0.0 END) * 100, 2) AS "deviationPct",
					SUM(CASE WHEN iv.vote != pm.majority_vote THEN 1 ELSE 0 END) AS "deviationCount",
					COUNT(*) AS "totalVotes"
				FROM individual_votes iv
				JOIN party_majority pm ON iv.besluit_id = pm.besluit_id AND iv.fractie_id = pm.fractie_id
				JOIN personen p ON iv.persoon_id = p.id
				JOIN fracties f ON iv.fractie_id = f.id
				GROUP BY iv.persoon_id, iv.fractie_id, p.roepnaam, p.voornamen, p.tussenvoegsel, p.achternaam, f.afkorting
				HAVING COUNT(*) >= 20
				ORDER BY "deviationPct" DESC
				LIMIT ${limit}
			`;

		return results
			.filter((r) => r.persoonId && r.fractieId && r.persoonNaam && r.fractieNaam)
			.map((r) => ({
				persoonId: r.persoonId!,
				fractieId: r.fractieId!,
				persoonNaam: r.persoonNaam!,
				fractieNaam: r.fractieNaam!,
				deviationPct: Number(r.deviationPct),
				deviationCount: Number(r.deviationCount),
				totalVotes: Number(r.totalVotes),
			}));
	}),

	getTopicTrends: os.analysis.getTopicTrends.handler(async ({ input }) => {
		const { dateFrom, dateTo } = input || {};

		const results = dateFrom && dateTo
			? await sql<{
					categoryId: string;
					categoryName: string;
					motionCount: string;
					acceptedCount: string;
					rejectedCount: string;
				}>`
				SELECT
					zc.category_id AS "categoryId",
					mc.name AS "categoryName",
					COUNT(DISTINCT z.id) AS "motionCount",
					COUNT(DISTINCT CASE WHEN b.besluit_tekst LIKE 'Aangenomen%' THEN z.id END) AS "acceptedCount",
					COUNT(DISTINCT CASE WHEN b.besluit_tekst LIKE 'Verworpen%' THEN z.id END) AS "rejectedCount"
				FROM zaak_categories zc
				JOIN zaken z ON zc.zaak_id = z.id
				JOIN motion_categories mc ON zc.category_id = mc.id
				LEFT JOIN besluiten b ON b.zaak_id = z.id
				WHERE z.soort = 'Motie'
				  AND z.gestart_op BETWEEN ${dateFrom} AND ${dateTo}
				GROUP BY zc.category_id, mc.name
				ORDER BY "motionCount" DESC
			`
			: await sql<{
					categoryId: string;
					categoryName: string;
					motionCount: string;
					acceptedCount: string;
					rejectedCount: string;
				}>`
				SELECT
					zc.category_id AS "categoryId",
					mc.name AS "categoryName",
					COUNT(DISTINCT z.id) AS "motionCount",
					COUNT(DISTINCT CASE WHEN b.besluit_tekst LIKE 'Aangenomen%' THEN z.id END) AS "acceptedCount",
					COUNT(DISTINCT CASE WHEN b.besluit_tekst LIKE 'Verworpen%' THEN z.id END) AS "rejectedCount"
				FROM zaak_categories zc
				JOIN zaken z ON zc.zaak_id = z.id
				JOIN motion_categories mc ON zc.category_id = mc.id
				LEFT JOIN besluiten b ON b.zaak_id = z.id
				WHERE z.soort = 'Motie'
				GROUP BY zc.category_id, mc.name
				ORDER BY "motionCount" DESC
			`;

		return results.map((r) => ({
			categoryId: r.categoryId,
			categoryName: r.categoryName,
			motionCount: Number(r.motionCount),
			acceptedCount: Number(r.acceptedCount),
			rejectedCount: Number(r.rejectedCount),
		}));
	}),

	getPartyTopicVoting: os.analysis.getPartyTopicVoting.handler(
		async ({ input }) => {
			const { dateFrom, dateTo, fractieId = "" } = input || {};

			const results = dateFrom && dateTo
				? await sql<{ fractieId: string | null; categoryId: string; fractieNaam: string | null; categoryName: string; votesFor: string | null; votesAgainst: string | null; totalVotes: string; forPct: string }>`
					SELECT
						mv.fractie_id AS "fractieId",
						zc.category_id AS "categoryId",
						f.afkorting AS "fractieNaam",
						mc.name AS "categoryName",
						SUM(CASE WHEN mv.vote_type = 'Voor' THEN 1 ELSE 0 END) AS "votesFor",
						SUM(CASE WHEN mv.vote_type = 'Tegen' THEN 1 ELSE 0 END) AS "votesAgainst",
						COUNT(*) AS "totalVotes",
						ROUND(AVG(CASE WHEN mv.vote_type = 'Voor' THEN 1.0 ELSE 0.0 END) * 100, 2) AS "forPct"
					FROM majority_party_votes mv
					JOIN zaak_categories zc ON mv.zaak_id = zc.zaak_id
					JOIN motion_categories mc ON zc.category_id = mc.id
					JOIN fracties f ON mv.fractie_id = f.id
					WHERE (${fractieId} = '' OR mv.fractie_id = ${fractieId})
					  AND mv.gestart_op BETWEEN ${dateFrom} AND ${dateTo}
					GROUP BY mv.fractie_id, zc.category_id, f.afkorting, mc.name
					HAVING COUNT(*) >= 5
					ORDER BY "totalVotes" DESC
				`
				: await sql<{ fractieId: string | null; categoryId: string; fractieNaam: string | null; categoryName: string; votesFor: string | null; votesAgainst: string | null; totalVotes: string; forPct: string }>`
					SELECT
						mv.fractie_id AS "fractieId",
						zc.category_id AS "categoryId",
						f.afkorting AS "fractieNaam",
						mc.name AS "categoryName",
						SUM(CASE WHEN mv.vote_type = 'Voor' THEN 1 ELSE 0 END) AS "votesFor",
						SUM(CASE WHEN mv.vote_type = 'Tegen' THEN 1 ELSE 0 END) AS "votesAgainst",
						COUNT(*) AS "totalVotes",
						ROUND(AVG(CASE WHEN mv.vote_type = 'Voor' THEN 1.0 ELSE 0.0 END) * 100, 2) AS "forPct"
					FROM majority_party_votes mv
					JOIN zaak_categories zc ON mv.zaak_id = zc.zaak_id
					JOIN motion_categories mc ON zc.category_id = mc.id
					JOIN fracties f ON mv.fractie_id = f.id
					WHERE (${fractieId} = '' OR mv.fractie_id = ${fractieId})
					GROUP BY mv.fractie_id, zc.category_id, f.afkorting, mc.name
					HAVING COUNT(*) >= 5
					ORDER BY "totalVotes" DESC
				`;

			return results
				.filter((r) => r.fractieId && r.fractieNaam)
				.map((r) => ({
					fractieId: r.fractieId!,
					categoryId: r.categoryId,
					fractieNaam: r.fractieNaam!,
					categoryName: r.categoryName,
					votesFor: Number(r.votesFor),
					votesAgainst: Number(r.votesAgainst),
					totalVotes: Number(r.totalVotes),
					forPct: Number(r.forPct),
				}));
		},
	),
};
