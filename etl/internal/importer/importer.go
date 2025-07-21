package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"etl/internal/models"
	"etl/pkg/odata"
	"etl/pkg/storage"
)

type SimpleImporter struct {
	client  *odata.Client
	storage storage.Storage
	stats   *models.ImportStats
}

func NewImporter(client *odata.Client, storage storage.Storage) *SimpleImporter {
	return &SimpleImporter{
		client:  client,
		storage: storage,
		stats:   models.NewImportStats(),
	}
}

func (imp *SimpleImporter) GetStats() *models.ImportStats {
	return imp.stats
}

func (imp *SimpleImporter) ImportMotiesWithVotes(ctx context.Context) error {
	log.Println("Starting import of motions with votes...")
	startTime := time.Now()

	err := imp.processAllMotions(ctx)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	totalDuration := time.Since(startTime)
	log.Printf("Import complete in %v", totalDuration)

	imp.stats.Finalize()
	return nil
}

func (imp *SimpleImporter) processAllMotions(ctx context.Context) error {
	skip := 0
	pageNum := 1

	for {
		log.Printf("Processing page %d (skip=%d)...", pageNum, skip)

		// Fetch page of zaken
		zaken, err := imp.fetchZakenPage(ctx, skip)
		if err != nil {
			return fmt.Errorf("fetching page %d: %w", pageNum, err)
		}

		if len(zaken) == 0 {
			log.Printf("No more data found, import complete")
			break
		}

		entities := imp.extractEntities(zaken)

		if err := imp.saveEntities(ctx, entities); err != nil {
			return fmt.Errorf("saving entities from page %d: %w", pageNum, err)
		}

		// Update stats
		imp.updateStats(entities)

		log.Printf("Page %d complete: processed %d zaken", pageNum, len(zaken))

		skip += len(zaken)
		pageNum++
	}

	return nil
}

func (imp *SimpleImporter) fetchZakenPage(ctx context.Context, skip int) ([]models.Zaak, error) {
	data, err := imp.client.GetMotiesWithVotes(ctx, skip, 0)
	if err != nil {
		return nil, err
	}

	response, err := odata.ParseODataResponse(data)
	if err != nil {
		return nil, err
	}

	zakenData, err := json.Marshal(response.Value)
	if err != nil {
		return nil, err
	}

	var zaken []models.Zaak
	if err := json.Unmarshal(zakenData, &zaken); err != nil {
		return nil, err
	}

	return zaken, nil
}

type ExtractedEntities struct {
	Zaken             []models.Zaak
	Besluiten         []models.Besluit
	Stemmingen        []models.Stemming
	Personen          []models.Persoon
	Fracties          []models.Fractie
	ZaakActors        []models.ZaakActor
	Kamerstukdossiers []models.Kamerstukdossier
}

func (imp *SimpleImporter) extractEntities(zaken []models.Zaak) ExtractedEntities {
	entities := ExtractedEntities{
		Zaken: zaken,
	}

	// Use maps to deduplicate by ID
	persoonMap := make(map[string]models.Persoon)
	fractieMap := make(map[string]models.Fractie)
	kamerstukdossierMap := make(map[string]models.Kamerstukdossier)

	for _, zaak := range zaken {
		// Extract besluiten
		for _, besluit := range zaak.Besluit {
			// Set FK relationship
			besluit.ZaakID = &zaak.ID
			entities.Besluiten = append(entities.Besluiten, besluit)

			// Extract stemmingen from each besluit
			for _, stemming := range besluit.Stemming {
				// Set FK relationships
				stemming.BesluitID = &besluit.ID

				if stemming.Persoon != nil {
					stemming.PersoonID = &stemming.Persoon.ID
					persoonMap[stemming.Persoon.ID] = *stemming.Persoon
				}

				if stemming.Fractie != nil {
					stemming.FractieID = &stemming.Fractie.ID
					fractieMap[stemming.Fractie.ID] = *stemming.Fractie
				}

				entities.Stemmingen = append(entities.Stemmingen, stemming)
			}
		}

		for _, zaakActor := range zaak.ZaakActor {
			zaakActor.ZaakID = &zaak.ID

			if zaakActor.Persoon != nil {
				zaakActor.PersoonID = &zaakActor.Persoon.ID
				persoonMap[zaakActor.Persoon.ID] = *zaakActor.Persoon
			}

			if zaakActor.Fractie != nil {
				zaakActor.FractieID = &zaakActor.Fractie.ID
				fractieMap[zaakActor.Fractie.ID] = *zaakActor.Fractie
			}

			entities.ZaakActors = append(entities.ZaakActors, zaakActor)
		}

		// Extract kamerstukdossiers
		for _, dossier := range zaak.Kamerstukdossier {
			// Set FK relationship
			dossier.ZaakID = &zaak.ID
			kamerstukdossierMap[dossier.ID] = dossier
		}
	}

	// Convert maps to slices
	for _, persoon := range persoonMap {
		entities.Personen = append(entities.Personen, persoon)
	}

	for _, fractie := range fractieMap {
		entities.Fracties = append(entities.Fracties, fractie)
	}

	for _, dossier := range kamerstukdossierMap {
		entities.Kamerstukdossiers = append(entities.Kamerstukdossiers, dossier)
	}

	return entities
}

func (imp *SimpleImporter) saveEntities(ctx context.Context, entities ExtractedEntities) error {
	// Save in dependency order to maintain FK constraints

	// 1. Independent entities first
	if err := imp.storage.SavePersonen(ctx, entities.Personen); err != nil {
		return fmt.Errorf("saving personen: %w", err)
	}

	if err := imp.storage.SaveFracties(ctx, entities.Fracties); err != nil {
		return fmt.Errorf("saving fracties: %w", err)
	}

	// 2. Zaken (independent)
	if err := imp.storage.SaveZaken(ctx, entities.Zaken); err != nil {
		return fmt.Errorf("saving zaken: %w", err)
	}

	// 3. Entities that depend on zaken
	if err := imp.storage.SaveZaakActors(ctx, entities.ZaakActors); err != nil {
		return fmt.Errorf("saving zaak actors: %w", err)
	}

	if err := imp.storage.SaveKamerstukdossiers(ctx, entities.Kamerstukdossiers); err != nil {
		return fmt.Errorf("saving kamerstukdossiers: %w", err)
	}

	if err := imp.storage.SaveBesluiten(ctx, entities.Besluiten); err != nil {
		return fmt.Errorf("saving besluiten: %w", err)
	}

	// 4. Entities that depend on besluiten
	if err := imp.storage.SaveStemmingen(ctx, entities.Stemmingen); err != nil {
		return fmt.Errorf("saving stemmingen: %w", err)
	}

	return nil
}

func (imp *SimpleImporter) updateStats(entities ExtractedEntities) {
	imp.stats.TotalZaken += len(entities.Zaken)
	imp.stats.TotalBesluiten += len(entities.Besluiten)
	imp.stats.TotalStemmingen += len(entities.Stemmingen)
	imp.stats.TotalPersonen += len(entities.Personen)
	imp.stats.TotalFracties += len(entities.Fracties)
	imp.stats.TotalKamerstukdossiers += len(entities.Kamerstukdossiers)

	// Track zaak types
	for _, zaak := range entities.Zaken {
		imp.stats.ZakenByType[zaak.Soort]++
	}
}
