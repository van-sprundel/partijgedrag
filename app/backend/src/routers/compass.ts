import { implement } from "@orpc/server";
import type { fracties, stemmingen } from "@prisma/client";
import type { Party, UserAnswer, Vote, VoteType } from "../contracts/index.js";
import { apiContract } from "../contracts/index.js";
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
			createdAt: session.createdAt,
		};
	}),

	getResults: os.compass.getResults.handler(async ({ input }) => {
		const session = await db.user_sessions.findUnique({
			where: { id: input.sessionId },
		});

		if (!session || !session.results) {
			return null;
		}

		const results = session.results as any;
		return {
			id: session.id,
			totalAnswers: results.totalAnswers,
			partyResults: results.partyResults,
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
			throw new Error("Motion not found");
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
			bulletPoints: [],
			// dossier?.bullet_points
			// ? Array.isArray(dossier.bullet_points!)
			// ? dossier.bullet_points! : []
			// : [],
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
			verwijderd: { not: true },
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
				name: party.naam_nl || party.afkorting,
				shortName: party.afkorting,
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

			const majorityVote = Object.entries(voteCounts).reduce((a, b) =>
				a[1] > b[1] ? a : b,
			)[0];

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

	// Convert to array and sort by score
	return Array.from(partyScores.values())
		.filter((result) => result.totalVotes > 0) // Only include parties with votes
		.sort((a, b) => b.score - a.score)
		.map((result) => ({
			party: result.party,
			score: Math.round(result.score * 100) / 100,
			agreement: Math.round(result.agreement * 100) / 100,
			totalVotes: result.totalVotes,
			matchingVotes: result.matchingVotes,
		}));
}
