import type { Politician } from "../../contracts/index.js";
import { sqlOneOrNull } from "../db/sql-tag.js";

export async function getPoliticianById(id: string) {
	return sqlOneOrNull<Politician>`
        SELECT
            id,
            voornamen as "firstName",
            achternaam as "lastName",
            CONCAT(voornamen, ' ', achternaam) as "fullName",
            bijgewerkt as "updatedAt"
        FROM "personen"
        WHERE id = ${id}
    `;
}
