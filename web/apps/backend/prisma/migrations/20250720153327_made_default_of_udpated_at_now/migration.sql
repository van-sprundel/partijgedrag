/*
  Warnings:

  - Made the column `updated_at` on table `besluiten` required. This step will fail if there are existing NULL values in that column.
  - Made the column `updated_at` on table `zaken` required. This step will fail if there are existing NULL values in that column.

*/
-- AlterTable
ALTER TABLE "besluiten" ALTER COLUMN "updated_at" SET NOT NULL,
ALTER COLUMN "updated_at" SET DEFAULT CURRENT_TIMESTAMP;

-- AlterTable
ALTER TABLE "fracties" ALTER COLUMN "updated_at" SET DEFAULT CURRENT_TIMESTAMP;

-- AlterTable
ALTER TABLE "individuele_stemming" ALTER COLUMN "updated_at" SET DEFAULT CURRENT_TIMESTAMP;

-- AlterTable
ALTER TABLE "stemmingen" ALTER COLUMN "updated_at" SET DEFAULT CURRENT_TIMESTAMP;

-- AlterTable
ALTER TABLE "voting_results" ALTER COLUMN "updated_at" SET DEFAULT CURRENT_TIMESTAMP;

-- AlterTable
ALTER TABLE "zaak_actors" ALTER COLUMN "updated_at" SET DEFAULT CURRENT_TIMESTAMP;

-- AlterTable
ALTER TABLE "zaken" ALTER COLUMN "updated_at" SET NOT NULL,
ALTER COLUMN "updated_at" SET DEFAULT CURRENT_TIMESTAMP;
