import { sqlOneOrNull } from "../db/sql-tag.js";

export async function getPoliticianById(id: string) {
  return sqlOneOrNull<{
    id: string;
    firstName: string | null;
    lastName: string | null;
    fullName: unknown;
    updatedAt: Date | null;
  }>`
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
