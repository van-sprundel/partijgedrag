package storage

import (
	"context"
	"encoding/json"
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

func (s *PostgresStorage) GetZakenForSimplifying(ctx context.Context, limit int) ([]models.Zaak, error) {
	var zaken []models.Zaak

	q := s.db.WithContext(ctx).
		Table("zaken AS z").
		Select("z.id, z.titel, z.bullet_points").
		Joins("LEFT JOIN besluiten b ON b.zaak_id = z.id").
		Where("z.simplified_bullet_points IS NULL").
		Where("z.bullet_points IS NOT NULL")

	if limit > 0 {
		q = q.Limit(limit)
	}

	if err := q.Find(&zaken).Error; err != nil {
		return nil, fmt.Errorf("failed to load filtered zaken: %w", err)
	}
	log.Printf("Found %d zakent to simpolify", len(zaken))
	return zaken, nil
}

func (s *PostgresStorage) UpdateZakenForSimplifying(ctx context.Context, zaken []models.Zaak) error {
	if len(zaken) == 0 {
		return nil
	}

	log.Printf("🔄 Updating %d simplified zaken...", len(zaken))

	tx := s.db.WithContext(ctx).Begin()
	for _, z := range zaken {
		updates := map[string]interface{}{}

		if len(z.SimplifiedBulletPoints) > 0 {
			bpJSON, err := json.Marshal(z.SimplifiedBulletPoints)
			if err != nil {
				log.Printf("⚠️ could not marshal simplified bullet points for zaak %s: %v", z.ID, err)
				continue
			}
			updates["simplified_bullet_points"] = string(bpJSON)
		}

		if z.SimplifiedTitel != nil && *z.SimplifiedTitel != "" {
			updates["simplified_title"] = *z.SimplifiedTitel
		}

		if len(updates) == 0 {
			continue
		}

		if err := tx.Model(&models.Zaak{}).
			Where("id = ?", z.ID).
			Updates(updates).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update zaak %s: %w", z.ID, err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit zaak updates: %w", err)
	}

	log.Println("✅ Simplified zaken updated successfully.")
	return nil
}

