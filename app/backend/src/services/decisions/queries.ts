import { sql } from "../db/sql-tag.js";

export type DecisionRow = {
	id: string;
	caseId: string | null;
};

export async function getDecisionsByCaseIds(caseIds: string[]) {
	if (caseIds.length === 0) {
		return [];
	}
	return sql<DecisionRow>`
        SELECT id, case_id as "caseId"
        FROM decisions
        WHERE case_id IN (SELECT unnest(${caseIds}::text[]))
    `;
}
