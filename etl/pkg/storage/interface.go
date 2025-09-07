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
	UpdateKamerstukdossierBulletPoints(ctx context.Context, id string, bulletPointsJSON string, documentURL string) error

	SaveCategories(ctx context.Context, categories []models.MotionCategory) error
	SaveZaakCategories(ctx context.Context, zaakCategories []models.ZaakCategory) error

	GetAllCategories(ctx context.Context) ([]models.MotionCategory, error)
	GetCategoriesByType(ctx context.Context, categoryType string) ([]models.MotionCategory, error)

	AssignCategoryToZaak(ctx context.Context, zaakID, categoryID string) error
	RemoveCategoryFromZaak(ctx context.Context, zaakID, categoryID string) error

	GetZakenForEnrichment(ctx context.Context) ([]models.Zaak, error)

	// Utility operations
	Close() error
	Ping(ctx context.Context) error
}

func NewStorage(config config.StorageConfig) (Storage, error) {
	return NewPostgresStorage(config)
}
