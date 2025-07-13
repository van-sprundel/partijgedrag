/*
  Warnings:

  - You are about to drop the `PasswordResetRequest` table. If the table is not empty, all the data it contains will be lost.
  - You are about to drop the `User` table. If the table is not empty, all the data it contains will be lost.
  - You are about to drop the `UserSession` table. If the table is not empty, all the data it contains will be lost.

*/
-- DropForeignKey
ALTER TABLE "PasswordResetRequest" DROP CONSTRAINT "PasswordResetRequest_userId_fkey";

-- DropForeignKey
ALTER TABLE "UserSession" DROP CONSTRAINT "UserSession_userId_fkey";

-- DropTable
DROP TABLE "PasswordResetRequest";

-- DropTable
DROP TABLE "User";

-- DropTable
DROP TABLE "UserSession";

-- CreateTable
CREATE TABLE "users" (
    "id" SERIAL NOT NULL,
    "email" TEXT NOT NULL,
    "password" TEXT NOT NULL
);

-- CreateTable
CREATE TABLE "user_sessions" (
    "id" SERIAL NOT NULL,
    "userId" INTEGER NOT NULL,
    "expiresAt" TIMESTAMPTZ(3) NOT NULL,
    "accessToken" TEXT NOT NULL,
    "refreshToken" TEXT NOT NULL,

    CONSTRAINT "user_sessions_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "password_reset_requests" (
    "id" SERIAL NOT NULL,
    "userId" INTEGER NOT NULL,
    "token" TEXT NOT NULL,
    "expiresAt" TIMESTAMPTZ(3) NOT NULL,

    CONSTRAINT "password_reset_requests_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "zaken" (
    "id" TEXT NOT NULL,
    "nummer" TEXT NOT NULL,
    "onderwerp" TEXT NOT NULL,
    "soort" TEXT NOT NULL,
    "titel" TEXT NOT NULL,
    "citeertitel" TEXT,
    "alias" TEXT,
    "status" TEXT NOT NULL,
    "datum" TIMESTAMP(3),
    "gestart_op" TIMESTAMP(3),
    "organisatie" TEXT NOT NULL,
    "grondslagvoorhang" TEXT,
    "termijn" TEXT,
    "vergaderjaar" TEXT NOT NULL,
    "volgnummer" INTEGER,
    "huidige_behandelstatus" TEXT,
    "afgedaan" BOOLEAN NOT NULL,
    "groot_project" BOOLEAN NOT NULL,
    "gewijzigd_op" TIMESTAMP(3),
    "api_gewijzigd_op" TIMESTAMP(3),
    "verwijderd" BOOLEAN NOT NULL,
    "kabinetsappreciatie" TEXT NOT NULL,
    "datum_afgedaan" TIMESTAMP(3),
    "kamer" TEXT NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "zaken_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "besluiten" (
    "id" TEXT NOT NULL,
    "agendapunt_id" TEXT,
    "stemmings_soort" TEXT,
    "besluit_soort" TEXT,
    "besluit_tekst" TEXT,
    "opmerking" TEXT,
    "status" TEXT,
    "agendapunt_zaak_besluit_volgorde" INTEGER,
    "gewijzigd_op" TIMESTAMP(3),
    "api_gewijzigd_op" TIMESTAMP(3),
    "verwijderd" BOOLEAN NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "besluiten_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "stemmingen" (
    "id" TEXT NOT NULL,
    "besluit_id" TEXT,
    "persoon_id" TEXT,
    "fractie_id" TEXT,
    "soort" TEXT,
    "fractie_grootte" INTEGER,
    "actor_naam" TEXT,
    "actor_fractie" TEXT,
    "vergissing" BOOLEAN NOT NULL,
    "sid_actor_lid" TEXT,
    "sid_actor_fractie" TEXT,
    "gewijzigd_op" TIMESTAMP(3),
    "api_gewijzigd_op" TIMESTAMP(3),
    "verwijderd" BOOLEAN NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "stemmingen_pkey" PRIMARY KEY ("id")
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
    "geboortedatum" TIMESTAMP(3),
    "geboorteplaats" TEXT,
    "geboorteland" TEXT,
    "overlijdensdatum" TIMESTAMP(3),
    "overlijdensplaats" TEXT,
    "overlijdensland" TEXT,
    "woonplaats" TEXT,
    "land" TEXT,
    "bijgewerkt" TIMESTAMP(3),
    "verwijderd" BOOLEAN NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "personen_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "fracties" (
    "id" TEXT NOT NULL,
    "nummer" INTEGER,
    "afkorting" TEXT,
    "naam_nl" TEXT,
    "naam_en" TEXT,
    "aantal_zetels" INTEGER,
    "aantal_stemmen" INTEGER,
    "datum_actief" TIMESTAMP(3),
    "datum_inactief" TIMESTAMP(3),
    "content_type" TEXT,
    "content_length" INTEGER,
    "gewijzigd_op" TIMESTAMP(3),
    "api_gewijzigd_op" TIMESTAMP(3),
    "verwijderd" BOOLEAN NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "fracties_pkey" PRIMARY KEY ("id")
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
    "bijgewerkt" TIMESTAMP(3),
    "verwijderd" BOOLEAN NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "zaak_actors_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "raw_odata" (
    "id" SERIAL NOT NULL,
    "type" TEXT NOT NULL,
    "entity_id" TEXT NOT NULL,
    "data" JSONB NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "raw_odata_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "voting_results" (
    "id" SERIAL NOT NULL,
    "document_id" TEXT NOT NULL,
    "besluit_id" TEXT NOT NULL,
    "besluit_tekst" TEXT NOT NULL,
    "voting_type" TEXT NOT NULL,
    "party_votes" JSONB NOT NULL,
    "date" TIMESTAMP(3),
    "status" TEXT NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "voting_results_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "individuele_stemming" (
    "id" SERIAL NOT NULL,
    "besluit_id" TEXT NOT NULL,
    "person_id" TEXT NOT NULL,
    "person_name" TEXT NOT NULL,
    "fractie_id" TEXT NOT NULL,
    "fractie_name" TEXT NOT NULL,
    "vote_type" TEXT NOT NULL,
    "is_correction" BOOLEAN NOT NULL,
    "date" TIMESTAMP(3),
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "individuele_stemming_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "users_id_key" ON "users"("id");

-- CreateIndex
CREATE UNIQUE INDEX "user_sessions_userId_accessToken_key" ON "user_sessions"("userId", "accessToken");

-- CreateIndex
CREATE UNIQUE INDEX "password_reset_requests_userId_key" ON "password_reset_requests"("userId");

-- AddForeignKey
ALTER TABLE "user_sessions" ADD CONSTRAINT "user_sessions_userId_fkey" FOREIGN KEY ("userId") REFERENCES "users"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "password_reset_requests" ADD CONSTRAINT "password_reset_requests_userId_fkey" FOREIGN KEY ("userId") REFERENCES "users"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
