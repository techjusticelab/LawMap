package etl

import (
	"context"
	"fmt"
	"time"

	dgraph "lawmap/internal/domain/graph"
)

// Result represents the output of an ETL job.
type Result struct {
	NodesCreated   int
	NodesUpdated   int
	EdgesCreated   int
	ItemsProcessed int
	Errors         []error
	StartTime      time.Time
	EndTime        time.Time
	Metadata       map[string]any
}

// Status represents the current state of an ETL job.
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCanceled  Status = "canceled"
)

// Job represents a scheduled or running ETL job.
type Job struct {
	ID         string
	SourceName string
	Status     Status
	Result     *Result
	StartedAt  time.Time
	FinishedAt time.Time
	Error      error
}

// Extractor defines the interface for data extraction from external sources.
type Extractor interface {
	// Extract fetches raw data from the source and returns it as a byte stream.
	Extract(ctx context.Context) ([]byte, error)

	// Name returns the name of this extractor.
	Name() string
}

// Transformer defines the interface for transforming raw data into graph nodes and edges.
type Transformer interface {
	// Transform converts raw data into nodes and edges.
	Transform(ctx context.Context, data []byte) ([]*dgraph.Node, []*dgraph.Edge, error)

	// Name returns the name of this transformer.
	Name() string
}

// Loader defines the interface for loading transformed data into storage.
type Loader interface {
	// Load persists nodes and edges to the graph store.
	Load(ctx context.Context, nodes []*dgraph.Node, edges []*dgraph.Edge) error

	// Name returns the name of this loader.
	Name() string
}

// Pipeline combines extraction, transformation, and loading in sequence.
type Pipeline struct {
	Extractor   Extractor
	Transformer Transformer
	Loader      Loader
}

// Run executes the ETL pipeline and returns the result.
func (p *Pipeline) Run(ctx context.Context) (*Result, error) {
	result := &Result{
		StartTime: time.Now(),
		Metadata:  make(map[string]any),
	}

	// Extract
	data, err := p.Extractor.Extract(ctx)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("extraction failed: %w", err))
		result.EndTime = time.Now()
		return result, err
	}

	// Transform
	nodes, edges, err := p.Transformer.Transform(ctx, data)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("transformation failed: %w", err))
		result.EndTime = time.Now()
		return result, err
	}
	result.ItemsProcessed = len(nodes)

	// Load
	if err := p.Loader.Load(ctx, nodes, edges); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("loading failed: %w", err))
		result.EndTime = time.Now()
		return result, err
	}

	result.NodesCreated = len(nodes)
	result.EdgesCreated = len(edges)
	result.EndTime = time.Now()
	result.Metadata["extractor"] = p.Extractor.Name()
	result.Metadata["transformer"] = p.Transformer.Name()
	result.Metadata["loader"] = p.Loader.Name()

	return result, nil
}

// Registry holds available ETL pipelines by name.
type Registry struct {
	pipelines map[string]*Pipeline
}

// NewRegistry creates a new pipeline registry.
func NewRegistry() *Registry {
	return &Registry{
		pipelines: make(map[string]*Pipeline),
	}
}

// Register adds a pipeline to the registry.
func (r *Registry) Register(name string, pipeline *Pipeline) {
	r.pipelines[name] = pipeline
}

// Get retrieves a pipeline by name.
func (r *Registry) Get(name string) (*Pipeline, bool) {
	p, ok := r.pipelines[name]
	return p, ok
}

// List returns all registered pipeline names.
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.pipelines))
	for name := range r.pipelines {
		names = append(names, name)
	}
	return names
}
