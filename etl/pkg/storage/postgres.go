package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"etl/internal/config"
	"etl/internal/models"
	"etl/pkg/odata"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type GormPostgresStorage struct {
	db *gorm.DB
}

func NewGormPostgresStorage(config config.StorageConfig) (*GormPostgresStorage, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		config.Host, config.Username, config.Password, config.Database, config.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.AutoMigrate(
		&models.Zaak{},
		&models.Besluit{},
		&models.Stemming{},
		&models.Persoon{},
		&models.Fractie{},
		&models.ZaakActor{},
		&models.RawOData{},
		&models.VotingResult{},
		&models.IndividueleStemming{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return &GormPostgresStorage{
		db: db,
	}, nil
}

func (s *GormPostgresStorage) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *GormPostgresStorage) SaveODataZaak(ctx context.Context, zaakData interface{}) error {
	odataZaak, err := s.convertToODataZaak(zaakData)
	if err != nil {
		return fmt.Errorf("failed to convert zaak data: %w", err)
	}

	if err := s.saveRawOData(ctx, "zaak", odataZaak.ID, zaakData); err != nil {
		log.Printf("Warning: failed to save raw zaak data: %v", err)
	}

	dbZaak := s.mapZaakToDB(odataZaak)

	// save the zaak with upsert
	if err := s.db.WithContext(ctx).Save(dbZaak).Error; err != nil {
		return fmt.Errorf("failed to save zaak: %w", err)
	}

	// save related ZaakActors
	for _, zaakActor := range odataZaak.ZaakActor {
		if err := s.saveZaakActor(ctx, odataZaak.ID, &zaakActor); err != nil {
			log.Printf("Warning: failed to save zaak actor: %v", err)
		}
	}

	return nil
}

func (s *GormPostgresStorage) SaveODataBesluit(ctx context.Context, besluitData interface{}) error {
	odataBesluit, err := s.convertToODataBesluit(besluitData)
	if err != nil {
		return fmt.Errorf("failed to convert besluit data: %w", err)
	}

	if err := s.saveRawOData(ctx, "besluit", odataBesluit.ID, besluitData); err != nil {
		log.Printf("Warning: failed to save raw besluit data: %v", err)
	}

	dbBesluit := s.mapBesluitToDB(odataBesluit)
	return s.db.WithContext(ctx).Save(dbBesluit).Error
}

func (s *GormPostgresStorage) SaveODataStemming(ctx context.Context, stemmingData interface{}) error {
	odataStemming, err := s.convertToODataStemming(stemmingData)
	if err != nil {
		return fmt.Errorf("failed to convert stemming data: %w", err)
	}

	if err := s.saveRawOData(ctx, "stemming", odataStemming.ID, stemmingData); err != nil {
		log.Printf("Warning: failed to save raw stemming data: %v", err)
	}

	if odataStemming.Persoon != nil {
		if err := s.savePersoon(ctx, odataStemming.Persoon); err != nil {
			log.Printf("Warning: failed to save persoon: %v", err)
		}
	}

	if odataStemming.Fractie != nil {
		if err := s.saveFractie(ctx, odataStemming.Fractie); err != nil {
			log.Printf("Warning: failed to save fractie: %v", err)
		}
	}

	dbStemming := s.mapStemmingToDB(odataStemming)
	return s.db.WithContext(ctx).Save(dbStemming).Error
}

func (s *GormPostgresStorage) SaveVotingResult(ctx context.Context, result interface{}) error {
	odataResult, err := s.convertToODataVotingResult(result)
	if err != nil {
		return fmt.Errorf("failed to convert voting result: %w", err)
	}

	dbResult := s.mapVotingResultToDB(odataResult)
	return s.db.WithContext(ctx).Save(dbResult).Error
}

func (s *GormPostgresStorage) SaveIndividueleStemming(ctx context.Context, vote interface{}) error {
	odataVote, err := s.convertToODataIndividueleStemming(vote)
	if err != nil {
		return fmt.Errorf("failed to convert individual vote: %w", err)
	}

	dbVote := s.mapIndividueleStemmingToDB(odataVote)
	return s.db.WithContext(ctx).Save(dbVote).Error
}

func (s *GormPostgresStorage) Ping(ctx context.Context) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (s *GormPostgresStorage) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	tables := map[string]interface{}{
		"zaken":            &models.Zaak{},
		"besluiten":        &models.Besluit{},
		"stemmingen":       &models.Stemming{},
		"personen":         &models.Persoon{},
		"fracties":         &models.Fractie{},
		"zaak_actors":      &models.ZaakActor{},
		"voting_results":   &models.VotingResult{},
		"individual_votes": &models.IndividueleStemming{},
	}

	for tableName, model := range tables {
		var count int64
		if err := s.db.WithContext(ctx).Model(model).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("failed to count %s: %w", tableName, err)
		}
		stats[tableName] = count
	}

	return stats, nil
}

