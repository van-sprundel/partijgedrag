import { implement } from "@orpc/server";
import { apiContract, type VoteType } from "../contracts/index.js";
import { db } from "../lib/db.js";
import { mapZaakToMotion } from "../utils/motion.js";

const os = implement(apiContract);

export const motionRouter = {
	getAll: os.motions.getAll.handler(async ({ input }) => {
		const { limit, offset, category, status } = input;

		const where: any = {
			soort: "Motie",
		};
		if (category) {
			where.soort = category;
		}
		if (status) {
			where.status = status;
		}

		const [zaken, total] = await Promise.all([
			db.zaken.findMany({
				where,
				include: {
					kamerstukdossiers: true,
				},
				orderBy: { datum: "desc" },
				take: limit,
				skip: offset,
			}),
			db.zaken.count({ where }),
		]);

		const motions = zaken.map((zaak) => {
			const dossier = zaak.kamerstukdossiers?.[0];
			return mapZaakToMotion(zaak, dossier);
		});

		return {
			motions,
			total,
			hasMore: offset + limit < total,
		};
	}),

	getById: os.motions.getById.handler(async ({ input }) => {
		const zaak = await db.zaken.findUnique({
			where: { id: input.id },
			include: {
				kamerstukdossiers: true,
			},
		});

		if (!zaak) {
			return null;
		}

		const dossier = zaak.kamerstukdossiers?.[0];
		return mapZaakToMotion(zaak, dossier);
	}),

	getForCompass: os.motions.getForCompass.handler(async ({ input }) => {
		const { count, excludeIds = [] } = input;

		const zaken = await db.zaken.findMany({
			where: {
				soort: "Motie",
				id: {
					notIn: excludeIds,
				},
			},
			include: {
				kamerstukdossiers: true,
			},
			orderBy: { datum: "desc" },
			take: count,
		});

		const motions = zaken.map((zaak) => {
			const dossier = zaak.kamerstukdossiers?.[0];
			return mapZaakToMotion(zaak, dossier);
		});

		return motions;
	}),

	getVotes: os.motions.getVotes.handler(async ({ input }) => {
		const votesWithRelations = await db.stemmingen.findMany({
			where: {
				besluit: {
					zaak_id: input.motionId,
				},
			},
			include: {
				fractie: true,
				persoon: true,
			},
			orderBy: [
				{ fractie: { naam_nl: "asc" } },
				{ persoon: { achternaam: "asc" } },
			],
		});

		const votes = votesWithRelations.map((vote) => ({
			id: vote.id,
			motionId: input.motionId,
			partyId: vote.fractie_id || "",
			politicianId: vote.persoon_id || "",
			voteType: vote.soort as VoteType,
			reasoning: null,
			createdAt: vote.gewijzigd_op || new Date(),
			updatedAt: vote.api_gewijzigd_op || new Date(),
			party: vote.fractie
				? {
						id: vote.fractie.id,
						name: vote.fractie.naam_nl || vote.fractie.afkorting || "",
						shortName: vote.fractie.afkorting || "",
						color: null,
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
						fullName: `${vote.persoon.voornamen || ""} ${
							vote.persoon.tussenvoegsel || ""
						} ${vote.persoon.achternaam || ""}`.trim(),
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
		const partyPositions = Array.from(partyVoteMap.values()).map(
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
					position: majorityVote as VoteType,
					count,
				};
			},
		);

		return { votes, partyPositions };
	}),
};
