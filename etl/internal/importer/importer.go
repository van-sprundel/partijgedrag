package importer

import (
	"context"
	"encoding/json"
	"etl/pkg/odata"
	"etl/pkg/storage"
	"fmt"
	"log"
)

type Importer struct {
	client  *odata.Client
	storage storage.Storage
	stats   *odata.ImportStats
}

func NewImporter(client *odata.Client, storage storage.Storage) *Importer {
	return &Importer{
		client:  client,
		storage: storage,
		stats:   odata.NewImportStats(),
	}
}

func (imp *Importer) GetStats() *odata.ImportStats {
	return imp.stats
}

// ImportMotiesWithVotes imports all motions with their associated decisions and votes
func (imp *Importer) ImportMotiesWithVotes(ctx context.Context) error {
	log.Println("Starting import of motions with votes...")

	skip := 0
	totalProcessed := 0
	var nextLink string

	for {
		select {
		case <-ctx.Done():
			log.Printf("Import cancelled, processed %d motions", totalProcessed)
			return ctx.Err()
		default:
		}

		var data []byte
		var err error

		if nextLink != "" {
			log.Printf("Fetching motions page with nextLink")
			data, err = imp.client.MakeRequest(ctx, nextLink)
		} else {
			log.Printf("Fetching motions page with skip: %d", skip)
			data, err = imp.client.GetMotiesWithVotes(ctx, skip, 0) // pageSize not used
		}

		if err != nil {
			return fmt.Errorf("fetching motions: %w", err)
		}

		response, err := odata.ParseODataResponse(data)
		if err != nil {
			return fmt.Errorf("parsing OData response: %w", err)
		}

		// Parse the value as an array of Zaak objects
		zakenData, err := json.Marshal(response.Value)
		if err != nil {
			return fmt.Errorf("marshalling zaken data: %w", err)
		}

		var zaken []odata.Zaak
		if err := json.Unmarshal(zakenData, &zaken); err != nil {
			return fmt.Errorf("unmarshalling zaken: %w", err)
		}

		if len(zaken) == 0 {
			log.Println("No more motions to process")
			break
		}

		// Process each zaak (motion)
		for _, zaak := range zaken {
			if err := imp.processZaak(ctx, zaak); err != nil {
				log.Printf("Error processing zaak %s: %v", zaak.ID, err)
				imp.stats.AddError(fmt.Sprintf("Error processing zaak %s: %v", zaak.ID, err))
			}
		}

		totalProcessed += len(zaken)
		log.Printf("Processed %d motions in this batch, %d total", len(zaken), totalProcessed)

		// Check if there are more pages
		if !response.HasNextPage() {
			log.Printf("No more pages, import complete. Total processed: %d", totalProcessed)
			break
		}

		// Use the next link for proper skiptoken pagination
		nextLink = response.NextLink
		skip += len(zaken) // Update skip based on actual items received
	}

	imp.stats.Finalize()
	return nil
}

func (imp *Importer) processZaak(ctx context.Context, zaak odata.Zaak) error {
	log.Printf("Processing zaak: %s - %s (%s)", zaak.Nummer, zaak.Titel, zaak.Soort)

	if err := imp.storage.SaveODataZaak(ctx, zaak); err != nil {
		return fmt.Errorf("saving zaak: %w", err)
	}

	// associated decisions
	for _, besluit := range zaak.Besluit {
		if err := imp.processBesluit(ctx, besluit); err != nil {
			log.Printf("Error processing besluit %s for zaak %s: %v", besluit.ID, zaak.ID, err)
		}
	}

	imp.stats.IncrementZaakType(zaak.Soort)

	return nil
}

func (imp *Importer) processBesluit(ctx context.Context, besluit odata.Besluit) error {
	log.Printf("Processing besluit: %s - %s", besluit.ID, besluit.BesluitTekst)

	if err := imp.storage.SaveODataBesluit(ctx, besluit); err != nil {
		return fmt.Errorf("saving besluit: %w", err)
	}

	for _, stemming := range besluit.Stemming {
		if err := imp.processStemming(ctx, stemming); err != nil {
			log.Printf("Error processing stemming %s for besluit %s: %v", stemming.ID, besluit.ID, err)
		}
	}

	imp.stats.TotalBesluiten++

	// create and save voting result
	votingResult := imp.createVotingResult(besluit)
	if err := imp.storage.SaveVotingResult(ctx, votingResult); err != nil {
		log.Printf("Error saving voting result: %v", err)
	}

	return nil
}

func (imp *Importer) processStemming(ctx context.Context, stemming odata.Stemming) error {
	log.Printf("Processing stemming: %s - %s", stemming.ID, stemming.Soort)

	if err := imp.storage.SaveODataStemming(ctx, stemming); err != nil {
		return fmt.Errorf("saving stemming: %w", err)
	}

	individueleStemming := imp.createIndividueleStemming(stemming)
	if err := imp.storage.SaveIndividueleStemming(ctx, individueleStemming); err != nil {
		log.Printf("Error saving individual vote: %v", err)
	}

	imp.stats.TotalStemmingen++

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

	// check if related zaak
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
			zaak := stemming.Besluit.Zaak[0] // TODO we should get the correct zaak, not just the first
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
