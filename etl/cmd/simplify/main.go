package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"etl/internal/cmdutils"
	"etl/pkg/storage"
)

func main() {
	var (
		configPath = flag.String("config", "configs/config.yaml", "Path to configuration file")
		modelName  = flag.String("model", "gemma3:4b", "Name of the Ollama model to use (e.g., mistral, llama3, gemma)")
		limit      = flag.Int("limit", 10, "Optional: limit number of zaken to process (0 = no limit)")
	)
	flag.Parse()

	log.Printf("Starting simplify...")

	// Load configuration
	cfg, err := cmdutils.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}
	log.Printf("Successfully loaded config")

	// Parse database config (supports DATABASE_URL or yaml config)
	storageConfig, err := cmdutils.ParseStorageConfig(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to parse storage config: %v", err)
	}
	log.Printf("Successfully parsed storage config")

	store, err := storage.NewStorage(*storageConfig)
	if err != nil {
		log.Fatalf("❌ Failed to initialize storage: %v", err)
	}
	log.Printf("Successfully initialized storage")
	defer store.Close()

	log.Println("🔗 Connected to database")

	// Migrate to ensure schema consistency
	log.Println("🚀 Migrating database...")
	if err := store.Migrate(context.Background()); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}
	log.Println("✅ Database migration complete")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("🛑 Received interrupt signal, shutting down gracefully...")
		cancel()
	}()

	start := time.Now()
	log.Printf("✨ Starting simplification using model: %s", *modelName)

	// Simplify only zaken that have bullet points and no besluit
	err = store.SimplifyCasesWithOllamaFiltered(ctx, *modelName, *limit)
	if err != nil {
		log.Fatalf("❌ Simplification failed: %v", err)
	}

	elapsed := time.Since(start)
	log.Printf("✅ Simplification completed in %s", elapsed)

	printSummary(elapsed, *modelName)
}

func printSummary(duration time.Duration, model string) {
	fmt.Println("\n=== Simplification Summary ===")
	fmt.Printf("Model Used: %s\n", model)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Println("All zaken with bullet_points and no besluit have been processed.")
	fmt.Println("Simplified results saved to simplified_bullet_points column.\n")
}
