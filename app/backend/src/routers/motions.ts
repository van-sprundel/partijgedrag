import { implement } from "@orpc/server";
import { apiContract, type Party, type VoteType } from "../contracts/index.js";
import { db } from "../lib/db.js";
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
		const { limit, offset, category, status, withVotes } = input;
		const rows = await getAllMotions(
			limit,
			offset,
			category,
			status,
			withVotes,
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
		return getMotionById(input.id);
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
		return getRecentMotions(limit);
	}),

	getCategories: os.motions.getCategories.handler(async ({ input }) => {
		const { type } = input;
		return getMotionCategories(type);
	}),

	getVotes: os.motions.getVotes.handler(async ({ input }) => {
		// Get decisions for this motion
		const { getDecisionsByCaseIds } = await import(
			"../services/decisions/queries.js"
		);
		const decisions = await getDecisionsByCaseIds([input.motionId]);
		const decisionIds = decisions.map((d) => d.id);

		// Get votes for all decisions of this motion
		const { getVotesByDecisionIds } = await import(
			"../services/votes/queries.js"
		);
		const votesWithRelations = await getVotesByDecisionIds(decisionIds);

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
		const { getPartiesByIdsOrNames } = await import(
			"../services/parties/queries.js"
		);
		const parties = await getPartiesByIdsOrNames(partyIds, []);

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
