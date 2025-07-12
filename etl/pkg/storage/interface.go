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

	SaveODataZaakBatch(ctx context.Context, zaken []interface{}) error
	SaveODataBesluitBatch(ctx context.Context, besluiten []interface{}) error
	SaveODataStemmingBatch(ctx context.Context, stemmingen []interface{}) error
	SaveVotingResultBatch(ctx context.Context, results []interface{}) error
	SaveIndividueleStemingBatch(ctx context.Context, votes []interface{}) error

	Close() error
}

func NewStorage(config config.StorageConfig) (Storage, error) {
	return NewGormPostgresStorage(config)
}