func (s *GormPostgresStorage) saveRawOData(ctx context.Context, entityType, entityID string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	rawData := &models.RawOData{
		Type:     entityType,
		EntityID: entityID,
		Data:     string(jsonData),
	}

	return s.db.WithContext(ctx).Save(rawData).Error
}

func (s *GormPostgresStorage) savePersoon(ctx context.Context, persoon *odata.Persoon) error {
	if persoon == nil {
		return nil
	}

	dbPersoon := s.mapPersoonToDB(persoon)
	return s.db.WithContext(ctx).Save(dbPersoon).Error
}

func (s *GormPostgresStorage) saveFractie(ctx context.Context, fractie *odata.Fractie) error {
	if fractie == nil {
		return nil
	}

	dbFractie := s.mapFractieToDB(fractie)
	return s.db.WithContext(ctx).Save(dbFractie).Error
}

func (s *GormPostgresStorage) saveZaakActor(ctx context.Context, zaakID string, zaakActor *odata.ZaakActor) error {
	if zaakActor == nil {
		return nil
	}

	if zaakActor.Persoon != nil {
		if err := s.savePersoon(ctx, zaakActor.Persoon); err != nil {
			log.Printf("Warning: failed to save persoon for zaak actor: %v", err)
		}
	}

	if zaakActor.Fractie != nil {
		if err := s.saveFractie(ctx, zaakActor.Fractie); err != nil {
			log.Printf("Warning: failed to save fractie for zaak actor: %v", err)
		}
	}

	dbZaakActor := s.mapZaakActorToDB(zaakActor, zaakID)
	return s.db.WithContext(ctx).Save(dbZaakActor).Error
}

func (s *GormPostgresStorage) convertToODataZaak(data interface{}) (*odata.Zaak, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var zaak odata.Zaak
	if err := json.Unmarshal(jsonData, &zaak); err != nil {
		return nil, err
	}

	return &zaak, nil
}

func (s *GormPostgresStorage) convertToODataBesluit(data interface{}) (*odata.Besluit, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var besluit odata.Besluit
	if err := json.Unmarshal(jsonData, &besluit); err != nil {
		return nil, err
	}

	return &besluit, nil
}

func (s *GormPostgresStorage) convertToODataStemming(data interface{}) (*odata.Stemming, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var stemming odata.Stemming
	if err := json.Unmarshal(jsonData, &stemming); err != nil {
		return nil, err
	}

	return &stemming, nil
}

func (s *GormPostgresStorage) convertToODataVotingResult(data interface{}) (*odata.VotingResult, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var result odata.VotingResult
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *GormPostgresStorage) convertToODataIndividueleStemming(data interface{}) (*odata.IndividueleStemming, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var vote odata.IndividueleStemming
	if err := json.Unmarshal(jsonData, &vote); err != nil {
		return nil, err
	}

	return &vote, nil
}

