package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"etl/internal/models"
)

type FileStorage struct {
	basePath string
}

func NewFileStorage(basePath string) *FileStorage {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		panic(fmt.Sprintf("failed to create base storage directory: %v", err))
	}

	return &FileStorage{
		basePath: basePath,
	}
}

func (fs *FileStorage) ensureDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

func (fs *FileStorage) buildFilePath(subPath, filename string) (string, error) {
	fullPath := filepath.Join(fs.basePath, subPath)
	if err := fs.ensureDir(fullPath); err != nil {
		return "", err
	}
	return filepath.Join(fullPath, filename), nil
}

// save a single kamerstukdossier
func (fs *FileStorage) SaveKamerstukdossier(ctx context.Context, dossier *models.Kamerstukdossier) error {
	filename := fmt.Sprintf("dossier_%s.json", dossier.Nummer)
	if dossier.Toevoeging != nil && *dossier.Toevoeging != "" {
		filename = fmt.Sprintf("dossier_%s_%s.json", dossier.Nummer, *dossier.Toevoeging)
	}

	filePath, err := fs.buildFilePath("kamerstukdossiers", filename)
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(dossier, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal dossier: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write dossier file: %w", err)
	}

	return nil
}

// save multiple kamerstukdossiers in a batch
func (fs *FileStorage) SaveKamerstukdossiers(ctx context.Context, dossiers []*models.Kamerstukdossier) error {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("dossiers_batch_%s.json", timestamp)
	filePath, err := fs.buildFilePath("kamerstukdossiers", filename)
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(dossiers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal dossiers batch: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write dossiers batch file: %w", err)
	}

	return nil
}

// save a document with its metadata
func (fs *FileStorage) SaveDocument(ctx context.Context, dossierNummer string, volgnummer int, data []byte) error {
	filename := fmt.Sprintf("document_%s_%d.xml", dossierNummer, volgnummer)
	filePath, err := fs.buildFilePath(filepath.Join("documents", dossierNummer), filename)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write document file: %w", err)
	}

	return nil
}

// retrieve a kamerstukdossier by its nummer
func (fs *FileStorage) GetKamerstukdossier(ctx context.Context, nummer string) (*models.Kamerstukdossier, error) {
	filename := fmt.Sprintf("dossier_%s.json", nummer)
	filePath := filepath.Join(fs.basePath, "kamerstukdossiers", filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read dossier file: %w", err)
	}

	var dossier models.Kamerstukdossier
	if err := json.Unmarshal(data, &dossier); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dossier: %w", err)
	}

	return &dossier, nil
}

// save a parsed document structure
func (fs *FileStorage) SaveParsedDocument(ctx context.Context, doc *models.ParsedDocument) error {
	filename := fmt.Sprintf("parsed_doc_%s.json", doc.ID)
	filePath, err := fs.buildFilePath("parsed_documents", filename)
	if err != nil {
		return err
	}

	jsonData, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal parsed document: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write parsed document file: %w", err)
	}

	return nil
}

// retrieve a parsed document by its ID
func (fs *FileStorage) GetParsedDocument(ctx context.Context, id string) (*models.ParsedDocument, error) {
	filename := fmt.Sprintf("parsed_doc_%s.json", id)
	filePath := filepath.Join(fs.basePath, "parsed_documents", filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read parsed document file: %w", err)
	}

	var doc models.ParsedDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parsed document: %w", err)
	}

	return &doc, nil
}

func (fs *FileStorage) Close() error {
	return nil
}
