package cmdutils

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"etl/internal/config"

	"gopkg.in/yaml.v3"
)

// LoadConfig reads and parses the YAML configuration file from the given path.
func LoadConfig(path string) (*config.Config, error) {
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

// ParseStorageConfig determines the final storage configuration.
// It prioritizes the DATABASE_URL environment variable if it's set,
// otherwise it falls back to the configuration from the YAML file.
func ParseStorageConfig(cfg *config.Config) (*config.StorageConfig, error) {
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