//func (s *PostgresStorage) SimplifyCasesWithOllamaFiltered(ctx context.Context, model string, limit int) error {
//	type CaseRow struct {
//		ID           string          `gorm:"column:id"`
//		Title        string          `gorm:"column:titel"`
//		BulletPoints json.RawMessage `gorm:"column:bullet_points"`
//	}
//
//	type SimplifiedCase struct {
//		SimplifiedTitle        string `json:"simplified_title"`
//		SimplifiedBulletPoints string `json:"simplified_bullet_points"`
//	}
//
//	var cases []CaseRow
//
//	query := `
//	SELECT z.id, z.titel, z.bullet_points
//	FROM zaken z
//	LEFT JOIN besluiten b ON b.zaak_id = z.id
//	WHERE z.simplified_bullet_points IS NULL
//	AND z.bullet_points IS NOT NULL
//`
//	if limit > 0 {
//		query += fmt.Sprintf(" LIMIT %d", limit)
//	}
//
//	if err := s.db.WithContext(ctx).Raw(query).Scan(&cases).Error; err != nil {
//		return fmt.Errorf("failed to load filtered cases: %w", err)
//	}
//
//	log.Printf("🧾 Found %d cases to simplify", len(cases))
//
//	for _, c := range cases {
//		var bulletPoints []string
//		if err := json.Unmarshal(c.BulletPoints, &bulletPoints); err != nil {
//			log.Printf("❌ case %s: invalid bullet_points json", c.ID)
//			continue
//		}
//
//		simplifiedCase, err := callOllamaSimplifyCase(model, c.Title, bulletPoints)
//		if err != nil {
//			log.Printf("ollama error for case %s: %v", c.ID, err)
//			continue
//		}
//		// Add "vz:" prefix to bullet points that start with "verzoekt"
//		for i, bp := range bulletPoints {
//			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(bp)), "verzoekt") {
//				if i < len(simplifiedCase.SimplifiedBulletPoints) {
//					simplifiedCase.SimplifiedBulletPoints[i] = "vz: " + simplifiedCase.SimplifiedBulletPoints[i]
//				}
//			}
//		}
//
//		simplifiedJSON, _ := json.Marshal(simplifiedCase.SimplifiedBulletPoints)
//
//		if err := s.db.WithContext(ctx).
//			Model(&models.Zaak{}).
//			Where("id = ?", c.ID).
//			Updates(map[string]interface{}{
//				"simplified_bullet_points": simplifiedJSON,
//				"simplified_title":         simplifiedCase.SimplifiedTitle,
//			}).Error; err != nil {
//			log.Printf("❌ failed to update case %s: %v", c.ID, err)
//		} else {
//			log.Printf("✅ simplified case %s", c.ID)
//		}
//	}
//
//	return nil
//}
//
////func callOllamaSimplifyCase(model string, title string, bulletPoints []string) (*struct {
////	SimplifiedTitle        string   `json:"simplified_title"`
////	SimplifiedBulletPoints []string `json:"simplified_bullet_points"`
////}, error) {
////	if strings.TrimSpace(title) == "" && len(bulletPoints) == 0 {
////		return &struct {
////			SimplifiedTitle        string   `json:"simplified_title"`
////			SimplifiedBulletPoints []string `json:"simplified_bullet_points"`
////		}{
////			SimplifiedTitle:        "",
////			SimplifiedBulletPoints: []string{},
////		}, nil
////	}
////
////	bpJSON, _ := json.Marshal(bulletPoints)
////
////	prompt := fmt.Sprintf(`Je bent een taalassistent gespecialiseerd in het eenvoudig en duidelijk herschrijven van politieke teksten in het Nederlands.
////
////**Doel:**
////Herschrijf de volgende titel en bijbehorende bullet points in eenvoudiger Nederlands.
////Behoud de oorspronkelijke betekenis en nuance, maar gebruik korte, begrijpelijke zinnen voor een breed publiek.
////
////**Belangrijk:**
////- Maak de tekst vriendelijk en helder, zonder jargon.
////- Kort waar mogelijk, maar verlies geen inhoud.
////- Geef **uitsluitend** het JSON-object terug en niets anders.
////
////**Verwacht JSON-formaat:**
////{
////  "simplified_title": "Nieuwe titel in eenvoudiger Nederlands",
////  "simplified_bullet_points": [
////    "Eerste vereenvoudigde bullet point",
////    "Tweede vereenvoudigde bullet point"
////  ]
////}
////
////**Invoer:**
////Titel:
////%s
////
////Bullet points:
////%s
////`, title, string(bpJSON))
////	req := map[string]string{
////		"model":  model,
////		"prompt": prompt,
////	}
////	body, _ := json.Marshal(req)
////
////	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(body))
////	if err != nil {
////		return nil, err
////	}
////	defer resp.Body.Close()
////
////	var out strings.Builder
////	decoder := json.NewDecoder(resp.Body)
////	for decoder.More() {
////		var chunk struct {
////			Response string `json:"response"`
////			Done     bool   `json:"done"`
////		}
////		if err := decoder.Decode(&chunk); err != nil {
////
////			io.Copy(io.Discard, resp.Body)
////			return nil, err
////		}
////
////		out.WriteString(chunk.Response)
////		if chunk.Done {
////			break
////		}
////	}
////	raw := out.String()
////
////	// Find the first '{' and the last '}'
////	start := strings.Index(raw, "{")
////	end := strings.LastIndex(raw, "}")
////
////	if start == -1 || end == -1 || start >= end {
////		log.Printf("⚠️ Could not find JSON object in Ollama response:\n%s", raw)
////		return nil, fmt.Errorf("no valid JSON found in response")
////	}
////
////	jsonStr := raw[start : end+1]
////
////	// Then decode
////	var simplified struct {
////		SimplifiedTitle        string   `json:"simplified_title"`
////		SimplifiedBulletPoints []string `json:"simplified_bullet_points"`
////	}
////	if err := json.Unmarshal([]byte(jsonStr), &simplified); err != nil {
////		return nil, fmt.Errorf("failed to parse cleaned ollama JSON: %w", err)
////	}
////
////	return &simplified, nil
////}
