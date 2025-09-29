import type {
	Decision,
	PartyResult,
	UserAnswer,
	UserSession,
	Vote,
} from "../../contracts/index.js";
import { sql, sqlOne, sqlOneOrNull } from "../db/sql-tag.js";

export async function createUserSession(
	answers: UserAnswer[],
	results: {
		totalAnswers: number;
		partyResults: PartyResult[];
		createdAt: Date;
	},
) {
	return sqlOne<UserSession>`
        INSERT INTO "user_sessions" (answers, results)
        VALUES (${JSON.stringify(answers)}, ${JSON.stringify(results)})
        RETURNING id, answers, results, created_at as "createdAt", updated_at as "updatedAt"
    `;
}

export async function getUserSessionById(sessionId: string) {
	return sqlOneOrNull<UserSession>`
        SELECT
            id, answers, results, created_at as "createdAt", updated_at as "updatedAt"
        FROM "user_sessions"
        WHERE id = ${sessionId}
    `;
}

export async function getDecisionIdsByCaseId(motionId: string) {
	return sql<{ id: string }[]>`
        SELECT id FROM "besluiten"
        WHERE "zaak_id" = ${motionId}
    `;
}

export async function getVotesByDecisionIds(decisionIds: string[]) {
	if (decisionIds.length === 0) {
		return [];
	}
	return sql<Vote[]>`
        SELECT
            id,
            besluit_id as "motionId",
            fractie_id as "partyId",
            persoon_id as "politicianId",
            soort as "voteType",
            gewijzigd_op as "createdAt",
            api_gewijzigd_op as "updatedAt"
        FROM "stemmingen"
        WHERE "besluit_id" IN (${decisionIds})
    `;
}

export async function getDecisionsByCaseIds(motionIds: string[]) {
	if (motionIds.length === 0) {
		return [];
	}
	return sql<Pick<Decision, "id" | "caseId">[]>`
        SELECT id, zaak_id as "caseId" FROM "besluiten"
        WHERE "zaak_id" IN (${motionIds})
    `;
}
