package importer

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"etl/internal/models"
	"etl/pkg/odata"
	"etl/pkg/storage"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

type Importer struct {
	client      *odata.Client
	storage     storage.Storage
	stats       *odata.ImportStats
	concurrency int
	batchSize   int
}

func NewImporter(client *odata.Client, storage storage.Storage) *Importer {
	return &Importer{
		client:      client,
		storage:     storage,
		stats:       odata.NewImportStats(),
		concurrency: runtime.NumCPU() * 2,
		batchSize:   1000, // for DB operations
	}
}

func NewImporterWithConfig(client *odata.Client, storage storage.Storage, concurrency int, batchSize int) *Importer {
	return &Importer{
		client:      client,
		storage:     storage,
		stats:       odata.NewImportStats(),
		concurrency: concurrency,
		batchSize:   batchSize,
	}
}

func (imp *Importer) GetStats() *odata.ImportStats {
	return imp.stats
}

func (imp *Importer) ImportMotiesWithVotes(ctx context.Context) error {
	log.Println("Starting streaming import of motions with votes...")
	startTime := time.Now()

	log.Printf("Starting streaming fetch and process with %d workers...", imp.concurrency)
	err := imp.fetchAndProcessMotionsStreaming(ctx)
	if err != nil {
		return fmt.Errorf("streaming import failed: %w", err)
	}

	totalDuration := time.Since(startTime)
	log.Printf("Streaming import complete in %v", totalDuration)

	imp.stats.Finalize()
	return nil
}

func (imp *Importer) fetchAllMotionsConcurrent(ctx context.Context) ([]odata.Zaak, error) {
	firstPageData, err := imp.client.GetMotiesWithVotes(ctx, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("fetching first page: %w", err)
	}

	firstResponse, err := odata.ParseODataResponse(firstPageData)
	if err != nil {
		return nil, fmt.Errorf("parsing first page response: %w", err)
	}

	firstZakenData, err := json.Marshal(firstResponse.Value)
	if err != nil {
		return nil, fmt.Errorf("marshalling first zaken data: %w", err)
	}

	var firstZaken []odata.Zaak
	if err := json.Unmarshal(firstZakenData, &firstZaken); err != nil {
		return nil, fmt.Errorf("unmarshalling first zaken: %w", err)
	}

	if len(firstZaken) == 0 {
		return []odata.Zaak{}, nil
	}

	type fetchJob struct {
		skip     int
		nextLink string
		pageNum  int
	}

	type fetchResult struct {
		zaken   []odata.Zaak
		pageNum int
		err     error
	}

	jobs := make(chan fetchJob, imp.concurrency*2)
	results := make(chan fetchResult, imp.concurrency*2)

	var wg sync.WaitGroup
	for i := 0; i < imp.concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobs {
				select {
				case <-ctx.Done():
					results <- fetchResult{err: ctx.Err()}
					return
				default:
				}

				zaken, err := imp.fetchPage(ctx, job.skip, job.nextLink)
				results <- fetchResult{
					zaken:   zaken,
					pageNum: job.pageNum,
					err:     err,
				}
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	pageSize := len(firstZaken)
	currentSkip := pageSize
	currentNextLink := firstResponse.NextLink
	pageNum := 1

	go func() {
		defer close(jobs)

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if currentNextLink == "" && currentSkip == pageSize {
				// no more pages
				break
			}

			jobs <- fetchJob{
				skip:     currentSkip,
				nextLink: currentNextLink,
				pageNum:  pageNum,
			}

			currentSkip += pageSize
			currentNextLink = "" // get the actual nextLink in response
			pageNum++

			// safety limit, prob not needed
			// if pageNum > 100 {
			// 	break
			// }
		}
	}()

	allZaken := make([]odata.Zaak, 0)
	allZaken = append(allZaken, firstZaken...)

	pagesReceived := make(map[int][]odata.Zaak)
	nextExpectedPage := 1
	totalFetched := len(firstZaken)

	for result := range results {
		if result.err != nil {
			if result.err == ctx.Err() {
				break
			}
			log.Printf("Error fetching page %d: %v", result.pageNum, result.err)
			continue
		}

		if len(result.zaken) == 0 {
			log.Printf("Page %d returned no results, stopping", result.pageNum)
			break
		}

		pagesReceived[result.pageNum] = result.zaken
		totalFetched += len(result.zaken)

		for {
			if pageData, exists := pagesReceived[nextExpectedPage]; exists {
				allZaken = append(allZaken, pageData...)
				delete(pagesReceived, nextExpectedPage)
				nextExpectedPage++
				log.Printf("Processed page %d, total fetched: %d", nextExpectedPage-1, totalFetched)
			} else {
				break
			}
		}

		if len(result.zaken) < pageSize {
			log.Printf("Page %d had %d items (< %d), likely last page", result.pageNum, len(result.zaken), pageSize)
			break
		}
	}

	log.Printf("Concurrent fetch complete: %d motions total", len(allZaken))
	return allZaken, err
}

