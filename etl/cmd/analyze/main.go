package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"etl/internal/analysis"
	"etl/internal/cmdutils"
	"etl/internal/models"
	"etl/pkg/storage"
)

func main() {
	var (
		configPath = flag.String("config", "configs/config.yaml", "Path to configuration file")
		period     = flag.String("period", "all", "Time period to analyze (e.g., '2024', '2023', 'all')")
		output     = flag.String("output", "text", "Output format: text, json")
		limit      = flag.Int("limit", 50, "Limit for results (e.g., top N deviations)")
		partyID    = flag.String("party", "", "Filter by party ID (for party-specific analyses)")
	)

	// Subcommands
	coalitionCmd := flag.NewFlagSet("coalition", flag.ExitOnError)
	deviationCmd := flag.NewFlagSet("deviation", flag.ExitOnError)
	topicsCmd := flag.NewFlagSet("topics", flag.ExitOnError)
	partyCmd := flag.NewFlagSet("party", flag.ExitOnError)
	reportCmd := flag.NewFlagSet("report", flag.ExitOnError)
	listPartiesCmd := flag.NewFlagSet("list-parties", flag.ExitOnError)
	refreshCmd := flag.NewFlagSet("refresh", flag.ExitOnError)

	flag.Parse()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Handle subcommand
	subcommand := os.Args[1]
	var subArgs []string
	if len(os.Args) > 2 {
		subArgs = os.Args[2:]
	}

	// Re-parse with subcommand args to allow flags after subcommand
	switch subcommand {
	case "coalition":
		coalitionCmd.StringVar(configPath, "config", "configs/config.yaml", "Path to configuration file")
		coalitionCmd.StringVar(period, "period", "all", "Time period")
		coalitionCmd.StringVar(output, "output", "text", "Output format")
		coalitionCmd.Parse(subArgs)
	case "deviation":
		deviationCmd.StringVar(configPath, "config", "configs/config.yaml", "Path to configuration file")
		deviationCmd.StringVar(period, "period", "all", "Time period")
		deviationCmd.StringVar(output, "output", "text", "Output format")
		deviationCmd.IntVar(limit, "limit", 50, "Number of results")
		deviationCmd.Parse(subArgs)
	case "topics":
		topicsCmd.StringVar(configPath, "config", "configs/config.yaml", "Path to configuration file")
		topicsCmd.StringVar(period, "period", "all", "Time period")
		topicsCmd.StringVar(output, "output", "text", "Output format")
		topicsCmd.Parse(subArgs)
	case "party":
		partyCmd.StringVar(configPath, "config", "configs/config.yaml", "Path to configuration file")
		partyCmd.StringVar(period, "period", "all", "Time period")
		partyCmd.StringVar(output, "output", "text", "Output format")
		partyCmd.StringVar(partyID, "id", "", "Party ID (required)")
		partyCmd.Parse(subArgs)
	case "report":
		reportCmd.StringVar(configPath, "config", "configs/config.yaml", "Path to configuration file")
		reportCmd.StringVar(period, "period", "all", "Time period")
		reportCmd.StringVar(output, "output", "text", "Output format")
		reportCmd.Parse(subArgs)
	case "list-parties":
		listPartiesCmd.StringVar(configPath, "config", "configs/config.yaml", "Path to configuration file")
		listPartiesCmd.Parse(subArgs)
	case "refresh":
		refreshCmd.StringVar(configPath, "config", "configs/config.yaml", "Path to configuration file")
		refreshCmd.Parse(subArgs)
	case "help", "-h", "--help":
		printUsage()
		os.Exit(0)
	default:
		fmt.Printf("Unknown command: %s\n", subcommand)
		printUsage()
		os.Exit(1)
	}

	// Initialize storage
	cfg, err := cmdutils.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	storageConfig, err := cmdutils.ParseStorageConfig(cfg)
	if err != nil {
		log.Fatalf("Failed to parse storage config: %v", err)
	}

	store, err := storage.NewStorage(*storageConfig)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	svc := analysis.NewService(store)

	// Execute command
	switch subcommand {
	case "coalition":
		runCoalition(ctx, svc, *period, *output)
	case "deviation":
		runDeviation(ctx, svc, *period, *limit, *output)
	case "topics":
		runTopics(ctx, svc, *period, *output)
	case "party":
		if *partyID == "" {
			fmt.Println("Error: -id flag is required for party analysis")
			fmt.Println("Use 'analyze list-parties' to see available parties")
			os.Exit(1)
		}
		runParty(ctx, svc, *partyID, *period, *output)
	case "report":
		runReport(ctx, svc, *period, *output)
	case "list-parties":
		runListParties(ctx, store)
	case "refresh":
		runRefresh(ctx, store)
	}
}

func printUsage() {
	fmt.Println(`Partijgedrag Analysis Tool

Usage:
  analyze <command> [flags]

Commands:
  coalition     Show coalition alignment between parties
  deviation     Show MPs who deviate from party line most often
  topics        Show motion trends by topic/category
  party         Show how a specific party votes on topics
  report        Generate a full analysis report
  list-parties  List all active parties with their IDs
  refresh       Refresh materialized views (run after ETL import)

Common Flags:
  -config string   Path to configuration file (default "configs/config.yaml")
  -period string   Time period: year like "2024" or "all" (default "all")
  -output string   Output format: "text" or "json" (default "text")

Examples:
  analyze refresh                    # Run this first after importing data!
  analyze coalition -period 2024
  analyze deviation -limit 20
  analyze topics -period all -output json
  analyze party -id <party-uuid>
  analyze report -period 2024`)
}

