package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"etl/internal/config"
	"etl/internal/models"
	"etl/pkg/storage"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

func main() {
	var (
		configPath = flag.String("config", "configs/config.yaml", "Path to configuration file")
		action     = flag.String("action", "seed", "Action to perform: seed, list, add, enrich")
		name       = flag.String("name", "", "Category name (for add action)")
		catType    = flag.String("type", "", "Category type: generic, hot_topic, or empty for general")
		desc       = flag.String("desc", "", "Category description")
	)
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	storageConfig, err := parseStorageConfig(cfg)
	if err != nil {
		log.Fatalf("Failed to parse storage config: %v", err)
	}

	store, err := storage.NewStorage(*storageConfig)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	switch *action {
	case "seed":
		if err := seedInitialCategories(ctx, store); err != nil {
			log.Fatalf("Failed to seed categories: %v", err)
		}
	case "list":
		if err := listCategories(ctx, store); err != nil {
			log.Fatalf("Failed to list categories: %v", err)
		}
	case "add":
		if *name == "" {
			log.Fatal("Name is required for add action")
		}
		if err := addCategory(ctx, store, *name, *catType, *desc); err != nil {
			log.Fatalf("Failed to add category: %v", err)
		}
	case "enrich":
		if err := enrichZaken(ctx, store); err != nil {
			log.Fatalf("Failed to enrich zaken: %v", err)
		}
	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}

func seedInitialCategories(ctx context.Context, store storage.Storage) error {
	fmt.Println("Seeding initial categories...")

	genericTopics := []string{
		"Bestuur",
		"Cultuur en recreatie",
		"Economie",
		"Financiën",
		"Huisvesting",
		"Internationaal",
		"Landbouw",
		"Migratie en integratie",
		"Natuur en milieu",
		"Onderwijs en wetenschap",
		"Openbare orde en veiligheid",
		"Recht",
		"Ruimte en infrastructuur",
		"Sociale zekerheid",
		"Verkeer",
		"Werk",
		"Zorg en gezondheid",
	}

	hotTopics := []string{
		"Vaccinaties",
		"Immigratie",
		"Oorlog",
		"Corona",
		"Klimaatverandering",
		"Woningcrisis",
		"Inflatie",
	}

	var categories []models.MotionCategory
	now := time.Now()

	genericTopicsWithKeywords := map[string][]string{
		"Bestuur":                     {"bestuur", "governance", "regering", "kabinet", "minister", "staatssecretaris"},
		"Cultuur en recreatie":        {"cultuur", "kunst", "sport", "recreatie", "museum", "theater", "bibliotheek"},
		"Economie":                    {"economie", "economisch", "handel", "industrie", "ondernemerschap", "mkb", "bedrijven"},
		"Financiën":                   {"financiën", "belasting", "btw", "budget", "begroting", "schuld", "deficit"},
		"Huisvesting":                 {"wonen", "huur", "woningbouw", "huisvesting", "hypotheek", "woningnood", "woningmarkt"},
		"Internationaal":              {"internationaal", "europa", "eu", "buitenland", "ontwikkelingssamenwerking", "defensie"},
		"Landbouw":                    {"landbouw", "boer", "vee", "gewas", "agrarisch", "voedsel", "mest"},
		"Migratie en integratie":      {"migratie", "integratie", "vluchteling", "asiel", "immigrant", "inburgering"},
		"Natuur en milieu":            {"natuur", "milieu", "klimaat", "co2", "duurzaam", "energie", "vervuiling"},
		"Onderwijs en wetenschap":     {"onderwijs", "school", "universiteit", "student", "wetenschap", "onderzoek", "educatie"},
		"Openbare orde en veiligheid": {"veiligheid", "politie", "criminaliteit", "terrorisme", "orde", "handhaving"},
		"Recht":                       {"recht", "rechtspraak", "rechter", "wet", "juridisch", "justitie", "advocaat"},
		"Ruimte en infrastructuur":    {"infrastructuur", "weg", "spoor", "bouw", "ruimtelijk", "planning", "transport"},
		"Sociale zekerheid":           {"sociale zekerheid", "uitkering", "aow", "wajong", "bijstand", "pensioen"},
		"Verkeer":                     {"verkeer", "auto", "fiets", "openbaar vervoer", "trein", "bus", "file"},
		"Werk":                        {"werk", "werkgelegenheid", "baan", "arbeidsmarkt", "cao", "vakbond", "werknemer"},
		"Zorg en gezondheid":          {"zorg", "gezondheid", "medisch", "ziekenhuis", "dokter", "medicijn", "patiënt"},
	}

	for _, topic := range genericTopics {
		genericType := "generic"
		keywords := genericTopicsWithKeywords[topic]
		categories = append(categories, models.MotionCategory{
			ID:          uuid.New().String(),
			Name:        topic,
			Type:        &genericType,
			Description: nil,
			Keywords:    keywords,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}

	hotTopicsWithKeywords := map[string][]string{
		"Vaccinaties":        {"vaccinatie", "vaccin", "inenting", "immunisatie", "prik"},
		"Immigratie":         {"immigratie", "migratie", "asielzoeker", "vluchtelingen", "grenzen"},
		"Oorlog":             {"oorlog", "conflict", "militair", "defensie", "wapen", "vrede", "oekraïne", "rusland"},
		"Corona":             {"corona", "covid", "pandemie", "lockdown", "mondkapje", "quarantaine"},
		"Klimaatverandering": {"klimaatverandering", "opwarming", "broeikas", "klimaat", "duurzaamheid"},
		"Woningcrisis":       {"woningcrisis", "woningtekort", "betaalbaar wonen", "huurprijzen"},
		"Inflatie":           {"inflatie", "prijsstijging", "koopkracht", "duurte"},
	}

	for _, topic := range hotTopics {
		hotType := "hot_topic"
		keywords := hotTopicsWithKeywords[topic]
		categories = append(categories, models.MotionCategory{
			ID:          uuid.New().String(),
			Name:        topic,
			Type:        &hotType,
			Description: nil,
			Keywords:    keywords,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}

	if err := store.SaveCategories(ctx, categories); err != nil {
		return fmt.Errorf("saving categories: %w", err)
	}

	fmt.Printf("Successfully seeded %d categories\n", len(categories))
	fmt.Printf("- %d generic topics\n", len(genericTopics))
	fmt.Printf("- %d hot topics\n", len(hotTopics))
	return nil
}

func listCategories(ctx context.Context, store storage.Storage) error {
	fmt.Println("All categories:")

	categories, err := store.GetAllCategories(ctx)
	if err != nil {
		return err
	}

	if len(categories) == 0 {
		fmt.Println("No categories found")
		return nil
	}

	fmt.Println("\nGeneric Topics:")
	for _, cat := range categories {
		if cat.Type != nil && *cat.Type == "generic" {
			desc := ""
			if cat.Description != nil {
				desc = fmt.Sprintf(" - %s", *cat.Description)
			}
			fmt.Printf("  - %s%s\n", cat.Name, desc)
		}
	}

	fmt.Println("\nHot Topics:")
	for _, cat := range categories {
		if cat.Type != nil && *cat.Type == "hot_topic" {
			desc := ""
			if cat.Description != nil {
				desc = fmt.Sprintf(" - %s", *cat.Description)
			}
			fmt.Printf("  - %s%s\n", cat.Name, desc)
		}
	}

	fmt.Println("\nGeneral Keywords:")
	for _, cat := range categories {
		if cat.Type == nil || (*cat.Type != "generic" && *cat.Type != "hot_topic") {
			desc := ""
			if cat.Description != nil {
				desc = fmt.Sprintf(" - %s", *cat.Description)
			}
			fmt.Printf("  - %s%s\n", cat.Name, desc)
		}
	}

	fmt.Printf("\nTotal: %d categories\n", len(categories))
	return nil
}

func addCategory(ctx context.Context, store storage.Storage, name, catType, desc string) error {
	var typePtr *string
	if catType != "" {
		typePtr = &catType
	}

	var descPtr *string
	if desc != "" {
		descPtr = &desc
	}

	now := time.Now()
	category := models.MotionCategory{
		ID:          uuid.New().String(),
		Name:        name,
		Type:        typePtr,
		Description: descPtr,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := store.SaveCategories(ctx, []models.MotionCategory{category}); err != nil {
		return fmt.Errorf("saving category: %w", err)
	}

	typeDesc := "general keyword"
	if typePtr != nil {
		typeDesc = *typePtr
	}

	fmt.Printf("Successfully added category '%s' as %s\n", name, typeDesc)
	return nil
}

func enrichZaken(ctx context.Context, store storage.Storage) error {
	fmt.Println("Starting enrichment of zaken with categories...")

	categories, err := store.GetAllCategories(ctx)
	if err != nil {
		return fmt.Errorf("getting categories: %w", err)
	}

	if len(categories) == 0 {
		fmt.Println("No categories found. Run with -action=seed first.")
		return nil
	}

	zaken, err := store.GetZakenForEnrichment(ctx)
	if err != nil {
		return fmt.Errorf("getting zaken: %w", err)
	}

	fmt.Printf("Found %d zaken to analyze and %d categories\n", len(zaken), len(categories))

	enriched := 0
	for _, zaak := range zaken {
		matches := findCategoryMatches(zaak, categories)

		if len(matches) > 0 {
			for _, categoryID := range matches {
				if err := store.AssignCategoryToZaak(ctx, zaak.ID, categoryID); err != nil {
					log.Printf("Warning: failed to assign category to zaak %s: %v", zaak.ID, err)
				} else {
					enriched++
				}
			}
		}
	}

	fmt.Printf("Enrichment complete: assigned %d category relationships\n", enriched)
	return nil
}

func findCategoryMatches(zaak models.Zaak, categories []models.MotionCategory) []string {
	var matches []string

	searchText := ""
	if zaak.Titel != nil {
		searchText += strings.ToLower(*zaak.Titel) + " "
	}
	if zaak.Onderwerp != nil {
		searchText += strings.ToLower(*zaak.Onderwerp) + " "
	}

	for _, category := range categories {
		for _, keyword := range category.Keywords {
			if strings.Contains(searchText, strings.ToLower(keyword)) {
				matches = append(matches, category.ID)
				break
			}
		}
	}

	return matches
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
		if cfg.Storage.Type == "postgres" && cfg.Storage.Password == "" {
			fmt.Print("Database password: ")
			fmt.Scanln(&cfg.Storage.Password)
		}
		return &cfg.Storage, nil
	}

	return &cfg.Storage, nil
}
