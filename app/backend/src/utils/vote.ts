import type { VoteType } from "../contracts/index.js";

// Helper function to map database vote types to your VoteType enum
export const mapVoteTypeFromDB = (dbVoteType: string | null): VoteType => {
	if (!dbVoteType) return "ABSTAIN" as VoteType;

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
};
