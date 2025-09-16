import { implement, ORPCError } from "@orpc/server";
import type { Party as PartyModel } from "@prisma/client";
import type { Party, UserAnswer, Vote, VoteType } from "../contracts/index.js";
import { apiContract, CompassResultSchema } from "../contracts/index.js";
import { db } from "../lib/db.js";
import {
	mapCaseToMotion,
	mapPartyToContract,
	mapVoteToContract,
} from "../utils/mappers.js";

const os = implement(apiContract);

export const compassRouter = {
	submitAnswers: os.compass.submitAnswers.handler(async ({ input }) => {
		const { answers } = input;

		// Calculate party alignment scores
		const partyResults = await calculatePartyAlignment(answers);

		// Save session
		const session = await db.userSession.create({
			data: {
				answers: answers,
				results: {
					totalAnswers: answers.length,
					partyResults,
					createdAt: new Date(),
				},
			},
		});

		return {
			id: session.id,
			totalAnswers: answers.length,
			partyResults,
			createdAt: new Date(session.createdAt),
		};
	}),

	getResults: os.compass.getResults.handler(async ({ input }) => {
		const session = await db.userSession.findUnique({
			where: { id: input.sessionId },
		});

		if (!session || !session.results) {
			return null;
		}

		const results = CompassResultSchema.parse({
			...(session.results as object),
			id: session.id,
		});

		// Get detailed motion breakdown if requested
		const answers = session.answers as UserAnswer[];
		const motionDetails = await getMotionVoteDetails(answers);

		return {
			id: session.id,
			totalAnswers: results.totalAnswers,
			partyResults: results.partyResults,
			motionDetails,
			createdAt: session.createdAt,
		};
	}),

	getMotionDetails: os.compass.getMotionDetails.handler(async ({ input }) => {
		const { motionId, includeVotes } = input;

		const zaak = await db.case.findUnique({
			where: { id: motionId },
		});

		if (!zaak) {
			throw new ORPCError("NOT_FOUND", { message: "Motion not found" });
		}

		const motion = mapCaseToMotion(zaak);

		let votes: Vote[] = [];
		let partyPositions: {
			party: Party;
			position: VoteType;
			count: number;
		}[] = [];

		if (includeVotes) {
			const votesWithRelations = await db.vote.findMany({
				where: {
					decision: {
						caseId: motionId,
					},
				},
				include: {
					politician: true,
					party: true,
				},
			});

			votes = votesWithRelations.map((v) => mapVoteToContract(v));

			const partyVoteMap = new Map<
				string,
				{ party: PartyModel; votes: VoteType[] }
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

			partyPositions = Array.from(partyVoteMap.values()).map(
				({ party, votes: partyVotes }) => {
					const voteCounts = partyVotes.reduce(
						(acc, vote) => {
							acc[vote] = (acc[vote] || 0) + 1;
							return acc;
						},
						{} as Record<string, number>,
					);

					const majorityVoteEntry =
						Object.keys(voteCounts).length > 0
							? Object.entries(voteCounts).reduce((a, b) =>
									a[1] > b[1] ? a : b,
								)
							: [];

					const majorityVote =
						(majorityVoteEntry[0] as VoteType) ?? ("NEUTRAL" as const);
					const count = majorityVoteEntry[1] ?? null;

					return {
						party: mapPartyToContract(party),
						position: majorityVote,
						count,
					};
				},
			);
		}

		return {
			motion,
			votes: includeVotes ? votes : undefined,
			partyPositions: includeVotes ? partyPositions : undefined,
		};
	}),
};

