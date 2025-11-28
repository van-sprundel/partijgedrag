import { implement } from "@orpc/server";
import { apiContract, type Party, type VoteType } from "../contracts/index.js";
import { sql } from "../services/db/sql-tag.js";
import {
  getAllMotions,
  getForCompass,
  getForCompassCount,
  getMotionById,
  getMotionCategories,
  getMotionStatistics,
  getRecentMotions,
} from "../services/motions/queries.js";
import { mapPartyToContract } from "../utils/mappers.js";

function mapVoteType(dutchVoteType: string | null): VoteType {
  switch (dutchVoteType) {
    case "Voor":
      return "FOR";
    case "Tegen":
      return "AGAINST";
    case "Niet deelgenomen":
      return "NEUTRAL";
    default:
      return "NEUTRAL";
  }
}

const os = implement(apiContract);

export const motionRouter = {
  getAll: os.motions.getAll.handler(async ({ input }) => {
    const { limit, offset, category, status, withVotes, search } = input;
    const rows = await getAllMotions(
      limit,
      offset,
      category,
      status,
      withVotes,
      search,
    );
    const total = rows[0]?.total ? parseInt(rows[0].total, 10) : 0;
    const motions = rows.map((r) => ({ ...r, total: undefined }));

    return {
      motions,
      total,
      hasMore: offset + limit < total,
    };
  }),

  getById: os.motions.getById.handler(async ({ input }) => {
    const result = await getMotionById(input.id);
    if (!result) return null;

    const bulletPoints = Array.isArray(result.bulletPoints)
      ? result.bulletPoints.filter((bp): bp is string => typeof bp === 'string')
      : [];

    return {
      ...result,
      title: result.title ?? "Untitled Motion",
      status: result.status ?? "unknown",
      bulletPoints,
      createdAt: result.createdAt ?? new Date(),
      updatedAt: result.updatedAt ?? new Date(),
    };
  }),

  getForCompass: os.motions.getForCompass.handler(async ({ input }) => {
    const { count, excludeIds, categoryIds, after } = input;
    return getForCompass(count, excludeIds, categoryIds, after);
  }),

  getForCompassCount: os.motions.getForCompassCount.handler(
    async ({ input }) => {
      const { categoryIds, after } = input;
      return getForCompassCount(categoryIds, after);
    },
  ),

  getRecent: os.motions.getRecent.handler(async ({ input }) => {
    const { limit } = input;
    const results = await getRecentMotions(limit);

    return results.map(r => {
      const bulletPoints = Array.isArray(r.bulletPoints)
        ? r.bulletPoints.filter((bp): bp is string => typeof bp === 'string')
        : [];

      return {
        ...r,
        title: r.title ?? "Untitled Motion",
        status: r.status ?? "unknown",
        bulletPoints,
        createdAt: r.createdAt ?? new Date(),
        updatedAt: r.updatedAt ?? new Date(),
      };
    });
  }),

  getCategories: os.motions.getCategories.handler(async ({ input }) => {
    const { type } = input;
    const results = await getMotionCategories(type);

    return results.map(r => ({
      ...r,
      type: (r.type ?? "generic") as "generic" | "hot_topic",
      keywords: r.keywords ?? [],
      createdAt: r.createdAt ?? new Date(),
      updatedAt: r.updatedAt ?? new Date(),
    }));
  }),

  getVotes: os.motions.getVotes.handler(async ({ input }) => {
    // Get decisions for this motion
    const decisions = await sql<{ id: string; caseId: string | null }>`
			SELECT id, zaak_id as "caseId"
			FROM besluiten
			WHERE zaak_id = ${input.motionId}
		`;
    const decisionIds = decisions.map((d) => d.id);

    if (decisionIds.length === 0) {
      return { votes: [], partyPositions: [] };
    }

    // Get votes for all decisions of this motion
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
			WHERE besluit_id = ANY(${decisionIds})
			  AND vergissing IS DISTINCT FROM true
		`;

    // Map votes to contract format
    const votes = votesWithRelations.map((v) => ({
      id: v.id,
      motionId: v.decisionId || "",
      partyId: v.partyId || "",
      politicianId: v.politicianId || "",
      voteType: mapVoteType(v.type),
      reasoning: null,
      createdAt: v.updatedAt || new Date(),
      updatedAt: v.apiUpdatedAt || new Date(),
    }));

    // Get unique party IDs and fetch party data
    const partyIds = votesWithRelations
      .map((v) => v.partyId)
      .filter((p) => p !== null) as string[];

    if (partyIds.length === 0) {
      return { votes, partyPositions: [] };
    }

    const partiesFromDb = await sql<{
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
      logoData: any | null;
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
			WHERE id = ANY(${partyIds})
		`;
    const parties = partiesFromDb.map(mapPartyToContract);

    const partyVoteMap = new Map<
      string,
      { party: Party; votes: VoteType[]; partySize: number }
    >();

    votesWithRelations.forEach((vote) => {
      if (vote.partyId) {
        const party = parties.find((p) => p.id === vote.partyId);
        if (party) {
          if (!partyVoteMap.has(vote.partyId)) {
            partyVoteMap.set(vote.partyId, {
              party: party,
              votes: [],
              partySize: Number(vote.partySize || 0),
            });
          }
          if (vote.type) {
            partyVoteMap.get(vote.partyId)?.votes.push(mapVoteType(vote.type));
          }
        }
      }
    });

    const partyPositions = Array.from(partyVoteMap.values()).map(
      ({ party, votes: partyVotes, partySize }) => {
        if (partyVotes.length === 0) {
          return {
            party: party,
            position: "NEUTRAL" as const,
            count: partySize,
          };
        }
        const voteCounts = partyVotes.reduce(
          (acc, vote) => {
            acc[vote] = (acc[vote] || 0) + 1;
            return acc;
          },
          {} as Record<VoteType, number>,
        );

        const majorityVoteEntry = (Object.keys(voteCounts) as VoteType[])
          .map((vote) => [vote, voteCounts[vote]] as const)
          .reduce((a, b) => (a[1] > b[1] ? a : b));

        const position = majorityVoteEntry[0] as VoteType;

        return {
          party,
          position,
          count: partySize,
        };
      },
    );

    return { votes, partyPositions };
  }),

  getStatistics: os.motions.getStatistics.handler(async () => {
    return getMotionStatistics();
  }),
};
