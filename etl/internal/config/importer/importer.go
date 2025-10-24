package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"etl/internal/api"
	"etl/internal/models"
	"etl/internal/parser"
	"etl/pkg/odata"
	"etl/pkg/storage"
)

type SimpleImporter struct {
	client    *odata.Client
	apiClient *api.Client
	storage   storage.Storage
	stats     *models.ImportStats
	parser    *parser.DocumentParser
}

func NewImporter(client *odata.Client, storage storage.Storage, apiClient *api.Client) *SimpleImporter {
	return &SimpleImporter{
		client:    client,
		apiClient: apiClient,
		storage:   storage,
		stats:     models.NewImportStats(),
		parser:    parser.NewDocumentParser(),
	}
}

func (imp *SimpleImporter) GetStats() *models.ImportStats {
	return imp.stats
}

func (imp *SimpleImporter) ImportMotiesWithVotes(ctx context.Context, after *time.Time) error {
	if after != nil {
		log.Printf("Starting import of motions with votes modified after %s...", after.Format(time.RFC3339))
	} else {
		log.Println("Starting import of motions with votes...")
	}
	startTime := time.Now()

	err := imp.processAllMotions(ctx, after)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	totalDuration := time.Since(startTime)
	log.Printf("Import complete in %v", totalDuration)

	imp.stats.Finalize()
	return nil
}

func (imp *SimpleImporter) processAllMotions(ctx context.Context, after *time.Time) error {
	skip := 0
	pageNum := 1
	totalProcessed := 0
	startTime := time.Now()
	lastBatchSize := 0

	// Try to get total count for better progress tracking
	totalCount := 0
	if count, err := imp.client.GetMotiesCount(ctx, after); err == nil {
		totalCount = count
		log.Printf("Estimated total records to process: %d", totalCount)
	} else {
		log.Printf("Could not get total count (will show relative progress): %v", err)
	}

	for {
		batchStartTime := time.Now()
		log.Printf("Processing page %d (skip=%d)...", pageNum, skip)

		// Fetch page of zaken
		zaken, err := imp.fetchZakenPageAfter(ctx, skip, after)
		if err != nil {
			return fmt.Errorf("fetching page %d: %w", pageNum, err)
		}

		if len(zaken) == 0 {
			log.Printf("No more data found, import complete")
			break
		}

		entities := imp.extractEntities(ctx, zaken)

		if err := imp.saveEntities(ctx, entities); err != nil {
			return fmt.Errorf("saving entities from page %d: %w", pageNum, err)
		}

		// Update stats
		imp.updateStats(entities)

		batchDuration := time.Since(batchStartTime)
		totalProcessed += len(zaken)
		avgDuration := time.Since(startTime) / time.Duration(pageNum)

		// Progress indicators
		progressMsg := fmt.Sprintf("Page %d complete: processed %d zaken in %v (total: %d)",
			pageNum, len(zaken), batchDuration.Round(time.Second), totalProcessed)

		// Add percentage and ETA if we have total count
		if totalCount > 0 {
			percentage := float64(totalProcessed) / float64(totalCount) * 100
			progressMsg += fmt.Sprintf(" [%.1f%%]", percentage)

			if totalProcessed > 0 {
				avgTimePerRecord := time.Since(startTime) / time.Duration(totalProcessed)
				remainingRecords := totalCount - totalProcessed
				estimatedTimeLeft := time.Duration(remainingRecords) * avgTimePerRecord
				progressMsg += fmt.Sprintf(" | ETA: %v", estimatedTimeLeft.Round(time.Minute))
			}
		}

		// Add batch size trend to help estimate progress
		if lastBatchSize > 0 {
			trend := ""
			if len(zaken) < lastBatchSize {
				trend = " (↓ smaller batch - likely approaching end)"
			} else if len(zaken) > lastBatchSize {
				trend = " (↑ larger batch)"
			} else {
				trend = " (→ same size)"
			}
			progressMsg += trend
		}

		log.Printf("%s | avg batch time: %v", progressMsg, avgDuration.Round(time.Second))
		lastBatchSize = len(zaken)

		skip += len(zaken)
		pageNum++
	}

	return nil
}