// fetch and processes motions page by page to avoid memory issues
func (imp *Importer) fetchAndProcessMotionsStreaming(ctx context.Context) error {
	skip := 0
	pageNum := 1
	totalProcessed := 0

	for {
		log.Printf("Fetching and processing page %d (skip=%d)...", pageNum, skip)

		pageZaken, err := imp.fetchPage(ctx, skip, "")
		if err != nil {
			return fmt.Errorf("fetching page %d: %w", pageNum, err)
		}

		if len(pageZaken) == 0 {
			log.Printf("No more data found, streaming complete")
			break
		}

		if err := imp.processMotionsBatch(ctx, pageZaken); err != nil {
			return fmt.Errorf("processing page %d: %w", pageNum, err)
		}

		totalProcessed += len(pageZaken)
		log.Printf("Page %d complete: processed %d motions (total: %d)", pageNum, len(pageZaken), totalProcessed)

		skip += len(pageZaken)
		pageNum++

		// time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Streaming import complete: processed %d total motions across %d pages", totalProcessed, pageNum-1)
	return nil
}

func (imp *Importer) fetchPage(ctx context.Context, skip int, nextLink string) ([]odata.Zaak, error) {
	var data []byte
	var err error

	if nextLink != "" {
		data, err = imp.client.MakeRequest(ctx, nextLink)
	} else {
		data, err = imp.client.GetMotiesWithVotes(ctx, skip, 0)
	}

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

	var zaken []odata.Zaak
	if err := json.Unmarshal(zakenData, &zaken); err != nil {
		return nil, err
	}

	return zaken, nil
}

func (imp *Importer) processMotionsBatch(ctx context.Context, zaken []odata.Zaak) error {
	if len(zaken) == 0 {
		return nil
	}

	log.Printf("Processing batch of %d motions...", len(zaken))

	var allBesluiten []odata.Besluit
	var allStemmingen []odata.Stemming
	var allVotingResults []odata.VotingResult
	var allIndividueleStemmingen []odata.IndividueleStemming
	kamerstukdossierMap := make(map[string]odata.Kamerstukdossier) // deduplicate by ID
	var allDocumentInfos []odata.DocumentInfo

	for _, zaak := range zaken {
		imp.stats.IncrementZaakType(zaak.Soort)

		if len(zaak.Kamerstukdossier) > 0 {
			for _, dossier := range zaak.Kamerstukdossier {
				kamerstukdossierMap[dossier.ID] = dossier

				docInfos := imp.fetchDocumentsForDossier(ctx, zaak.ID, dossier)
				allDocumentInfos = append(allDocumentInfos, docInfos...)
			}
		}

		for _, besluit := range zaak.Besluit {
			besluit.ZaakID = zaak.ID
			allBesluiten = append(allBesluiten, besluit)
			imp.stats.TotalBesluiten++

			votingResult := imp.createVotingResult(besluit)
			allVotingResults = append(allVotingResults, votingResult)

			for _, stemming := range besluit.Stemming {
				stemming.BesluitID = besluit.ID
				allStemmingen = append(allStemmingen, stemming)
				imp.stats.TotalStemmingen++

				individueleStemming := imp.createIndividueleStemming(stemming)
				allIndividueleStemmingen = append(allIndividueleStemmingen, individueleStemming)
			}
		}
	}

	allKamerstukdossiers := make([]odata.Kamerstukdossier, 0, len(kamerstukdossierMap))
	for _, dossier := range kamerstukdossierMap {
		allKamerstukdossiers = append(allKamerstukdossiers, dossier)
	}

	log.Printf("Extracted from batch: %d zaken, %d besluiten, %d stemmingen, %d voting results, %d individual votes, %d kamerstukdossiers (deduplicated), %d documents",
		len(zaken), len(allBesluiten), len(allStemmingen), len(allVotingResults), len(allIndividueleStemmingen), len(allKamerstukdossiers), len(allDocumentInfos))

	log.Printf("Saving batch to database...")

	if err := imp.batchSaveZaken(ctx, zaken); err != nil {
		return fmt.Errorf("batch saving zaken: %w", err)
	}

	if err := imp.batchSaveBesluiten(ctx, allBesluiten); err != nil {
		return fmt.Errorf("batch saving besluiten: %w", err)
	}

	if err := imp.batchSaveStemmingen(ctx, allStemmingen); err != nil {
		return fmt.Errorf("batch saving stemmingen: %w", err)
	}

	if err := imp.batchSaveVotingResults(ctx, allVotingResults); err != nil {
		return fmt.Errorf("batch saving voting results: %w", err)
	}

	if err := imp.batchSaveIndividueleStemmingen(ctx, allIndividueleStemmingen); err != nil {
		return fmt.Errorf("batch saving individual stemmingen: %w", err)
	}

	if err := imp.batchSaveKamerstukdossiers(ctx, allKamerstukdossiers); err != nil {
		return fmt.Errorf("batch saving kamerstukdossiers: %w", err)
	}

	if err := imp.batchSaveDocumentInfos(ctx, allDocumentInfos); err != nil {
		return fmt.Errorf("batch saving document infos: %w", err)
	}

	log.Printf("Batch processing completed successfully")
	return nil
}

func (imp *Importer) batchSaveZaken(ctx context.Context, zaken []odata.Zaak) error {
	log.Printf("Batch saving %d zaken...", len(zaken))

	for i := 0; i < len(zaken); i += imp.batchSize {
		end := i + imp.batchSize
		if end > len(zaken) {
			end = len(zaken)
		}

		batch := make([]interface{}, end-i)
		for j, zaak := range zaken[i:end] {
			batch[j] = zaak
		}

		if err := imp.storage.SaveODataZaakBatch(ctx, batch); err != nil {
			return fmt.Errorf("saving zaken batch %d-%d: %w", i, end-1, err)
		}

		log.Printf("Saved zaken batch %d-%d", i, end-1)
	}

	return nil
}

func (imp *Importer) batchSaveBesluiten(ctx context.Context, besluiten []odata.Besluit) error {
	log.Printf("Batch saving %d besluiten...", len(besluiten))

	for i := 0; i < len(besluiten); i += imp.batchSize {
		end := i + imp.batchSize
		if end > len(besluiten) {
			end = len(besluiten)
		}

		batch := make([]interface{}, end-i)
		for j, besluit := range besluiten[i:end] {
			batch[j] = besluit
		}

		if err := imp.storage.SaveODataBesluitBatch(ctx, batch); err != nil {
			return fmt.Errorf("saving besluiten batch %d-%d: %w", i, end-1, err)
		}

		log.Printf("Saved besluiten batch %d-%d", i, end-1)
	}

	return nil
}

func (imp *Importer) batchSaveStemmingen(ctx context.Context, stemmingen []odata.Stemming) error {
	log.Printf("Batch saving %d stemmingen...", len(stemmingen))

	for i := 0; i < len(stemmingen); i += imp.batchSize {
		end := i + imp.batchSize
		if end > len(stemmingen) {
			end = len(stemmingen)
		}

		batch := make([]interface{}, end-i)
		for j, stemming := range stemmingen[i:end] {
			batch[j] = stemming
		}

		if err := imp.storage.SaveODataStemmingBatch(ctx, batch); err != nil {
			return fmt.Errorf("saving stemmingen batch %d-%d: %w", i, end-1, err)
		}

		log.Printf("Saved stemmingen batch %d-%d", i, end-1)
	}

	return nil
}

func (imp *Importer) batchSaveVotingResults(ctx context.Context, results []odata.VotingResult) error {
	log.Printf("Batch saving %d voting results...", len(results))

	for i := 0; i < len(results); i += imp.batchSize {
		end := i + imp.batchSize
		if end > len(results) {
			end = len(results)
		}

		batch := make([]interface{}, end-i)
		for j, result := range results[i:end] {
			batch[j] = result
		}

		if err := imp.storage.SaveVotingResultBatch(ctx, batch); err != nil {
			return fmt.Errorf("saving voting results batch %d-%d: %w", i, end-1, err)
		}

		log.Printf("Saved voting results batch %d-%d", i, end-1)
	}

	return nil
}

func (imp *Importer) batchSaveIndividueleStemmingen(ctx context.Context, votes []odata.IndividueleStemming) error {
	log.Printf("Batch saving %d individual votes...", len(votes))

	for i := 0; i < len(votes); i += imp.batchSize {
		end := i + imp.batchSize
		if end > len(votes) {
			end = len(votes)
		}

		batch := make([]interface{}, end-i)
		for j, vote := range votes[i:end] {
			batch[j] = vote
		}

		if err := imp.storage.SaveIndividueleStemingBatch(ctx, batch); err != nil {
			return fmt.Errorf("saving individual votes batch %d-%d: %w", i, end-1, err)
		}

		log.Printf("Saved individual votes batch %d-%d", i, end-1)
	}

	return nil
}

func (imp *Importer) createVotingResult(besluit odata.Besluit) odata.VotingResult {
	partyVotes := make(map[string]string)

	for _, stemming := range besluit.Stemming {
		if stemming.Fractie != nil {
			partyVotes[stemming.Fractie.NaamNL] = stemming.Soort
		}
	}

	result := odata.VotingResult{
		BesluitID:    besluit.ID,
		BesluitTekst: besluit.BesluitTekst,
		BesluitSoort: besluit.BesluitSoort,
		VotingType:   getStringValue(besluit.StemmingsSoort),
		PartyVotes:   partyVotes,
		Date:         besluit.GewijzigdOp,
		Status:       besluit.Status,
		ZaakID:       besluit.ZaakID,
	}

	return result
}

func (imp *Importer) createIndividueleStemming(stemming odata.Stemming) odata.IndividueleStemming {
	vote := odata.IndividueleStemming{
		PersonName:   stemming.ActorNaam,
		FractieName:  stemming.ActorFractie,
		VoteType:     stemming.Soort,
		IsCorrection: stemming.Vergissing,
		Date:         &stemming.GewijzigdOp,
		BesluitID:    stemming.BesluitID,
	}

	return vote
}

func (imp *Importer) fetchDocumentsForDossier(ctx context.Context, zaakID string, dossier odata.Kamerstukdossier) []odata.DocumentInfo {
	var documentInfos []odata.DocumentInfo

	if dossier.HoogsteVolgnummer <= 0 {
		return []odata.DocumentInfo{}
	}

	skippedCount := 0
	fetchedCount := 0

	volgnummer := dossier.HoogsteVolgnummer

	exists, err := imp.storage.DocumentExists(ctx, dossier.Nummer.String(), volgnummer)
	if err != nil {
		log.Printf("Warning: failed to check existing document for dossier %s-%d: %v", dossier.Nummer.String(), volgnummer, err)
	}

	if exists {
		skippedCount++
		log.Printf("Skipping document %s-%d (already cached)", dossier.Nummer.String(), volgnummer)
	} else {
		docInfo := odata.DocumentInfo{
			ZaakID:        zaakID,
			DossierNummer: dossier.Nummer.String(),
			Volgnummer:    volgnummer,
			URL:           imp.client.BuildDocumentURL(dossier.Nummer.String(), volgnummer),
			FetchedAt:     time.Now(),
			Success:       false,
		}

		xmlData, err := imp.client.FetchDocument(ctx, dossier.Nummer.String(), volgnummer)
		if err != nil {
			log.Printf("Warning: failed to fetch document %s-%d: %v", dossier.Nummer.String(), volgnummer, err)
			docInfo.Error = err.Error()
			documentInfos = append(documentInfos, docInfo)
		} else {
			// parse XML to JSON
			parsedContent, err := imp.parseXMLToJSON(xmlData)
			if err != nil {
				log.Printf("Warning: failed to parse XML for document %s-%d: %v", dossier.Nummer.String(), volgnummer, err)
				docInfo.Error = err.Error()
				documentInfos = append(documentInfos, docInfo)
			} else {
				docInfo.Content = parsedContent
				docInfo.Success = true
				documentInfos = append(documentInfos, docInfo)
				fetchedCount++
			}
		}
	}

	log.Printf("Processed dossier %s: fetched %d document, skipped %d cached document",
		dossier.Nummer.String(), fetchedCount, skippedCount)
	return documentInfos
}

func (imp *Importer) parseXMLToJSON(xmlData []byte) (map[string]interface{}, error) {
	var officielePublicatie models.OfficielePublicatie
	if err := xml.Unmarshal(xmlData, &officielePublicatie); err == nil {
		jsonData, err := json.Marshal(officielePublicatie)
		if err != nil {
			return nil, err
		}

		var result map[string]interface{}
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return nil, err
		}

		result["document_type"] = "officiele_publicatie"
		return result, nil
	}

	var kamerDocument models.KamerDocument
	if err := xml.Unmarshal(xmlData, &kamerDocument); err == nil {
		jsonData, err := json.Marshal(kamerDocument)
		if err != nil {
			return nil, err
		}

		var result map[string]interface{}
		if err := json.Unmarshal(jsonData, &result); err != nil {
			return nil, err
		}

		result["document_type"] = "kamer_document"
		return result, nil
	}

	return map[string]interface{}{
		"document_type": "raw_xml",
		"raw_content":   string(xmlData),
	}, nil
}

