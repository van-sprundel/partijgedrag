import { implement, ORPCError } from "@orpc/server";
import type { Party, UserAnswer, Vote, VoteType } from "../contracts/index.js";
import { apiContract, CompassResultSchema } from "../contracts/index.js";
import { getDecisionsByCaseIds } from "../services/decisions/queries.js";
import { getMotionsByIds } from "../services/motions/queries.js";
import { getActiveParties, getPartiesByIdsOrNames } from "../services/parties/queries.js";
import {
	createUserSession,
	getUserSessionById,
} from "../services/sessions/queries.js";
import { getVotesByDecisionIds } from "../services/votes/queries.js";
import { mapCaseToMotion } from "../utils/mappers.js";

const os = implement(apiContract);

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

export const compassRouter = {
	submitAnswers: os.compass.submitAnswers.handler(async ({ input }) => {
		const { answers } = input;

		const partyResults = await calculatePartyAlignment(answers);
		const session = await createUserSession(
			answers,
			{
				totalAnswers: answers.length,
				partyResults,
				createdAt: new Date(),
			}
		);

		return {
			id: session.id,
			totalAnswers: answers.length,
			partyResults,
			createdAt: new Date(session.created_at),
		};
	}),

	getResults: os.compass.getResults.handler(async ({ input }) => {
		const session = await getUserSessionById(input.sessionId);

		if (!session || !session.results) {
			return null;
		}

		const results = CompassResultSchema.parse({
			...(session.results as object),
			id: session.id,
		});

		const answers = session.answers as UserAnswer[];
		const motionDetails = await getMotionVoteDetails(answers);

		return {
			id: session.id,
			totalAnswers: results.totalAnswers,
			partyResults: results.partyResults,
			motionDetails,
			createdAt: session.created_at,
		};
	}),

	getMotionDetails: os.compass.getMotionDetails.handler(async ({ input }) => {
		const { motionId, includeVotes } = input;

		const motions = await getMotionsByIds([motionId]);

		if (!motions || motions.length === 0) {
			throw new ORPCError("NOT_FOUND", { message: "Motion not found" });
		}

		const motion = motions[0];

		let votes: Vote[] = [];
		let partyPositions: {
			party: Party;
			position: VoteType;
			count: number;
		}[] = [];

		if (includeVotes) {
			const decisions = await getDecisionsByCaseIds([motionId]);
			const decisionIds = decisions.map((d) => d.id);

			const votesWithRelations = await getVotesByDecisionIds(decisionIds);

			votes = votesWithRelations.map((v) => ({
				id: v.id,
				motionId: v.motionId,
				partyId: v.partyId || "",
				politicianId: v.politicianId || "",
				voteType: mapVoteType(v.voteType),
				reasoning: null,
				createdAt: v.createdAt,
				updatedAt: v.updatedAt,
			}));

			const partyVoteMap = new Map<
				string,
				{ party: Party; votes: VoteType[]; partySize: number }
			>();

			const partyIds = votesWithRelations
				.map((v) => v.partyId)
				.filter((p) => p !== null) as string[];
			const parties = await getPartiesByIdsOrNames(partyIds, []);

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
							partyVoteMap
								.get(vote.partyId)
								?.votes.push(mapVoteType(vote.type));
						}
					}
				}
			});

			partyPositions = Array.from(partyVoteMap.values()).map(
				({ party, votes: partyVotes, partySize }) => {
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

					return {
						party,
						position: majorityVote,
						count: partySize,
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

	const decisions = await getDecisionsByCaseIds(motionIds);
	const decisionIds = decisions.map((d) => d.id);

	const votesRaw = await getVotesByDecisionIds(decisionIds);
	const votes = votesRaw.map((v) => ({
		...v,
		type: v.voteType,
	}));

	const parties = await getActiveParties();

	const partyNameMap = new Map<string, Party>();
	parties.forEach((p) => {
		if (p.name) partyNameMap.set(p.name, p);
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
			party,
			totalVotes: 0,
			matchingVotes: 0,
			score: 0,
			agreement: 0,
		});
	});

	answers.forEach((answer) => {
		const decision = decisions.find((d) => d.caseId === answer.motionId);
		if (!decision) return;

		const motionVotes = votes.filter((v) => v.decisionId === decision.id);

		const partyPositions = new Map<string, VoteType[]>();

		motionVotes.forEach((vote) => {
			const party = vote.partyId
				? parties.find((p) => p.id === vote.partyId)
				: vote.actorParty
					? partyNameMap.get(vote.actorParty)
					: undefined;

			if (party?.id && vote.type) {
				const partyId = party.id;
				if (!partyPositions.has(partyId)) {
					partyPositions.set(partyId, []);
				}
				partyPositions.get(partyId)?.push(mapVoteType(vote.type));
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

	const motions = await getMotionsByIds(motionIds);

	const decisions = await getDecisionsByCaseIds(motionIds);
	const decisionIds = decisions.map((d) => d.id);

	const votesRaw = await getVotesByDecisionIds(decisionIds);
	const votes = votesRaw.map((v) => ({
		...v,
		type: v.voteType,
	}));

	const partyIdsFromVotes = votes
		.map((v) => v.partyId)
		.filter((p) => p !== null) as string[];
	const partyNamesFromVotes = votes
		.map((v) => v.actorParty)
		.filter((p) => p !== null) as string[];

	const parties = await getPartiesByIdsOrNames(partyIdsFromVotes, partyNamesFromVotes);

	const partyNameMap = new Map<string, Party>();
	parties.forEach((p) => {
		if (p.name) partyNameMap.set(p.name, p);
	});

	return answers.map((answer) => {
		const motion = motions.find((m) => m.id === answer.motionId);
		const decision = decisions.find((d) => d.caseId === answer.motionId);

		const motionVotes = decision
			? votes.filter((v) => v.decisionId === decision.id)
			: [];

		const partyVotes = new Map<
			string,
			{
				party: Party;
				votes: VoteType[];
				majorityVote: VoteType;
				partySize: number;
			}
		>();

		motionVotes.forEach((vote) => {
			const party = vote.partyId
				? parties.find((p) => p.id === vote.partyId)
				: vote.actorParty
					? partyNameMap.get(vote.actorParty)
					: undefined;

			if (party?.id && vote.type) {
				const partyId = party.id;
				if (!partyVotes.has(partyId)) {
					partyVotes.set(partyId, {
						party: party,
						votes: [],
						majorityVote: "NEUTRAL",
						partySize: Number(vote.partySize || 0),
					});
				}
				partyVotes.get(partyId)?.votes.push(mapVoteType(vote.type));
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
			({ party, majorityVote, partySize }) => ({
				party,
				position: majorityVote,
				voteCount: partySize,
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
			motion: motion || null,
			partyPositions,
		};
	});
}
