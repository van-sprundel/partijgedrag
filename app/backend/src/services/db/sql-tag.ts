import { sql as sqlTag } from "@ts-safeql/sql-tag";
import type { QueryResultRow } from "pg";
import { useClient } from "./client.js";

/**
 * SQL Tagged template that automatically binds values and runs it.
 */
export async function sql<T extends QueryResultRow = never>(
	strings: TemplateStringsArray,
	...values: unknown[]
) {
	using client = await useClient();
	const query = sqlTag(strings, ...values);
	const { rows } = await client.query<T>(query);
	return rows;
}

/**
 * SQL Tagged template that returns exactly one or throws an error.
 */
export async function sqlOne<T extends QueryResultRow = never>(
	strings: TemplateStringsArray,
	...values: unknown[]
) {
	const rows = await sql<T>(strings, ...values);

	if (rows.length !== 1) {
		throw new Error(`Expected 1 row, received ${rows.length}`);
	}

	return rows[0];
}

/**
 * SQL Tagged template that returns either one or null.
 */
export const sqlOneOrNull = async <T extends QueryResultRow>(
	strings: TemplateStringsArray,
	...values: unknown[]
): Promise<T | null> =>
	await sql<T>(strings, ...values).then((r) => r[0] ?? null);
