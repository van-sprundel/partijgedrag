-- AddColumn: Add stemming_datum (voting date) to besluiten (decisions) table
ALTER TABLE "besluiten" ADD COLUMN "stemming_datum" TIMESTAMPTZ(6);

-- CreateIndex: Add index on stemming_datum for better query performance
CREATE INDEX "idx_besluiten_stemming_datum" ON "besluiten"("stemming_datum");

-- Comment: This field stores the actual voting date extracted from Agendapunt->Activiteit
COMMENT ON COLUMN "besluiten"."stemming_datum" IS 'Actual voting date extracted from Agendapunt->Activiteit.Datum';
