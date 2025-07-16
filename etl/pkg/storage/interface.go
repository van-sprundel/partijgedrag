package storage

import (
	"context"
	"etl/internal/config"
)

type Storage interface {
	SaveODataZaak(ctx context.Context, zaak interface{}) error
	SaveODataBesluit(ctx context.Context, besluit interface{}) error
	SaveODataStemming(ctx context.Context, stemming interface{}) error
	SaveVotingResult(ctx context.Context, result interface{}) error
	SaveIndividueleStemming(ctx context.Context, vote interface{}) error
	SaveKamerstukdossier(ctx context.Context, dossier interface{}) error
	SaveDocumentInfo(ctx context.Context, docInfo interface{}) error

	SaveODataZaakBatch(ctx context.Context, zaken []interface{}) error
	SaveODataBesluitBatch(ctx context.Context, besluiten []interface{}) error
	SaveODataStemmingBatch(ctx context.Context, stemmingen []interface{}) error
	SaveVotingResultBatch(ctx context.Context, results []interface{}) error
	SaveIndividueleStemingBatch(ctx context.Context, votes []interface{}) error
	SaveKamerstukdossierBatch(ctx context.Context, dossiers []interface{}) error
	SaveDocumentInfoBatch(ctx context.Context, docInfos []interface{}) error

	DocumentExists(ctx context.Context, dossierNummer string, volgnummer int) (bool, error)
	GetExistingDocumentNumbers(ctx context.Context, dossierNummer string) ([]int, error)

	Close() error
}

func NewStorage(config config.StorageConfig) (Storage, error) {
	return NewGormPostgresStorage(config)
}
