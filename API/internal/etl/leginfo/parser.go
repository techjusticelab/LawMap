package leginfo

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"

	dgraph "lawmap/internal/domain/graph"
	"lawmap/internal/pkg/log"
	"lawmap/internal/pkg/parse"
)

// Parser transforms LegInfo HTML into graph nodes and edges.
type Parser struct {
	jurisdiction string
}

// NewParser creates a new LegInfo parser.
func NewParser() *Parser {
	return &Parser{
		jurisdiction: "CA",
	}
}

// Name returns the parser name.
func (p *Parser) Name() string {
	return "leginfo-parser"
}

// Transform converts extracted LegInfo data into graph nodes and edges.
func (p *Parser) Transform(ctx context.Context, data []byte) ([]*dgraph.Node, []*dgraph.Edge, error) {
	logger := log.Default().WithField("component", "leginfo-parser")

	// Step 1: Unmarshal ExtractedData
	var extracted ExtractedData
	if err := json.Unmarshal(data, &extracted); err != nil {
		return nil, nil, fmt.Errorf("unmarshal extracted data: %w", err)
	}

	logger.Info("Starting transformation", map[string]any{
		"code":     extracted.Code,
		"sections": len(extracted.Sections),
	})

	// Step 2: Parse each section to create section nodes
	var sectionNodes []*dgraph.Node
	var parseErrors []error

	for i, sectionData := range extracted.Sections {
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		default:
		}

		node, err := p.ParseSection([]byte(sectionData.HTML), extracted.Code, sectionData.HierarchyCtx)
		if err != nil {
			logger.Warn("Failed to parse section", map[string]any{
				"section": sectionData.Number,
				"error":   err.Error(),
			})
			parseErrors = append(parseErrors, fmt.Errorf("section %s: %w", sectionData.Number, err))
			continue
		}

		sectionNodes = append(sectionNodes, node)

		// Progress logging
		if (i+1)%10 == 0 {
			logger.Debug("Parse progress", map[string]any{
				"parsed": i + 1,
				"total":  len(extracted.Sections),
			})
		}
	}

	logger.Info("Parsed sections", map[string]any{
		"successful": len(sectionNodes),
		"failed":     len(parseErrors),
	})

	// Step 3: Build hierarchy nodes (CODE, TITLE, CHAPTER, etc.)
	hierarchyBuilder := NewHierarchyBuilder()
	hierarchyNodes := hierarchyBuilder.BuildHierarchyNodes(sectionNodes)

	logger.Info("Built hierarchy nodes", map[string]any{
		"hierarchy_nodes": len(hierarchyNodes),
	})

	// Step 4: Combine all nodes
	allNodes := append(hierarchyNodes, sectionNodes...)

	// Step 5: Build PARENT_OF edges
	edges := hierarchyBuilder.BuildParentOfEdges(allNodes)

	logger.Info("Transformation complete", map[string]any{
		"total_nodes":      len(allNodes),
		"section_nodes":    len(sectionNodes),
		"hierarchy_nodes":  len(hierarchyNodes),
		"edges":            len(edges),
		"parse_errors":     len(parseErrors),
	})

	// If too many errors, might want to fail
	if len(parseErrors) > len(extracted.Sections)/2 {
		return nil, nil, fmt.Errorf("too many parse failures: %d/%d sections failed", len(parseErrors), len(extracted.Sections))
	}

	return allNodes, edges, nil
}