func (imp *Importer) batchSaveKamerstukdossiers(ctx context.Context, dossiers []odata.Kamerstukdossier) error {
	if len(dossiers) == 0 {
		return nil
	}

	log.Printf("Batch saving %d kamerstukdossiers...", len(dossiers))

	for i := 0; i < len(dossiers); i += imp.batchSize {
		end := i + imp.batchSize
		if end > len(dossiers) {
			end = len(dossiers)
		}

		batch := make([]interface{}, end-i)
		for j, dossier := range dossiers[i:end] {
			batch[j] = dossier
		}

		if err := imp.storage.SaveKamerstukdossierBatch(ctx, batch); err != nil {
			return fmt.Errorf("saving kamerstukdossiers batch %d-%d: %w", i, end-1, err)
		}

		log.Printf("Saved kamerstukdossiers batch %d-%d", i, end-1)
	}

	return nil
}

func (imp *Importer) batchSaveDocumentInfos(ctx context.Context, docInfos []odata.DocumentInfo) error {
	if len(docInfos) == 0 {
		return nil
	}

	log.Printf("Batch saving %d document infos...", len(docInfos))

	for i := 0; i < len(docInfos); i += imp.batchSize {
		end := i + imp.batchSize
		if end > len(docInfos) {
			end = len(docInfos)
		}

		batch := make([]interface{}, end-i)
		for j, docInfo := range docInfos[i:end] {
			batch[j] = docInfo
		}

		if err := imp.storage.SaveDocumentInfoBatch(ctx, batch); err != nil {
			return fmt.Errorf("saving document infos batch %d-%d: %w", i, end-1, err)
		}

		log.Printf("Saved document infos batch %d-%d", i, end-1)
	}

	return nil
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
