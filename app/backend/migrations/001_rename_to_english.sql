-- Migration: Rename all tables and columns from Dutch to English
-- Date: 2025-10-27

-- ================================================
-- RENAME TABLES
-- ================================================

ALTER TABLE "besluiten" RENAME TO "decisions";
ALTER TABLE "fracties" RENAME TO "parties";
ALTER TABLE "kamerstukdossiers" RENAME TO "parliamentary_documents";
ALTER TABLE "zaak_kamerstukdossiers" RENAME TO "case_parliamentary_documents";
ALTER TABLE "personen" RENAME TO "politicians";
ALTER TABLE "stemmingen" RENAME TO "votes";
ALTER TABLE "zaak_actors" RENAME TO "case_actors";
ALTER TABLE "zaken" RENAME TO "cases";
ALTER TABLE "zaak_categories" RENAME TO "case_categories";

-- ================================================
-- RENAME COLUMNS - decisions (formerly besluiten)
-- ================================================

ALTER TABLE "decisions" RENAME COLUMN "agendapunt_id" TO "agenda_point_id";
ALTER TABLE "decisions" RENAME COLUMN "zaak_id" TO "case_id";
ALTER TABLE "decisions" RENAME COLUMN "stemmings_soort" TO "voting_type";
ALTER TABLE "decisions" RENAME COLUMN "besluit_soort" TO "decision_type";
ALTER TABLE "decisions" RENAME COLUMN "besluit_tekst" TO "decision_text";
ALTER TABLE "decisions" RENAME COLUMN "opmerking" TO "comment";
ALTER TABLE "decisions" RENAME COLUMN "agendapunt_zaak_besluit_volgorde" TO "agenda_point_case_decision_order";
ALTER TABLE "decisions" RENAME COLUMN "gewijzigd_op" TO "updated_at";
ALTER TABLE "decisions" RENAME COLUMN "api_gewijzigd_op" TO "api_updated_at";

-- ================================================
-- RENAME COLUMNS - parties (formerly fracties)
-- ================================================

ALTER TABLE "parties" RENAME COLUMN "nummer" TO "number";
ALTER TABLE "parties" RENAME COLUMN "afkorting" TO "short_name";
ALTER TABLE "parties" RENAME COLUMN "aantal_zetels" TO "seats";
ALTER TABLE "parties" RENAME COLUMN "aantal_stemmen" TO "votes_count";
ALTER TABLE "parties" RENAME COLUMN "datum_actief" TO "active_from";
ALTER TABLE "parties" RENAME COLUMN "datum_inactief" TO "active_to";
ALTER TABLE "parties" RENAME COLUMN "gewijzigd_op" TO "updated_at";
ALTER TABLE "parties" RENAME COLUMN "api_gewijzigd_op" TO "api_updated_at";
ALTER TABLE "parties" RENAME COLUMN "verwijderd" TO "removed";

-- ================================================
-- RENAME COLUMNS - parliamentary_documents (formerly kamerstukdossiers)
-- ================================================

ALTER TABLE "parliamentary_documents" RENAME COLUMN "nummer" TO "number";
ALTER TABLE "parliamentary_documents" RENAME COLUMN "titel" TO "title";
ALTER TABLE "parliamentary_documents" RENAME COLUMN "citeertitel" TO "citation_title";
ALTER TABLE "parliamentary_documents" RENAME COLUMN "toevoeging" TO "addition";
ALTER TABLE "parliamentary_documents" RENAME COLUMN "hoogste_volgnummer" TO "highest_sequence_number";
ALTER TABLE "parliamentary_documents" RENAME COLUMN "afgesloten" TO "closed";
ALTER TABLE "parliamentary_documents" RENAME COLUMN "datum_aangemaakt" TO "created_at";
ALTER TABLE "parliamentary_documents" RENAME COLUMN "datum_gesloten" TO "closed_at";
ALTER TABLE "parliamentary_documents" RENAME COLUMN "kamer" TO "chamber";
ALTER TABLE "parliamentary_documents" RENAME COLUMN "bijgewerkt" TO "updated_at";
ALTER TABLE "parliamentary_documents" RENAME COLUMN "api_gewijzigd_op" TO "api_updated_at";

-- ================================================
-- RENAME COLUMNS - case_parliamentary_documents (formerly zaak_kamerstukdossiers)
-- ================================================

