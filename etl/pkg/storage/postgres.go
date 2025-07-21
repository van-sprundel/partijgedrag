package storage

import (
	"context"
	"fmt"
	"log"

	"etl/internal/config"
	"etl/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type PostgresStorage struct {
	db *gorm.DB
}

func NewPostgresStorage(config config.StorageConfig) (*PostgresStorage, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=UTC",
		config.Host, config.Username, config.Password, config.Database, config.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate all models
	if err := db.AutoMigrate(
		&models.Zaak{},
		&models.Besluit{},
		&models.Stemming{},
		&models.Persoon{},
		&models.Fractie{},
		&models.ZaakActor{},
		&models.Kamerstukdossier{},
		&models.DocumentInfo{},
		&models.ZaakDocument{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) SaveZaken(ctx context.Context, zaken []models.Zaak) error {
	if len(zaken) == 0 {
		return nil
	}

	log.Printf("Saving %d zaken...", len(zaken))

	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(zaken, 1000).Error
}

func (s *PostgresStorage) SaveBesluiten(ctx context.Context, besluiten []models.Besluit) error {
	if len(besluiten) == 0 {
		return nil
	}

	log.Printf("Saving %d besluiten...", len(besluiten))

	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(besluiten, 1000).Error
}

func (s *PostgresStorage) SaveStemmingen(ctx context.Context, stemmingen []models.Stemming) error {
	if len(stemmingen) == 0 {
		return nil
	}

	log.Printf("Saving %d stemmingen...", len(stemmingen))

	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(stemmingen, 1000).Error
}

func (s *PostgresStorage) SavePersonen(ctx context.Context, personen []models.Persoon) error {
	if len(personen) == 0 {
		return nil
	}

	log.Printf("Saving %d personen...", len(personen))

	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(personen, 1000).Error
}

func (s *PostgresStorage) SaveFracties(ctx context.Context, fracties []models.Fractie) error {
	if len(fracties) == 0 {
		return nil
	}

	log.Printf("Saving %d fracties...", len(fracties))

	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(fracties, 1000).Error
}

func (s *PostgresStorage) SaveZaakActors(ctx context.Context, zaakActors []models.ZaakActor) error {
	if len(zaakActors) == 0 {
		return nil
	}

	log.Printf("Saving %d zaak actors...", len(zaakActors))

	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(zaakActors, 1000).Error
}

func (s *PostgresStorage) SaveKamerstukdossiers(ctx context.Context, dossiers []models.Kamerstukdossier) error {
	if len(dossiers) == 0 {
		return nil
	}

	log.Printf("Saving %d kamerstukdossiers...", len(dossiers))

	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).CreateInBatches(dossiers, 1000).Error
}

func (s *PostgresStorage) SaveDocumentInfo(ctx context.Context, docInfo models.DocumentInfo) error {
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoNothing: true, // Don't overwrite existing documents
	}).Create(&docInfo).Error
}

func (s *PostgresStorage) DocumentExists(ctx context.Context, dossierNummer string, volgnummer int) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&models.DocumentInfo{}).
		Where("dossier_nummer = ? AND volgnummer = ?", dossierNummer, volgnummer).
		Count(&count).Error

	return count > 0, err
}

func (s *PostgresStorage) LinkZaakToDocument(ctx context.Context, zaakID string, documentID uint) error {
	zaakDoc := models.ZaakDocument{
		ZaakID:     zaakID,
		DocumentID: documentID,
	}

	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&zaakDoc).Error
}

func (s *PostgresStorage) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *PostgresStorage) Ping(ctx context.Context) error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (s *PostgresStorage) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	tables := map[string]interface{}{
		"zaken":             &models.Zaak{},
		"besluiten":         &models.Besluit{},
		"stemmingen":        &models.Stemming{},
		"personen":          &models.Persoon{},
		"fracties":          &models.Fractie{},
		"zaak_actors":       &models.ZaakActor{},
		"kamerstukdossiers": &models.Kamerstukdossier{},
		"document_info":     &models.DocumentInfo{},
		"zaak_documents":    &models.ZaakDocument{},
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
