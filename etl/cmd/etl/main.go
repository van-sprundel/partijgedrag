package main

import (
	"context"
	"encoding/json"
	"etl/internal/config"
	"etl/internal/importer"
	"etl/pkg/odata"
	"etl/pkg/storage"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	var (
		configPath = flag.String("config", "configs/config.yaml", "Path to configuration file")
	)
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// parse DATABASE_URL if provided, otherwise use config values
	storageConfig, err := parseStorageConfig(cfg)
	if err != nil {
		log.Fatalf("Failed to parse storage config: %v", err)
	}

	store, err := storage.NewStorage(*storageConfig)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	client := odata.NewClient(cfg.API)

	concurrency := cfg.Import.Concurrency
	if concurrency <= 0 {
		concurrency = 8
	}

	batchSize := cfg.Import.BatchSize
	if batchSize <= 0 {
		batchSize = 1000
	}

	var imp *importer.Importer
	log.Printf("Using high-performance mode: %d workers, batch size %d", concurrency, batchSize)
	imp = importer.NewImporterWithConfig(client, store, concurrency, batchSize)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// gracefu shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received interrupt signal, shutting down...")
		cancel()
	}()

	if cfg.Import.TestMode {
		if err := testMotionsQuery(ctx, client); err != nil {
			log.Fatalf("Test query failed: %v", err)
		}
	} else {
		if err := imp.ImportMotiesWithVotes(ctx); err != nil {
			log.Fatalf("Import failed: %v", err)
		}
	}

	if cfg.Import.ShowStats {
		printStats(imp.GetStats())
	}
}

func loadConfig(path string) (*config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &cfg, nil
}

func parseStorageConfig(cfg *config.Config) (*config.StorageConfig, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// prompt for password if not provided
		if cfg.Storage.Type == "postgres" && cfg.Storage.Password == "" {
			fmt.Print("Database password: ")
			fmt.Scanln(&cfg.Storage.Password)
		}
		return &cfg.Storage, nil
	}

	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing DATABASE_URL: %w", err)
	}

	storageConfig := &config.StorageConfig{
		Type:     "postgres",
		Host:     parsedURL.Hostname(),
		Database: strings.TrimPrefix(parsedURL.Path, "/"),
		Username: parsedURL.User.Username(),
	}

	if port := parsedURL.Port(); port != "" {
		storageConfig.Port = parsePort(port)
	} else {
		storageConfig.Port = 5432
	}

	if password, ok := parsedURL.User.Password(); ok {
		storageConfig.Password = password
	}

	return storageConfig, nil
}

func parsePort(port string) int {
	i, err := strconv.Atoi(port)
	if err != nil {
		return 5432
	}
	return i
}

func testMotionsQuery(ctx context.Context, client *odata.Client) error {
	log.Println("Testing motions query...")

	data, err := client.GetMotiesWithVotes(ctx, 0, 1)
	if err != nil {
		return fmt.Errorf("fetching motions: %w", err)
	}

	response, err := odata.ParseODataResponse(data)
	if err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	prettyData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling response: %w", err)
	}

	fmt.Println("Sample motions response:")
	fmt.Println(string(prettyData))
	return nil
}

func printStats(stats *odata.ImportStats) {
	fmt.Println("\n=== Import Statistics ===")
	fmt.Printf("Total Zaken: %d\n", stats.TotalZaken)
	fmt.Printf("Total Besluiten: %d\n", stats.TotalBesluiten)
	fmt.Printf("Total Stemmingen: %d\n", stats.TotalStemmingen)
	fmt.Printf("Total Personen: %d\n", stats.TotalPersonen)
	fmt.Printf("Total Fracties: %d\n", stats.TotalFracties)
	fmt.Printf("Processing Errors: %d\n", stats.ProcessingErrors)
	fmt.Printf("Processing Duration: %v\n", stats.ProcessingDuration)
	fmt.Printf("Start Time: %s\n", stats.ProcessingStartTime.Format(time.RFC3339))
	fmt.Printf("End Time: %s\n", stats.ProcessingEndTime.Format(time.RFC3339))

	if len(stats.ZakenByType) > 0 {
		fmt.Println("\nZaken by Type:")
		for zaakType, count := range stats.ZakenByType {
			fmt.Printf("  %s: %d\n", zaakType, count)
		}
	}

	if len(stats.BesluitenByType) > 0 {
		fmt.Println("\nBesluiten by Type:")
		for besluitType, count := range stats.BesluitenByType {
			fmt.Printf("  %s: %d\n", besluitType, count)
		}
	}

	if len(stats.StemmingByType) > 0 {
		fmt.Println("\nStemming by Type:")
		for stemmingType, count := range stats.StemmingByType {
			fmt.Printf("  %s: %d\n", stemmingType, count)
		}
	}

	if len(stats.ErrorDetails) > 0 {
		fmt.Println("\nError Details:")
		for i, error := range stats.ErrorDetails {
			fmt.Printf("  %d: %s\n", i+1, error)
		}
	}
}
