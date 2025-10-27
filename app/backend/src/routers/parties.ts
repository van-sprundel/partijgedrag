import { implement, ORPCError } from "@orpc/server";
import { apiContract } from "../contracts/index.js";
import {
	getAllParties,
	getPartyById,
	getVotesByPartyAndMotionIds,
} from "../services/parties/queries.js";

const os = implement(apiContract);

export const partyRouter = {
	getAll: os.parties.getAll.handler(async ({ input }) => {
		const { activeOnly } = input;
		return getAllParties(activeOnly);
	}),

	getById: os.parties.getById.handler(async ({ input }) => {
		return getPartyById(input.id);
	}),

	getWithVotes: os.parties.getWithVotes.handler(async ({ input }) => {
		const { partyId, motionIds } = input;

		const party = await getPartyById(partyId);

		if (!party) {
			throw new ORPCError("NOT_FOUND", { message: "Party not found" });
		}

		const votes = await getVotesByPartyAndMotionIds(partyId, motionIds);

		return {
			party,
			votes,
		};
	}),
};