func (imp *SimpleImporter) fetchZakenPage(ctx context.Context, skip int) ([]models.Zaak, error) {
	return imp.fetchZakenPageAfter(ctx, skip, nil)
}

func (imp *SimpleImporter) fetchZakenPageAfter(ctx context.Context, skip int, after *time.Time) ([]models.Zaak, error) {
	data, err := imp.client.GetMotiesWithVotesAfter(ctx, skip, 0, after)
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

func (imp *SimpleImporter) extractEntities(ctx context.Context, zaken []models.Zaak) ExtractedEntities {
	entities := ExtractedEntities{
		Zaken: zaken,
	}

	persoonMap := make(map[string]models.Persoon)
	fractieMap := make(map[string]models.Fractie)
	kamerstukdossierMap := make(map[string]models.Kamerstukdossier)

	for i := range zaken {
		zaak := &zaken[i]
		for _, besluit := range zaak.Besluit {
			besluit.ZaakID = &zaak.ID
			entities.Besluiten = append(entities.Besluiten, besluit)

			for _, stemming := range besluit.Stemming {
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
			} else {
				zaakActor.PersoonID = nil
			}

			if zaakActor.Fractie != nil {
				zaakActor.FractieID = &zaakActor.Fractie.ID
				fractieMap[zaakActor.Fractie.ID] = *zaakActor.Fractie
			} else {
				zaakActor.FractieID = nil
			}

			entities.ZaakActors = append(entities.ZaakActors, zaakActor)
		}

		// Find and set DID from Document
		// Match only on Onderwerp (subject) since multiple motions can share the same Titel
		var targetDoc *models.Document
		for _, dossier := range zaak.Kamerstukdossier {
			for _, doc := range dossier.Document {
				if zaak.Onderwerp != nil && strings.EqualFold(strings.TrimSpace(doc.Onderwerp), strings.TrimSpace(*zaak.Onderwerp)) {
					targetDoc = &doc
					break
				}
			}
			if targetDoc != nil {
				break
			}
		}

		if targetDoc != nil {
			zaak.DID = &targetDoc.DocumentNummer
		}

		for _, dossier := range zaak.Kamerstukdossier {
			kamerstukdossierMap[dossier.ID] = dossier
		}
	}

	for _, persoon := range persoonMap {
		entities.Personen = append(entities.Personen, persoon)
	}

	for _, fractie := range fractieMap {
		entities.Fracties = append(entities.Fracties, fractie)
	}

	// Fetch logos for fracties that don't have one yet
	imp.fetchFractieLogos(ctx, entities.Fracties)

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

	if err := imp.storage.SaveKamerstukdossiers(ctx, entities.Kamerstukdossiers); err != nil {
		return fmt.Errorf("saving kamerstukdossiers: %w", err)
	}

	// 2. Zaken (independent)
	if err := imp.storage.SaveZaken(ctx, entities.Zaken); err != nil {
		return fmt.Errorf("saving zaken: %w", err)
	}

	// Process documents for motions
	if err := imp.processMotionDocuments(ctx, entities.Zaken); err != nil {
		log.Printf("Warning: failed to process documents for some motions: %v", err)
	}

	// 3. Entities that depend on zaken
	if err := imp.storage.SaveZaakActors(ctx, entities.ZaakActors); err != nil {
		return fmt.Errorf("saving zaak actors: %w", err)
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
		imp.stats.ZakenByType[*zaak.Soort]++
	}
}

func (imp *SimpleImporter) processMotionDocuments(ctx context.Context, zaken []models.Zaak) error {
	log.Printf("Processing documents for %d zaken...", len(zaken))
	processed := 0
	errors := 0

	for _, zaak := range zaken {
		if *zaak.Soort != "Motie" {
			continue
		}

		if err := imp.processMotionDocument(ctx, zaak); err != nil {
			log.Printf("Error processing document for zaak %s: %v", zaak.ID, err)
			errors++
		} else {
			processed++
		}
	}

	log.Printf("Document processing complete for batch: %d processed, %d errors", processed, errors)
	return nil
}

func (imp *SimpleImporter) fetchFractieLogos(ctx context.Context, fracties []models.Fractie) {
	var wg sync.WaitGroup
	for i := range fracties {
		if len(fracties[i].LogoData) == 0 {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				if logoData := imp.fetchFractieLogo(ctx, fracties[i].ID); logoData != nil {
					fracties[i].LogoData = logoData
				}
			}(i)
		}
	}
	wg.Wait()
}

