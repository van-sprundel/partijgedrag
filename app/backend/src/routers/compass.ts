import { implement, ORPCError } from "@orpc/server";
import type { fracties, stemmingen } from "@prisma/client";
import type { Party, UserAnswer, Vote, VoteType } from "../contracts/index.js";
import { apiContract, CompassResultSchema } from "../contracts/index.js";
import { db } from "../lib/db.js";

const os = implement(apiContract);

export const compassRouter = {
	submitAnswers: os.compass.submitAnswers.handler(async ({ input }) => {
		const { answers } = input;

		// Calculate party alignment scores
		const partyResults = await calculatePartyAlignment(answers);

		// Save session
		const session = await db.user_sessions.create({
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
		const session = await db.user_sessions.findUnique({
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

		// Find the motion in zaken (since motie is a type of zaak)
		const zaak = await db.zaken.findUnique({
			where: { id: motionId },
			include: {
				kamerstukdossiers: true,
			},
		});

		if (!zaak) {
			throw new ORPCError("NOT_FOUND", { message: "Motion not found" });
		}

		// Get the related kamerstukdossier for additional info
		const dossier = zaak.kamerstukdossiers[0];

		// Map zaak to motion format
		const motion = {
			id: zaak.id,
			title: zaak.titel || zaak.onderwerp || "Untitled Motion",
			description: zaak.onderwerp,
			shortTitle: zaak.citeertitel,
			motionNumber: zaak.nummer,
			date: zaak.datum,
			status: zaak.status || "unknown",
			category: zaak.soort,
			bulletPoints:
				dossier &&
				dossier.bullet_points != null &&
				Array.isArray(dossier.bullet_points)
					? dossier.bullet_points.filter(
							(bp): bp is string => typeof bp === "string",
						)
					: [],
			originalId: zaak.id,
			createdAt: zaak.gestart_op || new Date(),
			updatedAt: zaak.gewijzigd_op || new Date(),
		};

		let votes: any[] = [];
		let partyPositions: any[] = [];

		if (includeVotes) {
			// Get votes through the proper relationship chain
			const votesWithRelations = await db.stemmingen.findMany({
				where: {
					besluit: {
						zaak_id: motionId,
					},
				},
				include: {
					persoon: true,
					fractie: true,
				},
			});

			// Map votes to expected format
			votes = votesWithRelations.map((vote) => ({
				id: vote.id,
				motionId: motionId,
				partyId: vote.fractie_id || "",
				politicianId: vote.persoon_id || "",
				voteType: mapVoteTypeFromDB(vote.soort || ""),
				reasoning: null, // Not available in your schema
				createdAt: vote.gewijzigd_op || new Date(),
				updatedAt: vote.api_gewijzigd_op || new Date(),
				// Optional nested objects
				party: vote.fractie
					? {
							id: vote.fractie.id,
							name: vote.fractie.naam_nl || vote.fractie.afkorting || "",
							shortName: vote.fractie.afkorting || "",
							color: null, // Not available in your schema
							seats: Number(vote.fractie.aantal_zetels) || 0,
							activeFrom: vote.fractie.datum_actief,
							activeTo: vote.fractie.datum_inactief,
							createdAt: vote.fractie.gewijzigd_op || new Date(),
							updatedAt: vote.fractie.api_gewijzigd_op || new Date(),
						}
					: undefined,
				politician: vote.persoon
					? {
							id: vote.persoon.id,
							firstName: vote.persoon.voornamen || "",
							lastName: vote.persoon.achternaam || "",
							fullName:
								`${vote.persoon.voornamen || ""} ${vote.persoon.tussenvoegsel || ""} ${vote.persoon.achternaam || ""}`.trim(),
							// Add other politician fields as needed
						}
					: undefined,
			}));

			// Group votes by party (fractie)
			const partyVoteMap = new Map<string, { party: any; votes: string[] }>();

			votesWithRelations.forEach((vote) => {
				if (vote.fractie_id && vote.fractie) {
					if (!partyVoteMap.has(vote.fractie_id)) {
						partyVoteMap.set(vote.fractie_id, {
							party: vote.fractie,
							votes: [],
						});
					}
					if (vote.soort) {
						partyVoteMap.get(vote.fractie_id)?.votes.push(vote.soort);
					}
				}
			});

			// Calculate majority position for each party
			partyPositions = Array.from(partyVoteMap.values()).map(
				({ party, votes: partyVotes }) => {
					const voteCounts = partyVotes.reduce(
						(acc, vote) => {
							acc[vote] = (acc[vote] || 0) + 1;
							return acc;
						},
						{} as Record<string, number>,
					);

					// Find majority vote
					const majorityVoteEntry = Object.entries(voteCounts).reduce((a, b) =>
						a[1] > b[1] ? a : b,
					);

					const majorityVote = majorityVoteEntry[0];
					const count = majorityVoteEntry[1];

					return {
						party: {
							id: party.id,
							name: party.naam_nl || party.afkorting || "",
							shortName: party.afkorting || "",
							color: null, // Not available in your schema
							seats: Number(party.aantal_zetels) || 0,
							activeFrom: party.datum_actief,
							activeTo: party.datum_inactief,
							createdAt: party.gewijzigd_op || new Date(),
							updatedAt: party.api_gewijzigd_op || new Date(),
						},
						position: mapVoteTypeFromDB(majorityVote),
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

// Helper function to map database vote types to your VoteType enum
function mapVoteTypeFromDB(dbVoteType: string | null): VoteType {
	if (!dbVoteType) return "ABSTAIN" as VoteType;

	// Map Dutch vote types to your enum
	switch (dbVoteType.toLowerCase()) {
		case "voor":
		case "for":
			return "FOR" as VoteType;
		case "tegen":
		case "against":
			return "AGAINST" as VoteType;
		case "onthouding":
		case "abstain":
			return "ABSTAIN" as VoteType;
		default:
			return "ABSTAIN" as VoteType;
	}
}

async function calculatePartyAlignment(answers: UserAnswer[]) {
	const motionIds = answers.map((a) => a.motionId);

	// Get all votes for the motions in question
	const votes = await db.stemmingen.findMany({
		where: {
			besluit: {
				zaak_id: { in: motionIds },
			},
		},
		include: {
			fractie: true,
			besluit: true,
		},
	});

	// Get all active parties (fracties)
	const parties = await db.fracties.findMany({
		where: {
			OR: [{ datum_inactief: null }, { datum_inactief: { gte: new Date() } }],
		},
	});

	// Calculate scores for each party
	const partyScores = new Map<
		string,
		{
			party: any;
			totalVotes: number;
			matchingVotes: number;
			score: number;
			agreement: number;
		}
	>();

	// Initialize party scores
	parties.forEach((party) => {
		partyScores.set(party.id, {
			party: {
				id: party.id,
				name: party.naam_nl || party.afkorting || "",
				shortName: party.afkorting || "",
				color: null,
				seats: Number(party.aantal_zetels) || 0,
				activeFrom: party.datum_actief ? new Date(party.datum_actief) : null,
				activeTo: party.datum_inactief ? new Date(party.datum_inactief) : null,
				createdAt: party.gewijzigd_op
					? new Date(party.gewijzigd_op)
					: new Date(),
				updatedAt: party.api_gewijzigd_op
					? new Date(party.api_gewijzigd_op)
					: new Date(),
			},
			totalVotes: 0,
			matchingVotes: 0,
			score: 0,
			agreement: 0,
		});
	});

	// Calculate alignment for each answer
	answers.forEach((answer) => {
		// Get votes for this specific motion
		const motionVotes = votes.filter(
			(v) => v.besluit?.zaak_id === answer.motionId,
		);

		// Group votes by party to find majority position
		const partyPositions = new Map<string, string[]>();

		motionVotes.forEach((vote) => {
			if (vote.fractie_id && vote.soort) {
				if (!partyPositions.has(vote.fractie_id)) {
					partyPositions.set(vote.fractie_id, []);
				}
				partyPositions.get(vote.fractie_id)?.push(vote.soort);
			}
		});

		// Calculate majority vote for each party and compare with user answer
		partyPositions.forEach((votes, partyId) => {
			const partyScore = partyScores.get(partyId);
			if (!partyScore) return;

			partyScore.totalVotes++;

			// Find majority vote for this party
			const voteCounts = votes.reduce(
				(acc, vote) => {
					acc[vote] = (acc[vote] || 0) + 1;
					return acc;
				},
				{} as Record<string, number>,
			);

			const majorityVote =
				Object.keys(voteCounts).length > 0
					? Object.entries(voteCounts).reduce((a, b) =>
							a[1] > b[1] ? a : b,
						)[0]
					: null;

			// Convert to comparable format
			const userSupports = answer.answer === "agree";
			const partySupports = mapVoteTypeFromDB(majorityVote) === "FOR";

			// Award points for alignment
			if (
				(userSupports && partySupports) ||
				(!userSupports && !partySupports)
			) {
				partyScore.matchingVotes++;
				partyScore.score += answer.answer === "neutral" ? 0.5 : 1;
			} else if (answer.answer === "neutral") {
				partyScore.score += 0.5;
			}

			// Calculate agreement percentage
			partyScore.agreement =
				partyScore.totalVotes > 0
					? (partyScore.matchingVotes / partyScore.totalVotes) * 100
					: 0;
		});
	});

	// Add debugging information
	console.log(`\n=== Vote Analysis Debug ===`);
	console.log(`Total motions answered by user: ${answers.length}`);
	console.log(`Total votes found in database: ${votes.length}`);

	// Check which motions have votes
	const motionsWithVotes = new Set(
		votes.map((v) => v.besluit?.zaak_id).filter(Boolean),
	);
	console.log(
		`Motions with votes in DB: ${motionsWithVotes.size}/${answers.length}`,
	);

	answers.forEach((answer, index) => {
		const hasVotes = motionsWithVotes.has(answer.motionId);
		console.log(
			`Motion ${index + 1} (${answer.motionId}): ${hasVotes ? "HAS VOTES" : "NO VOTES"} - User answered: ${answer.answer}`,
		);
	});

	// Add logging to understand vote coverage
	console.log("\nVote coverage per party:");
	partyScores.forEach((score, partyId) => {
		if (score.totalVotes > 0) {
			console.log(
				`${score.party.shortName}: ${score.matchingVotes}/${score.totalVotes} matches (${Math.round(score.agreement)}% agreement) - Score: ${score.score}`,
			);
		}
	});

	// Filter out parties with very low vote coverage (less than 25% of motions)
	const minVotesThreshold = Math.max(1, Math.floor(answers.length * 0.25));
	console.log(
		`\nMinimum vote threshold: ${minVotesThreshold}/${answers.length} motions`,
	);

	partyScores.forEach((score, partyId) => {
		if (score.totalVotes > 0 && score.totalVotes < minVotesThreshold) {
			console.log(
				`Filtering out ${score.party.shortName} due to low vote coverage: ${score.totalVotes}/${answers.length} motions`,
			);
		}
	});

	// Convert to array and apply improved sorting
	const results = Array.from(partyScores.values())
		.filter((result) => result.totalVotes >= minVotesThreshold) // Filter out parties with insufficient vote coverage
		.sort((a, b) => {
			// Primary sort: agreement percentage (descending)
			if (Math.abs(b.agreement - a.agreement) > 0.1) {
				return b.agreement - a.agreement;
			}
			// Secondary sort: total votes (descending) - prefer parties that voted on more issues when agreement is similar
			if (Math.abs(b.totalVotes - a.totalVotes) > 0) {
				return b.totalVotes - a.totalVotes;
			}
			// Tertiary sort: raw score (descending)
			return b.score - a.score;
		});

	console.log("\nFinal ranking:");
	results.slice(0, 5).forEach((result, index) => {
		console.log(
			`${index + 1}. ${result.party.shortName}: ${result.matchingVotes}/${result.totalVotes} (${Math.round(result.agreement)}%) - Score: ${result.score}`,
		);
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

	// Get motion details
	const motions = await db.zaken.findMany({
		where: { id: { in: motionIds } },
		include: {
			kamerstukdossiers: true,
		},
	});

	// Get all votes for these motions
	const votes = await db.stemmingen.findMany({
		where: {
			besluit: {
				zaak_id: { in: motionIds },
			},
		},
		include: {
			fractie: true,
			persoon: true,
			besluit: true,
		},
	});

	return answers.map((answer) => {
		const motion = motions.find((m) => m.id === answer.motionId);
		const motionVotes = votes.filter(
			(v) => v.besluit?.zaak_id === answer.motionId,
		);

		// Group votes by party
		const partyVotes = new Map<
			string,
			{ party: any; votes: string[]; majorityVote: string }
		>();

		motionVotes.forEach((vote) => {
			if (vote.fractie_id && vote.fractie && vote.soort) {
				if (!partyVotes.has(vote.fractie_id)) {
					partyVotes.set(vote.fractie_id, {
						party: vote.fractie,
						votes: [],
						majorityVote: "",
					});
				}
				partyVotes.get(vote.fractie_id)?.votes.push(vote.soort);
			}
		});

		// Calculate majority vote for each party
		partyVotes.forEach((partyData, partyId) => {
			const voteCounts = partyData.votes.reduce(
				(acc, vote) => {
					acc[vote] = (acc[vote] || 0) + 1;
					return acc;
				},
				{} as Record<string, number>,
			);

			const majorityVote =
				Object.keys(voteCounts).length > 0
					? Object.entries(voteCounts).reduce((a, b) =>
							a[1] > b[1] ? a : b,
						)[0]
					: null;

			partyData.majorityVote = mapVoteTypeFromDB(majorityVote);
		});

		// Convert to array format
		const partyPositions = Array.from(partyVotes.values()).map(
			({ party, votes, majorityVote }) => ({
				party: {
					id: party.id,
					name: party.naam_nl || party.afkorting || "",
					shortName: party.afkorting || "",
					color: null,
					seats: Number(party.aantal_zetels) || 0,
					activeFrom: party.datum_actief ? new Date(party.datum_actief) : null,
					activeTo: party.datum_inactief
						? new Date(party.datum_inactief)
						: null,
					createdAt: party.gewijzigd_op
						? new Date(party.gewijzigd_op)
						: new Date(),
					updatedAt: party.api_gewijzigd_op
						? new Date(party.api_gewijzigd_op)
						: new Date(),
				},
				position: majorityVote,
				voteCount: votes.length,
				agreesWithUser:
					(answer.answer === "agree" && majorityVote === "FOR") ||
					(answer.answer === "disagree" && majorityVote === "AGAINST") ||
					answer.answer === "neutral",
			}),
		);

		const dossier = motion?.kamerstukdossiers?.[0];

		return {
			motionId: answer.motionId,
			userAnswer: answer.answer,
			motion: motion
				? {
						id: motion.id,
						title: motion.titel || motion.onderwerp || "Untitled Motion",
						description: motion.onderwerp,
						shortTitle: motion.citeertitel,
						motionNumber: motion.nummer,
						date: motion.datum,
						status: motion.status || "unknown",
						category: motion.soort,
						bulletPoints:
							dossier?.bullet_points && Array.isArray(dossier.bullet_points)
								? dossier.bullet_points.filter(
										(bp): bp is string => typeof bp === "string",
									)
								: [],
						originalId: motion.id,
						createdAt: motion.gestart_op || new Date(),
						updatedAt: motion.gewijzigd_op || new Date(),
					}
				: null,
			partyPositions,
		};
	});
}
