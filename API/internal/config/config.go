package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the full application configuration.
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Storage   StorageConfig   `yaml:"storage"`
	Sources   []Source        `yaml:"sources"`
	Scheduler SchedulerConfig `yaml:"scheduler"`
	Logging   LoggingConfig   `yaml:"logging"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port         int    `yaml:"port"`
	Host         string `yaml:"host"`
	ReadTimeout  int    `yaml:"read_timeout"`  // seconds
	WriteTimeout int    `yaml:"write_timeout"` // seconds
}

// StorageConfig defines where and how data is stored.
type StorageConfig struct {
	Backend  string           `yaml:"backend"` // "memory", "postgres", etc.
	Postgres PostgresConfig   `yaml:"postgres"`
	Graph    GraphStorage     `yaml:"graph"`
	Blob     BlobStorage      `yaml:"blob"`
	Index    IndexStorage     `yaml:"index"`
}

// PostgresConfig holds PostgreSQL connection settings.
type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

// GraphStorage configures the graph database.
type GraphStorage struct {
	Path       string            `yaml:"path"`        // for file-based backends
	Connection string            `yaml:"connection"`  // for network backends
	Options    map[string]string `yaml:"options"`
}

// BlobStorage configures document/text storage.
type BlobStorage struct {
	Path       string            `yaml:"path"`
	Connection string            `yaml:"connection"`
	Options    map[string]string `yaml:"options"`
}

// IndexStorage configures full-text search index.
type IndexStorage struct {
	Path       string            `yaml:"path"`
	Connection string            `yaml:"connection"`
	Options    map[string]string `yaml:"options"`
}

// Source represents an external data source for ETL.
type Source struct {
	Name          string   `yaml:"name"`
	Jurisdictions []string `yaml:"jurisdictions"`
	Codes         []string `yaml:"codes"`
	Kind          string   `yaml:"kind"` // "bulk", "api", "web", "mixed"
	URLs          []string `yaml:"urls"`
	Schedule      string   `yaml:"schedule"`      // cron expression or "monthly"
	Enabled       bool     `yaml:"enabled"`
	FetchOptions  map[string]any `yaml:"fetch_options"` // source-specific options
}

// SchedulerConfig controls ETL job execution.
type SchedulerConfig struct {
	Enabled          bool   `yaml:"enabled"`
	DefaultSchedule  string `yaml:"default_schedule"`  // e.g., "monthly"
	MaxConcurrent    int    `yaml:"max_concurrent"`
	RetryAttempts    int    `yaml:"retry_attempts"`
	RetryDelayMinutes int   `yaml:"retry_delay_minutes"`
}

// LoggingConfig controls logging behavior.
type LoggingConfig struct {
	Level  string `yaml:"level"`  // "debug", "info", "warn", "error"
	Format string `yaml:"format"` // "text", "json"
	Output string `yaml:"output"` // "stderr", "stdout", or file path
}

// Load reads a YAML config file and returns a Config struct.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config YAML: %w", err)
	}

	// Apply defaults
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 30
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 30
	}
	if cfg.Storage.Backend == "" {
		cfg.Storage.Backend = "memory"
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "text"
	}
	if cfg.Logging.Output == "" {
		cfg.Logging.Output = "stderr"
	}
	if cfg.Scheduler.DefaultSchedule == "" {
		cfg.Scheduler.DefaultSchedule = "monthly"
	}
	if cfg.Scheduler.MaxConcurrent == 0 {
		cfg.Scheduler.MaxConcurrent = 3
	}
	if cfg.Scheduler.RetryAttempts == 0 {
		cfg.Scheduler.RetryAttempts = 3
	}
	if cfg.Scheduler.RetryDelayMinutes == 0 {
		cfg.Scheduler.RetryDelayMinutes = 5
	}

	return &cfg, nil
}

// LoadOrDefault loads config from path, or returns a default config if the file doesn't exist.
func LoadOrDefault(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		// Return sensible defaults
		return &Config{
			Server: ServerConfig{
				Port:         8080,
				Host:         "",
				ReadTimeout:  30,
				WriteTimeout: 30,
			},
			Storage: StorageConfig{
				Backend: "memory",
			},
			Scheduler: SchedulerConfig{
				Enabled:          false,
				DefaultSchedule:  "monthly",
				MaxConcurrent:    3,
				RetryAttempts:    3,
				RetryDelayMinutes: 5,
			},
			Logging: LoggingConfig{
				Level:  "info",
				Format: "text",
				Output: "stderr",
			},
		}
	}
	return cfg
}
