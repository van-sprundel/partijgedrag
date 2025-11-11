package main

import (
	"context"
	"etl/internal/cmdutils"
	"etl/internal/llm"
	"etl/internal/services"
	"etl/pkg/storage"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env file not found, relying on existing environment variables")
	}

	var (
		configPath    = flag.String("config", "configs/config.yaml", "Path to configuration file")
		client        = flag.String("client", "ollama", "LLM client to use: 'ollama' or 'openai'")
		modelName     = flag.String("model", "gemma3:4b", "Name of the model to use")
		limit         = flag.Int("limit", 1, "Optional: limit number of zaken to process (0 = no limit)")
		concurrency   = flag.Int("concurrency", 1, "Optional: concurrency to use")
		saveBatchSize = flag.Int("save-batch-size", 500, "Optional: maximum number of items to save in a batch")
	)
	flag.Parse()

	if *limit != 0 && *concurrency > *limit {
		log.Fatalf("❌ Limit should not be smaller than concurrency")

	}

	// Load config (adjust this to your config loader)
	cfg, err := cmdutils.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Initialize storage
	store, err := storage.NewStorage(cfg.Storage)
	if err != nil {
		log.Fatalf("❌ Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize Ollama client
	var llmClient llm.LLMClient

	switch *client {
	case "ollama":
		llmClient = llm.NewOllamaClient("http://localhost:11434", *modelName)
	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			log.Fatal("OPENAI API key not provided. Set OPENAI_API_KEY or provide in config.")
		}
		llmClient = llm.NewOpenAIClient(apiKey, *modelName)
	default:
		log.Fatalf("❌ Unknown LLM client: %s", *client)
	}

	// Initialize simplifier service
	simplifier := service.NewSimplifierService(store, llmClient)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		log.Println("🛑 Interrupt received, shutting down...")
		cancel()
	}()

	// Run simplification
	log.Printf("✨ Starting simplification using %d...", client)
	start := time.Now()
	if *concurrency > 1 {
		if err := simplifier.SimplifyCasesConcurrent(ctx, *limit, *concurrency, *saveBatchSize); err != nil {
			log.Fatalf("❌Concurrent simplification failed: %v", err)
		}
	} else {
		if err := simplifier.SimplifyCases(ctx, *limit); err != nil {
			log.Fatalf("❌ Simplification failed: %v", err)
		}
	}

	log.Println("✅ Simplification completed successfully!")
	elapsed := time.Since(start)
	printSummary(elapsed, *client, *modelName)
}

func printSummary(duration time.Duration, client string, model string) {
	fmt.Println("\n=== Simplification Summary ===")
	fmt.Printf("Client used: %d\n", client)
	fmt.Printf("Model Used: %s\n", model)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Println("All zaken with bullet_points and no besluit have been processed.")
	fmt.Println("Simplified results saved to simplified_bullet_points column.\n")
}
