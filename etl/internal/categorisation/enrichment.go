package categorisation

import (
	"context"
	"fmt"
	"log"
	"strings"

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

	enriched := 0
	for _, zaak := range zaken {
		matches := s.findCategoryMatches(zaak, categories)

		if len(matches) > 0 {
			for _, categoryID := range matches {
				if err := s.store.AssignCategoryToZaak(ctx, zaak.ID, categoryID); err != nil {
					log.Printf("Warning: failed to assign category to zaak %s: %v", zaak.ID, err)
				} else {
					enriched++
				}
			}
		}
	}

	log.Printf("Enrichment complete: assigned %d category relationships", enriched)
	return nil
}

func (s *Service) findCategoryMatches(zaak models.Zaak, categories []models.MotionCategory) []string {
	var matches []string

	searchText := ""
	if zaak.Titel != nil {
		searchText += strings.ToLower(*zaak.Titel) + " "
	}
	if zaak.Onderwerp != nil {
		searchText += strings.ToLower(*zaak.Onderwerp) + " "
	}

	for _, category := range categories {
		for _, keyword := range []string(category.Keywords) {
			if strings.Contains(searchText, keyword) {
				matches = append(matches, category.ID)
				break
			}
		}
	}

	return matches
}