func runCoalition(ctx context.Context, svc *analysis.Service, period, output string) {
	result, err := svc.CoalitionAnalysis(ctx, period)
	if err != nil {
		log.Fatalf("Coalition analysis failed: %v", err)
	}

	if output == "json" {
		jsonStr, _ := svc.ToJSON(result)
		fmt.Println(jsonStr)
	} else {
		alignments := result.Data.([]models.CoalitionAlignment)
		fmt.Println(analysis.FormatCoalitionAlignments(alignments))

		// Also show matrix
		matrixResult, err := svc.CoalitionMatrix(ctx, period)
		if err == nil {
			matrixData := matrixResult.Data.(map[string]interface{})
			parties := matrixData["parties"].([]string)
			matrix := matrixData["matrix"].(map[string]map[string]float64)
			fmt.Println(analysis.FormatCoalitionMatrix(parties, matrix))
		}
	}
}

func runDeviation(ctx context.Context, svc *analysis.Service, period string, limit int, output string) {
	result, err := svc.MPDeviationAnalysis(ctx, period, limit)
	if err != nil {
		log.Fatalf("Deviation analysis failed: %v", err)
	}

	if output == "json" {
		jsonStr, _ := svc.ToJSON(result)
		fmt.Println(jsonStr)
	} else {
		deviations := result.Data.([]models.MPDeviation)
		fmt.Println(analysis.FormatMPDeviations(deviations))
	}
}

func runTopics(ctx context.Context, svc *analysis.Service, period, output string) {
	result, err := svc.TopicTrendAnalysis(ctx, period)
	if err != nil {
		log.Fatalf("Topic analysis failed: %v", err)
	}

	if output == "json" {
		jsonStr, _ := svc.ToJSON(result)
		fmt.Println(jsonStr)
	} else {
		trends := result.Data.([]models.TopicTrend)
		fmt.Println(analysis.FormatTopicTrends(trends))
	}
}

func runParty(ctx context.Context, svc *analysis.Service, partyID, period, output string) {
	result, err := svc.PartyTopicAnalysis(ctx, partyID, period)
	if err != nil {
		log.Fatalf("Party analysis failed: %v", err)
	}

	if output == "json" {
		jsonStr, _ := svc.ToJSON(result)
		fmt.Println(jsonStr)
	} else {
		voting := result.Data.([]models.PartyTopicVoting)
		fmt.Println(analysis.FormatPartyTopicVoting(voting))
	}
}

func runReport(ctx context.Context, svc *analysis.Service, period, output string) {
	results, err := svc.FullReport(ctx, period)
	if err != nil {
		log.Fatalf("Report generation failed: %v", err)
	}

	if output == "json" {
		for name, result := range results {
			fmt.Printf("=== %s ===\n", name)
			jsonStr, _ := svc.ToJSON(result)
			fmt.Println(jsonStr)
		}
	} else {
		fmt.Printf("\n========================================\n")
		fmt.Printf("  PARTIJGEDRAG ANALYSIS REPORT\n")
		fmt.Printf("  Period: %s\n", period)
		fmt.Printf("========================================\n")

		// Coalition
		if result, ok := results["coalition"]; ok {
			alignments := result.Data.([]models.CoalitionAlignment)
			fmt.Println(analysis.FormatCoalitionAlignments(alignments))
		}

		// Matrix
		if result, ok := results["matrix"]; ok {
			matrixData := result.Data.(map[string]interface{})
			parties := matrixData["parties"].([]string)
			matrix := matrixData["matrix"].(map[string]map[string]float64)
			fmt.Println(analysis.FormatCoalitionMatrix(parties, matrix))
		}

		// Deviations
		if result, ok := results["deviations"]; ok {
			deviations := result.Data.([]models.MPDeviation)
			fmt.Println(analysis.FormatMPDeviations(deviations))
		}

		// Topics
		if result, ok := results["topics"]; ok {
			trends := result.Data.([]models.TopicTrend)
			fmt.Println(analysis.FormatTopicTrends(trends))
		}
	}
}

func runListParties(ctx context.Context, store storage.Storage) {
	fracties, err := store.GetActiveFracties(ctx)
	if err != nil {
		log.Fatalf("Failed to get parties: %v", err)
	}

	fmt.Println("\n=== Active Parties ===")
	fmt.Printf("%-40s %-10s %s\n", "ID", "Abbr", "Name")
	fmt.Println(string(make([]byte, 70)))
	for _, f := range fracties {
		name := ""
		if f.NaamNL != nil {
			name = *f.NaamNL
		}
		abbr := ""
		if f.Afkorting != nil {
			abbr = *f.Afkorting
		}
		fmt.Printf("%-40s %-10s %s\n", f.ID, abbr, name)
	}
}

func runRefresh(ctx context.Context, store storage.Storage) {
	fmt.Println("Refreshing materialized views for analysis...")
	if err := store.RefreshMaterializedViews(ctx); err != nil {
		log.Fatalf("Failed to refresh views: %v", err)
	}
	fmt.Println("Done! Coalition analysis should now have data.")
}