// ParseSection extracts section data from LegInfo HTML with full hierarchy context.
func (p *Parser) ParseSection(data []byte, code string, ctx HierarchyContext) (*dgraph.Node, error) {
	doc, err := html.Parse(strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	// Extract section number and title
	sectionNum, title := p.extractSectionInfo(doc)
	if sectionNum == "" {
		return nil, fmt.Errorf("could not extract section number")
	}

	// Extract section text
	text := p.extractSectionText(doc)

	// Extract effective date and history
	effectiveDate, history := p.extractHistory(doc)

	// Build canonical ID with full hierarchy
	normalizedCode := parse.NormalizeCode(code)
	tocParser := NewTOCParser()
	id := tocParser.BuildCanonicalID(ctx, sectionNum)

	// Build props with full hierarchy info
	props := map[string]any{
		"jurisdiction": p.jurisdiction,
		"code":         normalizedCode,
		"section_num":  sectionNum,
	}

	if ctx.Title > 0 {
		props["title_num"] = ctx.Title
	}
	if ctx.Division > 0 {
		props["division_num"] = ctx.Division
	}
	if ctx.Part > 0 {
		props["part_num"] = ctx.Part
	}
	if ctx.Chapter > 0 {
		props["chapter_num"] = ctx.Chapter
	}
	if ctx.Article > 0 {
		props["article_num"] = ctx.Article
	}

	if history != "" {
		props["history"] = history
	}

	node := &dgraph.Node{
		ID:       id,
		Labels:   []string{"SECTION"},
		Title:    title,
		Citation: fmt.Sprintf("%s § %s", normalizedCode, sectionNum),
		Text:     text,
		Props:    props,
		Version: &dgraph.Version{
			FetchedAt:     time.Now().Format(time.RFC3339),
			EffectiveDate: effectiveDate,
			Hash:          "",
		},
		Sources: []dgraph.SourceMeta{
			{
				Name:        "LegInfo",
				URL:         fmt.Sprintf("https://leginfo.legislature.ca.gov/faces/codes_displaySection.xhtml?lawCode=%s&sectionNum=%s", code, sectionNum),
				RetrievedAt: time.Now().Format(time.RFC3339),
			},
		},
	}

	return node, nil
}

// extractHistory extracts effective date and legislative history from HTML.
func (p *Parser) extractHistory(n *html.Node) (effectiveDate string, history string) {
	var historyText strings.Builder

	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.TextNode {
			text := strings.TrimSpace(node.Data)

			// Look for patterns like "(Added by Stats. 1872, Ch. 3)"
			if strings.Contains(text, "Stats.") || strings.Contains(text, "Amended") || strings.Contains(text, "Added") {
				historyText.WriteString(text)
				historyText.WriteString(" ")

				// Try to extract year as effective date
				if effectiveDate == "" {
					re := regexp.MustCompile(`(\d{4})`)
					if m := re.FindStringSubmatch(text); m != nil {
						effectiveDate = m[1] + "-01-01" // Default to January 1st
					}
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(n)

	history = strings.TrimSpace(historyText.String())
	return
}

// extractSectionInfo extracts section number and title from HTML.
func (p *Parser) extractSectionInfo(n *html.Node) (string, string) {
	// Look for patterns like "Section 3342" or "3342."
	var sectionNum, title string

	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.TextNode {
			text := strings.TrimSpace(node.Data)

			// Try to match "Section XXXX" or "XXXX."
			re := regexp.MustCompile(`(?i)(?:Section\s+)?(\d+[\.\w]*)\.\s*(.*)`)
			if m := re.FindStringSubmatch(text); m != nil {
				if sectionNum == "" {
					sectionNum = m[1]
					if len(m) > 2 {
						title = strings.TrimSpace(m[2])
					}
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(n)
	return sectionNum, title
}

// extractSectionText extracts the main text content from HTML.
func (p *Parser) extractSectionText(n *html.Node) string {
	var textParts []string

	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		// Look for div or p tags that might contain section text
		if node.Type == html.ElementNode && (node.Data == "p" || node.Data == "div") {
			text := p.getTextContent(node)
			text = strings.TrimSpace(text)
			if text != "" && len(text) > 20 {
				textParts = append(textParts, text)
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(n)

	// Deduplicate and join
	seen := make(map[string]bool)
	var unique []string
	for _, part := range textParts {
		if !seen[part] {
			seen[part] = true
			unique = append(unique, part)
		}
	}

	return strings.Join(unique, "\n\n")
}

// getTextContent recursively extracts all text from a node.
func (p *Parser) getTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	var parts []string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		parts = append(parts, p.getTextContent(c))
	}

	return strings.Join(parts, " ")
}
