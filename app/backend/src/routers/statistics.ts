import { implement } from "@orpc/server";
import { Prisma } from "@prisma/client";
import {
	apiContract,
	type PartyCategoryLikeness,
	type PartyFocusCategory,
	type PartyLikeness,
} from "../contracts/index.js";
import { db } from "../lib/db.js";
import { mapPartyToContract } from "../utils/mappers.js";

const os = implement(apiContract);

export const statisticsRouter = {
	getPartyLikenessMatrix: os.statistics.getPartyLikenessMatrix.handler(
		async ({ input }) => {
			const { dateFrom, dateTo } = input || {};

			const partyFilter = Prisma.sql`
				p1.datum_inactief IS NULL
				AND p2.datum_inactief IS NULL
			`;

			const dateFilter =
				dateFrom && dateTo
					? Prisma.sql`AND pl.gestart_op BETWEEN ${dateFrom} AND ${dateTo}`
					: Prisma.empty;

			const results: PartyLikeness[] = await db.$queryRaw`
        SELECT
            p1.id AS "party1Id",
            p1.afkorting AS "party1Name",
            p2.id AS "party2Id",
            p2.afkorting AS "party2Name",
            COUNT(*) AS "commonMotions",
            SUM(CASE WHEN pl.same_vote THEN 1 ELSE 0 END) AS "sameVotes",
            (SUM(CASE WHEN pl.same_vote THEN 1 ELSE 0 END)::float / NULLIF(COUNT(*), 0)::float) * 100 AS "likenessPercentage"
        FROM party_likeness_per_motion pl
        JOIN fracties p1 ON pl.fractie1_id = p1.id
        JOIN fracties p2 ON pl.fractie2_id = p2.id
        WHERE ${partyFilter} ${dateFilter}
        GROUP BY p1.id, p1.afkorting, p2.id, p2.afkorting
        ORDER BY p1.afkorting, "likenessPercentage" DESC;
      `;

			return results.map((r) => ({
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

		const party = await db.party.findUnique({
			where: { id: partyId },
		});

		if (!party) {
			return null;
		}

		const dateFilter =
			dateFrom && dateTo
				? Prisma.sql`AND z.gestart_op BETWEEN ${dateFrom} AND ${dateTo}`
				: Prisma.empty;

		const results: PartyFocusCategory[] = await db.$queryRaw`
        SELECT
            mc.id AS "categoryId",
            mc.name AS "categoryName",
            mc.type AS "categoryType",
            COUNT(z.id) AS "motionCount"
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
        WHERE
            f.id = ${partyId}
            AND za.relatie = 'Indiener'
            AND z.soort = 'Motie'
            ${dateFilter}
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

			const dateFilter =
				dateFrom && dateTo
					? Prisma.sql`WHERE gestart_op BETWEEN ${dateFrom} AND ${dateTo}`
					: Prisma.empty;

			const results: PartyCategoryLikeness[] = await db.$queryRaw`
      WITH MotionMajorityVotes AS (
        SELECT * FROM majority_party_votes ${dateFilter}
      ),
      MotionCategoryVotes AS (
          SELECT
              mmv.zaak_id,
              mmv.fractie_id,
              mmv.vote_type,
              zc.category_id
          FROM MotionMajorityVotes mmv
          JOIN zaak_categories zc ON mmv.zaak_id = zc.zaak_id
      )
      SELECT
          mc.id AS "categoryId",
          mc.name AS "categoryName",
          p2.id AS "party2Id",
          p2.afkorting AS "party2Name",
          (SUM(CASE WHEN mcv1.vote_type = mcv2.vote_type THEN 1 ELSE 0 END)::float / NULLIF(COUNT(mcv2.zaak_id), 0)::float) * 100 AS "likenessPercentage"
      FROM
          (SELECT * FROM MotionCategoryVotes WHERE fractie_id = ${partyId}) mcv1
      JOIN
          MotionCategoryVotes mcv2 ON mcv1.zaak_id = mcv2.zaak_id AND mcv1.category_id = mcv2.category_id
      JOIN
          fracties p2 ON mcv2.fractie_id = p2.id
      JOIN
          motion_categories mc ON mcv1.category_id = mc.id
      WHERE
          mcv1.fractie_id != mcv2.fractie_id
          AND p2.datum_inactief IS NULL
          AND p2.naam_nl NOT IN ('Groep Van Haga', 'Fractie Den Haan', 'Lid Omtzigt', 'Lid Gündoğan')
      GROUP BY
          mc.id, mc.name, p2.id, p2.afkorting
      ORDER BY
          mc.name, p2.afkorting;
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
