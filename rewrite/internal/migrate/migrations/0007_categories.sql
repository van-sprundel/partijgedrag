CREATE TABLE IF NOT EXISTS categories (
  category_key text PRIMARY KEY,
  jurisdiction_key text NOT NULL REFERENCES jurisdictions(jurisdiction_key),
  name text NOT NULL,
  kind text NOT NULL CHECK (kind IN ('generic', 'hot_topic')),
  keywords text[] NOT NULL DEFAULT '{}'::text[],
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (jurisdiction_key, name)
);

CREATE TABLE IF NOT EXISTS motion_categories (
  motion_key text NOT NULL REFERENCES motions(motion_key),
  category_key text NOT NULL REFERENCES categories(category_key),
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (motion_key, category_key)
);

CREATE INDEX IF NOT EXISTS motion_categories_category_idx
  ON motion_categories (category_key, motion_key);

ALTER TABLE motions
ADD COLUMN IF NOT EXISTS categorized_at timestamptz;

CREATE INDEX IF NOT EXISTS motions_categorize_backlog_idx
  ON motions (jurisdiction_key)
  WHERE categorized_at IS NULL AND source_deleted = false;

INSERT INTO categories (category_key, jurisdiction_key, name, kind, keywords)
VALUES
  ('bestuur', 'nl-tweede-kamer', 'Bestuur', 'generic', ARRAY['bestuur', 'governance', 'regering', 'kabinet', 'minister', 'staatssecretaris', 'overheid', 'gemeente', 'provincie']),
  ('cultuur-en-recreatie', 'nl-tweede-kamer', 'Cultuur en recreatie', 'generic', ARRAY['cultuur', 'kunst', 'sport', 'recreatie', 'museum', 'theater', 'bibliotheek', 'media', 'erfgoed']),
  ('economie', 'nl-tweede-kamer', 'Economie', 'generic', ARRAY['economie', 'economisch', 'handel', 'industrie', 'ondernemerschap', 'mkb', 'bedrijven', 'concurrentie', 'marktwerking']),
  ('financien', 'nl-tweede-kamer', 'Financiën', 'generic', ARRAY['financiën', 'belasting', 'btw', 'budget', 'begroting', 'schuld', 'deficit', 'lastenverlichting', 'koopkracht']),
  ('huisvesting', 'nl-tweede-kamer', 'Huisvesting', 'generic', ARRAY['wonen', 'huur', 'woningbouw', 'huisvesting', 'hypotheek', 'woningnood', 'woningmarkt', 'verhuurder', 'huurder', 'leegstand']),
  ('internationaal', 'nl-tweede-kamer', 'Internationaal', 'generic', ARRAY['internationaal', 'europa', 'eu', 'buitenland', 'ontwikkelingssamenwerking', 'defensie', 'verdrag', 'mensenrechten']),
  ('landbouw', 'nl-tweede-kamer', 'Landbouw', 'generic', ARRAY['landbouw', 'boer', 'vee', 'gewas', 'agrarisch', 'voedsel', 'mest', 'visserij', 'pesticiden', 'stikstof']),
  ('migratie-en-integratie', 'nl-tweede-kamer', 'Migratie en integratie', 'generic', ARRAY['migratie', 'integratie', 'vluchteling', 'asiel', 'immigrant', 'inburgering', 'statushouder']),
  ('natuur-en-milieu', 'nl-tweede-kamer', 'Natuur en milieu', 'generic', ARRAY['natuur', 'milieu', 'klimaat', 'co2', 'duurzaam', 'energie', 'vervuiling', 'biodiversiteit', 'opwarming', 'uitstoot', 'windmolens', 'zonne-energie']),
  ('onderwijs-en-wetenschap', 'nl-tweede-kamer', 'Onderwijs en wetenschap', 'generic', ARRAY['onderwijs', 'school', 'universiteit', 'student', 'wetenschap', 'onderzoek', 'educatie', 'leraar', 'onderwijskwaliteit']),
  ('openbare-orde-en-veiligheid', 'nl-tweede-kamer', 'Openbare orde en veiligheid', 'generic', ARRAY['veiligheid', 'politie', 'criminaliteit', 'terrorisme', 'orde', 'handhaving', 'brandweer', 'hulpdiensten', 'rampenbestrijding']),
  ('recht', 'nl-tweede-kamer', 'Recht', 'generic', ARRAY['recht', 'rechtspraak', 'rechter', 'wet', 'juridisch', 'justitie', 'advocaat', 'privacy', 'discriminatie', 'grondwet']),
  ('ruimte-en-infrastructuur', 'nl-tweede-kamer', 'Ruimte en infrastructuur', 'generic', ARRAY['infrastructuur', 'weg', 'spoor', 'bouw', 'ruimtelijk', 'planning', 'transport', 'luchtvaart', 'schiphol', 'openbare ruimte']),
  ('sociale-zekerheid', 'nl-tweede-kamer', 'Sociale zekerheid', 'generic', ARRAY['sociale zekerheid', 'uitkering', 'aow', 'wajong', 'bijstand', 'pensioen', 'armoede', 'schulden', 'mantelzorg']),
  ('verkeer', 'nl-tweede-kamer', 'Verkeer', 'generic', ARRAY['verkeer', 'auto', 'fiets', 'openbaar vervoer', 'trein', 'bus', 'file', 'verkeersveiligheid', 'mobiliteit']),
  ('werk', 'nl-tweede-kamer', 'Werk', 'generic', ARRAY['werk', 'werkgelegenheid', 'baan', 'arbeidsmarkt', 'cao', 'vakbond', 'werknemer', 'zzp''er', 'flexwerk', 'thuiswerken']),
  ('zorg-en-gezondheid', 'nl-tweede-kamer', 'Zorg en gezondheid', 'generic', ARRAY['zorg', 'gezondheid', 'medisch', 'ziekenhuis', 'dokter', 'medicijn', 'patiënt', 'preventie', 'jeugdzorg', 'ouderenzorg', 'ggz', 'vaccinatie']),
  ('immigratie', 'nl-tweede-kamer', 'Immigratie', 'hot_topic', ARRAY['immigratie', 'migratie', 'asielzoeker', 'vluchtelingen', 'grenzen', 'azc', 'asielbeleid', 'gezinshereniging']),
  ('oorlog', 'nl-tweede-kamer', 'Oorlog', 'hot_topic', ARRAY['oorlog', 'conflict', 'militair', 'defensie', 'wapen', 'vrede', 'oekraïne', 'rusland', 'gaza', 'israel', 'palestina', 'humanitaire hulp', 'wederopbouw']),
  ('klimaatverandering', 'nl-tweede-kamer', 'Klimaatverandering', 'hot_topic', ARRAY['klimaatverandering', 'opwarming', 'broeikas', 'klimaat', 'duurzaamheid', 'energietransitie', 'klimaatdoelen']),
  ('woningcrisis', 'nl-tweede-kamer', 'Woningcrisis', 'hot_topic', ARRAY['woningcrisis', 'woningtekort', 'betaalbaar wonen', 'huurprijzen', 'sociale huur', 'starters']),
  ('inflatie', 'nl-tweede-kamer', 'Inflatie', 'hot_topic', ARRAY['inflatie', 'prijsstijging', 'koopkracht', 'duurte', 'energieprijzen', 'boodschappen'])
ON CONFLICT (category_key) DO UPDATE
SET name = EXCLUDED.name,
    kind = EXCLUDED.kind,
    keywords = EXCLUDED.keywords,
    updated_at = now();
