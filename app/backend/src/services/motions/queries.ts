import type {
	Motion,
	MotionCategory,
	Party,
	Vote,
} from "../../contracts/index.js";
import { sql, sqlOne, sqlOneOrNull } from "../db/sql-tag.js";

export async function getForCompass(
	count: number,
	excludeIds: string[] | undefined,
	categoryIds: string[] | undefined,
	after: Date | undefined,
) {
	return sql<Motion[]>`
        SELECT
            id,
            onderwerp as title,
            citeertitel as "shortTitle",
            nummer as "motionNumber",
            datum as date,
            status,
            soort as category,
            bullet_points as "bulletPoints",
            document_url as "documentURL",
            did,
            gestart_op as "createdAt",
            gewijzigd_op as "updatedAt"
        FROM "zaken"
        WHERE "soort" = 'Motie'
        AND "bullet_points" IS NOT NULL AND jsonb_array_length("bullet_points") > 0
        AND EXISTS (SELECT 1 FROM jsonb_array_elements_text("bullet_points") AS elem WHERE elem ILIKE 'verzoekt%')
        AND EXISTS (SELECT 1 FROM besluiten b JOIN stemmingen s ON b.id = s.besluit_id WHERE b.zaak_id = zaken.id AND s.fractie_id IS NOT NULL)
        AND (${excludeIds}::text[] IS NULL OR id NOT IN (SELECT unnest(${excludeIds}::text[])))
        AND (${after}::timestamp IS NULL OR "gestart_op" >= ${after})
        AND (${categoryIds}::text[] IS NULL OR EXISTS (
            SELECT 1 FROM "zaak_categories"
            WHERE "zaak_id" = "zaken".id
            AND "category_id" IN (SELECT unnest(${categoryIds}::text[]))
        ))
        ORDER BY RANDOM()
        LIMIT ${count}
    `;
}

export async function getAllMotions(
	limit: number,
	offset: number,
	category?: string,
	status?: string,
) {
	return sql<({ total: string } & Motion)[]>`
        WITH subset AS (
            SELECT
                id,
                onderwerp as title,
                citeertitel as "shortTitle",
                nummer as "motionNumber",
                datum as date,
                status,
                soort as category,
                bullet_points as "bulletPoints",
                document_url as "documentURL",
                did,
                gestart_op as "createdAt",
                gewijzigd_op as "updatedAt"
            FROM "zaken"
            WHERE "soort" = 'Motie'
            AND (${category}::text IS NULL OR "soort" = ${category})
            AND (${status}::text IS NULL OR status = ${status})
        )
        SELECT *, (SELECT count(*) FROM subset) as total
        FROM subset
        ORDER BY date DESC
        LIMIT ${limit}
        OFFSET ${offset}
    `;
}

export async function getMotionById(id: string) {
	return sqlOneOrNull<Motion>`
        SELECT
            id,
            onderwerp as title,
            citeertitel as "shortTitle",
            nummer as "motionNumber",
            datum as date,
            status,
            soort as category,
            bullet_points as "bulletPoints",
            document_url as "documentURL",
            did,
            gestart_op as "createdAt",
            gewijzigd_op as "updatedAt"
        FROM "zaken"
        WHERE id = ${id}
    `;
}

export async function getMotionsByIds(motionIds: string[]) {
	return sql<Motion[]>`
        SELECT
            id,
            onderwerp as title,
            citeertitel as "shortTitle",
            nummer as "motionNumber",
            datum as date,
            status,
            soort as category,
            bullet_points as "bulletPoints",
            document_url as "documentURL",
            did,
            gestart_op as "createdAt",
            gewijzigd_op as "updatedAt"
        FROM "zaken"
        WHERE id IN (${motionIds})
    `;
}

export async function getMotionCategories(
	type: "all" | "generic" | "hot_topic",
) {
	return sql<MotionCategory[]>`
        SELECT
            id,
            name,
            type,
            description,
            keywords,
            created_at as "createdAt",
            updated_at as "updatedAt"
        FROM "motion_categories"
        WHERE ${type} = 'all' OR type = ${type}
        ORDER BY type ASC, name ASC
    `;
}

export async function getVotesByDecisionId(decisionId: string) {
	return sql<Vote[]>`
        SELECT
            id,
            besluit_id as "motionId",
            fractie_id as "partyId",
            persoon_id as "politicianId",
            soort as "voteType",
            gewijzigd_op as "createdAt",
            api_gewijzigd_op as "updatedAt"
        FROM "stemmingen"
        WHERE "besluit_id" = ${decisionId}
        ORDER BY "fractie_id" ASC, "persoon_id" ASC
    `;
}

export async function getPartiesByIds(partyIds: string[]) {
	if (partyIds.length === 0) {
		return [];
	}
	return sql<Party[]>`
        SELECT
            id,
            naam_nl as name,
            afkorting as "shortName",
            aantal_zetels as seats,
            datum_actief as "activeFrom",
            datum_inactief as "activeTo",
            content_type as "contentType",
            encode(logo_data, 'base64') as "logoData",
            gewijzigd_op as "createdAt",
            api_gewijzigd_op as "updatedAt"
        FROM "fracties"
        WHERE id IN (${partyIds})
    `;
}

export async function getMotionStatistics() {
	return sqlOne<{
		count: number;
		firstMotionDate: Date | null;
		lastMotionDate: Date | null;
	}>`
        SELECT
            COUNT(id)::int as count,
            MIN("gestart_op") as "firstMotionDate",
            MAX("gestart_op") as "lastMotionDate"
        FROM "zaken"
        WHERE "soort" = 'Motie'
    `;
}
