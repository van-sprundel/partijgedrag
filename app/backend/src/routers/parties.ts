import { implement, ORPCError } from "@orpc/server";
import type { VoteType } from "../contracts/index.js";
import { apiContract } from "../contracts/index.js";
import { db } from "../lib/db.js";

const os = implement(apiContract);

function mapZaakToMotion(zaak: any, dossier?: any) {
	return {
		id: zaak.id,
		title: zaak.titel || zaak.onderwerp || "Untitled Motion",
		description: zaak.onderwerp,
		shortTitle: zaak.citeertitel,
		motionNumber: zaak.nummer,
		date: zaak.datum,
		status: zaak.status || "unknown",
		category: zaak.soort,
		bulletPoints: dossier?.bullet_points
			? Array.isArray(dossier.bullet_points)
				? dossier.bullet_points
				: []
			: [],
		originalId: zaak.id,
		createdAt: zaak.gestart_op || new Date(),
		updatedAt: zaak.gewijzigd_op || new Date(),
	};
}

function mapFractieToParty(fractie: any) {
	return {
		id: fractie.id,
		name: fractie.naam_nl || fractie.afkorting || "",
		shortName: fractie.afkorting || "",
		color: fractie.kleur || null,
		seats: Number(fractie.aantal_zetels) || 0,
		activeFrom: fractie.datum_actief,
		activeTo: fractie.datum_inactief,
		createdAt: fractie.gewijzigd_op || new Date(),
		updatedAt: fractie.api_gewijzigd_op || new Date(),
	};
}

export const partyRouter = {
	getAll: os.parties.getAll.handler(async ({ input }) => {
		const { activeOnly } = input;

		const where: any = {};
		if (activeOnly) {
			where.OR = [
				{ datum_inactief: null },
				{ datum_inactief: { gte: new Date() } },
			];
			where.verwijderd = { not: true };
		}

		const parties = await db.fracties.findMany({
			where,
			orderBy: { naam_nl: "asc" },
		});

		return parties.map(mapFractieToParty);
	}),

	getById: os.parties.getById.handler(async ({ input }) => {
		const party = await db.fracties.findUnique({
			where: { id: input.id },
		});

		if (!party) {
			return null;
		}

		return mapFractieToParty(party);
	}),

	getWithVotes: os.parties.getWithVotes.handler(async ({ input }) => {
		const { partyId, motionIds } = input;

		const party = await db.fracties.findUnique({
			where: { id: partyId },
		});

		if (!party) {
			throw new ORPCError("NOT_FOUND", { message: "Party not found" });
		}

		const where: any = {
			fractie_id: partyId,
		};

		if (motionIds && motionIds.length > 0) {
			where.besluit = {
				zaak_id: { in: motionIds },
			};
		}

		const votesWithRelations = await db.stemmingen.findMany({
			where,
			include: {
				persoon: true,
				besluit: {
					include: {
						zaak: {
							include: {
								kamerstukdossiers: true,
							},
						},
					},
				},
			},
			orderBy: { gewijzigd_op: "desc" },
		});

		const votes = votesWithRelations.map((vote) => {
			const motion = vote.besluit?.zaak
				? mapZaakToMotion(
						vote.besluit.zaak,
						vote.besluit.zaak.kamerstukdossiers?.[0],
					)
				: undefined;

			return {
				id: vote.id,
				motionId: motion?.id || "",
				partyId: vote.fractie_id || "",
				politicianId: vote.persoon_id || "",
				voteType: (vote.soort as VoteType) || "",
				reasoning: null,
				createdAt: vote.gewijzigd_op || new Date(),
				updatedAt: vote.api_gewijzigd_op || new Date(),
				motion,
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
			};
		});

		return {
			party: mapFractieToParty(party),
			votes,
		};
	}),
};
