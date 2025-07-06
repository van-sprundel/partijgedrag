package storage

import (
	"context"
	"etl/internal/models"
)

type Storage interface {
	SaveKamerstukdossier(ctx context.Context, dossier *models.Kamerstukdossier) error
	SaveKamerstukdossiers(ctx context.Context, dossiers []*models.Kamerstukdossier) error
	SaveDocument(ctx context.Context, dossierNummer string, volgnummer int, data []byte) error
	SaveParsedDocument(ctx context.Context, doc *models.ParsedDocument) error
	GetKamerstukdossier(ctx context.Context, nummer string) (*models.Kamerstukdossier, error)
	GetParsedDocument(ctx context.Context, id string) (*models.ParsedDocument, error)

	Close() error
}

type StorageConfig struct {
	Type     string `yaml:"type"`
	Path     string `yaml:"path"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func NewStorage(config StorageConfig) (Storage, error) {
	switch config.Type {
	case "file":
		return NewFileStorage(config.Path), nil
	// TODO:
	// case "postgres":
	//     return NewPostgresStorage(config)
	default:
		return NewFileStorage(config.Path), nil
	}
}
