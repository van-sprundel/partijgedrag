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

	// Dev makes `serve` read templates and static files from disk on every
	// request instead of the embedded copies, so edits only need a browser
	// refresh. Requires running from the repo root.
	Dev bool

	TweedeKamerODataBaseURL string
	TweedeKamerBatchSize    int
	TweedeKamerMaxPages     int
	TweedeKamerInitialSince time.Time
	CursorOverlap           time.Duration

	// SyncInterval makes `serve` run a full tweedekamer sync on this interval,
	// so a plain container deployment stays fresh without an external cron.
	// Zero disables the built-in scheduler.
	SyncInterval            time.Duration
	SyncMotionVoteLimit     int
	SyncMotionDocumentLimit int
}

func Load() (Config, error) {
	_ = godotenv.Load()

	initialSince, err := time.Parse(time.RFC3339, getEnv("TWEEDE_KAMER_INITIAL_SINCE", "1970-01-01T00:00:00Z"))
	if err != nil {
		return Config{}, fmt.Errorf("parse TWEEDE_KAMER_INITIAL_SINCE: %w", err)
	}

	syncInterval, err := time.ParseDuration(getEnv("SYNC_INTERVAL", "1h"))
	if err != nil {
		return Config{}, fmt.Errorf("parse SYNC_INTERVAL: %w", err)
	}
	if syncInterval < 0 {
		return Config{}, fmt.Errorf("SYNC_INTERVAL must be 0 or greater")
	}

	return Config{
		DatabaseURL:             getEnv("DATABASE_URL", "postgres://etl_user:etl_password@localhost:5432/partijgedrag"),
		HTTPHost:                getEnv("HTTP_HOST", "0.0.0.0"),
		HTTPPort:                getEnvInt("HTTP_PORT", 3001),
		Dev:                     getEnvBool("DEV", false),
		TweedeKamerODataBaseURL: getEnv("TWEEDE_KAMER_ODATA_BASE_URL", "https://gegevensmagazijn.tweedekamer.nl/OData/v4/2.0"),
		TweedeKamerBatchSize:    getEnvInt("TWEEDE_KAMER_BATCH_SIZE", 100),
		TweedeKamerMaxPages:     getEnvInt("TWEEDE_KAMER_MAX_PAGES", 0),
		TweedeKamerInitialSince: initialSince,
		CursorOverlap:           time.Duration(getEnvInt("TWEEDE_KAMER_CURSOR_OVERLAP_MINUTES", 10)) * time.Minute,
		SyncInterval:            syncInterval,
		SyncMotionVoteLimit:     getEnvInt("SYNC_MOTION_VOTE_LIMIT", 250),
		SyncMotionDocumentLimit: getEnvInt("SYNC_MOTION_DOCUMENT_LIMIT", 500),
	}, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
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
