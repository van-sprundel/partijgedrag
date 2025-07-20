/*
  Warnings:

  - Made the column `fractie_id` on table `stemmingen` required. This step will fail if there are existing NULL values in that column.
  - Made the column `zaak_id` on table `zaak_actors` required. This step will fail if there are existing NULL values in that column.
  - Made the column `persoon_id` on table `zaak_actors` required. This step will fail if there are existing NULL values in that column.
  - Made the column `fractie_id` on table `zaak_actors` required. This step will fail if there are existing NULL values in that column.

*/
-- DropForeignKey
ALTER TABLE "stemmingen" DROP CONSTRAINT "stemmingen_fractie_id_fkey";

-- DropForeignKey
ALTER TABLE "zaak_actors" DROP CONSTRAINT "zaak_actors_fractie_id_fkey";

-- DropForeignKey
ALTER TABLE "zaak_actors" DROP CONSTRAINT "zaak_actors_persoon_id_fkey";

-- DropForeignKey
ALTER TABLE "zaak_actors" DROP CONSTRAINT "zaak_actors_zaak_id_fkey";

-- AlterTable
ALTER TABLE "stemmingen" ALTER COLUMN "fractie_id" SET NOT NULL;

-- AlterTable
ALTER TABLE "zaak_actors" ALTER COLUMN "zaak_id" SET NOT NULL,
ALTER COLUMN "persoon_id" SET NOT NULL,
ALTER COLUMN "fractie_id" SET NOT NULL;

-- AddForeignKey
ALTER TABLE "stemmingen" ADD CONSTRAINT "stemmingen_fractie_id_fkey" FOREIGN KEY ("fractie_id") REFERENCES "fracties"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "zaak_actors" ADD CONSTRAINT "zaak_actors_zaak_id_fkey" FOREIGN KEY ("zaak_id") REFERENCES "zaken"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "zaak_actors" ADD CONSTRAINT "zaak_actors_persoon_id_fkey" FOREIGN KEY ("persoon_id") REFERENCES "personen"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "zaak_actors" ADD CONSTRAINT "zaak_actors_fractie_id_fkey" FOREIGN KEY ("fractie_id") REFERENCES "fracties"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
