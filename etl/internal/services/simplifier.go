package service

import (
	"context"
	"encoding/json"
	"etl/internal/llm"
	"etl/internal/models"
	"etl/pkg/storage"
	"fmt"
	"log"
)

type SimplifierService struct {
	store storage.Storage
	llm   llm.LLMClient
}

type simplifyResult struct {
	idx  int
	zaak models.Zaak
	err  error
}

func NewSimplifierService(store storage.Storage, llm llm.LLMClient) *SimplifierService {
	return &SimplifierService{store: store, llm: llm}
}

func (s *SimplifierService) SimplifyCases(ctx context.Context, limit int) error {
	println("✨ Starting SimplifyCasesConcurrent")
	zaken, err := s.store.GetZakenForSimplifying(ctx, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch zaken for simplifying: %w", err)
	}

	if len(zaken) == 0 {
		log.Println("✨ No zaken found that need simplification.")
		return nil
	}

	log.Printf("🧾 Found %d zaken to simplify...", len(zaken))

	var simplifiedZaken []models.Zaak

	for _, z := range zaken {
		log.Printf("✨ Simplifying zaak %s", z.ID)
		var bulletPoints []string

		// Safely decode JSON bullet points
		if z.BulletPoints != nil && len(*z.BulletPoints) > 0 {
			if err := json.Unmarshal([]byte(*z.BulletPoints), &bulletPoints); err != nil {
				log.Printf("⚠️ Skipping zaak %s: invalid bullet_points JSON (%v)", z.ID, err)
				continue
			}
		}

		// Skip empty cases
		if len(bulletPoints) == 0 && (z.Titel == nil || *z.Titel == "") {
			continue
		}

		// Call the LLM client (Ollama or OpenAI)
		simplified, err := s.llm.SimplifyCase(*z.Titel, bulletPoints)
		if err != nil {
			log.Printf("❌ Failed to simplify zaak %s: %v", z.ID, err)
			continue
		}

		// Prepare new simplified zaak
		// Prepare new simplified zaak
		newZaak := z
		newZaak.SimplifiedTitel = &simplified.SimplifiedTitel

		// Assign simplified bullet points directly (pq.StringArray)
		newZaak.SimplifiedBulletPoints = simplified.SimplifiedBulletPoints

		simplifiedZaken = append(simplifiedZaken, newZaak)
	}

	if len(simplifiedZaken) == 0 {
		log.Println("ℹ️ No zaken were successfully simplified.")
		return nil
	}

	if err := s.store.UpdateZakenForSimplifying(ctx, simplifiedZaken); err != nil {
		return fmt.Errorf("failed to update simplified zaken: %w", err)
	}

	log.Printf("✅ Successfully simplified and updated %d zaken.", len(simplifiedZaken))
	return nil
}

func (s *SimplifierService) SimplifyCasesConcurrent(ctx context.Context, limit int, workers int, batchSize int) error {
	println("✨ Starting SimplifyCasesConcurrent")

	zaken, err := s.store.GetZakenForSimplifying(ctx, limit)
	if err != nil {
		return fmt.Errorf("failed to fetch zaken for simplifying: %w", err)
	}
	if len(zaken) == 0 {
		log.Println("✨ No zaken found that need simplification.")
		return nil
	}

	log.Printf("🧾 Found %d zaken to simplify...", len(zaken))

	jobs := make(chan int)
	results := make(chan simplifyResult)

	// start worker goroutines
	for w := 0; w < workers; w++ {
		go func() {
			for i := range jobs {
				select {
				case <-ctx.Done():
					return // exit goroutine if context is cancelled
				default:
					z := zaken[i]
					// ... rest of processing

					// decode bulletPoints
					var bulletPoints []string
					if z.BulletPoints != nil && len(*z.BulletPoints) > 0 {
						if err := json.Unmarshal([]byte(*z.BulletPoints), &bulletPoints); err != nil {
							results <- simplifyResult{i, z, fmt.Errorf("invalid bullet_points JSON: %w", err)}
							continue
						}
					}
					log.Printf("✨ Simplifying zaak %s", z.ID)
					simplified, err := s.llm.SimplifyCase(*z.Titel, bulletPoints)

					if err != nil {
						results <- simplifyResult{i, z, err}
						continue
					}

					newZaak := z
					newZaak.SimplifiedTitel = &simplified.SimplifiedTitel
					newZaak.SimplifiedBulletPoints = simplified.SimplifiedBulletPoints

					results <- simplifyResult{i, newZaak, nil}
				}

			}
		}()
	}

	// send jobs
	go func() {
		for i := range zaken {
			jobs <- i
		}
		close(jobs)
	}()

	// collect results
	var batch []models.Zaak
	for range zaken {
		res := <-results
		if res.err != nil {
			log.Printf("❌ Failed to simplify zaak %s: %v", res.zaak.ID, res.err)
			continue
		}

		batch = append(batch, res.zaak)
		if len(batch) >= batchSize {
			if err := s.store.UpdateZakenForSimplifying(ctx, batch); err != nil {
				return fmt.Errorf("failed to update batch: %w", err)
			}
			log.Printf("✅ Flushed %d simplified zaken", len(batch))
			batch = batch[:0] // reset
		}
	}

	// flush any remaining zaken
	if len(batch) > 0 {
		if err := s.store.UpdateZakenForSimplifying(ctx, batch); err != nil {
			return fmt.Errorf("failed to update final batch: %w", err)
		}
		log.Printf("✅ Flushed final %d simplified zaken", len(batch))
	}

	return nil
}
