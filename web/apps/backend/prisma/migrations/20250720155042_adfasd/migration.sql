-- DropForeignKey
ALTER TABLE "individuele_stemming" DROP CONSTRAINT "individuele_stemming_person_id_fkey";

-- AlterTable
ALTER TABLE "individuele_stemming" ALTER COLUMN "person_id" DROP NOT NULL;

-- AddForeignKey
ALTER TABLE "individuele_stemming" ADD CONSTRAINT "individuele_stemming_person_id_fkey" FOREIGN KEY ("person_id") REFERENCES "personen"("id") ON DELETE SET NULL ON UPDATE CASCADE;
