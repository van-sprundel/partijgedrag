package storage

import (
	"context"
	"fmt"
	"log"
	"time"

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

func (s *PostgresStorage) UpdateZaakBulletPoints(ctx context.Context, zaakID string, bulletPointsJSON string, documentURL string) error {
	return s.db.WithContext(ctx).
		Model(&models.Zaak{}).
		Where("id = ?", zaakID).
		Updates(map[string]interface{}{
			"bullet_points": bulletPointsJSON,
			"document_url":  documentURL,
		}).Error
}

func (s *PostgresStorage) SaveCategories(ctx context.Context, categories []models.MotionCategory) error {
	if len(categories) == 0 {
		return nil
	}

	log.Printf("Saving %d categories...", len(categories))

	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		UpdateAll: true,
	}).CreateInBatches(categories, 1000).Error
}

func (s *PostgresStorage) SaveZaakCategories(ctx context.Context, zaakCategories []models.ZaakCategory) error {
	if len(zaakCategories) == 0 {
		return nil
	}

	log.Printf("Saving %d zaak categories...", len(zaakCategories))

	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoNothing: true,
	}).CreateInBatches(zaakCategories, 1000).Error
}

func (s *PostgresStorage) GetAllCategories(ctx context.Context) ([]models.MotionCategory, error) {
	var categories []models.MotionCategory
	err := s.db.WithContext(ctx).Order("name").Find(&categories).Error
	return categories, err
}

func (s *PostgresStorage) GetCategoriesByType(ctx context.Context, categoryType string) ([]models.MotionCategory, error) {
	var categories []models.MotionCategory
	err := s.db.WithContext(ctx).Where("type = ?", categoryType).Order("name").Find(&categories).Error
	return categories, err
}

func (s *PostgresStorage) AssignCategoryToZaak(ctx context.Context, zaakID, categoryID string) error {
	zaakCategory := models.ZaakCategory{
		ZaakID:     zaakID,
		CategoryID: categoryID,
		CreatedAt:  time.Now(),
	}

	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&zaakCategory).Error
}

func (s *PostgresStorage) RemoveCategoryFromZaak(ctx context.Context, zaakID, categoryID string) error {
	return s.db.WithContext(ctx).
		Where("zaak_id = ? AND category_id = ?", zaakID, categoryID).
		Delete(&models.ZaakCategory{}).Error
}

func (s *PostgresStorage) GetZakenForEnrichment(ctx context.Context) ([]models.Zaak, error) {
	var zaken []models.Zaak

	err := s.db.WithContext(ctx).
		Where("soort = ? AND verwijderd = ? AND titel IS NOT NULL", "Motie", false).
		Where("NOT EXISTS (SELECT 1 FROM zaak_categories WHERE zaak_categories.zaak_id = zaken.id)").
		Find(&zaken).Error

	return zaken, err
}

func (s *PostgresStorage) CleanDatabase(ctx context.Context) error {
	log.Println("Cleaning orphaned rows from zaak_kamerstukdossiers...")
	query1 := `
		DELETE FROM zaak_kamerstukdossiers
		WHERE kamerstukdossier_id NOT IN (SELECT id FROM kamerstukdossiers);
	`
	if err := s.db.WithContext(ctx).Exec(query1).Error; err != nil {
		return fmt.Errorf("failed to delete orphaned rows by kamerstukdossier_id: %w", err)
	}

	query2 := `
		DELETE FROM zaak_kamerstukdossiers
		WHERE zaak_id NOT IN (SELECT id FROM zaken);
	`
	if err := s.db.WithContext(ctx).Exec(query2).Error; err != nil {
		return fmt.Errorf("failed to delete orphaned rows by zaak_id: %w", err)
	}

	log.Println("Finished cleaning orphaned rows.")
	return nil
}

func (s *PostgresStorage) Migrate(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(
		&models.Zaak{},
		&models.Besluit{},
		&models.Stemming{},
		&models.Persoon{},
		&models.Fractie{},
		&models.ZaakActor{},
		&models.Kamerstukdossier{},
		&models.MotionCategory{},
		&models.ZaakCategory{},
	)
}
