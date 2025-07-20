/*
  Warnings:

  - You are about to drop the column `agendapunt_id` on the `besluiten` table. All the data in the column will be lost.
  - You are about to drop the column `datum` on the `zaken` table. All the data in the column will be lost.
  - You are about to drop the column `datum_afgedaan` on the `zaken` table. All the data in the column will be lost.
  - You are about to drop the column `grondslagvoorhang` on the `zaken` table. All the data in the column will be lost.
  - You are about to drop the column `huidige_behandelstatus` on the `zaken` table. All the data in the column will be lost.
  - You are about to drop the column `kamer` on the `zaken` table. All the data in the column will be lost.
  - You are about to drop the column `termijn` on the `zaken` table. All the data in the column will be lost.

*/
-- AlterTable
ALTER TABLE "besluiten" DROP COLUMN "agendapunt_id",
ADD COLUMN     "zaak_id" TEXT;

-- AlterTable
ALTER TABLE "zaken" DROP COLUMN "datum",
DROP COLUMN "datum_afgedaan",
DROP COLUMN "grondslagvoorhang",
DROP COLUMN "huidige_behandelstatus",
DROP COLUMN "kamer",
DROP COLUMN "termijn",
ALTER COLUMN "kabinetsappreciatie" DROP NOT NULL;

-- AddForeignKey
ALTER TABLE "besluiten" ADD CONSTRAINT "besluiten_zaak_id_fkey" FOREIGN KEY ("zaak_id") REFERENCES "zaken"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "stemmingen" ADD CONSTRAINT "stemmingen_besluit_id_fkey" FOREIGN KEY ("besluit_id") REFERENCES "besluiten"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "stemmingen" ADD CONSTRAINT "stemmingen_persoon_id_fkey" FOREIGN KEY ("persoon_id") REFERENCES "personen"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "stemmingen" ADD CONSTRAINT "stemmingen_fractie_id_fkey" FOREIGN KEY ("fractie_id") REFERENCES "fracties"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "zaak_actors" ADD CONSTRAINT "zaak_actors_zaak_id_fkey" FOREIGN KEY ("zaak_id") REFERENCES "zaken"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "zaak_actors" ADD CONSTRAINT "zaak_actors_persoon_id_fkey" FOREIGN KEY ("persoon_id") REFERENCES "personen"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "zaak_actors" ADD CONSTRAINT "zaak_actors_fractie_id_fkey" FOREIGN KEY ("fractie_id") REFERENCES "fracties"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "voting_results" ADD CONSTRAINT "voting_results_besluit_id_fkey" FOREIGN KEY ("besluit_id") REFERENCES "besluiten"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "individuele_stemming" ADD CONSTRAINT "individuele_stemming_besluit_id_fkey" FOREIGN KEY ("besluit_id") REFERENCES "besluiten"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "individuele_stemming" ADD CONSTRAINT "individuele_stemming_person_id_fkey" FOREIGN KEY ("person_id") REFERENCES "personen"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "individuele_stemming" ADD CONSTRAINT "individuele_stemming_fractie_id_fkey" FOREIGN KEY ("fractie_id") REFERENCES "fracties"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
