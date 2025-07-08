package main

import (
	"context"
	"encoding/json"
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

	// Display final statistics
	stats := imp.GetStats()
	fmt.Println("\n=== Import Statistics ===")
	fmt.Printf("Total documents processed: %d\n", stats.TotalProcessed)
	fmt.Printf("Successfully parsed: %d\n", stats.SuccessfulParsed)
	fmt.Printf("Parse errors: %d\n", stats.ParseErrors)
	fmt.Printf("Storage errors: %d\n", stats.StorageErrors)

	if len(stats.DocumentTypes) > 0 {
		fmt.Println("\nDocument types encountered:")
		for docType, count := range stats.DocumentTypes {
			fmt.Printf("  %s: %d\n", docType, count)
		}
	}

	if len(stats.ErrorsByCategory) > 0 {
		fmt.Println("\nErrors by category:")
		for category, count := range stats.ErrorsByCategory {
			fmt.Printf("  %s: %d\n", category, count)
		}
	}

	if len(stats.ErrorDetails) > 0 {
		fmt.Printf("\nShowing last %d errors:\n", min(5, len(stats.ErrorDetails)))
		start := max(0, len(stats.ErrorDetails)-5)
		for i := start; i < len(stats.ErrorDetails); i++ {
			err := stats.ErrorDetails[i]
			fmt.Printf("  [%s] %s: %s - %s\n", err.Timestamp, err.DocumentID, err.ErrorType, err.ErrorMessage)
		}
	}

	// Save detailed stats to file
	if statsData, err := json.MarshalIndent(stats, "", "  "); err == nil {
		if err := os.WriteFile("import_stats.json", statsData, 0644); err == nil {
			fmt.Println("\nDetailed statistics saved to import_stats.json")
		}
	}

	fmt.Println("\nImport completed successfully")
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
