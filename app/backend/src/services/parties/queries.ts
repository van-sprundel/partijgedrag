import type { Party, Vote } from "../../contracts/index.js";
import { sql, sqlOneOrNull } from "../db/sql-tag.js";

export async function getActiveParties() {
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
        WHERE "datum_inactief" IS NULL OR "datum_inactief" >= NOW()
    `;
}

export async function getPartiesByIdsOrNames(
	partyIds: string[],
	partyNames: string[],
) {
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
        WHERE (${partyIds}::text[] IS NULL OR id IN (SELECT unnest(${partyIds}::text[])))
        OR (${partyNames}::text[] IS NULL OR "naam_nl" IN (SELECT unnest(${partyNames}::text[])))
    `;
}

export async function getAllParties(activeOnly: boolean) {
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
        WHERE "verwijderd" IS NOT TRUE
        AND (${activeOnly} = false OR ("datum_inactief" IS NULL OR "datum_inactief" >= NOW()))
        ORDER BY "naam_nl" ASC
    `;
}

export async function getPartyById(id: string) {
	return sqlOneOrNull<Party>`
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
            besluit_id as "motionId",
            fractie_id as "partyId",
            persoon_id as "politicianId",
            soort as "voteType",
            gewijzigd_op as "createdAt",
            api_gewijzigd_op as "updatedAt"
        FROM "stemmingen"
        WHERE "fractie_id" = ${partyId} AND "vergissing" IS NOT TRUE
        AND (${motionIds}::text[] IS NULL OR "besluit_id" IN (SELECT unnest(${motionIds}::text[])))
        ORDER BY "gewijzigd_op" DESC
    `;
}
