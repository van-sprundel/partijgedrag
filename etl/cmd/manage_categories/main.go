package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"etl/internal/categorisation"
	"etl/internal/cmdutils"
	"etl/internal/models"
	"etl/pkg/storage"

	"github.com/google/uuid"
	"github.com/lib/pq"
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

	cfg, err := cmdutils.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// parse DATABASE_URL if provided, otherwise use config values
	storageConfig, err := cmdutils.ParseStorageConfig(cfg)
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
		categorisationService := categorisation.NewService(store)
		if err := categorisationService.EnrichZaken(ctx); err != nil {
			log.Fatalf("Failed to enrich zaken: %v", err)
		}
	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}

func seedInitialCategories(ctx context.Context, store storage.Storage) error {
	fmt.Println("Seeding initial categories...")

	var categories []models.MotionCategory
	now := time.Now()

	genericTopics := map[string][]string{
		"Bestuur":                     {"bestuur", "governance", "regering", "kabinet", "minister", "staatssecretaris", "overheid", "gemeente", "provincie"},
		"Cultuur en recreatie":        {"cultuur", "kunst", "sport", "recreatie", "museum", "theater", "bibliotheek", "media", "erfgoed"},
		"Economie":                    {"economie", "economisch", "handel", "industrie", "ondernemerschap", "mkb", "bedrijven", "concurrentie", "marktwerking"},
		"Financiën":                   {"financiën", "belasting", "btw", "budget", "begroting", "schuld", "deficit", "lastenverlichting", "koopkracht"},
		"Huisvesting":                 {"wonen", "huur", "woningbouw", "huisvesting", "hypotheek", "woningnood", "woningmarkt", "verhuurder", "huurder", "leegstand"},
		"Internationaal":              {"internationaal", "europa", "eu", "buitenland", "ontwikkelingssamenwerking", "defensie", "verdrag", "mensenrechten"},
		"Landbouw":                    {"landbouw", "boer", "vee", "gewas", "agrarisch", "voedsel", "mest", "visserij", "pesticiden", "stikstof"},
		"Migratie en integratie":      {"migratie", "integratie", "vluchteling", "asiel", "immigrant", "inburgering", "statushouder"},
		"Natuur en milieu":            {"natuur", "milieu", "klimaat", "co2", "duurzaam", "energie", "vervuiling", "biodiversiteit", "opwarming", "uitstoot", "windmolens", "zonne-energie"},
		"Onderwijs en wetenschap":     {"onderwijs", "school", "universiteit", "student", "wetenschap", "onderzoek", "educatie", "leraar", "onderwijskwaliteit"},
		"Openbare orde en veiligheid": {"veiligheid", "politie", "criminaliteit", "terrorisme", "orde", "handhaving", "brandweer", "hulpdiensten", "rampenbestrijding"},
		"Recht":                       {"recht", "rechtspraak", "rechter", "wet", "juridisch", "justitie", "advocaat", "privacy", "discriminatie", "grondwet"},
		"Ruimte en infrastructuur":    {"infrastructuur", "weg", "spoor", "bouw", "ruimtelijk", "planning", "transport", "luchtvaart", "schiphol", "openbare ruimte"},
		"Sociale zekerheid":           {"sociale zekerheid", "uitkering", "aow", "wajong", "bijstand", "pensioen", "armoede", "schulden", "mantelzorg"},
		"Verkeer":                     {"verkeer", "auto", "fiets", "openbaar vervoer", "trein", "bus", "file", "verkeersveiligheid", "mobiliteit"},
		"Werk":                        {"werk", "werkgelegenheid", "baan", "arbeidsmarkt", "cao", "vakbond", "werknemer", "zzp'er", "flexwerk", "thuiswerken"},
		"Zorg en gezondheid":          {"zorg", "gezondheid", "medisch", "ziekenhuis", "dokter", "medicijn", "patiënt", "preventie", "jeugdzorg", "ouderenzorg", "ggz", "vaccinatie"},
	}

	for topic := range genericTopics {
		genericType := "generic"
		keywords := genericTopics[topic]
		categories = append(categories, models.MotionCategory{
			ID:          uuid.New().String(),
			Name:        topic,
			Type:        &genericType,
			Description: nil,
			Keywords:    pq.StringArray(keywords),
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}

	hotTopics := map[string][]string{
		"Immigratie":         {"immigratie", "migratie", "asielzoeker", "vluchtelingen", "grenzen", "azc"},
		"Oorlog":             {"oorlog", "conflict", "militair", "defensie", "wapen", "vrede", "oekraïne", "rusland", "Gaza", "Israel", "Palestina"},
		"Klimaatverandering": {"klimaatverandering", "opwarming", "broeikas", "klimaat", "duurzaamheid"},
		"Woningcrisis":       {"woningcrisis", "woningtekort", "betaalbaar wonen", "huurprijzen"},
		"Inflatie":           {"inflatie", "prijsstijging", "koopkracht", "duurte"},
	}

	for topic := range hotTopics {
		hotType := "hot_topic"
		keywords := hotTopics[topic]
		categories = append(categories, models.MotionCategory{
			ID:          uuid.New().String(),
			Name:        topic,
			Type:        &hotType,
			Description: nil,
			Keywords:    pq.StringArray(keywords),
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
		for _, keyword := range []string(category.Keywords) {
			if strings.Contains(searchText, keyword) {
				matches = append(matches, category.ID)
				break
			}
		}
	}

	return matches
}
