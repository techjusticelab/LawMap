package leginfo

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"lawmap/internal/pkg/log"
)

// TOCParser parses LegInfo table of contents HTML.
type TOCParser struct {
	jurisdiction string
	logger       *log.Logger
}

// NewTOCParser creates a new table of contents parser.
func NewTOCParser() *TOCParser {
	return &TOCParser{
		jurisdiction: "CA",
		logger:       log.Default().WithField("component", "leginfo-toc-parser"),
	}
}

// ParseCodeTOC parses the table of contents HTML and returns the code structure.
func (p *TOCParser) ParseCodeTOC(htmlContent string, code string) (*CodeStructure, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	structure := &CodeStructure{
		Code:         code,
		CodeName:     p.getCodeName(code),
		Jurisdiction: p.jurisdiction,
		Titles:       []TitleStructure{},
		Sections:     []SectionReference{},
	}

	// Parse the tree structure
	p.parseNode(doc, structure, HierarchyContext{Code: code})

	p.logger.Info("Parsed code TOC", map[string]any{
		"code":     code,
		"titles":   len(structure.Titles),
		"sections": len(structure.Sections),
	})

	return structure, nil
}

// parseNode recursively parses the HTML tree to extract hierarchy.
func (p *TOCParser) parseNode(n *html.Node, structure *CodeStructure, ctx HierarchyContext) {
	if n.Type == html.ElementNode {
		// Look for links that indicate sections
		if n.Data == "a" {
			href := p.getAttr(n, "href")
			text := p.getTextContent(n)

			// Check if this is a section link
			if strings.Contains(href, "sectionNum=") {
				section := p.parseSectionLink(href, text, ctx)
				if section != nil {
					structure.Sections = append(structure.Sections, *section)
				}
			}
		}

		// Look for hierarchy markers in text
		if n.Type == html.TextNode || n.Data == "span" || n.Data == "div" {
			text := p.getTextContent(n)
			p.updateContextFromText(text, &ctx)
		}
	}

	// Recursively parse children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p.parseNode(c, structure, ctx)
	}
}

// parseSectionLink extracts section information from a link.
func (p *TOCParser) parseSectionLink(href, text string, ctx HierarchyContext) *SectionReference {
	// Extract section number from URL
	// URL format: faces/codes_displaySection.xhtml?lawCode=CIV&sectionNum=3342
	re := regexp.MustCompile(`sectionNum=([^&]+)`)
	matches := re.FindStringSubmatch(href)
	if len(matches) < 2 {
		return nil
	}

	sectionNum := matches[1]

	// Build full URL
	baseURL := "https://leginfo.legislature.ca.gov"
	fullURL := baseURL + "/" + strings.TrimPrefix(href, "/")

	return &SectionReference{
		Number:       sectionNum,
		URL:          fullURL,
		HierarchyCtx: ctx,
	}
}

// updateContextFromText updates the hierarchy context based on text patterns.
func (p *TOCParser) updateContextFromText(text string, ctx *HierarchyContext) {
	text = strings.TrimSpace(text)

	// Match Title
	if m := regexp.MustCompile(`(?i)^Title\s+(\d+)`).FindStringSubmatch(text); m != nil {
		if num, err := strconv.Atoi(m[1]); err == nil {
			ctx.Title = num
		}
	}

	// Match Division
	if m := regexp.MustCompile(`(?i)^Division\s+(\d+)`).FindStringSubmatch(text); m != nil {
		if num, err := strconv.Atoi(m[1]); err == nil {
			ctx.Division = num
		}
	}

	// Match Part
	if m := regexp.MustCompile(`(?i)^Part\s+(\d+)`).FindStringSubmatch(text); m != nil {
		if num, err := strconv.Atoi(m[1]); err == nil {
			ctx.Part = num
		}
	}

	// Match Chapter
	if m := regexp.MustCompile(`(?i)^Chapter\s+(\d+)`).FindStringSubmatch(text); m != nil {
		if num, err := strconv.Atoi(m[1]); err == nil {
			ctx.Chapter = num
		}
	}

	// Match Article
	if m := regexp.MustCompile(`(?i)^Article\s+(\d+)`).FindStringSubmatch(text); m != nil {
		if num, err := strconv.Atoi(m[1]); err == nil {
			ctx.Article = num
		}
	}
}

// BuildCanonicalID constructs the canonical ID for a section based on its hierarchy.
func (p *TOCParser) BuildCanonicalID(ctx HierarchyContext, sectionNum string) string {
	parts := []string{p.jurisdiction, ctx.Code}

	if ctx.Title > 0 {
		parts = append(parts, fmt.Sprintf("T%02d", ctx.Title))
	}
	if ctx.Division > 0 {
		parts = append(parts, fmt.Sprintf("D%02d", ctx.Division))
	}
	if ctx.Part > 0 {
		parts = append(parts, fmt.Sprintf("P%02d", ctx.Part))
	}
	if ctx.Chapter > 0 {
		parts = append(parts, fmt.Sprintf("CH%02d", ctx.Chapter))
	}
	if ctx.Article > 0 {
		parts = append(parts, fmt.Sprintf("ART%02d", ctx.Article))
	}

	if sectionNum != "" {
		parts = append(parts, "§"+sectionNum)
	}

	return strings.Join(parts, ":")
}

// Helper functions

func (p *TOCParser) getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func (p *TOCParser) getTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	var parts []string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		parts = append(parts, p.getTextContent(c))
	}

	return strings.Join(parts, "")
}

func (p *TOCParser) getCodeName(code string) string {
	names := map[string]string{
		"BPC":  "Business and Professions Code",
		"CIV":  "Civil Code",
		"CCP":  "Code of Civil Procedure",
		"COM":  "Commercial Code",
		"CORP": "Corporations Code",
		"EDC":  "Education Code",
		"ELEC": "Elections Code",
		"EVID": "Evidence Code",
		"FAM":  "Family Code",
		"FIN":  "Financial Code",
		"FGC":  "Fish and Game Code",
		"FAC":  "Food and Agricultural Code",
		"GOV":  "Government Code",
		"HNC":  "Harbors and Navigation Code",
		"HSC":  "Health and Safety Code",
		"INS":  "Insurance Code",
		"LAB":  "Labor Code",
		"MVC":  "Military and Veterans Code",
		"PEN":  "Penal Code",
		"PROB": "Probate Code",
		"PCC":  "Public Contract Code",
		"PRC":  "Public Resources Code",
		"PUC":  "Public Utilities Code",
		"RTC":  "Revenue and Taxation Code",
		"SHC":  "Streets and Highways Code",
		"UIC":  "Unemployment Insurance Code",
		"VEH":  "Vehicle Code",
		"WAT":  "Water Code",
		"WIC":  "Welfare and Institutions Code",
		"CONS": "California Constitution",
		"CRC":  "California Rules of Court",
	}

	if name, ok := names[code]; ok {
		return name
	}
	return code
}

// ExtractSectionList returns a simple list of section numbers from HTML.
// This is a fallback parser for cases where the full tree structure is complex.
func (p *TOCParser) ExtractSectionList(htmlContent string, code string) ([]SectionReference, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	var sections []SectionReference
	ctx := HierarchyContext{Code: code}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			href := p.getAttr(n, "href")
			text := p.getTextContent(n)

			if strings.Contains(href, "sectionNum=") {
				if section := p.parseSectionLink(href, text, ctx); section != nil {
					sections = append(sections, *section)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)

	p.logger.Info("Extracted section list", map[string]any{
		"code":     code,
		"sections": len(sections),
	})

	return sections, nil
}
