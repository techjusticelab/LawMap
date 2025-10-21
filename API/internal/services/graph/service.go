package graph

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	dgraph "lawmap/internal/domain/graph"
	graphrepo "lawmap/internal/repo/graph"
	"lawmap/internal/pkg/log"
)

// Service provides business logic for graph operations.
type Service struct {
	repo   *graphrepo.MemoryStore
	logger *log.Logger
}

// New creates a new graph service.
func New(repo *graphrepo.MemoryStore) *Service {
	return &Service{
		repo:   repo,
		logger: log.Default().WithField("component", "graph-service"),
	}
}

// UpsertNode creates or updates a node in the graph.
// If a node with the same ID exists, it merges sources and updates version if content changed.
func (s *Service) UpsertNode(ctx context.Context, node *dgraph.Node) error {
	if node.ID == "" {
		return fmt.Errorf("node ID is required")
	}

	// Check if node exists
	existing, exists := s.repo.GetNode(node.ID)
	if !exists {
		// New node - just store it
		s.logger.Info("Creating new node", map[string]any{"id": node.ID})
		return s.createNode(node)
	}

	// Node exists - merge and update if changed
	s.logger.Debug("Node exists, checking for updates", map[string]any{"id": node.ID})

	// Compare content hash
	existingHash := s.computeHash(existing)
	newHash := s.computeHash(node)

	if existingHash == newHash {
		// Content unchanged, just merge sources
		s.logger.Debug("Node content unchanged", map[string]any{"id": node.ID})
		return s.mergeSources(existing, node)
	}

	// Content changed - update node
	s.logger.Info("Updating node with new content", map[string]any{
		"id":           node.ID,
		"old_hash":     existingHash,
		"new_hash":     newHash,
	})
	return s.updateNode(existing, node, newHash)
}

// CreateEdge creates a new edge in the graph if it doesn't exist.
func (s *Service) CreateEdge(ctx context.Context, edge *dgraph.Edge) error {
	if edge.FromID == "" || edge.ToID == "" {
		return fmt.Errorf("edge requires from_id and to_id")
	}
	if edge.EdgeType == "" {
		return fmt.Errorf("edge requires edge_type")
	}

	// Verify nodes exist
	if _, ok := s.repo.GetNode(edge.FromID); !ok {
		return fmt.Errorf("from node not found: %s", edge.FromID)
	}
	if _, ok := s.repo.GetNode(edge.ToID); !ok {
		return fmt.Errorf("to node not found: %s", edge.ToID)
	}

	// For the current memory store implementation, edges are stored directly
	// In a real implementation, we'd check for duplicates and handle upserts
	s.logger.Info("Creating edge", map[string]any{
		"from":      edge.FromID,
		"to":        edge.ToID,
		"edge_type": edge.EdgeType,
	})

	return nil // Placeholder - needs actual implementation in repo
}

// BuildHierarchy creates PARENT_OF edges between nodes based on their canonical IDs.
// For example: CA:CIV -> CA:CIV:T02 -> CA:CIV:T02:CH02 -> CA:CIV:T02:CH02:§3342
func (s *Service) BuildHierarchy(ctx context.Context, nodes []*dgraph.Node) ([]*dgraph.Edge, error) {
	// Group nodes by ID length to establish parent-child relationships
	var edges []*dgraph.Edge
	edgeID := 0

	for _, node := range nodes {
		parentID := s.getParentID(node.ID)
		if parentID == "" {
			continue // Root node, no parent
		}

		// Check if parent exists
		if _, ok := s.repo.GetNode(parentID); !ok {
			s.logger.Warn("Parent node not found, skipping edge", map[string]any{
				"child":  node.ID,
				"parent": parentID,
			})
			continue
		}

		edge := &dgraph.Edge{
			ID:       fmt.Sprintf("e%d", edgeID),
			EdgeType: "PARENT_OF",
			FromID:   parentID,
			ToID:     node.ID,
			Props:    make(map[string]any),
		}
		edges = append(edges, edge)
		edgeID++
	}

	s.logger.Info("Built hierarchy edges", map[string]any{"count": len(edges)})
	return edges, nil
}

// DetectCitations scans node text for citations and creates CITES edges.
func (s *Service) DetectCitations(ctx context.Context, node *dgraph.Node) ([]*dgraph.Edge, error) {
	// This would use the citation parser to extract references from text
	// For now, return empty slice as placeholder
	s.logger.Debug("Detecting citations", map[string]any{"node_id": node.ID})
	return []*dgraph.Edge{}, nil
}

// Helper methods

func (s *Service) createNode(node *dgraph.Node) error {
	// Set version metadata if not present
	if node.Version == nil {
		node.Version = &dgraph.Version{
			FetchedAt: time.Now().Format(time.RFC3339),
		}
	}
	if node.Version.Hash == "" {
		node.Version.Hash = s.computeHash(node)
	}

	// Actual creation would happen in repo layer
	return nil
}

func (s *Service) updateNode(existing, updated *dgraph.Node, newHash string) error {
	// In a real implementation:
	// 1. Create an AMENDS edge from updated to existing
	// 2. Update the node with new content
	// 3. Preserve old version for history

	updated.Version.Hash = newHash
	updated.Version.FetchedAt = time.Now().Format(time.RFC3339)

	// Merge sources
	return s.mergeSources(existing, updated)
}

func (s *Service) mergeSources(existing, updated *dgraph.Node) error {
	// Deduplicate sources by URL
	sourceMap := make(map[string]dgraph.SourceMeta)

	for _, src := range existing.Sources {
		sourceMap[src.URL] = src
	}
	for _, src := range updated.Sources {
		sourceMap[src.URL] = src
	}

	var merged []dgraph.SourceMeta
	for _, src := range sourceMap {
		merged = append(merged, src)
	}

	updated.Sources = merged
	return nil
}

func (s *Service) computeHash(node *dgraph.Node) string {
	// Hash the content that matters for change detection
	content := fmt.Sprintf("%s|%s|%s", node.Title, node.Citation, node.Text)
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash[:8]) // Use first 8 bytes
}

// getParentID extracts the parent ID from a canonical ID.
// CA:CIV:T02:CH02:§3342 -> CA:CIV:T02:CH02
// CA:CIV:T02:CH02 -> CA:CIV:T02
// CA:CIV:T02 -> CA:CIV
// CA:CIV -> CA
// CA -> ""
func (s *Service) getParentID(id string) string {
	// Find last colon
	lastColon := -1
	for i := len(id) - 1; i >= 0; i-- {
		if id[i] == ':' {
			lastColon = i
			break
		}
	}

	if lastColon == -1 {
		return "" // No parent
	}

	return id[:lastColon]
}
