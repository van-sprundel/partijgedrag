import { sql, sqlOneOrNull } from "../db/sql-tag.js";

export type UserSessionRow = {
	id: string;
	answers: unknown;
	results: unknown | null;
	created_at: Date;
	updated_at: Date;
};

export async function createUserSession(answers: unknown, results: unknown) {
	const rows = await sql<UserSessionRow>`
        INSERT INTO user_sessions (id, answers, results, created_at, updated_at)
        VALUES (gen_random_uuid(), ${answers}::jsonb, ${results}::jsonb, NOW(), NOW())
        RETURNING *
    `;
	return rows[0];
}

export async function getUserSessionById(id: string) {
	return sqlOneOrNull<UserSessionRow>`
        SELECT *
        FROM user_sessions
        WHERE id = ${id}
    `;
}
