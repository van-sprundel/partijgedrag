-- AddForeignKey
ALTER TABLE "besluiten" ADD CONSTRAINT "besluiten_zaak_id_fkey" FOREIGN KEY ("zaak_id") REFERENCES "zaken"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "kamerstukdossiers" ADD CONSTRAINT "kamerstukdossiers_zaak_id_fkey" FOREIGN KEY ("zaak_id") REFERENCES "zaken"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "stemmingen" ADD CONSTRAINT "stemmingen_besluit_id_fkey" FOREIGN KEY ("besluit_id") REFERENCES "besluiten"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "stemmingen" ADD CONSTRAINT "stemmingen_persoon_id_fkey" FOREIGN KEY ("persoon_id") REFERENCES "personen"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "stemmingen" ADD CONSTRAINT "stemmingen_fractie_id_fkey" FOREIGN KEY ("fractie_id") REFERENCES "fracties"("id") ON DELETE SET NULL ON UPDATE CASCADE;
