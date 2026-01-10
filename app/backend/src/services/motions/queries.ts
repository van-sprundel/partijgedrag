import type { Motion, Party } from "../../contracts/index.js";
import { sql, sqlOne, sqlOneOrNull } from "../db/sql-tag.js";

export async function getForCompass(
	count: number,
	excludeIds: string[] | undefined,
	categoryIds: string[] | undefined,
	after: Date | undefined,
	search: string | undefined,
	partyIds: string[] | undefined,
) {
	return sql<Motion>`
        WITH VoteCounts AS (
            SELECT
                b.zaak_id,
                SUM(CASE WHEN s.soort = 'Voor' THEN 1 ELSE 0 END) as voor_votes,
                SUM(CASE WHEN s.soort = 'Tegen' THEN 1 ELSE 0 END) as tegen_votes
            FROM stemmingen s
            JOIN besluiten b ON s.besluit_id = b.id
            WHERE s.soort IN ('Voor', 'Tegen') AND s.vergissing IS NOT TRUE
            GROUP BY b.zaak_id
        )
        SELECT
            z.id,
            z.onderwerp as title,
            z.citeertitel as "shortTitle",
            z.nummer as "motionNumber",
            z.datum as date,
            z.status,
            z.soort as category,
            z.bullet_points as "bulletPoints",
            z.document_url as "documentURL",
            z.did,
            z.gestart_op as "createdAt",
            z.gewijzigd_op as "updatedAt"
        FROM "zaken" z
        JOIN VoteCounts vc ON z.id = vc.zaak_id
        WHERE "soort" = 'Motie'
        AND "bullet_points" IS NOT NULL AND jsonb_array_length("bullet_points") > 0
        AND EXISTS (SELECT 1 FROM jsonb_array_elements_text("bullet_points") AS elem WHERE elem ILIKE 'verzoekt%')
        AND EXISTS (SELECT 1 FROM besluiten b JOIN stemmingen s ON b.id = s.besluit_id WHERE b.zaak_id = z.id AND s.fractie_id IS NOT NULL)
        AND (${excludeIds}::text[] IS NULL OR z.id NOT IN (SELECT unnest(${excludeIds}::text[])))
        AND (${after}::timestamp IS NULL OR z."gestart_op" >= ${after})
        AND (${categoryIds}::text[] IS NULL OR EXISTS (
            SELECT 1 FROM "zaak_categories"
            WHERE "zaak_id" = z.id
            AND "category_id" IN (SELECT unnest(${categoryIds}::text[]))
        ))
        AND (${search}::text IS NULL OR (
            z.onderwerp ILIKE '%' || ${search} || '%'
            OR z.citeertitel ILIKE '%' || ${search} || '%'
            OR z.nummer ILIKE '%' || ${search} || '%'
            OR EXISTS (
                SELECT 1 FROM jsonb_array_elements_text(z."bullet_points") AS elem
                WHERE elem ILIKE '%' || ${search} || '%'
            )
        ))
        AND (${partyIds}::text[] IS NULL OR array_length(${partyIds}::text[], 1) < 2 OR (
            SELECT COUNT(DISTINCT s.soort)
            FROM stemmingen s
            JOIN besluiten b ON s.besluit_id = b.id
            WHERE b.zaak_id = z.id
            AND s.fractie_id = ANY(${partyIds}::text[])
            AND s.soort IN ('Voor', 'Tegen')
            AND s.vergissing IS NOT TRUE
        ) > 1)
        ORDER BY ABS(vc.voor_votes - vc.tegen_votes) ASC, RANDOM()
        LIMIT ${count}
    `;
}

