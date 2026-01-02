package storage

import (
	"context"
	"fmt"
	"log"
	"strings"
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
	log.Println("Cleaning all data from database tables...")

	tables := []string{
		"zaken",
		"besluiten",
		"stemmingen",
		"personen",
		"fracties",
		"zaak_actors",
		"kamerstukdossiers",
		"zaak_categories",
		"zaak_kamerstukdossiers",
	}

	query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", strings.Join(tables, ", "))

	if err := s.db.WithContext(ctx).Exec(query).Error; err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}

	log.Println("Finished cleaning database.")
	return nil
}

func (s *PostgresStorage) Migrate(ctx context.Context) error {
	// Run GORM AutoMigrate for tables
	if err := s.db.WithContext(ctx).AutoMigrate(
		&models.Zaak{},
		&models.Besluit{},
		&models.Stemming{},
		&models.Persoon{},
		&models.Fractie{},
		&models.ZaakActor{},
		&models.Kamerstukdossier{},
		&models.MotionCategory{},
		&models.ZaakCategory{},
		&models.UserSession{},
		&models.Activiteit{},
		&models.Agendapunt{},
	); err != nil {
		return fmt.Errorf("failed to run auto-migration: %w", err)
	}

	// Create text search indexes
	if err := s.createTextSearchIndexes(ctx); err != nil {
		return fmt.Errorf("failed to create text search indexes: %w", err)
	}

	// Create materialized views if they don't exist
	if err := s.createMaterializedViews(ctx); err != nil {
		return fmt.Errorf("failed to create materialized views: %w", err)
	}

	return nil
}

func (s *PostgresStorage) createTextSearchIndexes(ctx context.Context) error {
	// Enable pg_trgm extension for trigram-based text search
	if err := s.db.WithContext(ctx).Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error; err != nil {
		return fmt.Errorf("failed to create pg_trgm extension: %w", err)
	}

	// Create GIN indexes for text search on zaken table
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_zaken_onderwerp_gin_trgm ON zaken USING gin (onderwerp gin_trgm_ops);",
		"CREATE INDEX IF NOT EXISTS idx_zaken_citeertitel_gin_trgm ON zaken USING gin (citeertitel gin_trgm_ops);",
		"CREATE INDEX IF NOT EXISTS idx_zaken_nummer_gin_trgm ON zaken USING gin (nummer gin_trgm_ops);",
		"CREATE INDEX IF NOT EXISTS idx_zaken_bullet_points_gin ON zaken USING gin (bullet_points);",
	}
	for _, idx := range indexes {
		if err := s.db.WithContext(ctx).Exec(idx).Error; err != nil {
			return fmt.Errorf("failed to create text search index: %w", err)
		}
	}

	log.Println("Text search indexes created/verified successfully")
	return nil
}

func (s *PostgresStorage) createMaterializedViews(ctx context.Context) error {
	// Create majority_party_votes materialized view
	createMajorityPartyVotes := `
		CREATE MATERIALIZED VIEW IF NOT EXISTS majority_party_votes AS
		SELECT DISTINCT
			b.zaak_id,
			z.gestart_op,
			f.id as fractie_id,
			s.soort AS vote_type
		FROM stemmingen s
		JOIN besluiten b ON s.besluit_id = b.id
		JOIN zaken z ON b.zaak_id = z.id
		JOIN fracties f ON (s.actor_fractie = f.naam_nl OR s.actor_fractie = f.afkorting)
		WHERE s.actor_fractie IS NOT NULL
		  AND s.soort IN ('Voor', 'Tegen')
		  AND z.soort = 'Motie'
		  AND f.datum_inactief IS NULL;
	`
	if err := s.db.WithContext(ctx).Exec(createMajorityPartyVotes).Error; err != nil {
		return fmt.Errorf("failed to create majority_party_votes view: %w", err)
	}

	// Create indexes for majority_party_votes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_majority_party_votes_zaak_id ON majority_party_votes(zaak_id);",
		"CREATE INDEX IF NOT EXISTS idx_majority_party_votes_fractie_id ON majority_party_votes(fractie_id);",
		"CREATE INDEX IF NOT EXISTS idx_majority_party_votes_gestart_op ON majority_party_votes(gestart_op);",
		"CREATE INDEX IF NOT EXISTS idx_majority_party_votes_vote_type ON majority_party_votes(vote_type);",
	}
	for _, idx := range indexes {
		if err := s.db.WithContext(ctx).Exec(idx).Error; err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Create party_likeness_per_motion materialized view
	createPartyLikeness := `
		CREATE MATERIALIZED VIEW IF NOT EXISTS party_likeness_per_motion AS
		SELECT
			mv1.fractie_id as fractie1_id,
			mv2.fractie_id as fractie2_id,
			mv1.zaak_id,
			mv1.gestart_op,
			(mv1.vote_type = mv2.vote_type) as same_vote
		FROM majority_party_votes mv1
		JOIN majority_party_votes mv2 ON mv1.zaak_id = mv2.zaak_id
		WHERE mv1.fractie_id < mv2.fractie_id;
	`
	if err := s.db.WithContext(ctx).Exec(createPartyLikeness).Error; err != nil {
		return fmt.Errorf("failed to create party_likeness_per_motion view: %w", err)
	}

	// Create indexes for party_likeness_per_motion
	likenessIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_plpm_gestart_op ON party_likeness_per_motion(gestart_op);",
		"CREATE INDEX IF NOT EXISTS idx_plpm_fractie1_id ON party_likeness_per_motion(fractie1_id);",
		"CREATE INDEX IF NOT EXISTS idx_plpm_fractie2_id ON party_likeness_per_motion(fractie2_id);",
		"CREATE INDEX IF NOT EXISTS idx_plpm_zaak_id ON party_likeness_per_motion(zaak_id);",
		"CREATE INDEX IF NOT EXISTS idx_plpm_same_vote ON party_likeness_per_motion(same_vote);",
	}
	for _, idx := range likenessIndexes {
		if err := s.db.WithContext(ctx).Exec(idx).Error; err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	log.Println("Materialized views created/verified successfully")
	return nil
}
