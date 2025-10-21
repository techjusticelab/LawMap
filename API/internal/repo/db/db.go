package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
	"lawmap/internal/pkg/log"
)

// DB wraps a SQL database connection with helper methods.
type DB struct {
	*sql.DB
	logger *log.Logger
}

// Config holds database connection configuration.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Connect establishes a connection to PostgreSQL.
func Connect(cfg Config) (*DB, error) {
	if cfg.Port == 0 {
		cfg.Port = 5432
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "require"
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	// Support connection strings directly (for Neon and other services)
	if strings.HasPrefix(cfg.Host, "postgres://") || strings.HasPrefix(cfg.Host, "postgresql://") {
		connStr = cfg.Host
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	logger := log.Default().WithField("component", "database")
	logger.Info("Connected to PostgreSQL database")

	return &DB{DB: db, logger: logger}, nil
}

// RunMigrations executes SQL migration files in order.
func (db *DB) RunMigrations(migrationsDir string) error {
	db.logger.Info("Running database migrations", map[string]any{"dir": migrationsDir})

	// Get all .sql files in migrations directory
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("glob migration files: %w", err)
	}

	if len(files) == 0 {
		db.logger.Warn("No migration files found", map[string]any{"dir": migrationsDir})
		return nil
	}

	sort.Strings(files)

	for _, file := range files {
		if err := db.runMigrationFile(file); err != nil {
			return fmt.Errorf("run migration %s: %w", filepath.Base(file), err)
		}
	}

	db.logger.Info("Migrations completed successfully", map[string]any{"count": len(files)})
	return nil
}

func (db *DB) runMigrationFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	db.logger.Debug("Running migration", map[string]any{"file": filepath.Base(path)})

	ctx := context.Background()
	_, err = db.ExecContext(ctx, string(content))
	if err != nil {
		return fmt.Errorf("execute migration: %w", err)
	}

	db.logger.Info("Migration applied", map[string]any{"file": filepath.Base(path)})
	return nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	db.logger.Info("Closing database connection")
	return db.DB.Close()
}

// Ping verifies the database connection is alive.
func (db *DB) Ping() error {
	return db.DB.Ping()
}
