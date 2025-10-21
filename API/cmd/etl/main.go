package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	dgraph "lawmap/internal/domain/graph"
	"lawmap/internal/etl"
	"lawmap/internal/etl/leginfo"
	"lawmap/internal/pkg/log"
	"lawmap/internal/repo/db"
	graphrepo "lawmap/internal/repo/graph"
)

func main() {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Parse flags
	source := flag.String("source", "leginfo", "ETL source (leginfo, ...)")
	jurisdiction := flag.String("jurisdiction", "CA", "Jurisdiction code (CA, US, ...)")
	code := flag.String("code", "", "Code to fetch (CIV, PEN, BPC, ...)")
	dryRun := flag.Bool("dry-run", false, "Parse and validate without loading to database")
	flag.Parse()

	logger := log.Default()

	// Connect to database
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" && !*dryRun {
		fmt.Fprintf(os.Stderr, "ERROR: DATABASE_URL not set (or use --dry-run)\n")
		os.Exit(1)
	}

	ctx := context.Background()

	// Run the ETL pipeline
	switch *source {
	case "leginfo":
		if err := runLegInfoETL(ctx, logger, connStr, *jurisdiction, *code, *dryRun); err != nil {
			logger.Error("ETL failed", map[string]any{"error": err.Error()})
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown source: %s\n", *source)
		os.Exit(1)
	}

	logger.Info("ETL completed successfully")
}

func runLegInfoETL(ctx context.Context, logger *log.Logger, connStr, jurisdiction, code string, dryRun bool) error {
	logger.Info("Starting LegInfo ETL", map[string]any{
		"jurisdiction": jurisdiction,
		"code":         code,
		"dry_run":      dryRun,
	})

	// Create fetcher with code configuration
	fetcherCfg := leginfo.FetcherConfig{
		BaseURL:         "https://leginfo.legislature.ca.gov",
		RateLimitPerSec: 2.0,
		TimeoutSeconds:  30,
		UserAgent:       "LawMap/0.1 ETL (+https://github.com/yourorg/lawmap)",
		Code:            code,
		MaxSections:     5, // Limit for testing - set to 0 for all sections
	}
	fetcher := leginfo.NewFetcher(fetcherCfg)

	// Create parser
	parser := leginfo.NewParser()

	// Create loader
	var loader *leginfo.Loader
	if !dryRun {
		database, err := db.Connect(db.Config{Host: connStr})
		if err != nil {
			return fmt.Errorf("connect to database: %w", err)
		}

		// Run migrations
		logger.Info("Running database migrations")
		migrationsDir := "migrations"
		if err := database.RunMigrations(migrationsDir); err != nil {
			return fmt.Errorf("run migrations: %w", err)
		}
		logger.Info("Migrations completed")

		store := graphrepo.NewPostgresStore(database)
		loader = leginfo.NewLoader(store)
	} else {
		// Dry run - use a no-op loader
		loader = leginfo.NewLoader(&noopLoader{})
	}

	// Create and run pipeline
	pipeline := &etl.Pipeline{
		Extractor:   fetcher,
		Transformer: parser,
		Loader:      loader,
	}

	result, err := pipeline.Run(ctx)
	if err != nil {
		return fmt.Errorf("pipeline run failed: %w", err)
	}

	duration := result.EndTime.Sub(result.StartTime)

	logger.Info("Pipeline result", map[string]any{
		"nodes_created":   result.NodesCreated,
		"edges_created":   result.EdgesCreated,
		"items_processed": result.ItemsProcessed,
		"duration":        duration.String(),
		"errors":          len(result.Errors),
	})

	if len(result.Errors) > 0 {
		logger.Warn("Pipeline completed with errors", map[string]any{
			"error_count": len(result.Errors),
		})
		for i, e := range result.Errors {
			logger.Error(fmt.Sprintf("Error %d", i+1), map[string]any{"error": e.Error()})
		}
	}

	return nil
}

// noopLoader is a no-op loader for dry-run mode
type noopLoader struct{}

func (n *noopLoader) UpsertNode(ctx context.Context, node *dgraph.Node) error {
	return nil
}

func (n *noopLoader) UpsertEdge(ctx context.Context, edge *dgraph.Edge) error {
	return nil
}
