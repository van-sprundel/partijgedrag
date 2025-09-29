import type {
	PartyCategoryLikeness,
	PartyFocusCategory,
	PartyLikeness,
} from "../../contracts/index.js";
import { sql } from "../db/sql-tag.js";

export async function getPartyLikenessMatrix(dateFrom?: Date, dateTo?: Date) {
	return sql<PartyLikeness>`
        WITH PartyVotes AS (
            SELECT DISTINCT
                b.zaak_id,
                f.id as fractie_id,
                f.afkorting as fractie_name,
                s.soort as vote_type
            FROM stemmingen s
            JOIN besluiten b ON s.besluit_id = b.id
            JOIN zaken z ON b.zaak_id = z.id
            JOIN fracties f ON (s.actor_fractie = f.naam_nl OR s.actor_fractie = f.afkorting)
            WHERE s.actor_fractie IS NOT NULL
                AND s.soort IN ('Voor', 'Tegen')
                AND z.soort = 'Motie'
                AND f.datum_inactief IS NULL
                AND (${dateFrom}::timestamp IS NULL OR z.gestart_op >= ${dateFrom})
                AND (${dateTo}::timestamp IS NULL OR z.gestart_op <= ${dateTo})
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
}

export async function getPartyFocus(
	partyId: string,
	dateFrom?: Date,
	dateTo?: Date,
) {
	return sql<PartyFocusCategory>`
        SELECT
            mc.id AS "categoryId",
            mc.name AS "categoryName",
            mc.type AS "categoryType",
            COUNT(z.id)::int AS "motionCount"
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
            AND (${dateFrom}::timestamp IS NULL OR z.gestart_op >= ${dateFrom})
            AND (${dateTo}::timestamp IS NULL OR z.gestart_op <= ${dateTo})
        GROUP BY
            mc.id, mc.name, mc.type
        ORDER BY
            "motionCount" DESC;
    `;
}

export async function getPartyCategoryLikeness(
	partyId: string,
	dateFrom?: Date,
	dateTo?: Date,
) {
	return sql<PartyCategoryLikeness>`
        WITH PartyVotesByCategory AS (
            SELECT DISTINCT
                b.zaak_id,
                f.id as fractie_id,
                s.soort as vote_type,
                zc.category_id
            FROM stemmingen s
            JOIN besluiten b ON s.besluit_id = b.id
            JOIN zaken z ON b.zaak_id = z.id
            JOIN fracties f ON (s.actor_fractie = f.naam_nl OR s.actor_fractie = f.afkorting)
            JOIN zaak_categories zc ON z.id = zc.zaak_id
            WHERE s.actor_fractie IS NOT NULL
                AND s.soort IN ('Voor', 'Tegen')
                AND z.soort = 'Motie'
                AND f.datum_inactief IS NULL
                AND (${dateFrom}::timestamp IS NULL OR z.gestart_op >= ${dateFrom})
                AND (${dateTo}::timestamp IS NULL OR z.gestart_op <= ${dateTo})
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
        WHERE f.datum_inactief IS NULL
        ORDER BY mc.name, f.afkorting;
    `;
}
