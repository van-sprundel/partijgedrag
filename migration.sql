-- CreateTable
CREATE TABLE "public"."user_sessions" (
    "id" TEXT NOT NULL,
    "answers" JSONB NOT NULL,
    "results" JSONB,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "user_sessions_pkey" PRIMARY KEY ("id")
);

-- AddForeignKey
ALTER TABLE "public"."besluiten" ADD CONSTRAINT "besluiten_zaak_id_fkey" FOREIGN KEY ("zaak_id") REFERENCES "public"."zaken"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "public"."kamerstukdossiers" ADD CONSTRAINT "kamerstukdossiers_zaak_id_fkey" FOREIGN KEY ("zaak_id") REFERENCES "public"."zaken"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "public"."stemmingen" ADD CONSTRAINT "stemmingen_besluit_id_fkey" FOREIGN KEY ("besluit_id") REFERENCES "public"."besluiten"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "public"."stemmingen" ADD CONSTRAINT "stemmingen_persoon_id_fkey" FOREIGN KEY ("persoon_id") REFERENCES "public"."personen"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "public"."stemmingen" ADD CONSTRAINT "stemmingen_fractie_id_fkey" FOREIGN KEY ("fractie_id") REFERENCES "public"."fracties"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "public"."zaak_actors" ADD CONSTRAINT "zaak_actors_zaak_id_fkey" FOREIGN KEY ("zaak_id") REFERENCES "public"."zaken"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "public"."zaak_actors" ADD CONSTRAINT "zaak_actors_persoon_id_fkey" FOREIGN KEY ("persoon_id") REFERENCES "public"."personen"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "public"."zaak_actors" ADD CONSTRAINT "zaak_actors_fractie_id_fkey" FOREIGN KEY ("fractie_id") REFERENCES "public"."fracties"("id") ON DELETE SET NULL ON UPDATE CASCADE;

