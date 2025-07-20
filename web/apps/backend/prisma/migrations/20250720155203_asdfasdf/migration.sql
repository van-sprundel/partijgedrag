-- DropForeignKey
ALTER TABLE "individuele_stemming" DROP CONSTRAINT "individuele_stemming_fractie_id_fkey";

-- AlterTable
ALTER TABLE "individuele_stemming" ALTER COLUMN "fractie_id" DROP NOT NULL;

-- AddForeignKey
ALTER TABLE "individuele_stemming" ADD CONSTRAINT "individuele_stemming_fractie_id_fkey" FOREIGN KEY ("fractie_id") REFERENCES "fracties"("id") ON DELETE SET NULL ON UPDATE CASCADE;
