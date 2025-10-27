import type { Vote } from "../../contracts/index.js";
import { sql } from "../db/sql-tag.js";

export async function getVotesByDecisionIds(decisionIds: string[]) {
	if (decisionIds.length === 0) {
		return [];
	}
	return sql<Vote>`
        SELECT
            id,
            decision_id as "motionId",
            party_id as "partyId",
            politician_id as "politicianId",
            type as "voteType",
            party_size as "partySize",
            updated_at as "createdAt",
            api_updated_at as "updatedAt"
        FROM votes
        WHERE decision_id IN (SELECT unnest(${decisionIds}::text[]))
        AND mistake IS NOT TRUE
    `;
}
