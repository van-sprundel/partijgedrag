import { implement, ORPCError } from "@orpc/server";
import type { Prisma } from "@prisma/client";
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
			where.decisionId = { in: motionIds };
		}

		const votesWithRelations = await db.vote.findMany({
			where,
			orderBy: { updatedAt: "desc" },
		});

		const votes = votesWithRelations.map((v) => mapVoteToContract(v));

		return {
			party: mapPartyToContract(party),
			votes,
		};
	}),
};
