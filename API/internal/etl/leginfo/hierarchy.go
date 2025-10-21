package leginfo

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	dgraph "lawmap/internal/domain/graph"
	"lawmap/internal/pkg/log"
)

// HierarchyBuilder builds hierarchy nodes and edges from section nodes.
type HierarchyBuilder struct {
	jurisdiction string
	logger       *log.Logger
}

// NewHierarchyBuilder creates a new hierarchy builder.
func NewHierarchyBuilder() *HierarchyBuilder {
	return &HierarchyBuilder{
		jurisdiction: "CA",
		logger:       log.Default().WithField("component", "leginfo-hierarchy"),
	}
}

// BuildHierarchyNodes creates all intermediate hierarchy nodes (CODE, TITLE, CHAPTER, etc.)
// from a list of section nodes.
func (h *HierarchyBuilder) BuildHierarchyNodes(sections []*dgraph.Node) []*dgraph.Node {
	// Extract unique hierarchy IDs from section canonical IDs
	hierarchyMap := make(map[string]*dgraph.Node)

	for _, section := range sections {
		// Parse the canonical ID to extract hierarchy levels
		// E.g., "CA:CIV:T02:CH02:§3342" -> ["CA", "CA:CIV", "CA:CIV:T02", "CA:CIV:T02:CH02"]
		levels := h.extractHierarchyLevels(section.ID)

		for _, level := range levels {
			if _, exists := hierarchyMap[level.ID]; !exists {
				hierarchyMap[level.ID] = h.createHierarchyNode(level)
			}
		}
	}

	// Convert map to slice
	var nodes []*dgraph.Node
	for _, node := range hierarchyMap {
		nodes = append(nodes, node)
	}

	h.logger.Info("Built hierarchy nodes", map[string]any{
		"count": len(nodes),
	})

	return nodes
}

// BuildParentOfEdges creates PARENT_OF edges for the hierarchy.
func (h *HierarchyBuilder) BuildParentOfEdges(allNodes []*dgraph.Node) []*dgraph.Edge {
	var edges []*dgraph.Edge
	edgeMap := make(map[string]bool) // Deduplicate edges

	for _, node := range allNodes {
		parentID := h.getParentID(node.ID)
		if parentID == "" {
			continue // Root node, no parent
		}

		// Check if parent exists
		parentExists := false
		for _, n := range allNodes {
			if n.ID == parentID {
				parentExists = true
				break
			}
		}

		if !parentExists {
			h.logger.Warn("Parent node not found, skipping edge", map[string]any{
				"child":  node.ID,
				"parent": parentID,
			})
			continue
		}

		// Create edge key for deduplication
		edgeKey := fmt.Sprintf("%s->%s", parentID, node.ID)
		if edgeMap[edgeKey] {
			continue // Edge already exists
		}

		// Extract order from node props
		order := h.extractOrder(node)

		edge := &dgraph.Edge{
			ID:       fmt.Sprintf("e_%s_%s", parentID, node.ID),
			EdgeType: "PARENT_OF",
			FromID:   parentID,
			ToID:     node.ID,
			Props: map[string]any{
				"order": order,
			},
		}

		edges = append(edges, edge)
		edgeMap[edgeKey] = true
	}

	h.logger.Info("Built PARENT_OF edges", map[string]any{
		"count": len(edges),
	})

	return edges
}

// extractHierarchyLevels extracts all hierarchy levels from a canonical ID.
// E.g., "CA:CIV:T02:CH02:§3342" returns:
// - {ID: "CA", Label: "JURISDICTION", ...}
// - {ID: "CA:CIV", Label: "CODE", ...}
// - {ID: "CA:CIV:T02", Label: "TITLE", ...}
// - {ID: "CA:CIV:T02:CH02", Label: "CHAPTER", ...}
func (h *HierarchyBuilder) extractHierarchyLevels(canonicalID string) []HierarchyLayer {
	parts := strings.Split(canonicalID, ":")
	var levels []HierarchyLayer

	for i := 0; i < len(parts); i++ {
		// Skip section markers (§)
		if strings.HasPrefix(parts[i], "§") {
			continue
		}

		// Build ID up to this level
		levelID := strings.Join(parts[:i+1], ":")

		// Determine label and title
		label, title, order := h.determineLevel(parts[i], levelID, parts)

		levels = append(levels, HierarchyLayer{
			ID:    levelID,
			Label: label,
			Title: title,
			Order: order,
		})
	}

	return levels
}

