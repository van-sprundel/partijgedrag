import type { Party, Vote } from "../../contracts/index.js";
import { sql, sqlOneOrNull } from "../db/sql-tag.js";

export async function getActiveParties() {
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
        WHERE active_to IS NULL OR active_to >= NOW()
    `;
}

export async function getPartiesByIdsOrNames(
	partyIds: string[],
	partyNames: string[],
) {
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
        WHERE (${partyIds}::text[] IS NULL OR id IN (SELECT unnest(${partyIds}::text[])))
        OR (${partyNames}::text[] IS NULL OR name_nl IN (SELECT unnest(${partyNames}::text[])))
    `;
}

export async function getAllParties(activeOnly: boolean) {
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
        WHERE removed IS NOT TRUE
        AND (${activeOnly} = false OR (active_to IS NULL OR active_to >= NOW()))
        ORDER BY name_nl ASC
    `;
}

export async function getPartyById(id: string) {
	return sqlOneOrNull<Party>`
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
        WHERE id = ${id}
    `;
}

export async function getVotesByPartyAndMotionIds(
	partyId: string,
	motionIds?: string[],
) {
	return sql<Vote>`
        SELECT
            id,
            decision_id as "motionId",
            party_id as "partyId",
            politician_id as "politicianId",
            type as "voteType",
            updated_at as "createdAt",
            api_updated_at as "updatedAt"
        FROM votes
        WHERE party_id = ${partyId} AND mistake IS NOT TRUE
        AND (${motionIds}::text[] IS NULL OR decision_id IN (SELECT unnest(${motionIds}::text[])))
        ORDER BY updated_at DESC
    `;
}
