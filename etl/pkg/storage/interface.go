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

	Close() error
}

func NewStorage(config config.StorageConfig) (Storage, error) {
	// Use GORM-based implementation for cleaner code
	return NewGormPostgresStorage(config)
}
