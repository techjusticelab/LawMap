package leginfo

import (
	"context"

	dgraph "lawmap/internal/domain/graph"
	"lawmap/internal/pkg/log"
)

// GraphWriter defines the interface for storing graph data.
type GraphWriter interface {
	UpsertNode(ctx context.Context, node *dgraph.Node) error
	UpsertEdge(ctx context.Context, edge *dgraph.Edge) error
}

// Loader persists LegInfo data to the graph store.
type Loader struct {
	store  GraphWriter
	logger *log.Logger
}

// NewLoader creates a new loader for the graph store.
func NewLoader(store GraphWriter) *Loader {
	return &Loader{
		store:  store,
		logger: log.Default().WithField("component", "leginfo-loader"),
	}
}

// Name returns the loader name.
func (l *Loader) Name() string {
	return "leginfo-loader"
}

// Load persists nodes and edges to the graph store.
func (l *Loader) Load(ctx context.Context, nodes []*dgraph.Node, edges []*dgraph.Edge) error {
	l.logger.Info("Loading data to graph store", map[string]any{
		"nodes": len(nodes),
		"edges": len(edges),
	})

	if len(nodes) == 0 && len(edges) == 0 {
		return nil
	}

	// Upsert all nodes
	for _, node := range nodes {
		if err := l.store.UpsertNode(ctx, node); err != nil {
			l.logger.Error("Failed to upsert node", map[string]any{
				"node_id": node.ID,
				"error":   err.Error(),
			})
			return err
		}
	}

	// Upsert all edges
	for _, edge := range edges {
		if err := l.store.UpsertEdge(ctx, edge); err != nil {
			l.logger.Error("Failed to upsert edge", map[string]any{
				"edge_id": edge.ID,
				"error":   err.Error(),
			})
			return err
		}
	}

	l.logger.Info("Successfully loaded data", map[string]any{
		"nodes_loaded": len(nodes),
		"edges_loaded": len(edges),
	})

	return nil
}

// LoadSingle loads a single node into the store.
func (l *Loader) LoadSingle(ctx context.Context, node *dgraph.Node) error {
	l.logger.Debug("Loading single node", map[string]any{"node_id": node.ID})
	return l.store.UpsertNode(ctx, node)
}
