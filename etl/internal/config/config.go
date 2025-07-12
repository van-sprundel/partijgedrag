package config

import (
	"time"
)

type Config struct {
	API     APIConfig     `yaml:"api"`
	Import  ImportConfig  `yaml:"import"`
	Logging LoggingConfig `yaml:"logging"`
	Storage StorageConfig `yaml:"storage"`
}

type APIConfig struct {
	ODataBaseURL    string        `yaml:"odata_base_url"`
	SyncFeedBaseURL string        `yaml:"syncfeed_base_url"`
	Timeout         time.Duration `yaml:"timeout"`
	UserAgent       string        `yaml:"user_agent"`
}

type ImportConfig struct {
	TestMode  bool `yaml:"test_mode"`
	ShowStats bool `yaml:"show_stats"`
	BatchSize int  `yaml:"batch_size"`
	MaxPages  int  `yaml:"max_pages"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type StorageConfig struct {
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"` // either prompted or set using DATABASE_URL env var
}
