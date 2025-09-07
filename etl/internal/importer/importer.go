package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
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

		entities := imp.extractEntities(zaken)

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

func (imp *SimpleImporter) extractEntities(zaken []models.Zaak) ExtractedEntities {
	entities := ExtractedEntities{
		Zaken: zaken,
	}

	persoonMap := make(map[string]models.Persoon)
	fractieMap := make(map[string]models.Fractie)
	kamerstukdossierMap := make(map[string]models.Kamerstukdossier)

	for _, zaak := range zaken {
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

		for _, dossier := range zaak.Kamerstukdossier {
			dossier.ZaakID = &zaak.ID
			kamerstukdossierMap[dossier.ID] = dossier
		}
	}

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

	// Process documents for kamerstukdossiers
	if err := imp.processDocumentsForDossiers(ctx, entities.Kamerstukdossiers); err != nil {
		log.Printf("Warning: failed to process documents for some kamerstukdossiers: %v", err)
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

func (imp *SimpleImporter) processDocumentsForDossiers(ctx context.Context, dossiers []models.Kamerstukdossier) error {
	if len(dossiers) == 0 {
		return nil
	}

	log.Printf("Processing documents for %d kamerstukdossiers...", len(dossiers))

	processed := 0
	errors := 0

	for _, dossier := range dossiers {
		if err := imp.processDossierDocument(ctx, dossier); err != nil {
			log.Printf("Error processing document for dossier %s: %v", dossier.ID, err)
			errors++
		} else {
			processed++
		}
	}

	log.Printf("Document processing complete for batch: %d processed, %d errors", processed, errors)
	return nil
}

func (imp *SimpleImporter) processDossierDocument(ctx context.Context, dossier models.Kamerstukdossier) error {
	// Get all potential volgnummers from the Document data we already have
	volgnummers := imp.getMotieVolgnummers(dossier)
	if len(volgnummers) == 0 {
		log.Printf("No motion documents found for dossier %s (checked %d documents)",
			imp.formatDossierNumber(dossier), len(dossier.Document))
		return nil
	}

	log.Printf("Found %d potential motion documents for dossier %s: %v",
		len(volgnummers), imp.formatDossierNumber(dossier), volgnummers)

	// Try each volgnummer in descending order until one works
	var docResponse *api.DocumentResponse
	var lastErr error

	for _, volgnummer := range volgnummers {
		log.Printf("Trying volgnummer %d for dossier %s", volgnummer, imp.formatDossierNumber(dossier))
		var err error
		docResponse, err = imp.apiClient.FetchDocument(ctx, dossier, volgnummer)
		if err == nil {
			log.Printf("Successfully fetched document with volgnummer %d", volgnummer)
			break
		}
		log.Printf("Failed to fetch volgnummer %d: %v", volgnummer, err)
		lastErr = err
	}

	if docResponse == nil {
		return fmt.Errorf("all volgnummers failed for dossier %s, last error: %w", imp.formatDossierNumber(dossier), lastErr)
	}

	result, err := imp.parser.ExtractBulletPoints(docResponse.XMLData, docResponse.URL)
	if err != nil {
		return fmt.Errorf("parsing document: %w", err)
	}

	// If result is nil, this document is not a motion (motie)
	if result == nil {
		log.Printf("Document at %s is not a motion (title doesn't contain 'motie'), skipping for dossier %s",
			docResponse.URL, imp.formatDossierNumber(dossier))
		return nil
	}

	log.Printf("Confirmed motion document: '%s' for dossier %s", result.Title, imp.formatDossierNumber(dossier))

	if len(result.BulletPoints) == 0 {
		return nil
	}

	// Convert to JSON bytes for proper JSONB storage
	bulletPointsJSON, err := json.Marshal(result.BulletPoints)
	if err != nil {
		return fmt.Errorf("marshaling bullet points: %w", err)
	}

	if err := imp.storage.UpdateKamerstukdossierBulletPoints(ctx, dossier.ID, string(bulletPointsJSON), result.URL); err != nil {
		return fmt.Errorf("updating bullet points: %w", err)
	}

	log.Printf("Successfully stored %d bullet points for motion '%s' (dossier %s)",
		len(result.BulletPoints), result.Title, imp.formatDossierNumber(dossier))
	return nil
}

func (imp *SimpleImporter) getMotieVolgnummers(dossier models.Kamerstukdossier) []int {
	// Get all volgnummers for documents that have Onderwerp starting with 'Motie' (case-insensitive)
	var volgnummers []int
	for _, doc := range dossier.Document {
		onderwerp := strings.ToLower(doc.Onderwerp)
		if strings.HasPrefix(onderwerp, "motie") || strings.Contains(onderwerp, "motie ") {
			log.Printf("Found motion document: '%s' (volgnummer: %d)", doc.Onderwerp, doc.Volgnummer)
			volgnummers = append(volgnummers, doc.Volgnummer)
		}
	}

	// Sort in descending order so we try highest volgnummer first
	for i := 0; i < len(volgnummers)-1; i++ {
		for j := i + 1; j < len(volgnummers); j++ {
			if volgnummers[i] < volgnummers[j] {
				volgnummers[i], volgnummers[j] = volgnummers[j], volgnummers[i]
			}
		}
	}

	return volgnummers
}

func (imp *SimpleImporter) formatDossierNumber(dossier models.Kamerstukdossier) string {
	nummer := dossier.Nummer.String()
	if dossier.Toevoeging != nil && *dossier.Toevoeging != "" {
		return fmt.Sprintf("%s-%s", nummer, *dossier.Toevoeging)
	}
	return nummer
}