func (s *GormPostgresStorage) convertToODataMotionSubmitter(data interface{}) (*odata.MotionSubmitter, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var submitter odata.MotionSubmitter
	if err := json.Unmarshal(jsonData, &submitter); err != nil {
		return nil, err
	}

	return &submitter, nil
}

func (s *GormPostgresStorage) mapZaakToDB(odata *odata.Zaak) *models.Zaak {
	if odata == nil {
		return nil
	}

	return &models.Zaak{
		ID:                    odata.ID,
		Nummer:                odata.Nummer,
		Onderwerp:             odata.Onderwerp,
		Soort:                 odata.Soort,
		Titel:                 odata.Titel,
		Citeertitel:           odata.Citeertitel,
		Alias:                 odata.Alias,
		Status:                odata.Status,
		Datum:                 s.mapCustomDate(odata.Datum),
		GestartOp:             &odata.GestartOp,
		Organisatie:           odata.Organisatie,
		Grondslagvoorhang:     odata.Grondslagvoorhang,
		Termijn:               odata.Termijn,
		Vergaderjaar:          odata.Vergaderjaar,
		Volgnummer:            &odata.Volgnummer,
		HuidigeBehandelstatus: odata.HuidigeBehandelstatus,
		Afgedaan:              odata.Afgedaan,
		GrootProject:          odata.GrootProject,
		GewijzigdOp:           &odata.GewijzigdOp,
		ApiGewijzigdOp:        &odata.ApiGewijzigdOp,
		Verwijderd:            odata.Verwijderd,
		Kabinetsappreciatie:   odata.Kabinetsappreciatie,
		DatumAfgedaan:         s.mapCustomDate(odata.DatumAfgedaan),
		Kamer:                 odata.Kamer,
	}
}

func (s *GormPostgresStorage) mapBesluitToDB(odata *odata.Besluit) *models.Besluit {
	if odata == nil {
		return nil
	}

	return &models.Besluit{
		ID:                            odata.ID,
		AgendapuntID:                  &odata.AgendapuntId,
		StemmingsSoort:                odata.StemmingsSoort,
		BesluitSoort:                  &odata.BesluitSoort,
		BesluitTekst:                  &odata.BesluitTekst,
		Opmerking:                     odata.Opmerking,
		Status:                        &odata.Status,
		AgendapuntZaakBesluitVolgorde: &odata.AgendapuntZaakBesluitVolgorde,
		GewijzigdOp:                   &odata.GewijzigdOp,
		ApiGewijzigdOp:                &odata.ApiGewijzigdOp,
		Verwijderd:                    odata.Verwijderd,
	}
}

func (s *GormPostgresStorage) mapStemmingToDB(odata *odata.Stemming) *models.Stemming {
	if odata == nil {
		return nil
	}

	var persoonID, fractieID *string

	if odata.PersoonId != nil {
		persoonID = odata.PersoonId
	}
	fractieID = &odata.FractieId

	return &models.Stemming{
		ID:              odata.ID,
		BesluitID:       &odata.BesluitId,
		PersoonID:       persoonID,
		FractieID:       fractieID,
		Soort:           &odata.Soort,
		FractieGrootte:  &odata.FractieGrootte,
		ActorNaam:       &odata.ActorNaam,
		ActorFractie:    &odata.ActorFractie,
		Vergissing:      odata.Vergissing,
		SidActorLid:     odata.SidActorLid,
		SidActorFractie: &odata.SidActorFractie,
		GewijzigdOp:     &odata.GewijzigdOp,
		ApiGewijzigdOp:  &odata.ApiGewijzigdOp,
		Verwijderd:      odata.Verwijderd,
	}
}

