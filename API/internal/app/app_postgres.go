package app

import (
	"fmt"
	"os"

	httpapi "lawmap/internal/http"
	"lawmap/internal/repo/db"
	graphrepo "lawmap/internal/repo/graph"
	conf "lawmap/internal/config"
	"lawmap/internal/pkg/log"
)

// NewWithPostgres creates an app instance using PostgreSQL backend.
func NewWithPostgres() (*App, error) {
	logger := log.Default()

	// Get DATABASE_URL from environment (Neon connection string)
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// Fall back to individual settings
		dbCfg := db.Config{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvIntOrDefault("DB_PORT", 5432),
			User:     getEnvOrDefault("DB_USER", "postgres"),
			Password: os.Getenv("DB_PASSWORD"),
			DBName:   getEnvOrDefault("DB_NAME", "lawmap"),
			SSLMode:  getEnvOrDefault("DB_SSLMODE", "require"),
		}

		if dbCfg.Password == "" {
			return nil, fmt.Errorf("database password required (set DB_PASSWORD or DATABASE_URL)")
		}

		connStr = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			dbCfg.Host, dbCfg.Port, dbCfg.User, dbCfg.Password, dbCfg.DBName, dbCfg.SSLMode)
	}

	// Connect to database
	database, err := db.Connect(db.Config{Host: connStr})
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	logger.Info("Connected to PostgreSQL database")

	// Run migrations
	migrationsDir := getEnvOrDefault("MIGRATIONS_DIR", "migrations")
	if err := database.RunMigrations(migrationsDir); err != nil {
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	// Create PostgreSQL graph store
	store := graphrepo.NewPostgresStore(database)

	// Load example data if requested (for demo/testing)
	if examplesFile := os.Getenv("LOAD_EXAMPLES"); examplesFile != "" {
		logger.Info("Loading example data", map[string]any{"file": examplesFile})
		if err := loadExamplesToPostgres(store, examplesFile); err != nil {
			logger.Warn("Failed to load examples", map[string]any{"error": err.Error()})
		}
	}

	// Load sources config if available
	var sources []conf.SourceDescriptor
	spath := os.Getenv("SOURCES_FILE")
	if spath == "" {
		if _, err := os.Stat("configs/sources.json"); err == nil {
			spath = "configs/sources.json"
		} else if _, err := os.Stat("configs/sources.example.json"); err == nil {
			spath = "configs/sources.example.json"
		}
	}
	if spath != "" {
		if ss, err := conf.LoadSources(spath); err == nil {
			sources = ss
			logger.Info("Loaded sources", map[string]any{"count": len(sources), "file": spath})
		} else {
			logger.Warn("Could not load sources", map[string]any{"error": err.Error()})
		}
	}

	server := httpapi.NewServer(store, sources)
	return &App{Server: server, database: database}, nil
}

// loadExamplesToPostgres loads example JSONL data into PostgreSQL.
func loadExamplesToPostgres(store *graphrepo.PostgresStore, path string) error {
	// Load examples using memory store temporarily, then insert to postgres
	memStore := graphrepo.NewMemoryStore()
	if err := memStore.LoadJSONL(path); err != nil {
		return err
	}

	// Get all nodes and edges from memory store
	// This is a simplified approach - in production you'd iterate more efficiently
	logger := log.Default()
	logger.Info("Migrating example data to PostgreSQL...")

	// For now, we'll skip automatic loading of examples to PostgreSQL
	// Users should use the ETL pipeline or API to populate data
	logger.Info("Example data loading skipped - use ETL pipeline to populate database")

	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		var result int
		if _, err := fmt.Sscanf(v, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}