ALTER TABLE "case_parliamentary_documents" RENAME COLUMN "kamerstukdossier_id" TO "parliamentary_document_id";
ALTER TABLE "case_parliamentary_documents" RENAME COLUMN "zaak_id" TO "case_id";

-- ================================================
-- RENAME COLUMNS - politicians (formerly personen)
-- ================================================

ALTER TABLE "politicians" RENAME COLUMN "titels" TO "titles";
ALTER TABLE "politicians" RENAME COLUMN "initialen" TO "initials";
ALTER TABLE "politicians" RENAME COLUMN "tussenvoegsel" TO "prefix";
ALTER TABLE "politicians" RENAME COLUMN "achternaam" TO "last_name";
ALTER TABLE "politicians" RENAME COLUMN "voornamen" TO "first_names";
ALTER TABLE "politicians" RENAME COLUMN "roepnaam" TO "nickname";
ALTER TABLE "politicians" RENAME COLUMN "geslacht" TO "gender";
ALTER TABLE "politicians" RENAME COLUMN "geboortedatum" TO "birth_date";
ALTER TABLE "politicians" RENAME COLUMN "geboorteplaats" TO "birth_place";
ALTER TABLE "politicians" RENAME COLUMN "geboorteland" TO "birth_country";
ALTER TABLE "politicians" RENAME COLUMN "overlijdensdatum" TO "death_date";
ALTER TABLE "politicians" RENAME COLUMN "overlijdensplaats" TO "death_place";
ALTER TABLE "politicians" RENAME COLUMN "overlijdensland" TO "death_country";
ALTER TABLE "politicians" RENAME COLUMN "woonplaats" TO "residence";
ALTER TABLE "politicians" RENAME COLUMN "land" TO "country";
ALTER TABLE "politicians" RENAME COLUMN "bijgewerkt" TO "updated_at";

-- ================================================
-- RENAME COLUMNS - votes (formerly stemmingen)
-- ================================================

ALTER TABLE "votes" RENAME COLUMN "besluit_id_raw" TO "decision_id_raw";
ALTER TABLE "votes" RENAME COLUMN "besluit_id" TO "decision_id";
ALTER TABLE "votes" RENAME COLUMN "soort" TO "type";
ALTER TABLE "votes" RENAME COLUMN "fractie_grootte" TO "party_size";
ALTER TABLE "votes" RENAME COLUMN "actor_naam" TO "actor_name";
ALTER TABLE "votes" RENAME COLUMN "actor_fractie" TO "actor_party";
ALTER TABLE "votes" RENAME COLUMN "vergissing" TO "mistake";
ALTER TABLE "votes" RENAME COLUMN "sid_actor_lid" TO "sid_actor_member";
ALTER TABLE "votes" RENAME COLUMN "sid_actor_fractie" TO "sid_actor_party";
ALTER TABLE "votes" RENAME COLUMN "persoon_id_raw" TO "politician_id_raw";
ALTER TABLE "votes" RENAME COLUMN "persoon_id" TO "politician_id";
ALTER TABLE "votes" RENAME COLUMN "fractie_id_raw" TO "party_id_raw";
ALTER TABLE "votes" RENAME COLUMN "fractie_id" TO "party_id";
ALTER TABLE "votes" RENAME COLUMN "gewijzigd_op" TO "updated_at";
ALTER TABLE "votes" RENAME COLUMN "api_gewijzigd_op" TO "api_updated_at";

-- ================================================
-- RENAME COLUMNS - case_actors (formerly zaak_actors)
-- ================================================

ALTER TABLE "case_actors" RENAME COLUMN "zaak_id" TO "case_id";
ALTER TABLE "case_actors" RENAME COLUMN "persoon_id" TO "politician_id";
ALTER TABLE "case_actors" RENAME COLUMN "fractie_id" TO "party_id";
ALTER TABLE "case_actors" RENAME COLUMN "relatie" TO "relation";
ALTER TABLE "case_actors" RENAME COLUMN "actor_naam" TO "actor_name";
ALTER TABLE "case_actors" RENAME COLUMN "actor_fractie" TO "actor_party";
ALTER TABLE "case_actors" RENAME COLUMN "bijgewerkt" TO "updated_at";

