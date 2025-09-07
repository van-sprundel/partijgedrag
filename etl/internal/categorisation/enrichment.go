package categorisation

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"etl/internal/models"
	"etl/pkg/storage"
)

type Service struct {
	store storage.Storage
}

func NewService(store storage.Storage) *Service {
	return &Service{store: store}
}

func (s *Service) EnrichZaken(ctx context.Context) error {
	log.Println("Starting enrichment of zaken with categories...")

	categories, err := s.store.GetAllCategories(ctx)
	if err != nil {
		return fmt.Errorf("getting categories: %w", err)
	}

	if len(categories) == 0 {
		log.Println("No categories found. Categories need to be seeded first.")
		return nil
	}

	zaken, err := s.store.GetZakenForEnrichment(ctx)
	if err != nil {
		return fmt.Errorf("getting zaken: %w", err)
	}

	log.Printf("Found %d zaken to analyze and %d categories", len(zaken), len(categories))

	var assignments []models.ZaakCategory
	now := time.Now()
	for _, zaak := range zaken {
		matches := s.findCategoryMatches(zaak, categories)
		for _, categoryID := range matches {
			assignments = append(assignments, models.ZaakCategory{
				ZaakID:     zaak.ID,
				CategoryID: categoryID,
				CreatedAt:  now,
			})
		}
	}

	if len(assignments) > 0 {
		if err := s.store.SaveZaakCategories(ctx, assignments); err != nil {
			log.Printf("Warning: failed to assign some categories in bulk: %v", err)
		}
	}

	log.Printf("Enrichment complete: assigned %d category relationships", len(assignments))
	return nil
}

func (s *Service) findCategoryMatches(zaak models.Zaak, categories []models.MotionCategory) []string {
	var matches []string

	var sb strings.Builder
	if zaak.Titel != nil {
		sb.WriteString(strings.ToLower(*zaak.Titel))
		sb.WriteString(" ")
	}
	if zaak.Onderwerp != nil {
		sb.WriteString(strings.ToLower(*zaak.Onderwerp))
		sb.WriteString(" ")
	}
	searchText := sb.String()

	if searchText == "" {
		return nil
	}

	for _, category := range categories {
		for _, keyword := range category.Keywords {
			re, err := regexp.Compile(`\b` + regexp.QuoteMeta(strings.ToLower(keyword)) + `\b`)
			if err != nil {
				log.Printf("Warning: could not compile regex for keyword '%s': %v", keyword, err)
				continue
			}
			if re.MatchString(searchText) {
				matches = append(matches, category.ID)
				break
			}
		}
	}

	return matches
}