func (s *GormPostgresStorage) mapPersoonToDB(odata *odata.Persoon) *models.Persoon {
	if odata == nil {
		return nil
	}

	return &models.Persoon{
		ID:                odata.ID,
		Titels:            &odata.Titels,
		Initialen:         &odata.Initialen,
		Tussenvoegsel:     &odata.Tussenvoegsel,
		Achternaam:        &odata.Achternaam,
		Voornamen:         &odata.Voornamen,
		Roepnaam:          &odata.Roepnaam,
		Geslacht:          &odata.Geslacht,
		Geboortedatum:     s.mapCustomDate(odata.Geboortedatum),
		Geboorteplaats:    &odata.Geboorteplaats,
		Geboorteland:      &odata.Geboorteland,
		Overlijdensdatum:  s.mapCustomDate(odata.Overlijdensdatum),
		Overlijdensplaats: &odata.Overlijdensplaats,
		Overlijdensland:   &odata.Overlijdensland,
		Woonplaats:        &odata.Woonplaats,
		Land:              &odata.Land,
		Bijgewerkt:        &odata.Bijgewerkt,
		Verwijderd:        odata.Verwijderd,
	}
}

func (s *GormPostgresStorage) mapFractieToDB(odata *odata.Fractie) *models.Fractie {
	if odata == nil {
		return nil
	}

	return &models.Fractie{
		ID:             odata.ID,
		Nummer:         &odata.Nummer,
		Afkorting:      &odata.Afkorting,
		NaamNL:         &odata.NaamNL,
		NaamEN:         &odata.NaamEN,
		AantalZetels:   &odata.AantalZetels,
		AantalStemmen:  &odata.AantalStemmen,
		DatumActief:    s.mapCustomDate(odata.DatumActief),
		DatumInactief:  s.mapCustomDate(odata.DatumInactief),
		ContentType:    &odata.ContentType,
		ContentLength:  &odata.ContentLength,
		GewijzigdOp:    &odata.GewijzigdOp,
		ApiGewijzigdOp: &odata.ApiGewijzigdOp,
		Verwijderd:     odata.Verwijderd,
	}
}

func (s *GormPostgresStorage) mapZaakActorToDB(odata *odata.ZaakActor, zaakID string) *models.ZaakActor {
	if odata == nil {
		return nil
	}

	var persoonID, fractieID *string

	if odata.Persoon != nil {
		persoonID = &odata.Persoon.ID
	}

	if odata.Fractie != nil {
		fractieID = &odata.Fractie.ID
	}

	return &models.ZaakActor{
		ID:           odata.ID,
		ZaakID:       &zaakID,
		PersoonID:    persoonID,
		FractieID:    fractieID,
		Relatie:      &odata.Relatie,
		ActorNaam:    &odata.ActorNaam,
		ActorFractie: &odata.ActorFractie,
		Bijgewerkt:   &odata.Bijgewerkt,
		Verwijderd:   odata.Verwijderd,
	}
}

func (s *GormPostgresStorage) mapVotingResultToDB(odata *odata.VotingResult) *models.VotingResult {
	if odata == nil {
		return nil
	}

	partyVotesJSON, _ := json.Marshal(odata.PartyVotes)

	return &models.VotingResult{
		DocumentID:   odata.ZaakID,
		BesluitID:    odata.BesluitID,
		BesluitTekst: odata.BesluitTekst,
		VotingType:   odata.VotingType,
		PartyVotes:   string(partyVotesJSON),
		Date:         &odata.Date,
		Status:       odata.Status,
	}
}

func (s *GormPostgresStorage) mapIndividueleStemmingToDB(odata *odata.IndividueleStemming) *models.IndividueleStemming {
	if odata == nil {
		return nil
	}

	return &models.IndividueleStemming{
		BesluitID:    odata.BesluitID,
		PersonID:     odata.PersonID,
		PersonName:   odata.PersonName,
		FractieID:    odata.FractieID,
		FractieName:  odata.FractieName,
		VoteType:     odata.VoteType,
		IsCorrection: odata.IsCorrection,
		Date:         odata.Date,
	}
}

func (s *GormPostgresStorage) mapCustomDate(customDate *odata.CustomDate) *time.Time {
	if customDate == nil || !customDate.Valid {
		return nil
	}
	return &customDate.Time
}