export async function getAllMotions(
	limit: number,
	offset: number,
	category?: string,
	status?: string,
	withVotes?: boolean,
	search?: string,
) {
	return sql<{ total: string } & Motion>`
        WITH subset AS (
            SELECT
                z.id,
                z.onderwerp as title,
                z.citeertitel as "shortTitle",
                z.nummer as "motionNumber",
                z.datum as date,
                z.status,
                z.soort as category,
                z.bullet_points as "bulletPoints",
                z.document_url as "documentURL",
                z.did,
                z.gestart_op as "createdAt",
                z.gewijzigd_op as "updatedAt"
            FROM "zaken" z
            WHERE z."soort" = 'Motie'
            AND z."bullet_points" IS NOT NULL AND jsonb_array_length(z."bullet_points") > 0
            AND (${category}::text IS NULL OR z."soort" = ${category})
            AND (${status}::text IS NULL OR z.status = ${status})
            AND (${withVotes}::boolean IS NULL OR ${withVotes} = false OR (
                ${withVotes} = true AND EXISTS (
                    SELECT 1 FROM "besluiten" b
                    JOIN "stemmingen" s ON b.id = s.besluit_id
                    WHERE b.zaak_id = z.id AND s.vergissing IS NOT TRUE
                )
            ))
            AND (${search}::text IS NULL OR (
                z.onderwerp ILIKE '%' || ${search} || '%'
                OR z.citeertitel ILIKE '%' || ${search} || '%'
                OR z.nummer ILIKE '%' || ${search} || '%'
                OR EXISTS (
                    SELECT 1 FROM jsonb_array_elements_text(z."bullet_points") AS elem
                    WHERE elem ILIKE '%' || ${search} || '%'
                )
            ))
        )
        SELECT *, (SELECT count(*) FROM subset) as total
        FROM subset
        ORDER BY "createdAt" DESC
        LIMIT ${limit}
        OFFSET ${offset}
    `;
}

export async function getMotionById(id: string) {
	return sqlOneOrNull<{
		id: string;
		title: string | null;
		shortTitle: string | null;
		motionNumber: string | null;
		date: Date | null;
		status: string | null;
		category: string | null;
		bulletPoints: unknown | null;
		documentURL: string | null;
		did: string | null;
		createdAt: Date | null;
		updatedAt: Date | null;
	}>`
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
	return sql<Motion>`
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

export async function getRecentMotions(limit: number) {
	return sql<{
		id: string;
		title: string | null;
		shortTitle: string | null;
		motionNumber: string | null;
		date: Date | null;
		status: string | null;
		category: string | null;
		bulletPoints: unknown | null;
		documentURL: string | null;
		did: string | null;
		createdAt: Date | null;
		updatedAt: Date | null;
	}>`
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
        ORDER BY "gestart_op" DESC
        LIMIT ${limit}
    `;
}

export async function getMotionCategories(
	type: "all" | "generic" | "hot_topic",
) {
	return sql<{
		id: string;
		name: string;
		type: string | null;
		description: string | null;
		keywords: string[] | null;
		createdAt: Date | null;
		updatedAt: Date | null;
	}>`
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
	return sql<{
		id: string;
		motionId: string | null;
		partyId: string | null;
		politicianId: string | null;
		voteType: string | null;
		partySize: string | null;
		createdAt: Date | null;
		updatedAt: Date | null;
	}>`
        SELECT
            id,
            besluit_id as "motionId",
            fractie_id as "partyId",
            persoon_id as "politicianId",
            soort as "voteType",
            fractie_grootte as "partySize",
            gewijzigd_op as "createdAt",
            api_gewijzigd_op as "updatedAt"
        FROM "stemmingen"
        WHERE "besluit_id" = ${decisionId} AND "vergissing" IS NOT TRUE
        ORDER BY "fractie_id" ASC, "persoon_id" ASC
    `;
}

export async function getVotesByMotionId(motionId: string) {
	return sql<{
		id: string;
		motionId: string;
		partyId: string;
		politicianId: string;
		voteType: string | null;
		partySize: string | null;
		createdAt: Date;
		updatedAt: Date;
	}>`
        WITH decision_ids AS (
            SELECT id FROM "besluiten"
            WHERE "zaak_id" = ${motionId}
        )
        SELECT
            s.id,
            COALESCE(s.besluit_id, '') as "motionId",
            COALESCE(s.fractie_id, '') as "partyId",
            COALESCE(s.persoon_id, '') as "politicianId",
            s.soort as "voteType",
            s.fractie_grootte as "partySize",
            COALESCE(s.gewijzigd_op, NOW()) as "createdAt",
            COALESCE(s.api_gewijzigd_op, NOW()) as "updatedAt"
        FROM "stemmingen" s
        WHERE s.besluit_id IN (SELECT id FROM decision_ids)
        AND s.vergissing IS NOT TRUE
        ORDER BY s.fractie_id ASC, s.persoon_id ASC
    `;
}

export async function getPartiesByIds(partyIds: string[]) {
	if (partyIds.length === 0) {
		return [];
	}
	return sql<Party>`
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
		firstMotionDate: string | null;
		lastMotionDate: string | null;
	}>`
        SELECT
            COUNT(id)::int as count,
            MIN("gestart_op") as "firstMotionDate",
            MAX("gestart_op") as "lastMotionDate"
        FROM "zaken"
        WHERE "soort" = 'Motie'
    `;
}

