package importer

import (
	"context"
	"encoding/json"
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

func NewImporterWithConfig(client *odata.Client, storage storage.Storage, concurrency, batchSize int) *Importer {
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
	log.Println("Starting high-performance import of motions with votes...")
	startTime := time.Now()

	// concurrent fetching of all data
	log.Printf("Phase 1: Concurrent fetching with %d workers...", imp.concurrency)
	allZaken, err := imp.fetchAllMotionsConcurrent(ctx)
	if err != nil {
		return fmt.Errorf("concurrent fetching failed: %w", err)
	}

	fetchDuration := time.Since(startTime)
	log.Printf("Phase 1 complete: Fetched %d motions in %v (%.2f motions/sec)",
		len(allZaken), fetchDuration, float64(len(allZaken))/fetchDuration.Seconds())

	// batch processing and storage
	log.Printf("Phase 2: Batch processing with batch size %d...", imp.batchSize)
	processStart := time.Now()
	err = imp.processAllMotionsBatch(ctx, allZaken)
	if err != nil {
		return fmt.Errorf("batch processing failed: %w", err)
	}

	processDuration := time.Since(processStart)
	totalDuration := time.Since(startTime)
	log.Printf("Phase 2 complete: Processed %d motions in %v (%.2f motions/sec)",
		len(allZaken), processDuration, float64(len(allZaken))/processDuration.Seconds())
	log.Printf("Total import time: %v (%.2f motions/sec overall)",
		totalDuration, float64(len(allZaken))/totalDuration.Seconds())

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
	return allZaken, nil
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

func (imp *Importer) processAllMotionsBatch(ctx context.Context, allZaken []odata.Zaak) error {
	var allBesluiten []odata.Besluit
	var allStemmingen []odata.Stemming
	var allVotingResults []odata.VotingResult
	var allIndividueleStemmingen []odata.IndividueleStemming

	for _, zaak := range allZaken {
		imp.stats.IncrementZaakType(zaak.Soort)

		for _, besluit := range zaak.Besluit {
			allBesluiten = append(allBesluiten, besluit)
			imp.stats.TotalBesluiten++

			votingResult := imp.createVotingResult(besluit)
			allVotingResults = append(allVotingResults, votingResult)

			for _, stemming := range besluit.Stemming {
				allStemmingen = append(allStemmingen, stemming)
				imp.stats.TotalStemmingen++

				// Create individual vote
				individueleStemming := imp.createIndividueleStemming(stemming)
				allIndividueleStemmingen = append(allIndividueleStemmingen, individueleStemming)
			}
		}
	}

	log.Printf("Extracted entities: %d zaken, %d besluiten, %d stemmingen, %d voting results, %d individual votes",
		len(allZaken), len(allBesluiten), len(allStemmingen), len(allVotingResults), len(allIndividueleStemmingen))

	if err := imp.batchSaveZaken(ctx, allZaken); err != nil {
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
		return fmt.Errorf("batch saving individual votes: %w", err)
	}

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
	}

	if len(besluit.Zaak) > 0 {
		firstZaak := besluit.Zaak[0]
		result.ZaakID = firstZaak.ID
		result.ZaakNummer = firstZaak.Nummer
		result.ZaakTitel = firstZaak.Titel
		result.ZaakSoort = firstZaak.Soort
	}

	return result
}

func (imp *Importer) createIndividueleStemming(stemming odata.Stemming) odata.IndividueleStemming {
	vote := odata.IndividueleStemming{
		PersonID:     getStringValue(stemming.SidActorLid),
		PersonName:   stemming.ActorNaam,
		FractieID:    stemming.SidActorFractie,
		FractieName:  stemming.ActorFractie,
		VoteType:     stemming.Soort,
		IsCorrection: stemming.Vergissing,
		Date:         &stemming.GewijzigdOp,
	}

	if stemming.Persoon != nil {
		vote.PersonID = stemming.Persoon.ID
		vote.PersonName = fmt.Sprintf("%s %s", stemming.Persoon.Voornamen, stemming.Persoon.Achternaam)
	}

	if stemming.Fractie != nil {
		vote.FractieID = stemming.Fractie.ID
		vote.FractieName = stemming.Fractie.NaamNL
	}

	if stemming.Besluit != nil {
		vote.BesluitID = stemming.Besluit.ID
		vote.BesluitTekst = stemming.Besluit.BesluitTekst

		if len(stemming.Besluit.Zaak) > 0 {
			zaak := stemming.Besluit.Zaak[0]
			vote.ZaakID = zaak.ID
			vote.ZaakNummer = zaak.Nummer
			vote.ZaakTitel = zaak.Titel
		}
	}

	return vote
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
