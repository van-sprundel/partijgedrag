import type {
  Decision,
  PartyResult,
  UserAnswer,
  UserSession,
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
        VALUES (${answers}::jsonb, ${results}::jsonb)
        RETURNING id, answers, results, "createdAt", "updatedAt"
    `;
}

export async function getUserSessionById(sessionId: string) {
  return sqlOneOrNull<{ id: string; answers: unknown; results: unknown | null; createdAt: Date; updatedAt: Date }>`
        SELECT
            id, answers, results, "createdAt", "updatedAt"
        FROM "user_sessions"
        WHERE id = ${sessionId}
    `;
}

export async function getDecisionIdsByCaseId(motionId: string) {
  return sql<{ id: string }>`
        SELECT id FROM "besluiten"
        WHERE "zaak_id" = ${motionId}
    `;
}

export async function getVotesByDecisionIds(decisionIds: string[]) {
  if (decisionIds.length === 0) {
    return [];
  }
  return sql<{ id: string; motionId: string | null; partyId: string | null; politicianId: string | null; voteType: string | null; createdAt: Date | null; updatedAt: Date | null }>`
        SELECT
            id,
            besluit_id as "motionId",
            fractie_id as "partyId",
            persoon_id as "politicianId",
            soort as "voteType",
            gewijzigd_op as "createdAt",
            api_gewijzigd_op as "updatedAt"
        FROM "stemmingen"
        WHERE "besluit_id" = ANY(${decisionIds}) AND "vergissing" IS NOT TRUE
    `;
}

export async function getDecisionsByCaseIds(motionIds: string[]) {
  if (motionIds.length === 0) {
    return [];
  }
  return sql<Pick<Decision, "id" | "caseId">>`
        SELECT id, zaak_id as "caseId" FROM "besluiten"
        WHERE "zaak_id" = ANY(${motionIds})
    `;
}
