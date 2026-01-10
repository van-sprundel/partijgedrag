import { implement, ORPCError } from "@orpc/server";
import { apiContract } from "../contracts/index.js";
import { sql, sqlOneOrNull } from "../services/db/sql-tag.js";
import { mapPartyToContract, mapVoteToContract } from "../utils/mappers.js";

const os = implement(apiContract);

export const partyRouter = {
  getAll: os.parties.getAll.handler(async ({ input }) => {
    const { activeOnly } = input;

    const parties = await sql<{
      id: string;
      number: string | null;
      shortName: string | null;
      nameNl: string | null;
      nameEn: string | null;
      seats: string | null;
      votesCount: string | null;
      activeFrom: Date | null;
      activeTo: Date | null;
      contentType: string | null;
      contentLength: string | null;
      updatedAt: Date | null;
      apiUpdatedAt: Date | null;
      logoData: unknown | null;
      removed: boolean | null;
    }>`
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
			WHERE verwijderd IS DISTINCT FROM true
			  AND (
				${!activeOnly} OR
				datum_inactief IS NULL OR
				datum_inactief >= NOW()
			  )
			ORDER BY naam_nl ASC
		`;

    return parties.map((p) => mapPartyToContract(p));
  }),

  getById: os.parties.getById.handler(async ({ input }) => {
    const party = await sqlOneOrNull<{
      id: string;
      number: string | null;
      shortName: string | null;
      nameNl: string | null;
      nameEn: string | null;
      seats: string | null;
      votesCount: string | null;
      activeFrom: Date | null;
      activeTo: Date | null;
      contentType: string | null;
      contentLength: string | null;
      updatedAt: Date | null;
      apiUpdatedAt: Date | null;
      logoData: unknown | null;
      removed: boolean | null;
    }>`
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
			WHERE id = ${input.id}
		`;

    if (!party) {
      return null;
    }

    return mapPartyToContract(party);
  }),

  getInRange: os.parties.getInRange.handler(async ({ input }) => {
    const { dateFrom, dateTo } = input;

    if (dateFrom > dateTo) {
      throw new ORPCError("BAD_REQUEST", {
        message: "`dateFrom` must be earlier than or the same as `dateTo`",
      });
    }

    const parties = await sql<{
      id: string;
      number: string | null;
      shortName: string | null;
      nameNl: string | null;
      nameEn: string | null;
      seats: string | null;
      votesCount: string | null;
      activeFrom: Date | null;
      activeTo: Date | null;
      contentType: string | null;
      contentLength: string | null;
      updatedAt: Date | null;
      apiUpdatedAt: Date | null;
      logoData: unknown | null;
      removed: boolean | null;
    }>`
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
			WHERE verwijderd IS DISTINCT FROM true
			  AND (datum_actief IS NULL OR datum_actief <= ${dateTo})
			  AND (datum_inactief IS NULL OR datum_inactief >= ${dateFrom})
			ORDER BY naam_nl ASC
		`;

    return parties.map((p) => mapPartyToContract(p));
  }),

  getWithVotes: os.parties.getWithVotes.handler(async ({ input }) => {
    const { partyId, motionIds } = input;

    const party = await sqlOneOrNull<{
      id: string;
      number: string | null;
      shortName: string | null;
      nameNl: string | null;
      nameEn: string | null;
      seats: string | null;
      votesCount: string | null;
      activeFrom: Date | null;
      activeTo: Date | null;
      contentType: string | null;
      contentLength: string | null;
      updatedAt: Date | null;
      apiUpdatedAt: Date | null;
      logoData: unknown | null;
      removed: boolean | null;
    }>`
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
      throw new ORPCError("NOT_FOUND", { message: "Party not found" });
    }

    const votesWithRelations = await sql<{
      id: string;
      decisionIdRaw: string | null;
      decisionId: string | null;
      type: string | null;
      partySize: string | null;
      actorName: string | null;
      actorParty: string | null;
      mistake: boolean | null;
      sidActorMember: string | null;
      sidActorParty: string | null;
      politicianIdRaw: string | null;
      politicianId: string | null;
      partyIdRaw: string | null;
      partyId: string | null;
      updatedAt: Date | null;
      apiUpdatedAt: Date | null;
    }>`
			SELECT
				id,
				besluit_id_raw as "decisionIdRaw",
				besluit_id as "decisionId",
				soort as type,
				fractie_grootte as "partySize",
				actor_naam as "actorName",
				actor_fractie as "actorParty",
				vergissing as mistake,
				sid_actor_lid as "sidActorMember",
				sid_actor_fractie as "sidActorParty",
				persoon_id_raw as "politicianIdRaw",
				persoon_id as "politicianId",
				fractie_id_raw as "partyIdRaw",
				fractie_id as "partyId",
				gewijzigd_op as "updatedAt",
				api_gewijzigd_op as "apiUpdatedAt"
			FROM stemmingen
			WHERE fractie_id = ${partyId}
			  AND (${!motionIds || motionIds.length === 0} OR besluit_id = ANY(${motionIds || []}))
			ORDER BY gewijzigd_op DESC
		`;

    const votes = votesWithRelations.map((v) => mapVoteToContract(v));

    return {
      party: mapPartyToContract(party),
      votes,
    };
  }),
};
