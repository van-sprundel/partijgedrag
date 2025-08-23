import { implement } from "@orpc/server";
import type { Party as PartyModel } from "@prisma/client";
import { apiContract, type VoteType } from "../contracts/index.js";
import { db } from "../lib/db.js";
import {
	mapCaseToMotion,
	mapPartyToContract,
	mapVoteToContract,
} from "../utils/mappers.js";

const os = implement(apiContract);

export const motionRouter = {
	getAll: os.motions.getAll.handler(async ({ input }) => {
		const { limit, offset, category, status } = input;

		const where: any = {
			type: "Motie",
		};
		if (category) {
			where.type = category;
		}
		if (status) {
			where.status = status;
		}

		const [cases, total] = await Promise.all([
			db.case.findMany({
				where,
				include: {
					parliamentaryDocuments: true,
				},
				orderBy: { date: "desc" },
				take: limit,
				skip: offset,
			}),
			db.case.count({ where }),
		]);

		const motions = cases.map((c) => {
			const dossier = c.parliamentaryDocuments?.[0];
			return mapCaseToMotion(c, dossier);
		});

		return {
			motions,
			total,
			hasMore: offset + limit < total,
		};
	}),

	getById: os.motions.getById.handler(async ({ input }) => {
		const c = await db.case.findUnique({
			where: { id: input.id },
			include: {
				parliamentaryDocuments: true,
			},
		});

		if (!c) {
			return null;
		}

		const dossier = c.parliamentaryDocuments?.[0];
		return mapCaseToMotion(c, dossier);
	}),

	getForCompass: os.motions.getForCompass.handler(async ({ input }) => {
		const { count, excludeIds = [] } = input;

		const cases = await db.case.findMany({
			where: {
				type: "Motie",
				id: {
					notIn: excludeIds,
				},
			},
			include: {
				parliamentaryDocuments: true,
			},
			orderBy: { date: "desc" },
			take: count,
		});

		const motions = cases.map((c) => {
			const dossier = c.parliamentaryDocuments?.[0];
			return mapCaseToMotion(c, dossier);
		});

		return motions;
	}),

	getVotes: os.motions.getVotes.handler(async ({ input }) => {
		const votesWithRelations = await db.vote.findMany({
			where: {
				decision: {
					caseId: input.motionId,
				},
			},
			include: {
				party: true,
				politician: true,
			},
			orderBy: [
				{ party: { nameNl: "asc" } },
				{ politician: { lastName: "asc" } },
			],
		});

		const votes = votesWithRelations.map((v) => mapVoteToContract(v));

		const partyVoteMap = new Map<
			string,
			{ party: PartyModel; votes: string[] }
		>();

		votesWithRelations.forEach((vote) => {
			if (vote.partyId && vote.party) {
				if (!partyVoteMap.has(vote.partyId)) {
					partyVoteMap.set(vote.partyId, {
						party: vote.party,
						votes: [],
					});
				}
				if (vote.type) {
					partyVoteMap.get(vote.partyId)?.votes.push(vote.type);
				}
			}
		});

		const partyPositions = Array.from(partyVoteMap.values()).map(
			({ party, votes: partyVotes }) => {
				if (partyVotes.length === 0) {
					return {
						party: mapPartyToContract(party),
						position: "NEUTRAL" as VoteType,
						count: 0,
					};
				}
				const voteCounts = partyVotes.reduce(
					(acc, vote) => {
						acc[vote] = (acc[vote] || 0) + 1;
						return acc;
					},
					{} as Record<string, number>,
				);

				const majorityVoteEntry = Object.entries(voteCounts).reduce((a, b) =>
					a[1] > b[1] ? a : b,
				);

				const majorityVote = majorityVoteEntry[0];
				const count = majorityVoteEntry[1];

				return {
					party: mapPartyToContract(party),
					position: majorityVote as VoteType,
					count,
				};
			},
		);

		return { votes, partyPositions };
	}),
};
