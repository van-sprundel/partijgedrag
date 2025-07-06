package main

import (
	"context"
	"etl/internal/api"
	"etl/internal/config"
	"etl/internal/parser"
	"etl/pkg/importer"
	"etl/pkg/storage"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"gopkg.in/yaml.v3"
)

func main() {
	cfg, err := loadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Received shutdown signal, stopping...")
		cancel()
	}()

	client := api.NewClient()
	parser := parser.NewParser()

	// Create storage based on configuration
	storageInstance, err := storage.NewStorage(storage.StorageConfig{
		Type: cfg.Storage.Type,
		Path: cfg.Storage.Path,
	})
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer storageInstance.Close()

	imp := importer.New(client, parser, storageInstance)

	for _, category := range cfg.Categories {
		if err := imp.ImportCategory(ctx, category, ""); err != nil {
			log.Fatalf("Import failed for %s: %v", category, err)
		}
		log.Printf("Successfully imported %s", category)
	}

	fmt.Println("Import completed successfully")
}

func loadConfig(path string) (*config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
