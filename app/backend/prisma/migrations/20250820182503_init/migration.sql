-- CreateTable
CREATE TABLE "besluiten" (
    "id" TEXT NOT NULL,
    "agendapunt_id" TEXT,
    "zaak_id" TEXT,
    "stemmings_soort" TEXT,
    "besluit_soort" TEXT,
    "besluit_tekst" TEXT,
    "opmerking" TEXT,
    "status" TEXT,
    "agendapunt_zaak_besluit_volgorde" BIGINT,
    "gewijzigd_op" TIMESTAMPTZ(6),
    "api_gewijzigd_op" TIMESTAMPTZ(6),
    "verwijderd" BOOLEAN,

    CONSTRAINT "besluiten_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "document_info" (
    "id" BIGSERIAL NOT NULL,
    "dossier_nummer" TEXT,
    "volgnummer" BIGINT,
    "url" TEXT,
    "content" JSONB,
    "fetched_at" TIMESTAMPTZ(6),
    "success" BOOLEAN,
    "error" TEXT,

    CONSTRAINT "document_info_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "fracties" (
    "id" TEXT NOT NULL,
    "nummer" BIGINT,
    "afkorting" TEXT,
    "naam_nl" TEXT,
    "naam_en" TEXT,
    "aantal_zetels" BIGINT,
    "aantal_stemmen" BIGINT,
    "datum_actief" TIMESTAMPTZ(6),
    "datum_inactief" TIMESTAMPTZ(6),
    "content_type" TEXT,
    "content_length" BIGINT,
    "gewijzigd_op" TIMESTAMPTZ(6),
    "api_gewijzigd_op" TIMESTAMPTZ(6),
    "verwijderd" BOOLEAN,

    CONSTRAINT "fracties_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "kamerstukdossiers" (
    "id" TEXT NOT NULL,
    "zaak_id" TEXT,
    "nummer" TEXT,
    "titel" TEXT,
    "citeertitel" TEXT,
    "alias" TEXT,
    "toevoeging" TEXT,
    "hoogste_volgnummer" BIGINT,
    "afgesloten" BOOLEAN,
    "datum_aangemaakt" TIMESTAMPTZ(6),
    "datum_gesloten" TIMESTAMPTZ(6),
    "kamer" TEXT,
    "bijgewerkt" TIMESTAMPTZ(6),
    "api_gewijzigd_op" TIMESTAMPTZ(6),
    "verwijderd" BOOLEAN,
    "bullet_points" JSONB,

    CONSTRAINT "kamerstukdossiers_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "personen" (
    "id" TEXT NOT NULL,
    "titels" TEXT,
    "initialen" TEXT,
    "tussenvoegsel" TEXT,
    "achternaam" TEXT,
    "voornamen" TEXT,
    "roepnaam" TEXT,
    "geslacht" TEXT,
    "geboortedatum" TIMESTAMPTZ(6),
    "geboorteplaats" TEXT,
    "geboorteland" TEXT,
    "overlijdensdatum" TIMESTAMPTZ(6),
    "overlijdensplaats" TEXT,
    "overlijdensland" TEXT,
    "woonplaats" TEXT,
    "land" TEXT,
    "bijgewerkt" TIMESTAMPTZ(6),
    "verwijderd" BOOLEAN,

    CONSTRAINT "personen_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "stemmingen" (
    "id" TEXT NOT NULL,
    "besluit_id_raw" TEXT,
    "besluit_id" TEXT,
    "soort" TEXT,
    "fractie_grootte" BIGINT,
    "actor_naam" TEXT,
    "actor_fractie" TEXT,
    "vergissing" BOOLEAN,
    "sid_actor_lid" TEXT,
    "sid_actor_fractie" TEXT,
    "persoon_id_raw" TEXT,
    "persoon_id" TEXT,
    "fractie_id_raw" TEXT,
    "fractie_id" TEXT,
    "gewijzigd_op" TIMESTAMPTZ(6),
    "api_gewijzigd_op" TIMESTAMPTZ(6),
    "verwijderd" BOOLEAN,

    CONSTRAINT "stemmingen_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "zaak_actors" (
    "id" TEXT NOT NULL,
    "zaak_id" TEXT,
    "persoon_id" TEXT,
    "fractie_id" TEXT,
    "relatie" TEXT,
    "actor_naam" TEXT,
    "actor_fractie" TEXT,
    "bijgewerkt" TIMESTAMPTZ(6),
    "verwijderd" BOOLEAN,

    CONSTRAINT "zaak_actors_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "zaak_documents" (
    "id" BIGSERIAL NOT NULL,
    "zaak_id" TEXT,
    "document_id" BIGINT,
    "created_at" TIMESTAMPTZ(6),

    CONSTRAINT "zaak_documents_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "zaken" (
    "id" TEXT NOT NULL,
    "nummer" TEXT,
    "onderwerp" TEXT,
    "soort" TEXT,
    "titel" TEXT,
    "citeertitel" TEXT,
    "alias" TEXT,
    "status" TEXT,
    "datum" TIMESTAMPTZ(6),
    "gestart_op" TIMESTAMPTZ(6),
    "organisatie" TEXT,
    "grondslagvoorhang" TEXT,
    "termijn" TEXT,
    "vergaderjaar" TEXT,
    "volgnummer" BIGINT,
    "huidige_behandelstatus" TEXT,
    "afgedaan" BOOLEAN,
    "groot_project" BOOLEAN,
    "gewijzigd_op" TIMESTAMPTZ(6),
    "api_gewijzigd_op" TIMESTAMPTZ(6),
    "verwijderd" BOOLEAN,
    "kabinetsappreciatie" TEXT,
    "datum_afgedaan" TIMESTAMPTZ(6),
    "kamer" TEXT,

    CONSTRAINT "zaken_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE INDEX "idx_besluiten_zaak_id" ON "besluiten"("zaak_id");

-- CreateIndex
CREATE INDEX "idx_document_info_dossier_nummer" ON "document_info"("dossier_nummer");

-- CreateIndex
CREATE UNIQUE INDEX "idx_document_unique" ON "document_info"("dossier_nummer", "volgnummer");

-- CreateIndex
CREATE INDEX "idx_kamerstukdossiers_zaak_id" ON "kamerstukdossiers"("zaak_id");

-- CreateIndex
CREATE INDEX "idx_stemmingen_besluit_id" ON "stemmingen"("besluit_id");

-- CreateIndex
CREATE INDEX "idx_stemmingen_fractie_id" ON "stemmingen"("fractie_id");

-- CreateIndex
CREATE INDEX "idx_stemmingen_persoon_id" ON "stemmingen"("persoon_id");

-- CreateIndex
CREATE INDEX "idx_zaak_actors_fractie_id" ON "zaak_actors"("fractie_id");

-- CreateIndex
CREATE INDEX "idx_zaak_actors_persoon_id" ON "zaak_actors"("persoon_id");

-- CreateIndex
CREATE INDEX "idx_zaak_actors_zaak_id" ON "zaak_actors"("zaak_id");

-- CreateIndex
CREATE INDEX "idx_zaak_documents_document_id" ON "zaak_documents"("document_id");

-- CreateIndex
CREATE INDEX "idx_zaak_documents_zaak_id" ON "zaak_documents"("zaak_id");

-- CreateIndex
CREATE UNIQUE INDEX "idx_zaak_document" ON "zaak_documents"("zaak_id", "document_id");
