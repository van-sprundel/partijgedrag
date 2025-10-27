import type {
	Motion,
	MotionCategory,
	Party,
	Politician,
	Vote,
} from "../../contracts/index.js";
import { sql, sqlOne, sqlOneOrNull } from "../db/sql-tag.js";

export async function getForCompass(
	count: number,
	excludeIds: string[] | undefined,
	categoryIds: string[] | undefined,
	after: Date | undefined,
) {
	return sql<Motion>`
        WITH VoteCounts AS (
            SELECT
                d.case_id,
                SUM(CASE WHEN v.type = 'Voor' THEN 1 ELSE 0 END) as voor_votes,
                SUM(CASE WHEN v.type = 'Tegen' THEN 1 ELSE 0 END) as tegen_votes
            FROM votes v
            JOIN decisions d ON v.decision_id = d.id
            WHERE v.type IN ('Voor', 'Tegen') AND v.mistake IS NOT TRUE
            GROUP BY d.case_id
        )
        SELECT
            c.id,
            c.subject as title,
            c.citation_title as "shortTitle",
            c.number as "motionNumber",
            c.date as date,
            c.status,
            c.type as category,
            c.bullet_points as "bulletPoints",
            c.document_url as "documentURL",
            c.did,
            c.started_at as "createdAt",
            c.updated_at as "updatedAt"
        FROM cases c
        JOIN VoteCounts vc ON c.id = vc.case_id
        WHERE c.type = 'Motie'
        AND c.bullet_points IS NOT NULL AND jsonb_array_length(c.bullet_points) > 0
        AND EXISTS (SELECT 1 FROM jsonb_array_elements_text(c.bullet_points) AS elem WHERE elem ILIKE 'verzoekt%')
        AND EXISTS (SELECT 1 FROM decisions d JOIN votes v ON d.id = v.decision_id WHERE d.case_id = c.id AND v.party_id IS NOT NULL)
        AND (${excludeIds}::text[] IS NULL OR c.id NOT IN (SELECT unnest(${excludeIds}::text[])))
        AND (${after}::timestamp IS NULL OR c.started_at >= ${after})
        AND (${categoryIds}::text[] IS NULL OR EXISTS (
            SELECT 1 FROM case_categories
            WHERE case_id = c.id
            AND category_id IN (SELECT unnest(${categoryIds}::text[]))
        ))
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
) {
	return sql<{ total: string } & Motion>`
        WITH subset AS (
            SELECT
                c.id,
                c.subject as title,
                c.citation_title as "shortTitle",
                c.number as "motionNumber",
                c.date as date,
                c.status,
                c.type as category,
                c.bullet_points as "bulletPoints",
                c.document_url as "documentURL",
                c.did,
                c.started_at as "createdAt",
                c.updated_at as "updatedAt"
            FROM cases c
            WHERE c.type = 'Motie'
            AND c.bullet_points IS NOT NULL AND jsonb_array_length(c.bullet_points) > 0
            AND (${category}::text IS NULL OR c.type = ${category})
            AND (${status}::text IS NULL OR c.status = ${status})
            AND (${withVotes}::boolean IS NULL OR ${withVotes} = false OR (
                ${withVotes} = true AND EXISTS (
                    SELECT 1 FROM decisions d
                    JOIN votes v ON d.id = v.decision_id
                    WHERE d.case_id = c.id AND v.mistake IS NOT TRUE
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
	return sqlOneOrNull<Motion>`
        SELECT
            id,
            subject as title,
            citation_title as "shortTitle",
            number as "motionNumber",
            date as date,
            status,
            type as category,
            bullet_points as "bulletPoints",
            document_url as "documentURL",
            did,
            started_at as "createdAt",
            updated_at as "updatedAt"
        FROM cases
        WHERE id = ${id}
    `;
}

export async function getMotionsByIds(motionIds: string[]) {
	return sql<Motion>`
        SELECT
            id,
            subject as title,
            citation_title as "shortTitle",
            number as "motionNumber",
            date as date,
            status,
            type as category,
            bullet_points as "bulletPoints",
            document_url as "documentURL",
            did,
            started_at as "createdAt",
            updated_at as "updatedAt"
        FROM cases
        WHERE id IN (${motionIds})
    `;
}

export async function getRecentMotions(limit: number) {
	return sql<Motion>`
        SELECT
            id,
            subject as title,
            citation_title as "shortTitle",
            number as "motionNumber",
            date as date,
            status,
            type as category,
            bullet_points as "bulletPoints",
            document_url as "documentURL",
            did,
            started_at as "createdAt",
            updated_at as "updatedAt"
        FROM cases
        WHERE type = 'Motie'
        ORDER BY started_at DESC
        LIMIT ${limit}
    `;
}

export async function getMotionCategories(
	type: "all" | "generic" | "hot_topic",
) {
	return sql<MotionCategory>`
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
	return sql<Vote>`
        SELECT
            id,
            decision_id as "motionId",
            party_id as "partyId",
            politician_id as "politicianId",
            type as "voteType",
            party_size as "partySize",
            updated_at as "createdAt",
            api_updated_at as "updatedAt"
        FROM votes
        WHERE decision_id = ${decisionId} AND mistake IS NOT TRUE
        ORDER BY party_id ASC, politician_id ASC
    `;
}

export async function getVotesByMotionId(motionId: string) {
	return sql<Vote>`
        WITH decision_ids AS (
            SELECT id FROM decisions
            WHERE case_id = ${motionId}
        )
        SELECT
            v.id,
            COALESCE(v.decision_id, '') as "motionId",
            COALESCE(v.party_id, '') as "partyId",
            COALESCE(v.politician_id, '') as "politicianId",
            v.type as "voteType",
            v.party_size as "partySize",
            COALESCE(v.updated_at, NOW()) as "createdAt",
            COALESCE(v.api_updated_at, NOW()) as "updatedAt"
        FROM votes v
        WHERE v.decision_id IN (SELECT id FROM decision_ids)
        AND v.mistake IS NOT TRUE
        ORDER BY v.party_id ASC, v.politician_id ASC
    `;
}

export async function getPartiesByIds(partyIds: string[]) {
	if (partyIds.length === 0) {
		return [];
	}
	return sql<Party>`
        SELECT
            id,
            name_nl as name,
            short_name as "shortName",
            seats as seats,
            active_from as "activeFrom",
            active_to as "activeTo",
            content_type as "contentType",
            encode(logo_data, 'base64') as "logoData",
            updated_at as "createdAt",
            api_updated_at as "updatedAt"
        FROM parties
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
            MIN(started_at) as "firstMotionDate",
            MAX(started_at) as "lastMotionDate"
        FROM cases
        WHERE type = 'Motie'
    `;
}

export async function getSubmitterByMotionId(motionId: string) {
	return sqlOneOrNull<Politician>`
        SELECT p.id, p.first_names as "firstName", p.last_name as "lastName", CONCAT(p.first_names, ' ', p.last_name) as "fullName", p.updated_at as "updatedAt"
        FROM politicians p
        JOIN case_actors ca ON p.id = ca.politician_id
        WHERE ca.case_id = ${motionId}
    `;
}

export async function getForCompassCount(
	categoryIds: string[] | undefined,
	after: Date | undefined,
) {
	return sqlOne<{ count: number }>`
        SELECT
            COUNT(c.id)::int as count
        FROM cases c
        WHERE c.type = 'Motie'
        AND c.bullet_points IS NOT NULL AND jsonb_array_length(c.bullet_points) > 0
        AND EXISTS (SELECT 1 FROM jsonb_array_elements_text(c.bullet_points) AS elem WHERE elem ILIKE 'verzoekt%')
        AND EXISTS (SELECT 1 FROM decisions d JOIN votes v ON d.id = v.decision_id WHERE d.case_id = c.id AND v.party_id IS NOT NULL)
        AND (${after}::timestamp IS NULL OR c.started_at >= ${after})
        AND (${categoryIds}::text[] IS NULL OR EXISTS (
            SELECT 1 FROM case_categories
            WHERE case_id = c.id
            AND category_id IN (SELECT unnest(${categoryIds}::text[]))
        ))
    `;
}
