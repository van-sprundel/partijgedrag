package importer

import (
	"context"
	"fmt"
	"log"

	"etl/internal/api"
	"etl/internal/models"
	"etl/internal/parser"
	"etl/pkg/storage"
)

type Importer struct {
	client  *api.Client
	parser  *parser.Parser
	storage storage.Storage
}

func New(client *api.Client, parser *parser.Parser, storage storage.Storage) *Importer {
	return &Importer{
		client:  client,
		parser:  parser,
		storage: storage,
	}
}

// imports entities for a specific category
func (i *Importer) ImportCategory(ctx context.Context, category string, startSkiptoken string) error {
	skiptoken := startSkiptoken
	totalProcessed := 0

	log.Printf("Starting import for category: %s", category)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Import cancelled, processed %d entries", totalProcessed)
			return ctx.Err()
		default:
		}

		log.Printf("Fetching page with skiptoken: %s", skiptoken)
		feedData, err := i.client.FetchFeed(ctx, category, skiptoken)
		if err != nil {
			return fmt.Errorf("fetching feed: %w", err)
		}

		feed, err := i.parser.ParseFeed(feedData)
		if err != nil {
			return fmt.Errorf("parsing feed: %w", err)
		}

		kamerstukdossiers := i.parser.ExtractKamerstukdossiers(feed)

		if len(kamerstukdossiers) > 0 {
			if err := i.storage.SaveKamerstukdossiers(ctx, kamerstukdossiers); err != nil {
				log.Printf("Error saving dossiers batch: %v", err)
			} else {
				log.Printf("Saved batch of %d dossiers", len(kamerstukdossiers))
			}
		}

		for _, dossier := range kamerstukdossiers {
			if err := i.processDossier(ctx, dossier); err != nil {
				log.Printf("Error processing dossier %s: %v", dossier.Nummer, err)
				continue
			}
		}

		totalProcessed += len(kamerstukdossiers)
		log.Printf("Processed %d entries in this batch, %d total", len(kamerstukdossiers), totalProcessed)

		paginationInfo := i.parser.GetPaginationInfo(feed)
		if !paginationInfo.HasMore {
			log.Printf("No more pages, import complete. Total processed: %d", totalProcessed)
			break
		}

		skiptoken = paginationInfo.NextSkiptoken

		// Rate limiting: Add delay between requests if API starts rate limiting
		// Consider implementing exponential backoff or concurrent processing with rate limits
		// time.Sleep(100 * time.Millisecond)
	}

	return nil
}

func (i *Importer) processDossier(ctx context.Context, dossier *models.Kamerstukdossier) error {
	log.Printf("Processing dossier: %s - %s", dossier.Nummer, dossier.Titel)

	// store the dossier metadata
	if err := i.storage.SaveKamerstukdossier(ctx, dossier); err != nil {
		log.Printf("Error saving dossier metadata: %v", err)
	}

	// fetch and store detailed documents
	if dossier.HoogsteVolgnummer > 0 {
		toevoeging := ""
		if dossier.Toevoeging != nil {
			toevoeging = *dossier.Toevoeging
		}
		docData, err := i.client.FetchDocument(ctx, dossier.Nummer, toevoeging, 1)
		if err != nil {
			return fmt.Errorf("fetching document: %w", err)
		}

		log.Printf("Fetched document for %s, size: %d bytes", dossier.Nummer, len(docData))

		if err := i.storage.SaveDocument(ctx, dossier.Nummer, 1, docData); err != nil {
			log.Printf("Error saving document: %v", err)
		}

		parsedDoc, err := i.parser.ParseDocument(docData)
		if err != nil {
			log.Printf("Error parsing document %s: %v", dossier.Nummer, err)
		} else {
			// save parsed document (this contains the full text)
			if err := i.storage.SaveParsedDocument(ctx, parsedDoc); err != nil {
				log.Printf("Error saving parsed document %s: %v", parsedDoc.ID, err)
			} else {
				log.Printf("Successfully parsed and saved document %s", parsedDoc.ID)
			}
		}
	}

	return nil
}

// import a specific dossier by nummer
func (i *Importer) ImportSingleDossier(ctx context.Context, nummer string, toevoeging string) error {
	log.Printf("Importing single dossier: %s", nummer)

	for volgnummer := 1; volgnummer <= 10; volgnummer++ { // Conservative limit: check API docs for actual max
		docData, err := i.client.FetchDocument(ctx, nummer, toevoeging, volgnummer)
		if err != nil {
			log.Printf("Document %s-%d not found, stopping", nummer, volgnummer)
			break
		}

		log.Printf("Processing document %s-%d", nummer, volgnummer)

		parsedDoc, err := i.parser.ParseDocument(docData)
		if err != nil {
			log.Printf("Error parsing document %s-%d: %v", nummer, volgnummer, err)
			continue
		}

		// save the document (has the full text)
		if err := i.storage.SaveParsedDocument(ctx, parsedDoc); err != nil {
			log.Printf("Error saving parsed document %s: %v", parsedDoc.ID, err)
		} else {
			log.Printf("Successfully parsed and saved document %s", parsedDoc.ID)
		}

		// Rate limiting
		// time.Sleep(50 * time.Millisecond)
	}

	return nil
}