-- ================================================
-- RENAME COLUMNS - cases (formerly zaken)
-- ================================================

ALTER TABLE "cases" RENAME COLUMN "nummer" TO "number";
ALTER TABLE "cases" RENAME COLUMN "onderwerp" TO "subject";
ALTER TABLE "cases" RENAME COLUMN "soort" TO "type";
ALTER TABLE "cases" RENAME COLUMN "titel" TO "title";
ALTER TABLE "cases" RENAME COLUMN "citeertitel" TO "citation_title";
ALTER TABLE "cases" RENAME COLUMN "datum" TO "date";
ALTER TABLE "cases" RENAME COLUMN "gestart_op" TO "started_at";
ALTER TABLE "cases" RENAME COLUMN "organisatie" TO "organization";
ALTER TABLE "cases" RENAME COLUMN "grondslagvoorhang" TO "basis";
ALTER TABLE "cases" RENAME COLUMN "termijn" TO "term";
ALTER TABLE "cases" RENAME COLUMN "vergaderjaar" TO "meeting_year";
ALTER TABLE "cases" RENAME COLUMN "volgnummer" TO "sequence_number";
ALTER TABLE "cases" RENAME COLUMN "huidige_behandelstatus" TO "current_treatment_status";
ALTER TABLE "cases" RENAME COLUMN "afgedaan" TO "completed";
ALTER TABLE "cases" RENAME COLUMN "groot_project" TO "large_project";
ALTER TABLE "cases" RENAME COLUMN "gewijzigd_op" TO "updated_at";
ALTER TABLE "cases" RENAME COLUMN "api_gewijzigd_op" TO "api_updated_at";
ALTER TABLE "cases" RENAME COLUMN "verwijderd" TO "removed";
ALTER TABLE "cases" RENAME COLUMN "kabinetsappreciatie" TO "cabinet_appreciation";
ALTER TABLE "cases" RENAME COLUMN "datum_afgedaan" TO "completed_at";
ALTER TABLE "cases" RENAME COLUMN "kamer" TO "chamber";

-- ================================================
-- RENAME COLUMNS - case_categories (formerly zaak_categories)
-- ================================================

ALTER TABLE "case_categories" RENAME COLUMN "zaak_id" TO "case_id";

-- ================================================
-- RENAME INDEXES
-- ================================================

ALTER INDEX "idx_besluiten_zaak_id" RENAME TO "idx_decisions_case_id";
ALTER INDEX "idx_stemmingen_besluit_id" RENAME TO "idx_votes_decision_id";
ALTER INDEX "idx_stemmingen_fractie_id" RENAME TO "idx_votes_party_id";
ALTER INDEX "idx_stemmingen_persoon_id" RENAME TO "idx_votes_politician_id";
ALTER INDEX "idx_zaak_actors_fractie_id" RENAME TO "idx_case_actors_party_id";
ALTER INDEX "idx_zaak_actors_persoon_id" RENAME TO "idx_case_actors_politician_id";
ALTER INDEX "idx_zaak_actors_zaak_id" RENAME TO "idx_case_actors_case_id";

-- ================================================
-- RENAME FOREIGN KEY CONSTRAINTS
-- ================================================

ALTER TABLE "case_parliamentary_documents"
    DROP CONSTRAINT "fk_zaak_kamerstukdossiers_kamerstukdossier",
    ADD CONSTRAINT "fk_case_parliamentary_documents_document"
        FOREIGN KEY ("parliamentary_document_id") REFERENCES "parliamentary_documents"("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

ALTER TABLE "case_parliamentary_documents"
    DROP CONSTRAINT "fk_zaak_kamerstukdossiers_zaak",
    ADD CONSTRAINT "fk_case_parliamentary_documents_case"
        FOREIGN KEY ("case_id") REFERENCES "cases"("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

ALTER TABLE "case_categories"
    DROP CONSTRAINT "case_categories_zaak_id_fkey",
    ADD CONSTRAINT "case_categories_case_id_fkey"
        FOREIGN KEY ("case_id") REFERENCES "cases"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE "case_categories"
    DROP CONSTRAINT "case_categories_category_id_fkey",
    ADD CONSTRAINT "case_categories_category_id_fkey"
        FOREIGN KEY ("category_id") REFERENCES "motion_categories"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