async function calculatePartyAlignment(answers: UserAnswer[]) {
	const motionIds = answers.map((a) => a.motionId);

	const votes = await db.vote.findMany({
		where: {
			decision: {
				caseId: { in: motionIds },
			},
		},
		include: {
			party: true,
			decision: true,
		},
	});

	const parties = await db.party.findMany({
		where: {
			OR: [{ activeTo: null }, { activeTo: { gte: new Date() } }],
		},
	});

	const partyScores = new Map<
		string,
		{
			party: Party;
			totalVotes: number;
			matchingVotes: number;
			score: number;
			agreement: number;
		}
	>();

	parties.forEach((party) => {
		partyScores.set(party.id, {
			party: mapPartyToContract(party),
			totalVotes: 0,
			matchingVotes: 0,
			score: 0,
			agreement: 0,
		});
	});

	answers.forEach((answer) => {
		const motionVotes = votes.filter(
			(v) => v.decision?.caseId === answer.motionId,
		);

		const partyPositions = new Map<string, VoteType[]>();

		motionVotes.forEach((vote) => {
			if (vote.partyId && vote.type) {
				if (!partyPositions.has(vote.partyId)) {
					partyPositions.set(vote.partyId, []);
				}
				partyPositions.get(vote.partyId)?.push(vote.type);
			}
		});

		partyPositions.forEach((votes, partyId) => {
			const partyScore = partyScores.get(partyId);
			if (!partyScore) return;

			partyScore.totalVotes++;

			const voteCounts = votes.reduce(
				(acc, vote) => {
					acc[vote] = (acc[vote] || 0) + 1;
					return acc;
				},
				{} as Record<VoteType, number>,
			);

			const majorityVote =
				Object.keys(voteCounts).length > 0
					? (Object.entries(voteCounts).reduce((a, b) =>
							a[1] > b[1] ? a : b,
						)[0] as VoteType)
					: null;

			const userSupports = answer.answer === "agree";
			const partySupports = majorityVote === "FOR";

			if (
				(userSupports && partySupports) ||
				(!userSupports && !partySupports)
			) {
				partyScore.matchingVotes++;
				partyScore.score += answer.answer === "neutral" ? 0.5 : 1;
			} else if (answer.answer === "neutral") {
				partyScore.score += 0.5;
			}

			partyScore.agreement =
				partyScore.totalVotes > 0
					? (partyScore.matchingVotes / partyScore.totalVotes) * 100
					: 0;
		});
	});

	const minVotesThreshold = Math.max(1, Math.floor(answers.length * 0.25));

	const results = Array.from(partyScores.values())
		.filter((result) => result.totalVotes >= minVotesThreshold)
		.sort((a, b) => {
			if (Math.abs(b.agreement - a.agreement) > 0.1) {
				return b.agreement - a.agreement;
			}
			if (Math.abs(b.totalVotes - a.totalVotes) > 0) {
				return b.totalVotes - a.totalVotes;
			}
			return b.score - a.score;
		});

	return results.map((result) => ({
		party: result.party,
		score: Math.round(result.score * 100) / 100,
		agreement: Math.round(result.agreement * 100) / 100,
		totalVotes: result.totalVotes,
		matchingVotes: result.matchingVotes,
	}));
}

async function getMotionVoteDetails(answers: UserAnswer[]) {
	const motionIds = answers.map((a) => a.motionId);

	const motions = await db.case.findMany({
		where: { id: { in: motionIds } },
	});

	const votes = await db.vote.findMany({
		where: {
			decision: {
				caseId: { in: motionIds },
			},
		},
		include: {
			party: true,
			politician: true,
			decision: true,
		},
	});

	return answers.map((answer) => {
		const motion = motions.find((m) => m.id === answer.motionId);
		const motionVotes = votes.filter(
			(v) => v.decision?.caseId === answer.motionId,
		);

		const partyVotes = new Map<
			string,
			{ party: PartyModel; votes: VoteType[]; majorityVote: VoteType }
		>();

		motionVotes.forEach((vote) => {
			if (vote.partyId && vote.party && vote.type) {
				if (!partyVotes.has(vote.partyId)) {
					partyVotes.set(vote.partyId, {
						party: vote.party,
						votes: [],
						majorityVote: "NEUTRAL",
					});
				}
				partyVotes.get(vote.partyId)?.votes.push(vote.type);
			}
		});

		partyVotes.forEach((partyData, _partyId) => {
			const voteCounts = partyData.votes.reduce(
				(acc, vote) => {
					acc[vote as VoteType] = (acc[vote as VoteType] || 0) + 1;
					return acc;
				},
				{} as Record<VoteType, number>,
			);

			const majorityVote =
				Object.keys(voteCounts).length > 0
					? Object.entries(voteCounts).reduce((a, b) =>
							a[1] > b[1] ? a : b,
						)[0]
					: null;

			partyData.majorityVote = (majorityVote ?? "NEUTRAL") as VoteType;
		});

		const partyPositions = Array.from(partyVotes.values()).map(
			({ party, votes, majorityVote }) => ({
				party: mapPartyToContract(party),
				position: majorityVote,
				voteCount: votes.length,
				agreesWithUser:
					majorityVote !== "NEUTRAL" &&
					((answer.answer === "agree" && majorityVote === "FOR") ||
						(answer.answer === "disagree" && majorityVote === "AGAINST") ||
						answer.answer === "neutral"),
			}),
		);

		return {
			motionId: answer.motionId,
			userAnswer: answer.answer,
			motion: motion ? mapCaseToMotion(motion) : null,
			partyPositions,
		};
	});
}
