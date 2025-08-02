package storage

import (
	"context"
	"etl/internal/config"
	"etl/internal/models"
)

type Storage interface {
	// Core entity operations
	SaveZaken(ctx context.Context, zaken []models.Zaak) error
	SaveBesluiten(ctx context.Context, besluiten []models.Besluit) error
	SaveStemmingen(ctx context.Context, stemmingen []models.Stemming) error
	SavePersonen(ctx context.Context, personen []models.Persoon) error
	SaveFracties(ctx context.Context, fracties []models.Fractie) error
	SaveZaakActors(ctx context.Context, zaakActors []models.ZaakActor) error
	SaveKamerstukdossiers(ctx context.Context, dossiers []models.Kamerstukdossier) error

	// Kamerstukdossier operations
	UpdateKamerstukdossierBulletPoints(ctx context.Context, id string, bulletPointsJSON string) error

	// Utility operations
	Close() error
	Ping(ctx context.Context) error
}

func NewStorage(config config.StorageConfig) (Storage, error) {
	return NewPostgresStorage(config)
}
