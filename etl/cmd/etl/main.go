package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"etl/internal/api"
	"etl/internal/config"
	"etl/internal/importer"
	"etl/internal/models"
	"etl/pkg/odata"
	"etl/pkg/storage"

	"gopkg.in/yaml.v3"
)

func main() {
	var (
		configPath = flag.String("config", "configs/config.yaml", "Path to configuration file")
		afterDate  = flag.String("after", "", "Only fetch records modified after this date\n\t\tSupported formats:\n\t\t- RFC3339: 2024-01-01T00:00:00Z\n\t\t- Keywords: today, yesterday, this-week, last-week, this-month, last-month")
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

	var afterTime *time.Time
	if *afterDate != "" {
		parsedTime, err := parseAfterDate(*afterDate)
		if err != nil {
			log.Fatalf("Failed to parse after date: %v", err)
		}
		afterTime = &parsedTime
		log.Printf("Filtering records modified after: %s", afterTime.Format(time.RFC3339))
	}

	client := odata.NewClient(cfg.API)
	apiClient := api.NewClient(cfg.API)

	imp := importer.NewImporter(client, store, apiClient)

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
		if err := imp.ImportMotiesWithVotes(ctx, afterTime); err != nil {
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

func printStats(stats *models.ImportStats) {
	fmt.Println("\n=== Import Statistics ===")
	fmt.Printf("Total Zaken: %d\n", stats.TotalZaken)
	fmt.Printf("Total Besluiten: %d\n", stats.TotalBesluiten)
	fmt.Printf("Total Stemmingen: %d\n", stats.TotalStemmingen)
	fmt.Printf("Total Personen: %d\n", stats.TotalPersonen)
	fmt.Printf("Total Fracties: %d\n", stats.TotalFracties)
	fmt.Printf("Total Kamerstukdossiers: %d\n", stats.TotalKamerstukdossiers)
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

	if len(stats.ErrorDetails) > 0 {
		fmt.Println("\nError Details:")
		for i, error := range stats.ErrorDetails {
			fmt.Printf("  %d: %s\n", i+1, error)
		}
	}
}

func parseAfterDate(dateStr string) (time.Time, error) {
	now := time.Now().UTC()

	switch strings.ToLower(dateStr) {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), nil
	case "yesterday":
		yesterday := now.AddDate(0, 0, -1)
		return time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.UTC), nil
	case "this-week":
		// Start of current week (Monday)
		weekday := int(now.Weekday())
		if weekday == 0 { // Sunday
			weekday = 7
		}
		daysBack := weekday - 1
		startOfWeek := now.AddDate(0, 0, -daysBack)
		return time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, time.UTC), nil
	case "last-week":
		// Start of last week (Monday)
		weekday := int(now.Weekday())
		if weekday == 0 { // Sunday
			weekday = 7
		}
		daysBack := weekday - 1 + 7 // Go back to start of last week
		startOfLastWeek := now.AddDate(0, 0, -daysBack)
		return time.Date(startOfLastWeek.Year(), startOfLastWeek.Month(), startOfLastWeek.Day(), 0, 0, 0, 0, time.UTC), nil
	case "this-month":
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC), nil
	case "last-month":
		lastMonth := now.AddDate(0, -1, 0)
		return time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.UTC), nil
	default:
		// Try to parse as RFC3339 format
		return time.Parse(time.RFC3339, dateStr)
	}
}
