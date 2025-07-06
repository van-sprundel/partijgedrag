// internal/config/config.go
package config

import (
	"time"
)

type Config struct {
	API        APIConfig        `yaml:"api"`
	Documents  DocumentsConfig  `yaml:"documents"`
	Categories []string         `yaml:"categories"`
	Pagination PaginationConfig `yaml:"pagination"`
	Logging    LoggingConfig    `yaml:"logging"`
	Storage    StorageConfig    `yaml:"storage"`
}

type APIConfig struct {
	BaseURL   string        `yaml:"base_url"`
	Timeout   time.Duration `yaml:"timeout"`
	UserAgent string        `yaml:"user_agent"`
}

type DocumentsConfig struct {
	BaseURL string `yaml:"base_url"`
}

type PaginationConfig struct {
	MaxPages  int `yaml:"max_pages"`
	BatchSize int `yaml:"batch_size"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type StorageConfig struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

// pkg/importer/importer.go