func (imp *SimpleImporter) fetchFractieLogo(ctx context.Context, fractieID string) []byte {
	logoURL := fmt.Sprintf("https://gegevensmagazijn.tweedekamer.nl/OData/v4/2.0/fractie/%s/resource", fractieID)

	data, err := imp.client.MakeRequest(ctx, logoURL)
	if err != nil {
		log.Printf("Failed to fetch logo for fractie %s: %v", fractieID, err)
		return nil
	}

	log.Printf("Successfully fetched logo for fractie %s (%d bytes)", fractieID, len(data))
	return data
}

func (imp *SimpleImporter) processMotionDocument(ctx context.Context, zaak models.Zaak) error {
	log.Printf("Attempting to process document for motion '%s' (%s)", *zaak.Onderwerp, zaak.ID)

	// check if we have a DID set
	if zaak.DID == nil {
		log.Printf("No DID found for motion '%s' (%s), skipping", *zaak.Onderwerp, zaak.ID)
		return nil
	}

	// find the document by DocumentNummer (DID)
	var targetDoc *models.Document
	var targetDossier *models.Kamerstukdossier

	for _, dossier := range zaak.Kamerstukdossier {
		for _, doc := range dossier.Document {
			if doc.DocumentNummer == *zaak.DID {
				targetDoc = &doc
				targetDossier = &dossier
				break
			}
		}
		if targetDoc != nil {
			break
		}
	}

	if targetDoc == nil {
		log.Printf("No document found with DID '%s' for motion '%s' (%s)", *zaak.DID, *zaak.Onderwerp, zaak.ID)
		return nil // Not an error, just no document to process
	}

	log.Printf("Found document '%s' (volgnummer %d) in dossier %s for motion '%s'",
		targetDoc.DocumentNummer, targetDoc.Volgnummer, targetDossier.ID, *zaak.Onderwerp)

	docResponse, err := imp.apiClient.FetchDocument(ctx, *targetDossier, targetDoc.Volgnummer)
	if err != nil {
		return fmt.Errorf("failed to fetch document %s (volgnummer %d) for dossier %s: %w",
			targetDoc.DocumentNummer, targetDoc.Volgnummer, targetDossier.ID, err)
	}

	result, err := imp.parser.ExtractBulletPoints(docResponse.XMLData, docResponse.URL)
	if err != nil {
		return fmt.Errorf("failed to parse document %s (volgnummer %d) for dossier %s: %w",
			targetDoc.DocumentNummer, targetDoc.Volgnummer, targetDossier.ID, err)
	}

	if result == nil {
		log.Printf("Document %s (volgnummer %d) at %s is not a motion, skipping for zaak %s",
			targetDoc.DocumentNummer, targetDoc.Volgnummer, docResponse.URL, zaak.ID)
		return nil
	}

	if len(result.BulletPoints) == 0 {
		log.Printf("Motion '%s' (zaak %s, document %s) has no bullet points.",
			result.Title, zaak.ID, targetDoc.DocumentNummer)
		return nil
	}

	bulletPointsJSON, err := json.Marshal(result.BulletPoints)
	if err != nil {
		return fmt.Errorf("marshaling bullet points for zaak %s: %w", zaak.ID, err)
	}

	if err := imp.storage.UpdateZaakBulletPoints(ctx, zaak.ID, string(bulletPointsJSON), result.URL); err != nil {
		return fmt.Errorf("updating bullet points for zaak %s: %w", zaak.ID, err)
	}

	log.Printf("Successfully stored %d bullet points for motion '%s' (zaak %s, document %s)",
		len(result.BulletPoints), result.Title, zaak.ID, targetDoc.DocumentNummer)

	return nil
}
