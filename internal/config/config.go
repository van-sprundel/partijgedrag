package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	HTTPHost    string
	HTTPPort    int

	TweedeKamerODataBaseURL string
	TweedeKamerBatchSize    int
	TweedeKamerMaxPages     int
	TweedeKamerInitialSince time.Time
	CursorOverlap           time.Duration
}

func Load() (Config, error) {
	_ = godotenv.Load()

	initialSince, err := time.Parse(time.RFC3339, getEnv("TWEEDE_KAMER_INITIAL_SINCE", "1970-01-01T00:00:00Z"))
	if err != nil {
		return Config{}, fmt.Errorf("parse TWEEDE_KAMER_INITIAL_SINCE: %w", err)
	}

	return Config{
		DatabaseURL:             getEnv("DATABASE_URL", "postgres://etl_user:etl_password@localhost:5432/partijgedrag"),
		HTTPHost:                getEnv("HTTP_HOST", "0.0.0.0"),
		HTTPPort:                getEnvInt("HTTP_PORT", 3001),
		TweedeKamerODataBaseURL: getEnv("TWEEDE_KAMER_ODATA_BASE_URL", "https://gegevensmagazijn.tweedekamer.nl/OData/v4/2.0"),
		TweedeKamerBatchSize:    getEnvInt("TWEEDE_KAMER_BATCH_SIZE", 100),
		TweedeKamerMaxPages:     getEnvInt("TWEEDE_KAMER_MAX_PAGES", 0),
		TweedeKamerInitialSince: initialSince,
		CursorOverlap:           time.Duration(getEnvInt("TWEEDE_KAMER_CURSOR_OVERLAP_MINUTES", 10)) * time.Minute,
	}, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