// determineLevel determines the label, title, and order for a hierarchy level.
func (h *HierarchyBuilder) determineLevel(part string, fullID string, allParts []string) (label, title string, order int) {
	// Jurisdiction (e.g., "CA")
	if part == h.jurisdiction {
		return "JURISDICTION", "California", 0
	}

	// Code (e.g., "CIV")
	if !strings.HasPrefix(part, "T") && !strings.HasPrefix(part, "D") &&
		!strings.HasPrefix(part, "P") && !strings.HasPrefix(part, "CH") &&
		!strings.HasPrefix(part, "ART") {
		tocParser := NewTOCParser()
		return "CODE", tocParser.getCodeName(part), 0
	}

	// Title (e.g., "T02")
	if strings.HasPrefix(part, "T") && len(part) >= 2 {
		numStr := part[1:]
		num, _ := strconv.Atoi(numStr)
		return "TITLE", fmt.Sprintf("Title %d", num), num
	}

	// Division (e.g., "D01")
	if strings.HasPrefix(part, "D") && len(part) >= 2 {
		numStr := part[1:]
		num, _ := strconv.Atoi(numStr)
		return "DIVISION", fmt.Sprintf("Division %d", num), num
	}

	// Part (e.g., "P01")
	if strings.HasPrefix(part, "P") && len(part) >= 2 {
		numStr := part[1:]
		num, _ := strconv.Atoi(numStr)
		return "PART", fmt.Sprintf("Part %d", num), num
	}

	// Chapter (e.g., "CH02")
	if strings.HasPrefix(part, "CH") {
		numStr := part[2:]
		num, _ := strconv.Atoi(numStr)
		return "CHAPTER", fmt.Sprintf("Chapter %d", num), num
	}

	// Article (e.g., "ART01")
	if strings.HasPrefix(part, "ART") {
		numStr := part[3:]
		num, _ := strconv.Atoi(numStr)
		return "ARTICLE", fmt.Sprintf("Article %d", num), num
	}

	return "UNKNOWN", part, 0
}

// createHierarchyNode creates a graph node for a hierarchy level.
func (h *HierarchyBuilder) createHierarchyNode(level HierarchyLayer) *dgraph.Node {
	// Extract code from ID (second part)
	parts := strings.Split(level.ID, ":")
	var code string
	if len(parts) > 1 {
		code = parts[1]
	}

	props := map[string]any{
		"jurisdiction": h.jurisdiction,
	}

	if code != "" {
		props["code"] = code
	}

	// Add specific hierarchy props based on label
	switch level.Label {
	case "TITLE":
		if level.Order > 0 {
			props["title_num"] = level.Order
		}
	case "DIVISION":
		if level.Order > 0 {
			props["division_num"] = level.Order
		}
	case "PART":
		if level.Order > 0 {
			props["part_num"] = level.Order
		}
	case "CHAPTER":
		if level.Order > 0 {
			props["chapter_num"] = level.Order
		}
	case "ARTICLE":
		if level.Order > 0 {
			props["article_num"] = level.Order
		}
	}

	return &dgraph.Node{
		ID:     level.ID,
		Labels: []string{level.Label},
		Title:  level.Title,
		Props:  props,
		Version: &dgraph.Version{
			FetchedAt: time.Now().Format(time.RFC3339),
		},
		Sources: []dgraph.SourceMeta{
			{
				Name:        "LegInfo",
				URL:         "https://leginfo.legislature.ca.gov/",
				RetrievedAt: time.Now().Format(time.RFC3339),
			},
		},
	}
}

// getParentID extracts the parent ID from a canonical ID.
// E.g., "CA:CIV:T02:CH02:§3342" -> "CA:CIV:T02:CH02"
func (h *HierarchyBuilder) getParentID(id string) string {
	lastColon := strings.LastIndex(id, ":")
	if lastColon == -1 {
		return "" // No parent
	}
	return id[:lastColon]
}

// extractOrder attempts to extract the order number from a node's properties.
func (h *HierarchyBuilder) extractOrder(node *dgraph.Node) int {
	// Try to extract from various numbered properties
	if val, ok := node.Props["title_num"].(int); ok {
		return val
	}
	if val, ok := node.Props["division_num"].(int); ok {
		return val
	}
	if val, ok := node.Props["part_num"].(int); ok {
		return val
	}
	if val, ok := node.Props["chapter_num"].(int); ok {
		return val
	}
	if val, ok := node.Props["article_num"].(int); ok {
		return val
	}

	// Try to extract from section number
	if sectionNum, ok := node.Props["section_num"].(string); ok {
		// Try to extract numeric part
		re := regexp.MustCompile(`^(\d+)`)
		if m := re.FindStringSubmatch(sectionNum); m != nil {
			if num, err := strconv.Atoi(m[1]); err == nil {
				return num
			}
		}
	}

	return 0
}
