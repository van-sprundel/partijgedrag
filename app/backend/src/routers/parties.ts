import { implement, ORPCError } from "@orpc/server";
import type { Party as PartyModel } from "@prisma/client";
import type { VoteType } from "../contracts/index.js";
import { apiContract } from "../contracts/index.js";
import { db } from "../lib/db.js";
import { mapPartyToContract, mapVoteToContract } from "../utils/mappers.js";

const os = implement(apiContract);

export const partyRouter = {
	getAll: os.parties.getAll.handler(async ({ input }) => {
		const { activeOnly } = input;

		const where: Prisma.PartyWhereInput = {};
		if (activeOnly) {
			where.OR = [{ activeTo: null }, { activeTo: { gte: new Date() } }];
			// where.removed = { not: true };
		}

		const parties = await db.party.findMany({
			where,
			orderBy: { nameNl: "asc" },
		});

		return parties.map((p) => mapPartyToContract(p));
	}),

	getById: os.parties.getById.handler(async ({ input }) => {
		const party = await db.party.findUnique({
			where: { id: input.id },
		});

		if (!party) {
			return null;
		}

		return mapPartyToContract(party);
	}),

	getWithVotes: os.parties.getWithVotes.handler(async ({ input }) => {
		const { partyId, motionIds } = input;

		const party = await db.party.findUnique({
			where: { id: partyId },
		});

		if (!party) {
			throw new ORPCError("NOT_FOUND", { message: "Party not found" });
		}

		const where: Prisma.VoteWhereInput = {
			partyId: partyId,
		};

		if (motionIds && motionIds.length > 0) {
			where.decision = {
				caseId: { in: motionIds },
			};
		}

		const votesWithRelations = await db.vote.findMany({
			where,
			include: {
				politician: true,
				decision: {
					include: {
						case: {
							include: {
								parliamentaryDocuments: true,
							},
						},
					},
				},
			},
			orderBy: { updatedAt: "desc" },
		});

		const votes = votesWithRelations.map((v) => mapVoteToContract(v));

		return {
			party: mapPartyToContract(party),
			votes,
		};
	}),

	getSimilarity: os.parties.getSimilarity.handler(async ({ input }) => {
		const { activeOnly, minMotions } = input;

		// Get parties
		const where: any = {};
		if (activeOnly) {
			where.OR = [{ activeTo: null }, { activeTo: { gte: new Date() } }];
		}

		const parties = await db.party.findMany({
			where,
			orderBy: { nameNl: "asc" },
		});

		// Get all votes with party and motion information
		const votes = await db.vote.findMany({
			where: {
				partyId: { in: parties.map((p) => p.id) },
				decision: {
					case: {
						type: "Motie",
					},
				},
			},
			include: {
				party: true,
				decision: {
					include: {
						case: true,
					},
				},
			},
		});

		// Group votes by motion and party
		const motionVotes = new Map<string, Map<string, VoteType[]>>();

		votes.forEach((vote) => {
			if (!vote.decision?.case || !vote.partyId || !vote.type) return;

			const motionId = vote.decision.case.id;
			const partyId = vote.partyId;

			if (!motionVotes.has(motionId)) {
				motionVotes.set(motionId, new Map());
			}

			const motionPartyVotes = motionVotes.get(motionId)!;
			if (!motionPartyVotes.has(partyId)) {
				motionPartyVotes.set(partyId, []);
			}

			motionPartyVotes.get(partyId)!.push(vote.type as VoteType);
		});

		// Calculate party positions for each motion (majority vote)
		const partyPositions = new Map<string, Map<string, VoteType>>();

		motionVotes.forEach((partyVotes, motionId) => {
			const positions = new Map<string, VoteType>();

			partyVotes.forEach((votes, partyId) => {
				if (votes.length === 0) {
					positions.set(partyId, "NEUTRAL");
					return;
				}

				const voteCounts = votes.reduce(
					(acc, vote) => {
						acc[vote] = (acc[vote] || 0) + 1;
						return acc;
					},
					{} as Record<VoteType, number>,
				);

				const majorityVote = Object.entries(voteCounts).reduce((a, b) =>
					a[1] > b[1] ? a : b,
				)[0] as VoteType;

				positions.set(partyId, majorityVote);
			});

			partyPositions.set(motionId, positions);
		});

		// Filter motions that have votes from at least 2 parties
		const validMotions = Array.from(partyPositions.entries())
			.filter(([_, positions]) => positions.size >= 2)
			.slice(0, minMotions * 10); // Take more motions than minimum to ensure we have enough data

		if (validMotions.length < minMotions) {
			return {
				similarities: [],
				totalMotions: validMotions.length,
			};
		}

		// Calculate similarities between all party pairs
		const similarities: any[] = [];

		for (let i = 0; i < parties.length; i++) {
			for (let j = i + 1; j < parties.length; j++) {
				const party1 = parties[i];
				const party2 = parties[j];

				let agreementCount = 0;
				let totalComparisons = 0;

				validMotions.forEach(([motionId, positions]) => {
					const pos1 = positions.get(party1.id);
					const pos2 = positions.get(party2.id);

					if (pos1 && pos2) {
						totalComparisons++;
						if (pos1 === pos2) {
							agreementCount++;
						}
					}
				});

				if (totalComparisons >= minMotions) {
					const similarity =
						totalComparisons > 0
							? Math.round((agreementCount / totalComparisons) * 100)
							: 0;

					similarities.push({
						party1: mapPartyToContract(party1),
						party2: mapPartyToContract(party2),
						similarity,
						agreementCount,
						totalComparisons,
					});
				}
			}
		}

		// Sort by similarity (highest first)
		similarities.sort((a, b) => b.similarity - a.similarity);

		return {
			similarities,
			totalMotions: validMotions.length,
		};
	}),
};