export async function getSubmitterByMotionId(motionId: string) {
	return sqlOneOrNull<{
		id: string;
		firstName: string | null;
		lastName: string | null;
		fullName: unknown;
		updatedAt: Date | null;
	}>`
        SELECT p.id, p.voornamen as "firstName", p.achternaam as "lastName", CONCAT(p.voornamen, ' ', p.achternaam) as "fullName", p.bijgewerkt as "updatedAt"
        FROM personen p
        JOIN zaak_actors za ON p.id = za.persoon_id
        WHERE za.zaak_id = ${motionId}
    `;
}

export async function getForCompassCount(
	categoryIds: string[] | undefined,
	after: Date | undefined,
	search: string | undefined,
	partyIds: string[] | undefined,
) {
	return sqlOne<{ count: number }>`
        SELECT
            COUNT(z.id)::int as count
        FROM "zaken" z
        WHERE "soort" = 'Motie'
        AND "bullet_points" IS NOT NULL AND jsonb_array_length("bullet_points") > 0
        AND EXISTS (SELECT 1 FROM jsonb_array_elements_text("bullet_points") AS elem WHERE elem ILIKE 'verzoekt%')
        AND EXISTS (SELECT 1 FROM besluiten b JOIN stemmingen s ON b.id = s.besluit_id WHERE b.zaak_id = z.id AND s.fractie_id IS NOT NULL)
        AND (${after}::timestamp IS NULL OR z."gestart_op" >= ${after})
        AND (${categoryIds}::text[] IS NULL OR EXISTS (
            SELECT 1 FROM "zaak_categories"
            WHERE "zaak_id" = z.id
            AND "category_id" IN (SELECT unnest(${categoryIds}::text[]))
        ))
        AND (${search}::text IS NULL OR (
            z.onderwerp ILIKE '%' || ${search} || '%'
            OR z.citeertitel ILIKE '%' || ${search} || '%'
            OR z.nummer ILIKE '%' || ${search} || '%'
            OR EXISTS (
                SELECT 1 FROM jsonb_array_elements_text(z."bullet_points") AS elem
                WHERE elem ILIKE '%' || ${search} || '%'
            )
        ))
        AND (${partyIds}::text[] IS NULL OR array_length(${partyIds}::text[], 1) < 2 OR (
            SELECT COUNT(DISTINCT s.soort)
            FROM stemmingen s
            JOIN besluiten b ON s.besluit_id = b.id
            WHERE b.zaak_id = z.id
            AND s.fractie_id = ANY(${partyIds}::text[])
            AND s.soort IN ('Voor', 'Tegen')
            AND s.vergissing IS NOT TRUE
        ) > 1)
    `;
}
